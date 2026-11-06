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

// Delete removes metadata in dependency order. Version content storage must be
// incorporated here when backups are implemented.
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
