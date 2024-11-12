package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken generates a JWT token containing the password hash and expiration time
func GenerateToken(password string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256) // Create a new token with HMAC-SHA256

	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = "user"                               // Set the subject of the token
	claims["exp"] = time.Now().Add(8 * time.Hour).Unix() // Set the token expiration time (8 hours)
	claims["hash"] = GeneratePasswordHash(password)      // Password hash for future validation

	// Sign the token using the password
	tokenString, err := token.SignedString([]byte(password))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken parses the JWT token and checks its validity
func ParseToken(tokenString, password string) (*jwt.Token, error) {
	// Parse the token and check the signing method
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(password), nil // Return the key (password)
	})
	if err != nil || !token.Valid {
		return nil, err // Error if the token is invalid
	}

	// Check the token claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Compare the password hash to confirm authenticity
	if claims["hash"] != GeneratePasswordHash(password) {
		return nil, fmt.Errorf("invalid token hash")
	}

	return token, nil
}

// GeneratePasswordHash generates a password hash using SHA-256
func GeneratePasswordHash(password string) string {
	hash := sha256.Sum256([]byte(password)) // Hash the password
	return hex.EncodeToString(hash[:])      // Convert the hash to a string
}
