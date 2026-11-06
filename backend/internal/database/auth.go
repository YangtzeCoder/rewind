package database

import (
	"context"
	"database/sql"
	"errors"
)

// ErrNotFound indicates that the requested administrator does not exist.
var ErrNotFound = errors.New("administrator not found")

type AuthStore struct{ db *sql.DB }

func NewAuthStore(db *sql.DB) *AuthStore { return &AuthStore{db: db} }

func (s *AuthStore) Admin(ctx context.Context) (username, passwordHash string, err error) {
	err = s.db.QueryRowContext(ctx, `SELECT username, password_hash FROM admin WHERE id = 1`).Scan(&username, &passwordHash)
	if errors.Is(err, sql.ErrNoRows) {
		err = ErrNotFound
	}
	return
}

func (s *AuthStore) PasswordHash(ctx context.Context, username string) (string, error) {
	var hash string
	err := s.db.QueryRowContext(ctx, `SELECT password_hash FROM admin WHERE username = ?`, username).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return hash, err
}

func (s *AuthStore) CreateAdmin(ctx context.Context, username, passwordHash string, now int64) error {
	_, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO admin (id, username, password_hash, created_at, updated_at) VALUES (1, ?, ?, ?, ?)`, username, passwordHash, now, now)
	return err
}

// CreateSession only succeeds if the credentials have not changed since verification.
func (s *AuthStore) CreateSession(ctx context.Context, username, passwordHash, tokenHash string, now, expires int64) (bool, error) {
	result, err := s.db.ExecContext(ctx, `INSERT INTO login_tokens (token_hash, created_at, expires_at, revoked_at)
SELECT ?, ?, ?, ? FROM admin WHERE username = ? AND password_hash = ?`, tokenHash, now, expires, now, username, passwordHash)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n == 1, err
}

func (s *AuthStore) RenewSession(ctx context.Context, tokenHash string, now, expires int64) (bool, error) {
	result, err := s.db.ExecContext(ctx, `UPDATE login_tokens SET expires_at = ? WHERE token_hash = ? AND revoked = 0 AND expires_at > ?`, expires, tokenHash, now)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n == 1, err
}

func (s *AuthStore) RevokeSession(ctx context.Context, tokenHash string, now int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE login_tokens SET revoked = 1, revoked_at = ? WHERE token_hash = ? AND revoked = 0`, now, tokenHash)
	return err
}

// UpdatePassword succeeds only if the password has not changed since verification.
func (s *AuthStore) UpdatePassword(ctx context.Context, oldHash, newHash string, now int64) (bool, error) {
	result, err := s.db.ExecContext(ctx, `UPDATE admin SET password_hash = ?, updated_at = ? WHERE id = 1 AND password_hash = ?`, newHash, now, oldHash)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n == 1, err
}
