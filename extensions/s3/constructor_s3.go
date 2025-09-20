// Package s3 provides a treeview.Tree constructor dedicated to AWS S3.
package s3

import (
	"context"
	"fmt"

	"github.com/Digital-Shane/treeview"
	"github.com/Digital-Shane/treeview/extensions/s3/internal/s3"
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
//   - treeview.WithProgressCallback: Invoked after each filesystem entry is processed (breadth-first per directory)
func NewTreeFromS3(ctx context.Context, path string, profile string,
	opts ...treeview.Option[treeview.FileInfo]) (*treeview.Tree[treeview.FileInfo], error) {
	cfg := treeview.NewMasterConfig(opts, treeview.WithProvider[treeview.FileInfo](treeview.NewDefaultNodeProvider(
		treeview.WithFileExtensionRules[treeview.FileInfo](),
	)))
	nodes, err := buildFileSystemTreeForS3(ctx, path, profile, cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", treeview.ErrFileSystem, err)
	}
	tree := treeview.NewTreeFromCfg(nodes, cfg)
	return tree, nil
}

func buildFileSystemTreeForS3(ctx context.Context, path string, profile string,
	cfg *treeview.MasterConfig[treeview.FileInfo]) ([]*treeview.Node[treeview.FileInfo], error) {
	info, err := s3.Info(ctx, path, s3.WithProfile(profile))
	if err != nil {
		return nil, pathError(treeview.ErrPathResolution, path, err)
	}
	total := 1
	rootNode := treeview.NewFileSystemNode(path, info)
	cfg.HandleExpansion(rootNode)
	if info.IsDir() {
		if err := scanDirS3(ctx, rootNode, 0, false, cfg, &total); err != nil {
			return nil, err
		}
	}
	return []*treeview.Node[treeview.FileInfo]{rootNode}, nil
}

// scanDirS3 scans a bucket or key and its subdirectories, creating Node[treeview.FileInfo] for each entry.
// It returns an error if the traversal cap is exceeded or if there is an error.
func scanDirS3(ctx context.Context, parent *treeview.Node[treeview.FileInfo], depth int, followSymlinks bool,
	cfg *treeview.MasterConfig[treeview.FileInfo], count *int) error {
	if cfg.HasDepthLimitBeenReached(depth) {
		return nil
	}
	entries, err := s3.ReadDir(ctx, parent.Data().Path)
	if err != nil {
		return pathError(treeview.ErrDirectoryScan, parent.Data().Path, err)
	}
	children := make([]*treeview.Node[treeview.FileInfo], 0, len(entries)) // preallocation of the capacity only.
	for _, entry := range entries {
		// Check for cancellation between entries
		if err := ctx.Err(); err != nil {
			return err
		}
		childPath := s3.Join(parent.Data().Path, entry.Name()) // entry.Name is the full key.
		info, err := entry.Info()
		if err != nil {
			return pathError(treeview.ErrFileSystem, childPath, err)
		}
		if cfg.ShouldFilter(treeview.FileInfo{
			FileInfo: info,
			Path:     childPath,
		}) {
			continue // Item was filtered out
		}
		childNode := treeview.NewFileSystemNode(childPath, info)
		cfg.HandleExpansion(childNode)
		*count++
		cfg.ReportProgress(*count, childNode)
		if cfg.HasTraversalCapBeenReached(*count) {
			return pathError(treeview.ErrTraversalLimit, childPath, nil)
		}
		if info.IsDir() {
			if err := scanDirS3(ctx, childNode, depth+1, followSymlinks, cfg, count); err != nil {
				return err
			}
		}
		children = append(children, childNode)
	}
	if len(children) > 0 {
		parent.SetChildren(children)
	}
	return nil
}

// pathError creates an error that includes path context.
// It's used internally for file system operations where the path is important.
func pathError(sentinel error, path string, cause error) error {
	if cause == nil {
		return fmt.Errorf("%w: %s", sentinel, path)
	}
	return fmt.Errorf("%w: %s: %w", sentinel, path, cause)
}
