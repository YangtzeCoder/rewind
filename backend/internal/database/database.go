// Package database initializes the SQLite metadata database.
package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Open creates the data directory and applies database migrations.
func Open(dataDir string) (*sql.DB, error) {
	if dataDir == "" {
		return nil, fmt.Errorf("data directory must not be empty")
	}
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}
	path, err := filepath.Abs(filepath.Join(dataDir, "rewind.db"))
	if err != nil {
		return nil, fmt.Errorf("resolve database path: %w", err)
	}
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(path)}
	q := u.Query()
	q.Add("_pragma", "foreign_keys(1)")
	q.Add("_pragma", "busy_timeout(5000)")
	u.RawQuery = q.Encode()
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize database: %w", err)
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var version int
	if err := tx.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	if version > 2 {
		return fmt.Errorf("unsupported database schema version %d", version)
	}
	if version == 0 {
		if _, err := tx.Exec(schemaV1); err != nil {
			return err
		}
		if _, err := tx.Exec("PRAGMA user_version = 1"); err != nil {
			return err
		}
		version = 1
	}
	if version == 1 {
		if _, err := tx.Exec(`ALTER TABLE admin ADD COLUMN username TEXT NOT NULL DEFAULT 'admin';
CREATE UNIQUE INDEX admin_username ON admin(username);`); err != nil {
			return err
		}
		if _, err := tx.Exec("PRAGMA user_version = 2"); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Times are UTC Unix milliseconds. revoked_at is initialized to created_at by
// the token creation code and is meaningful as a revocation time only if revoked.
const schemaV1 = `
CREATE TABLE admin (
 id INTEGER PRIMARY KEY CHECK (id = 1), password_hash TEXT NOT NULL,
 created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL
);
CREATE TABLE login_tokens (
 id INTEGER PRIMARY KEY, token_hash TEXT NOT NULL UNIQUE,
 created_at INTEGER NOT NULL, expires_at INTEGER NOT NULL,
 revoked INTEGER NOT NULL DEFAULT 0 CHECK (revoked IN (0, 1)), revoked_at INTEGER NOT NULL
);
CREATE TABLE api_tokens (
 id TEXT PRIMARY KEY NOT NULL, name TEXT NOT NULL, token_hash TEXT NOT NULL UNIQUE,
 created_at INTEGER NOT NULL
);
CREATE TABLE projects (
 id TEXT PRIMARY KEY NOT NULL, name TEXT NOT NULL, description TEXT NOT NULL DEFAULT '',
 created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL
);
CREATE TABLE files (
 id TEXT PRIMARY KEY NOT NULL, project_id TEXT NOT NULL REFERENCES projects(id),
 key TEXT NOT NULL COLLATE BINARY, created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL,
 UNIQUE (project_id, key)
);
CREATE TABLE versions (
 sequence INTEGER PRIMARY KEY AUTOINCREMENT, id TEXT NOT NULL UNIQUE,
 file_id TEXT NOT NULL REFERENCES files(id), size INTEGER NOT NULL CHECK (size >= 0),
 sha256 TEXT NOT NULL, description TEXT NOT NULL DEFAULT '' CHECK (length(description) <= 1024),
 created_at INTEGER NOT NULL
);
CREATE INDEX versions_file_sequence ON versions(file_id, sequence DESC);
`
