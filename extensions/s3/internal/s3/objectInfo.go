package s3

import (
	"context"
	"io/fs"
	"time"
)

const defaultMode = 0o755

// ObjectInfo holds the information related to an S3 object.  It implements the interface io/fs/FileInfo.
type ObjectInfo struct {
	// bucket provides the name of the bucket holding the object.
	bucket       string
	key          string
	lastModified time.Time
	size         int64
	storageClass string
	isDir        bool
}

// Info returns the ObjectInfo of the object at `path`.  If it is a directory, then IsDir() is true and Size() and
// ModTime() is meaningless.
func Info(ctx context.Context, path string, opts ...Option) (*ObjectInfo, error) {
	if !HasAccess(ctx, path, opts...) {
		return nil, ErrNoAccess
	}
	b, p := Parse(path)
	oi := &ObjectInfo{
		bucket:       b,
		key:          p,
		lastModified: time.Time{},
		size:         0,
		storageClass: "STANDARD",
		isDir:        true,
	}
	if p == "" {
		return oi, nil
	}
	oi.isDir = IsDir(ctx, path, opts...)
	if !oi.isDir {
		hoOut, err := getHeadObject(ctx, path, opts...)
		if err != nil {
			return nil, err
		}
		oi.lastModified = *hoOut.LastModified
		oi.size = *hoOut.ContentLength
		oi.storageClass = string(hoOut.StorageClass)
	}
	if oi.storageClass == "" {
		oi.storageClass = "STANDARD" // Provides storage class information of the object. Amazon S3 returns this header
		// for all objects except for S3 Standard storage class objects.
	}
	return oi, nil
}

// Name returns the base name of the object as defined by fs.FileInfo.
func (oi *ObjectInfo) Name() string {
	return Base(Join(oi.bucket, oi.key))
}

// Mode is mandatory for the fs.FileInfo interface.
func (oi *ObjectInfo) Mode() fs.FileMode {

	if oi.isDir {
		return fs.ModeDir | defaultMode
	}
	return fs.ModeIrregular
}

// ModTime returns the last modification time of the object.
func (oi *ObjectInfo) ModTime() time.Time {
	return oi.lastModified
}

// IsDir determines if the given ObjectInfo represents a directory based on its bucket and key attributes.
func (oi *ObjectInfo) IsDir() bool {
	return oi.isDir
}

// Sys is a placeholder method to satisfy the fs.FileInfo interface, always returning nil.
func (oi *ObjectInfo) Sys() any {
	return nil
}

// Bucket returns the name of the bucket holding the object.
func (oi *ObjectInfo) Bucket() string {
	return oi.bucket
}

// LastModified returns the last modification date of the object.
func (oi *ObjectInfo) LastModified() time.Time {
	return oi.lastModified
}

// Path returns the path of the object without the bucket.
func (oi *ObjectInfo) Path() string {
	return oi.key
}

// Size returns the size of the object.
func (oi *ObjectInfo) Size() int64 {
	return oi.size
}

// StorageClass returns the storage class of the object.  It has one of the following values: "STANDARD",
// "REDUCED_REDUNDANCY", "GLACIER", "STANDARD_IA", "ONEZONE_IA", "INTELLIGENT_TIERING", "DEEP_ARCHIVE", "OUTPOSTS",
// or "GLACIER_IR".
func (oi *ObjectInfo) StorageClass() string {
	return oi.storageClass
}
