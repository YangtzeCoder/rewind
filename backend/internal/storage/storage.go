// Package storage defines version content storage independently of its backend.
package storage

import (
	"context"
	"errors"
	"io"
)

var (
	ErrNotFound      = errors.New("version file not found")
	ErrAlreadyExists = errors.New("version file already exists")
	ErrInvalidKey    = errors.New("invalid version key")
	ErrStageClosed   = errors.New("staged file is closed")
)

type VersionKey struct {
	ProjectID string
	FileID    string
	VersionID string
}

type FileInfo struct {
	Size   int64
	SHA256 string
}

// StagedFile is a temporary upload. Abort is safe after commit and on repeated calls.
type StagedFile interface {
	Commit(ctx context.Context, key VersionKey) error
	Abort(ctx context.Context) error
}

type Store interface {
	Stage(ctx context.Context, src io.Reader) (StagedFile, FileInfo, error)
	Open(ctx context.Context, key VersionKey) (io.ReadCloser, error)
	Exists(ctx context.Context, key VersionKey) (bool, error)
	Delete(ctx context.Context, key VersionKey) error
	DeleteProject(ctx context.Context, projectID string) error
}
