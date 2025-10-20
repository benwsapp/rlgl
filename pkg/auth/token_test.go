package auth_test

import (
	"strings"
	"testing"

	"github.com/benwsapp/rlgl/pkg/auth"
)

func TestGenerateToken(t *testing.T) {
	t.Parallel()

	token, err := auth.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	if !strings.HasPrefix(token, "rlgl_") {
		t.Errorf("expected token to start with 'rlgl_', got %s", token)
	}

	expectedMinLength := len("rlgl_") + 43
	if len(token) < expectedMinLength {
		t.Errorf("expected token length >= %d, got %d", expectedMinLength, len(token))
	}
}

func TestGenerateTokenUnique(t *testing.T) {
	t.Parallel()

	token1, err := auth.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	token2, err := auth.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token1 == token2 {
		t.Error("expected tokens to be unique")
	}
}

func TestValidateTokenFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		token string
		valid bool
	}{
		{
			name:  "valid token",
			token: "rlgl_abc123def456",
			valid: true,
		},
		{
			name:  "empty token",
			token: "",
			valid: false,
		},
		{
			name:  "only prefix",
			token: "rlgl_",
			valid: false,
		},
		{
			name:  "missing prefix",
			token: "abc123def456",
			valid: false,
		},
		{
			name:  "wrong prefix",
			token: "test_abc123def456",
			valid: false,
		},
		{
			name:  "partial prefix",
			token: "rlg_abc123def456",
			valid: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := auth.ValidateTokenFormat(testCase.token)
			if result != testCase.valid {
				t.Errorf("ValidateTokenFormat(%q) = %v, want %v", testCase.token, result, testCase.valid)
			}
		})
	}
}

func TestGeneratedTokenIsValid(t *testing.T) {
	t.Parallel()

	token, err := auth.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if !auth.ValidateTokenFormat(token) {
		t.Errorf("generated token %q failed validation", token)
	}
}
