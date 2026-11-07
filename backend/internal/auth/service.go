// Package auth implements administrator authentication.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"rewind/internal/database"

	"golang.org/x/crypto/bcrypt"
)

var ErrUnauthorized = errors.New("unauthorized")

const sessionLifetime = 30 * time.Minute
const loginPrefix = "login_"

type Service struct {
	store                      *database.AuthStore
	initialAdminPasswordSource func() (string, error)
}

func NewService(store *database.AuthStore, initialAdminPasswordSource func() (string, error)) *Service {
	return &Service{store: store, initialAdminPasswordSource: initialAdminPasswordSource}
}

func (s *Service) Initialize(ctx context.Context) error {
	_, _, err := s.store.Admin(ctx)
	if err == nil {
		return nil
	}
	if !errors.Is(err, database.ErrNotFound) {
		return fmt.Errorf("check administrator: %w", err)
	}
	initialAdminPassword, err := s.initialAdminPasswordSource()
	if err != nil {
		return err
	}
	if initialAdminPassword == "" {
		return errors.New("initial administrator password is required; set REWIND_ADMIN_PASSWORD or REWIND_ADMIN_PASSWORD_FILE")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(initialAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash administrator password: %w", err)
	}
	if err := s.store.CreateAdmin(ctx, "admin", string(hash), time.Now().UnixMilli()); err != nil {
		return fmt.Errorf("create administrator: %w", err)
	}
	return nil
}

func (s *Service) Login(ctx context.Context, username, password string) (string, time.Time, error) {
	hash, err := s.store.PasswordHash(ctx, username)
	if errors.Is(err, database.ErrNotFound) {
		return "", time.Time{}, ErrUnauthorized
	}
	if err != nil {
		return "", time.Time{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return "", time.Time{}, ErrUnauthorized
	}
	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		return "", time.Time{}, err
	}
	token := loginPrefix + hex.EncodeToString(secret[:])
	now := time.Now().UTC()
	expires := now.Add(sessionLifetime)
	ok, err := s.store.CreateSession(ctx, username, hash, hashToken(token), now.UnixMilli(), expires.UnixMilli())
	if err != nil {
		return "", time.Time{}, err
	}
	if !ok {
		return "", time.Time{}, ErrUnauthorized
	}
	return token, expires, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) error {
	if !strings.HasPrefix(strings.ToLower(token), loginPrefix) {
		return ErrUnauthorized
	}
	now := time.Now().UnixMilli()
	ok, err := s.store.RenewSession(ctx, hashToken(token), now, now+sessionLifetime.Milliseconds())
	if err != nil {
		return err
	}
	if !ok {
		return ErrUnauthorized
	}
	return nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	return s.store.RevokeSession(ctx, hashToken(token), time.Now().UnixMilli())
}

func (s *Service) ChangePassword(ctx context.Context, current, next string) error {
	_, oldHash, err := s.store.Admin(ctx)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(oldHash), []byte(current)) != nil {
		return ErrUnauthorized
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(next), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	ok, err := s.store.UpdatePassword(ctx, oldHash, string(newHash), time.Now().UnixMilli())
	if err != nil {
		return err
	}
	if !ok {
		return ErrUnauthorized
	}
	return nil
}

func (s *Service) Username(ctx context.Context) (string, error) {
	username, _, err := s.store.Admin(ctx)
	return username, err
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(strings.ToLower(token)))
	return hex.EncodeToString(hash[:])
}
