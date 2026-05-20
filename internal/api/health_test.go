package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleHealth(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	HandleHealth(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, res.StatusCode)
	}

	expectedContentType := "text/plain; charset=utf-8"
	if contentType := res.Header.Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected Content-Type '%s', got '%s'", expectedContentType, contentType)
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expectedBody := "OK"
	if string(bodyBytes) != expectedBody {
		t.Errorf("Expected response body '%s', got '%s'", expectedBody, string(bodyBytes))
	}
}
