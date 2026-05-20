package auth

import (
	"net/http"
	"testing"
)

func TestGetBearerToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		authHeader     string
		expectedToken  string
		expectedStatus bool
	}{
		{
			name:           "Happy Path - Valid Bearer Token",
			authHeader:     "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			expectedToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			expectedStatus: true,
		},
		{
			name:           "Failure - Empty Header",
			authHeader:     "",
			expectedToken:  "",
			expectedStatus: false,
		},
		{
			name:           "Failure - Missing 'Bearer' Prefix (e.g., Basic Auth)",
			authHeader:     "Basic user:password",
			expectedToken:  "",
			expectedStatus: false,
		},
		{
			name:           "Failure - Only 'Bearer' Keyword (No Token)",
			authHeader:     "Bearer",
			expectedToken:  "",
			expectedStatus: false,
		},
		{
			name:           "Failure - Too Many Spaces (Malformed)",
			authHeader:     "Bearer token123 extraData",
			expectedToken:  "",
			expectedStatus: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			headers := http.Header{}
			if tc.authHeader != "" {
				headers.Set("Authorization", tc.authHeader)
			}

			token, ok := GetBearerToken(headers)

			if ok != tc.expectedStatus {
				t.Errorf("Expected status %v, got %v", tc.expectedStatus, ok)
			}

			if token != tc.expectedToken {
				t.Errorf("Expected token '%v', got '%v'", tc.expectedToken, token)
			}
		})
	}
}
