package helper

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 3         // Number of iterations
	argonMemory  = 64 * 1024 // 64 MB
	argonThreads = 2         // Parallelism
	argonKeyLen  = 32        // Output key length in bytes
	argonSaltLen = 16        // Salt length in bytes
)

// Returns the hashed password in the format: $argon2id$salt$hash
func HashPassword(password string) (string, error) {
	// Generate a random salt
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash the password using Argon2id
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		argonTime,
		argonMemory,
		argonThreads,
		argonKeyLen,
	)

	// Encode salt and hash as base64 and store in a parseable format
	saltEncoded := base64.RawStdEncoding.EncodeToString(salt)
	hashEncoded := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$salt$hash
	encodedHash := fmt.Sprintf("$argon2id$%s$%s", saltEncoded, hashEncoded)

	return encodedHash, nil
}

// VerifyPassword verifies a password against an Argon2id hash
// Returns true if the password matches the hash, false otherwise
func VerifyPassword(password, encodedHash string) (bool, error) {
	// Parse the encoded hash
	parts := strings.Split(encodedHash, "$")

	// Check if it's an Argon2id hash
	if len(parts) != 4 || parts[0] != "" || parts[1] != "argon2id" {
		return false, fmt.Errorf("invalid hash format")
	}

	// Decode salt and hash
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// Hash the provided password with the same salt
	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		argonTime,
		argonMemory,
		argonThreads,
		argonKeyLen,
	)

	// Constant-time comparison
	if len(hash) != len(computedHash) {
		return false, nil
	}

	match := true
	for i := range hash {
		if hash[i] != computedHash[i] {
			match = false
		}
	}

	return match, nil
}
