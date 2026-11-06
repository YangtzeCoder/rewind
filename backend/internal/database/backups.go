package database

import (
	"context"
	"database/sql"
	"errors"
)

type Backup struct {
	ProjectID, Key, FileID, VersionID string
	Size                              int64
	SHA256, Description               string
	OriginalFilename, ContentType     string
	CreatedAt                         int64
}

type BackupStore struct{ db *sql.DB }

func NewBackupStore(db *sql.DB) *BackupStore { return &BackupStore{db: db} }

// History returns versions in insertion order, newest first, for an existing file.
func (s *BackupStore) History(ctx context.Context, projectID, key string, limit, offset int64) ([]Backup, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var fileID string
	err = tx.QueryRowContext(ctx, `SELECT id FROM files WHERE project_id = ? AND key = ?`, projectID, key).Scan(&fileID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id, size, description, created_at FROM versions
WHERE file_id = ? ORDER BY sequence DESC LIMIT ? OFFSET ?`, fileID, limit, offset)
	if err != nil {
		return nil, err
	}
	versions := make([]Backup, 0)
	for rows.Next() {
		var b Backup
		if err := rows.Scan(&b.VersionID, &b.Size, &b.Description, &b.CreatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		versions = append(versions, b)
	}
	err = rows.Err()
	if closeErr := rows.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return versions, nil
}

// Version finds a version only when it belongs to the specified project.
func (s *BackupStore) Version(ctx context.Context, projectID, versionID string) (Backup, error) {
	var b Backup
	err := s.db.QueryRowContext(ctx, `SELECT f.id, v.id, v.size, v.original_filename, v.content_type
FROM versions v JOIN files f ON f.id = v.file_id
WHERE f.project_id = ? AND v.id = ?`, projectID, versionID).
		Scan(&b.FileID, &b.VersionID, &b.Size, &b.OriginalFilename, &b.ContentType)
	if errors.Is(err, sql.ErrNoRows) {
		return Backup{}, ErrNotFound
	}
	b.ProjectID = projectID
	return b, err
}

// Create holds a write transaction while publishing the content. This also
// serializes publication with project deletion. The caller must remove published
// content if the database commit fails after publish succeeds.
func (s *BackupStore) Create(ctx context.Context, b Backup, publish func(fileID string) error) (Backup, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Backup{}, err
	}
	defer tx.Rollback()
	// Start with a write, rather than upgrading a read transaction, so concurrent
	// uploads can wait for the writer without a SQLite snapshot upgrade conflict.
	err = tx.QueryRowContext(ctx, `INSERT INTO files (id, project_id, key, created_at, updated_at)
SELECT ?, id, ?, ?, ? FROM projects WHERE id = ?
ON CONFLICT (project_id, key) DO UPDATE SET updated_at = excluded.updated_at
RETURNING id`, b.FileID, b.Key, b.CreatedAt, b.CreatedAt, b.ProjectID).Scan(&b.FileID)
	if errors.Is(err, sql.ErrNoRows) {
		return Backup{}, ErrNotFound
	}
	if err != nil {
		return Backup{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO versions
(id, file_id, size, sha256, description, original_filename, content_type, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, b.VersionID, b.FileID, b.Size, b.SHA256, b.Description, b.OriginalFilename, b.ContentType, b.CreatedAt); err != nil {
		return Backup{}, err
	}
	if err := publish(b.FileID); err != nil {
		return Backup{}, err
	}
	if err := tx.Commit(); err != nil {
		return Backup{}, err
	}
	return b, nil
}
