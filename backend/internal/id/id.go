// Package id generates identifiers for persisted entities.
package id

import (
	"fmt"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

// Project generates a project identifier.
func Project() (string, error) { return generate("prj_", 16) }

// File generates a file identifier.
func File() (string, error) { return generate("file_", 16) }

// Version generates a version identifier.
func Version() (string, error) { return generate("ver_", 16) }

// APITokenID generates an identifier for managing an API token.
func APITokenID() (string, error) { return generate("tkn_", 16) }

// APITokenCredential generates a secret credential independent of its management ID.
func APITokenCredential() (string, error) { return generate("api_", 50) }

func generate(prefix string, length int) (string, error) {
	random, err := gonanoid.Generate("0123456789abcdefghijklmnopqrstuvwxyz", length)
	if err != nil {
		return "", fmt.Errorf("generate ID: %w", err)
	}
	return prefix + random, nil
}
