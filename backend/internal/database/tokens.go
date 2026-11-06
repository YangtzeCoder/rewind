package database

import (
	"context"
	"database/sql"
)

type APIToken struct {
	ID        string
	Name      string
	CreatedAt int64
}

type TokenStore struct{ db *sql.DB }

func NewTokenStore(db *sql.DB) *TokenStore { return &TokenStore{db: db} }

func (s *TokenStore) Create(ctx context.Context, token APIToken, hash string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO api_tokens (id, name, token_hash, created_at) VALUES (?, ?, ?, ?)`, token.ID, token.Name, hash, token.CreatedAt)
	return err
}

func (s *TokenStore) List(ctx context.Context, limit, offset int64) ([]APIToken, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, created_at FROM api_tokens ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tokens := make([]APIToken, 0)
	for rows.Next() {
		var token APIToken
		if err := rows.Scan(&token.ID, &token.Name, &token.CreatedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}

func (s *TokenStore) Delete(ctx context.Context, id string) (bool, error) {
	result, err := s.db.ExecContext(ctx, `DELETE FROM api_tokens WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n > 0, err
}
