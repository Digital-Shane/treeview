// Package s3 provides a treeview.Tree constructor dedicated to AWS S3.
package s3

import (
	"context"
	"errors"
	"fmt"

	"github.com/Digital-Shane/treeview"
	internals3 "github.com/Digital-Shane/treeview/extensions/s3/internal/s3"
)

// NewTreeFromS3 creates a new tree structure based on files fetched from an S3 path, using configurable options.
// Returns a pointer to a Tree structure or an error if an issue occurs during tree creation.
//
// Supported options:
// Build options:
//   - treeview.WithFilterFunc:   Filters items during tree building
//   - treeview.WithMaxDepth:     Limits tree depth during construction
//   - treeview.WithExpandFunc:   Sets initial expansion state for nodes
//   - treeview.WithTraversalCap: Limits total nodes processed (returns a partial tree + error if exceeded)
//   - treeview.WithProgressCallback: Invoked after each node is created during traversal
func NewTreeFromS3(ctx context.Context, path string, profile string,
	opts ...treeview.Option[treeview.FileInfo]) (*treeview.Tree[treeview.FileInfo], error) {
	allOpts := append([]treeview.Option[treeview.FileInfo]{treeview.WithProvider[treeview.FileInfo](treeview.NewDefaultNodeProvider(
		treeview.WithFileExtensionRules[treeview.FileInfo](),
	))}, opts...)

	tree, err := treeview.NewTreeFromWalker(ctx, &walker{path: path, profile: profile}, allOpts...)
	if err != nil {
		return tree, wrapConstructorError(err)
	}

	return tree, nil
}

func wrapConstructorError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, treeview.ErrTraversalLimit) {
		return err
	}
	if errors.Is(err, treeview.ErrFileSystem) {
		return err
	}
	return fmt.Errorf("%w: %w", treeview.ErrFileSystem, err)
}

type walker struct {
	path    string
	profile string
}

func (w *walker) Root(ctx context.Context) (treeview.WalkItem[treeview.FileInfo], error) {
	info, err := internals3.Info(ctx, w.path, internals3.WithProfile(w.profile))
	if err != nil {
		return treeview.WalkItem[treeview.FileInfo]{}, pathError(treeview.ErrPathResolution, w.path, err)
	}

	return treeview.WalkItem[treeview.FileInfo]{
		ID:   w.path,
		Name: info.Name(),
		Data: treeview.FileInfo{FileInfo: info, Path: w.path},
	}, nil
}

func (w *walker) Children(ctx context.Context, parent treeview.WalkItem[treeview.FileInfo]) ([]treeview.WalkItem[treeview.FileInfo], error) {
	if !parent.Data.IsDir() {
		return nil, nil
	}

	entries, err := internals3.ReadDir(ctx, parent.Data.Path)
	if err != nil {
		return nil, pathError(treeview.ErrDirectoryScan, parent.Data.Path, err)
	}

	children := make([]treeview.WalkItem[treeview.FileInfo], 0, len(entries))
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		childPath := internals3.Join(parent.Data.Path, entry.Name())
		info, err := entry.Info()
		if err != nil {
			return nil, pathError(treeview.ErrFileSystem, childPath, err)
		}

		children = append(children, treeview.WalkItem[treeview.FileInfo]{
			ID:   childPath,
			Name: info.Name(),
			Data: treeview.FileInfo{FileInfo: info, Path: childPath},
		})
	}

	return children, nil
}

// pathError creates an error that includes path context.
// It's used internally for file system operations where the path is important.
func pathError(sentinel error, path string, cause error) error {
	if cause == nil {
		return fmt.Errorf("%w: %s", sentinel, path)
	}
	return fmt.Errorf("%w: %s: %w", sentinel, path, cause)
}
