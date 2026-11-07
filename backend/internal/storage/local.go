package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Local struct {
	tmpDir string
	root   string
}

var _ Store = (*Local)(nil)

// NewLocal initializes local storage and clears uploads left by a previous run.
// Only one service instance may use a data directory at a time.
func NewLocal(dataDir string) (*Local, error) {
	if dataDir == "" {
		return nil, errors.New("data directory must not be empty")
	}
	s := &Local{tmpDir: filepath.Join(dataDir, "uploads-tmp"), root: filepath.Join(dataDir, "storage")}
	if err := os.RemoveAll(s.tmpDir); err != nil {
		return nil, fmt.Errorf("clear temporary uploads: %w", err)
	}
	for _, dir := range []string{s.tmpDir, s.root} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, fmt.Errorf("create storage directory: %w", err)
		}
	}
	return s, nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}

func (s *Local) Stage(ctx context.Context, src io.Reader) (StagedFile, FileInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, FileInfo{}, err
	}
	f, err := os.CreateTemp(s.tmpDir, "upload-*")
	if err != nil {
		return nil, FileInfo{}, err
	}
	path := f.Name()
	hash := sha256.New()
	size, copyErr := io.Copy(io.MultiWriter(f, hash), contextReader{ctx, src})
	closeErr := f.Close()
	if copyErr == nil {
		copyErr = ctx.Err()
	}
	if copyErr != nil || closeErr != nil {
		_ = os.Remove(path)
		return nil, FileInfo{}, errors.Join(copyErr, closeErr)
	}
	return &localStage{store: s, path: path}, FileInfo{Size: size, SHA256: hex.EncodeToString(hash.Sum(nil))}, nil
}

type localStage struct {
	store  *Local
	path   string
	mu     sync.Mutex
	closed bool
}

func (f *localStage) Commit(ctx context.Context, key VersionKey) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return ErrStageClosed
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	path, err := f.store.versionPath(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	// Linking creates the destination atomically without replacing an existing version.
	// Both directories are on the same filesystem; removing the temporary name
	// afterwards leaves the committed file in place without copying its contents.
	if err := os.Link(f.path, path); err != nil {
		if errors.Is(err, os.ErrExist) {
			return ErrAlreadyExists
		}
		return err
	}
	f.closed = true
	_ = os.Remove(f.path) // A leftover temporary name is cleared at startup.
	return nil
}

func (f *localStage) Abort(_ context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return nil
	}
	if err := os.Remove(f.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	f.closed = true
	return nil
}

func validID(value, prefix string) bool {
	if !strings.HasPrefix(value, prefix) || len(value) != len(prefix)+16 {
		return false
	}
	for _, c := range value[len(prefix):] {
		if !(c >= '0' && c <= '9') && !(c >= 'a' && c <= 'z') {
			return false
		}
	}
	return true
}

func (s *Local) versionPath(key VersionKey) (string, error) {
	if !validID(key.ProjectID, "prg_") || !validID(key.FileID, "file_") || !validID(key.VersionID, "ver_") {
		return "", ErrInvalidKey
	}
	return filepath.Join(s.root, key.ProjectID, key.FileID, key.VersionID), nil
}

func (s *Local) Open(ctx context.Context, key VersionKey) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path, err := s.versionPath(key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}
	return f, err
}

func (s *Local) Exists(ctx context.Context, key VersionKey) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	path, err := s.versionPath(key)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return err == nil, err
}

func (s *Local) Delete(ctx context.Context, key VersionKey) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path, err := s.versionPath(key)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (s *Local) DeleteProject(ctx context.Context, projectID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !validID(projectID, "prg_") {
		return ErrInvalidKey
	}
	return os.RemoveAll(filepath.Join(s.root, projectID))
}
