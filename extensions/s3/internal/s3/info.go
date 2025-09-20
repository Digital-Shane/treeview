package s3

import (
	"context"
	"io/fs"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/pkg/errors"
)

const (
	_cS3URI = "s3://"
)

// Base returns the last element of `path`.
// Trailing path separators are removed before extracting the last element.
// If the path is empty, Base returns "."
func Base(path string) string {
	if path == "" {
		return "."
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
	cl, err := newClientForBucket(*b, opts...)
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

// Parse returns the bucket name and the key of `path`
func Parse(path string) (string, string) {
	const (
		cNumEle = 2
	)
	path = strings.TrimPrefix(path, _cS3URI)
	p := strings.SplitAfterN(path, "/", cNumEle)
	p[0] = strings.TrimSuffix(p[0], "/")
	if len(p) == 1 {
		return p[0], ""
	}
	return p[0], p[1]
}

// InfoDir holds information about S3 repertory.
type InfoDir struct {
	creationDate time.Time
	name         string
	dir          bool
	bucket       string
}

// Bucket returns the name of the bucket.  The name does not have the prefix "s3://"
func (id InfoDir) Bucket() string {
	if id.bucket == "" {
		return id.name
	}
	return id.bucket
}

// CreationDate returns the creation of the repertory.
func (id InfoDir) CreationDate() time.Time {
	return id.creationDate
}

// IsDir receiver is true if the information is related to a key that points
// to a "subdirectory".
func (id InfoDir) IsDir() bool {
	return id.dir
}

// Name returns the name of the object without the bucket.
func (id InfoDir) Name() string {
	if id.bucket == "" {
		return ""
	}
	return id.name
}

// ReadDir lists all the keys in the path. It returns a slice of `fs.DirEntry`. It hides the pagination, i.e.,
// it may return more than 1,000 objects.
// This version is really compliant with fs.DirEntry.  In ReadDir fs.DirEntry.Name() does not return the base.
func ReadDir(ctx context.Context, path string, opts ...Option) ([]fs.DirEntry, error) {
	de, err := list(ctx, path, opts...)
	if err != nil {
		return nil, err
	}
	var a []fs.DirEntry
	for _, entry := range de {
		ee := entry // as we use a pointer, we need to dereference entry.
		ee.name = Base(entry.Name())
		a = append(a, &ee)
	}
	return a, nil
}

// DirEntry is an entry read from a directory.  It implements the fs.DirEntry interface.
type DirEntry struct {
	InfoDir
	size int64
}

// Info returns the FileInfo for the file or subdirectory described by the entry.
func (de *DirEntry) Info() (fs.FileInfo, error) {
	return de, nil
}

// Name returns the name of the file (or subdirectory) described by the entry.
func (de *DirEntry) Name() string {
	return de.name
}

// IsDir reports whether the entry describes a directory.
func (de *DirEntry) IsDir() bool {
	return de.InfoDir.IsDir()
}

// Mode systematically returns fs.ModeIrregular.  It is necessary for fs.DirEntry interface compliance.
func (de *DirEntry) Mode() fs.FileMode {
	return fs.ModeIrregular
}

// ModTime returns the modification time of the entry.
func (de *DirEntry) ModTime() time.Time {
	return de.creationDate
}

// Size returns the size of the entry.
func (de *DirEntry) Size() int64 {
	// s, err := Size(context.Background(), UnParse(de.bucket, de.name))
	// if err != nil {
	// 	return 0
	// }
	// return s
	return de.size
}

// Sys returns.  It is needed for fs.DirEntry interface compliance.
func (de *DirEntry) Sys() interface{} {
	return nil
}

// Type returns systematically fs.ModeIrregular.  Need for fs.DirEntry interface compliance.
func (de *DirEntry) Type() fs.FileMode {
	return fs.ModeIrregular
}

// -------------------------------------------

// Function list method lists all the keys in the path.
// It returns a slice of `DirEntry`. It hides the pagination, i.e., it may return more than 1,000 objects.
// It is not recursive but lists the directories.
func list(ctx context.Context, pth string, opts ...Option) ([]DirEntry, error) {
	objects, err := getList(ctx, pth, opts...)
	if err != nil {
		return nil, err
	}
	knownDirs := make(map[string]bool) // to store the already detected subdirectories.
	var list []DirEntry
	b, dir := Parse(pth)
	for _, o := range objects {
		oWithoutPrefix := strings.TrimPrefix(*o.Key, strings.TrimSuffix(dir, "/")+"/")
		el := strings.Split(oWithoutPrefix, "/")
		switch len(el) {
		case 0:
			// SHOULD NEVER HAPPEN
			return nil, errors.Wrapf(ErrMajorFailure, "key %s", *o.Key)
		case 1:
			id := DirEntry{
				InfoDir: InfoDir{
					creationDate: *o.LastModified,
					name:         *o.Key,
					dir:          false,
				},
				size: *o.Size,
			}
			id.bucket = b
			list = append(list, id)
		default:
			_, ok := knownDirs[el[0]]
			if ok {
				continue // the subdirectory is already known
			}
			knownDirs[el[0]] = true // store the name of the subdirectory to avoid a second time.
			trn := path.Join(b, dir, el[0])
			id := DirEntry{
				InfoDir: InfoDir{
					creationDate: *o.LastModified,
					name:         trn,
					dir:          true,
				},
				size: 0,
			}
			id.bucket = b
			list = append(list, id)
		}
	}
	return list, nil
}

func getList(ctx context.Context, path string, opts ...Option) ([]types.Object, error) {
	objects, _, err := getListPossiblyRecurse(ctx, path, false, opts...)
	return objects, err
}

// getListPossiblyRecurse lists all the keys in `path`. If `recurse` is true, it also lists the keys in the
// subdirectories.  It returns a slice of `types.Object`, and a slice of common prefixes.
func getListPossiblyRecurse(ctx context.Context, path string, recurse bool, opts ...Option) ([]types.Object,
	[]types.CommonPrefix, error) {
	b, p := parsePtr(path)
	c, err := newClientForBucket(*b, opts...)
	if err != nil {
		return nil, nil, err
	}

	var objects []types.Object
	var commonPrefixes []types.CommonPrefix
	var token *string
	token = nil
	for {
		lov2i := &awsS3.ListObjectsV2Input{
			Bucket:            b,
			Prefix:            p,
			ContinuationToken: token,
		}
		if recurse {
			lov2i.Delimiter = aws.String("/")
		}
		lov2o, err := c.ListObjectsV2(ctx, lov2i)
		if err != nil {
			return nil, nil, err
		}
		objects = append(objects, lov2o.Contents...)
		commonPrefixes = append(commonPrefixes, lov2o.CommonPrefixes...)

		if !*lov2o.IsTruncated {
			break
		}
		token = lov2o.NextContinuationToken
	}
	return objects, commonPrefixes, nil
}

func parsePtr(path string) (*string, *string) {
	a, b := Parse(path)
	return aws.String(a), aws.String(b)
}
