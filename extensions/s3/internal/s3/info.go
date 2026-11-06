package s3

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/pkg/errors"
)

// Base returns the last element of `path`.
// Trailing path separators are removed before extracting the last element.
// If the path is empty, Base returns "."
func Base(path string) string {
	if path == "" {
		return "."
	}
	if !strings.HasPrefix(path, URI) {
		return filepath.Base(path)
	}
	_, p := Parse(path)
	l := strings.Split(strings.TrimRight(p, "/"), "/")
	return l[len(l)-1]
}

// IsDir informs whether the path is a key with objects.  The path may end with "/".
// The root of an accessible bucket is a directory.
//
//	If the bucket is not accessible, it returns false.
func IsDir(ctx context.Context, path string, opts ...Option) bool {
	b, p := parsePtr(path)
	cl, err := NewClientForBucket(ctx, *b, opts...)
	if err != nil {
		return false
	}
	var token *string
	token = nil
	for {
		lov2i := &awsS3.ListObjectsV2Input{
			Bucket:            b,
			Prefix:            p,
			ContinuationToken: token,
		}
		lov2o, err := cl.ListObjectsV2(ctx, lov2i)
		if err != nil {
			return false
		}
		if strings.HasSuffix(path, "/") {
			return true // It exists, and it ends with /, thus it is a dir.
		}
		for _, o := range lov2o.Contents {
			oWithoutPrefix := strings.TrimPrefix(*o.Key, *p)
			if strings.HasPrefix(oWithoutPrefix, "/") {
				return true
			}
		}
		if !*lov2o.IsTruncated {
			break
		}
		token = lov2o.NextContinuationToken
	}
	return false
}

// ListTopObjectsAndPrefixes lists all objects and prefixes in the specified path and returns a slice of
// ObjectInfo.  It is not recursive.  The prefixes are reported as directories.
// It hides the pagination, i.e., it may return more than 1,000 objects.
func ListTopObjectsAndPrefixes(ctx context.Context, path string, opts ...Option) ([]ObjectInfo, error) {
	return getListObjectsNonRecurse(ctx, path, false, opts...)
}

// -------------------------------------------

func getListObjectsNonRecurse(ctx context.Context, path string, withoutDir bool, opts ...Option) ([]ObjectInfo, error) {
	b, p := parsePtr(path)
	c, err := NewClientForBucket(ctx, *b, opts...)
	if err != nil {
		return nil, err
	}
	var ois []ObjectInfo
	commonPrefixes := map[string]bool{}
	prefix := strings.TrimSuffix(*p, "/") + "/"
	if prefix == "/" {
		prefix = ""
	}
	var token *string
	token = nil
	for {
		lov2i := &awsS3.ListObjectsV2Input{
			Bucket:            b,
			Prefix:            aws.String(prefix),
			ContinuationToken: token,
		}
		lov2o, err := c.ListObjectsV2(ctx, lov2i)
		if err != nil {
			return nil, errors.Wrap(err, "ListObjectsV2")
		}
		for _, content := range lov2o.Contents {
			key := strings.TrimPrefix(*content.Key, prefix)
			if key == "" {
				continue // the folder marker of `path` itself
			}
			l := strings.Split(key, "/")
			if len(l) == 1 {
				ois = append(ois, typesObjectToObjectInfo(&content, *b))
				continue
			}
			if withoutDir {
				continue
			}
			_, ok := commonPrefixes[l[0]]
			if !ok {
				commonPrefixes[l[0]] = true
				ois = append(ois, ObjectInfo{
					bucket:       *b,
					key:          prefix + l[0] + "/",
					lastModified: time.Time{},
					size:         0,
					storageClass: "",
					isDir:        true,
				})
			}
		}
		if !*lov2o.IsTruncated {
			break
		}
		token = lov2o.NextContinuationToken
	}
	return ois, nil
}

func typesObjectToObjectInfo(to *types.Object, bucket string) ObjectInfo {
	return ObjectInfo{
		bucket:       bucket,
		key:          *to.Key,
		lastModified: *to.LastModified,
		size:         *to.Size,
		storageClass: string(to.StorageClass),
		isDir:        false,
	}
}
