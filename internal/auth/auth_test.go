package auth

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCheckPasswordHash(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name          string
		password      string
		hash          string
		wantErr       bool
		matchPassword bool
	}{
		{
			name:          "Correct password",
			password:      password1,
			hash:          hash1,
			wantErr:       false,
			matchPassword: true,
		},
		{
			name:          "Incorrect password",
			password:      "wrongPassword",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Password doesn't match different hash",
			password:      password1,
			hash:          hash2,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Empty password",
			password:      "",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Invalid hash",
			password:      password1,
			hash:          "invalidhash",
			wantErr:       true,
			matchPassword: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && match != tt.matchPassword {
				t.Errorf("CheckPasswordHash() expects %v, got %v", tt.matchPassword, match)
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	tokenSecret := rand.Text()
	userId1 := uuid.New()
	userId2 := uuid.New()
	token1, _ := MakeJWT(userId1, tokenSecret, time.Hour*24)
	token2, _ := MakeJWT(userId2, tokenSecret, time.Hour*24)

	tests := []struct {
		name        string
		userID      uuid.UUID
		token       string
		wantErr     bool
		matchUserID bool
	}{
		{
			name:        "Valid token",
			userID:      userId1,
			token:       token1,
			wantErr:     false,
			matchUserID: true,
		},
		{
			name:        "Invalid token",
			userID:      userId1,
			token:       "ThisIsNotARealToken",
			wantErr:     true,
			matchUserID: false,
		},
		{
			name:        "Invalid token for user",
			userID:      userId1,
			token:       token2,
			wantErr:     false,
			matchUserID: false,
		},
		{
			name:        "Empty token",
			userID:      userId1,
			token:       "",
			wantErr:     true,
			matchUserID: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ValidateJWT(test.token, tokenSecret)
			if (err != nil) != test.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, test.wantErr)
			}

			if !test.wantErr && test.matchUserID && result != test.userID {
				t.Errorf("ValidateJWT() expects %v, got %v", test.userID, result)
			}
		})
	}
}
