package database

import (
	"context"
	"database/sql"
	"errors"
)

type Project struct {
	ID, Name, Description string
	CreatedAt, UpdatedAt  int64
}

type ProjectStore struct{ db *sql.DB }

func NewProjectStore(db *sql.DB) *ProjectStore { return &ProjectStore{db: db} }

type Dashboard struct {
	ProjectCount, FileCount, VersionCount, TotalSize int64
}

// Dashboard reads all counts from the same SQLite snapshot. TotalSize counts
// every stored version, including multiple versions with identical content.
func (s *ProjectStore) Dashboard(ctx context.Context) (Dashboard, error) {
	var d Dashboard
	err := s.db.QueryRowContext(ctx, `SELECT
(SELECT COUNT(*) FROM projects),
(SELECT COUNT(*) FROM files),
(SELECT COUNT(*) FROM versions),
(SELECT COALESCE(SUM(size), 0) FROM versions)`).Scan(&d.ProjectCount, &d.FileCount, &d.VersionCount, &d.TotalSize)
	return d, err
}

type ProjectFile struct {
	ID, Key                string
	CreatedAt, UpdatedAt   int64
	VersionCount           int64
	LatestVersionID        sql.NullString
	LatestSize, LatestTime sql.NullInt64
}

// ListFiles verifies the project and reads the page in the same snapshot.
func (s *ProjectStore) ListFiles(ctx context.Context, projectID, keyFilter string, limit, offset int64) ([]ProjectFile, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM projects WHERE id = ?)`, projectID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}
	rows, err := tx.QueryContext(ctx, `SELECT f.id, f.key, f.created_at, f.updated_at,
(SELECT COUNT(*) FROM versions WHERE file_id = f.id), v.id, v.size, v.created_at
FROM files f LEFT JOIN versions v ON v.sequence =
 (SELECT sequence FROM versions WHERE file_id = f.id ORDER BY sequence DESC LIMIT 1)
WHERE f.project_id = ? AND (? = '' OR instr(f.key, ?) > 0)
ORDER BY f.updated_at DESC, f.id DESC LIMIT ? OFFSET ?`, projectID, keyFilter, keyFilter, limit, offset)
	if err != nil {
		return nil, err
	}
	files := make([]ProjectFile, 0)
	for rows.Next() {
		var f ProjectFile
		if err := rows.Scan(&f.ID, &f.Key, &f.CreatedAt, &f.UpdatedAt, &f.VersionCount, &f.LatestVersionID, &f.LatestSize, &f.LatestTime); err != nil {
			rows.Close()
			return nil, err
		}
		files = append(files, f)
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
	return files, nil
}

func (s *ProjectStore) Create(ctx context.Context, p Project) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO projects (id, name, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, p.ID, p.Name, p.Description, p.CreatedAt, p.UpdatedAt)
	return err
}

func (s *ProjectStore) List(ctx context.Context, limit, offset int64) ([]Project, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, description, created_at, updated_at FROM projects ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	projects := make([]Project, 0)
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (s *ProjectStore) Get(ctx context.Context, id string) (Project, error) {
	var p Project
	err := s.db.QueryRowContext(ctx, `SELECT id, name, description, created_at, updated_at FROM projects WHERE id = ?`, id).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrNotFound
	}
	return p, err
}

func (s *ProjectStore) Update(ctx context.Context, id string, name, description *string, updatedAt int64) (Project, error) {
	var p Project
	err := s.db.QueryRowContext(ctx, `UPDATE projects SET name = COALESCE(?, name), description = COALESCE(?, description), updated_at = ? WHERE id = ? RETURNING id, name, description, created_at, updated_at`, name, description, updatedAt, id).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrNotFound
	}
	return p, err
}

// Delete removes metadata in dependency order. The caller removes version content
// after this transaction commits, when new uploads can no longer target the project.
func (s *ProjectStore) Delete(ctx context.Context, id string) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM versions WHERE file_id IN (SELECT id FROM files WHERE project_id = ?)`, id); err != nil {
		return false, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM files WHERE project_id = ?`, id); err != nil {
		return false, err
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return n > 0, nil
}
