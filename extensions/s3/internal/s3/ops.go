package s3

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/retrypolicy"
	"github.com/pkg/errors"

	config2 "github.com/Digital-Shane/treeview/extensions/s3/internal/config"
	"github.com/Digital-Shane/treeview/extensions/s3/internal/localstack"
)

var (
	// ErrMajorFailure is returned when an unexpected serious error occurred.  Contact the development team.
	ErrMajorFailure = errors.New("major failure")
	// ErrNoAccess is returned when the access to the object or bucket is not available.
	ErrNoAccess = errors.New("no access")
)

// GetClient returns a client for the bucket of `path`.  It handles the option WithProfile.
func GetClient(path string, opts ...Option) (*awsS3.Client, error) {
	b, _ := Parse(path)
	return newClientForBucket(b, opts...)
}

type getObjecter interface {
	GetObject(ctx context.Context, params *awsS3.GetObjectInput,
		optFns ...func(*awsS3.Options)) (*awsS3.GetObjectOutput, error)
}

// HasAccess returns true if the object or prefix at`path`can be accessed even if it does not yet exist.
//
// It supports the option WithProfile, and WithRetry.
func HasAccess(ctx context.Context, path string, opts ...Option) bool {
	if !IsWithRetry(opts...) {
		return hasAccess(ctx, path, opts...)
	}
	return hasAccessWithRetry(ctx, path, opts...)
}

// Join joins any number of path elements into a single s3URI. If the argument
// list is empty, Join returns an empty string.  If an elems is "", it is skipped.
// The function supports ".." path elements.
func Join(elems ...string) string {
	if len(elems) == 0 {
		return ""
	}
	s := path.Join(elems...)
	s = strings.TrimPrefix(strings.TrimPrefix(s, _cS3URI), "s3:/")
	return _cS3URI + s
}

// Size returns the size of the object located at `path`.
//
// It supports the option WithProfile.
func Size(ctx context.Context, path string, opts ...Option) (int64, error) {
	goo, err := getObject(ctx, path, opts...)
	if err != nil {
		return 0, err
	}
	_ = goo.Body.Close()
	return *goo.ContentLength, nil
}

// ---------------------

var errFailedFailSafe = errors.New("failed ")

// bucketFinder generates a new client for the bucket `b` and returns its region. It gracefully handles
// the use of local stack. It supports WithProfile.
func bucketFinder(b string, opts ...Option) (string, error) {
	// emulates curl -sI foo.s3.amazonaws.com | awk '/^x-amz-b	-region:/ { print $2 }'
	region := localstack.DefaultRegion
	if !localstack.InUse() {
		var client http.Client
		address := fmt.Sprintf("http://%s.s3.amazonaws.com", b)
		resp, err := client.Head(address)
		if err != nil {
			return "", err
		}
		_ = resp.Body.Close()
		regions, ok := resp.Header["X-Amz-Bucket-Region"]
		if !ok || len(regions) == 0 {
			return "", errors.New("not found")
		}
		region = regions[0]
	}
	oo := []config.LoadOptionsFunc{config.WithRegion(region)}
	oo = append(oo, collectCfg(opts...)...)
	_, err := newClient(oo...)
	if err != nil {
		return "", err
	}
	return region, nil
}

// getHeadObject returns the `s3.HeadObjectOutput` of the object at `path`.
func getHeadObject(ctx context.Context, path string, opts ...Option) (*awsS3.HeadObjectOutput, error) {
	c, err := GetClient(path, opts...)
	if err != nil {
		return nil, err
	}
	b, p := Parse(path)
	hoIn := &awsS3.HeadObjectInput{
		Bucket: aws.String(b),
		Key:    aws.String(p),
	}
	return c.HeadObject(ctx, hoIn)
}

