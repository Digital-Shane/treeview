package localstack

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Digital-Shane/treeview/extensions/s3/internal/config"
	"github.com/Digital-Shane/treeview/extensions/s3/internal/localstack/exec"
)

const (
	// DefaultRegion is the default region used by localstack.
	DefaultRegion = "us-west-2"
	// Endpoint is the localstack endpoint.
	Endpoint = "http://localhost:4566"
	// Env is the environment variable name for the localstack use.  If set, the aws
	// library uses the localstack emulator.  The preferred way to set this variable is to use
	// `localstack.Use()` and `localstack.UseNot()`.
	Env = config.Env
)

// InUse returns true if the application uses the localstack emulator.  LocalStack may run but not used by
// the application.
func InUse() bool {
	return os.Getenv(Env) != ""
}

// Use indicates that the localstack emulator should be used.
func Use() error {
	config.ClearCached()
	return os.Setenv(Env, "true")
}

// UseNot indicates that the localstack emulator should not be used anymore.
func UseNot() error {
	config.ClearCached()
	return os.Unsetenv(Env)
}

const _localstackURL = "https://localhost.localstack.cloud:4566"

var _endpointURL = "--endpoint-url=" + _localstackURL

// ErrInvalidBucketName is returned when the bucket name is invalid.
var ErrInvalidBucketName = errors.New("invalid bucket name")

// Call is the equivalent to call `awslocal` from the command line.
func Call(args ...string) ([]byte, error) {
	a := []string{_endpointURL}
	return exec.Run("aws", exec.WithArgs(append(a, args...)...), exec.WithVerbose())
}

// CallSilent is the equivalent to call `awslocal` from the command line.
func CallSilent(args ...string) ([]byte, error) {
	a := []string{_endpointURL}
	return exec.Run("aws", exec.WithArgs(append(a, args...)...))
}

// CreateBucket creates the bucket `bktName` on the localstack.
// For testing purpose exclusively.
// It supports the option WithNoErrorIfExist.
func CreateBucket(bktName string, opt ...Option) error {
	return CreateBucketAt(bktName, DefaultRegion, opt...)
}

// CreateBucketAt creates the bucket `bktName` on the localstack in the AWS region.
// For testing purpose exclusively.
// It supports the option WithNoErrorIfExist.
func CreateBucketAt(bktName string, region string, opt ...Option) error {
	b := parse(bktName)
	if !IsBucketNameValid(b) {
		return ErrInvalidBucketName
	}
	if collectOptions(opt...).noErrIfExist {
		_, _ = CallSilent("s3api", "create-bucket",
			"--bucket", b,
			"--region", region,
			"--create-bucket-configuration", "{\"LocationConstraint\": \""+region+"\"}")
		return nil
	}
	_, err := Call("s3api", "create-bucket",
		"--bucket", b,
		"--region", region,
		"--create-bucket-configuration", "{\"LocationConstraint\": \""+region+"\"}")

	return err
}

// DeleteBucket deletes the bucket `bktName` on the localstack.
func DeleteBucket(bktName string) error {
	return DeleteBucketAt(bktName, DefaultRegion)
}

// DeleteBucketAt deletes the bucket `bktName` on the localstack.  The bucket does not need to be empty.
func DeleteBucketAt(bktName string, region string) error {
	type answer struct {
		Contents []struct {
			Key          string    `json:"Key"`
			LastModified time.Time `json:"LastModified"`
			ETag         string    `json:"ETag"`
			Size         int       `json:"Size"`
			StorageClass string    `json:"StorageClass"`
		} `json:"Contents"`
	}
	b := parse(bktName)
	// list all potential objects
	data, err := Call("s3api", "list-objects-v2", "--bucket", b, "--region", region)
	if err != nil {
		return err
	}
	if len(data) != 0 { // if there are objects
		var a answer
		err = json.Unmarshal(data, &a)
		if err != nil {
			return err
		}
		for _, c := range a.Contents { // delete all objects
			_, err = Call("s3api", "delete-object", "--bucket", b, "--key", c.Key, "--region", region)
			if err != nil {
				return err
			}
		}
	}
	// delete the emptied bucket
	_, err = Call("s3api", "delete-bucket", "--bucket", b, "--region", region)
	return err
}

// IsBucketNameValid returns true if the bucket name is syntactically valid.
func IsBucketNameValid(b string) bool {
	return regexp.MustCompile(`^[a-z0-9.\-]+$`).MatchString(b) && len(b) >= 1 && len(b) <= 63
}

// IsExist checks whether the object `objectPath` exists on the localstack.
func IsExist(objectPath string) bool {
	a, _ := exec.Run("aws", exec.WithArgs(_endpointURL, "s3", "ls", objectPath))
	return len(a) > 0
}

// IsRunning checks if the localstack is running.
func IsRunning() bool {
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get(_localstackURL)
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return true
}

// PutObject pushes the file `source` on the localstack as object `objectPath`.
// For testing purpose exclusively.  If it is already in the bucket, the operation is skipped.
func PutObject(objectPath string, source string) error {
	if IsExist(objectPath) {
		return nil
	}
	_, err := Call("s3", "cp", "--region", DefaultRegion, source, objectPath)
	return err
}

// parse returns the bucket name
func parse(path string) string {
	const (
		cNumEle = 2
	)
	path = strings.TrimPrefix(path, "s3://")
	p := strings.SplitAfterN(path, "/", cNumEle)
	p[0] = strings.TrimSuffix(p[0], "/")
	if len(p) == 1 {
		return p[0]
	}
	return p[0]
}
