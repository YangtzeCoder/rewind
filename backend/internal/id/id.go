// Package id generates identifiers for persisted entities.
package id

import (
	"fmt"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

// Project generates a project identifier.
func Project() (string, error) { return generate("PRG_") }

// File generates a file identifier.
func File() (string, error) { return generate("FILE_") }

// Version generates a version identifier.
func Version() (string, error) { return generate("VER_") }

// APIToken generates an API token identifier, not the token's secret.
func APIToken() (string, error) { return generate("TKN_") }

func generate(prefix string) (string, error) {
	random, err := gonanoid.Generate("0123456789abcdefghijklmnopqrstuvwxyz", 16)
	if err != nil {
		return "", fmt.Errorf("generate ID: %w", err)
	}
	return prefix + random, nil
}