// hasAccess returns true if the object or prefix at`path`can be accessed even if it does not yet exist.
//
// It supports the option WithProfile.
func hasAccess(ctx context.Context, path string, opts ...Option) bool {
	b, o := Parse(path)
	c, err := newClientForBucket(b, opts...)
	if err != nil {
		return false
	}
	_, err = c.HeadBucket(context.Background(), &awsS3.HeadBucketInput{Bucket: aws.String(b)})
	if err == nil {
		return true
	}
	o = strings.TrimSuffix(o, "/")
	if o == "" {
		return false
	}
	oo := strings.Split(o, "/")
	i := len(oo)
	for {
		key := strings.Join(oo[:i], "/") + "/"
		_, err = c.HeadObject(ctx, &awsS3.HeadObjectInput{Bucket: aws.String(b),
			Key: aws.String(key),
		})
		if err == nil {
			return true
		}
		i--
		if i == 0 {
			break
		}
	}
	return false
}

// newClient creates a new client of the default configuration initialized by the package with the options `opts`.
func newClient(opts ...config.LoadOptionsFunc) (*awsS3.Client, error) {
	cfg1, err := config2.NewConfig(opts...)
	if err != nil {
		return nil, err
	}
	if !localstack.InUse() {
		return awsS3.NewFromConfig(cfg1), nil
	}
	// See https://docs.localstack.cloud/aws/integrations/aws-sdks/go/
	return awsS3.NewFromConfig(cfg1, func(o *awsS3.Options) {
		o.UsePathStyle = true
		o.BaseEndpoint = aws.String(localstack.Endpoint)
	}), nil
}

// newClientForBucket creates a new client of the default configuration for the bucket `bucket`.  It supports the option
// WithProfile.
func newClientForBucket(bucket string, opts ...Option) (*awsS3.Client, error) {
	reg, err := bucketFinder(bucket, opts...)
	if err != nil {
		return nil, err
	}
	o := collectCfg(opts...)
	o = append(o, config.WithRegion(reg))
	return newClient(o...)
}

// getObject returns a GetObjectOutput from the object `path`.  It handles the option WithProfile and passes
// the other ones.
//
// It supports the option WithProfile.
func getObject(ctx context.Context, path string, opts ...Option) (*awsS3.GetObjectOutput, error) {
	c, err := GetClient(path, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "create client")
	}
	return getObject1(ctx, c, path, opts...)
}

// getObject1 returns a GetObjectOutput from the object `path` using the client `c`. It handles the options
// WithRange, and WithRetry.
// If WithRetry, the default triggered error is "slow down".  The option may add other ones.
func getObject1(ctx context.Context, c getObjecter, path string, opts ...Option) (*awsS3.GetObjectOutput, error) {
	o := collectOptions(opts...)
	b, p := parsePtr(path)

	goi := &awsS3.GetObjectInput{
		Bucket: b,
		Key:    p,
	}
	r1 := retrypolicy.NewBuilder[any]().WithMaxRetries(o.retryMaxAttempt).WithDelay(o.retryLapse).
		HandleIf(isRetryableError).Build()
	a, err := failsafe.Get(func() (any, error) {
		return c.GetObject(ctx, goi)
	}, r1)
	if err != nil {
		return nil, errors.WithMessagef(err, "key %s", path)
	}
	return a.(*awsS3.GetObjectOutput), nil
}

func getRetryPolicy(opts ...Option) retrypolicy.RetryPolicy[any] {
	o := collectOptions(opts...)
	retryPolicy := retrypolicy.NewBuilder[any]().WithMaxRetries(o.retryMaxAttempt).WithDelay(o.retryLapse).Build()
	return retryPolicy
}

func hasAccessWithRetry(ctx context.Context, path string, opts ...Option) bool {
	return failsafe.Run(func() error {
		if hasAccess(ctx, path, opts...) {
			return nil
		}
		return errFailedFailSafe
	}, getRetryPolicy(opts...)) == nil
}

func isRetryableError(_ any, err error) bool {
	if err == nil {
		return false
	}
	var ae smithy.APIError
	if errors.As(err, &ae) {
		switch ae.ErrorCode() { // Corrected dereferencing
		case "InternalError", "OperationAborted", "RequestTimeout", "ServiceUnavailable", "SlowDown":
			return true
		}
	}
	return false
}
