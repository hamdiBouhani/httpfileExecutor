package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gookit/greq/ext/httpfile"
)

// newTestRequest is a helper to quickly create an httpfile.HTTPRequest for testing.
func newTestRequest(method, url string, headers map[string]string) *httpfile.HTTPRequest {
	req := &httpfile.HTTPRequest{
		Method:  method,
		URL:     url,
		Headers: make(map[string]string),
	}
	for k, v := range headers {
		req.Headers[k] = v
	}
	return req
}

func TestApplyAuth_Bearer(t *testing.T) {
	// 1. Create a mock server that verifies the Authorization header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := "Bearer my-secret-token"
		if got := r.Header.Get("Authorization"); got != expected {
			t.Errorf("Authorization header = %q; want %q", got, expected)
		}
		// Ensure the custom helper header was removed
		if got := r.Header.Get("X-Auth-Type"); got != "" {
			t.Errorf("X-Auth-Type header should have been removed, got %q", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// 2. Create a request that uses bearer auth
	req := newTestRequest("GET", server.URL, map[string]string{
		"X-Auth-Type":  "bearer",
		"X-Auth-Token": "my-secret-token",
	})

	// 3. Execute and check for errors
	resp, err := executeRequest(server.Client(), req)
	if err != nil {
		t.Fatalf("executeRequest failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d; want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestApplyAuth_Basic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok {
			t.Error("BasicAuth failed")
		}
		if user != "alice" || pass != "secret" {
			t.Errorf("BasicAuth = %q/%q; want alice/secret", user, pass)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req := newTestRequest("GET", server.URL, map[string]string{
		"X-Auth-Type": "basic",
		"X-Auth-User": "alice",
		"X-Auth-Pass": "secret",
	})

	resp, err := executeRequest(server.Client(), req)
	if err != nil {
		t.Fatalf("executeRequest failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d; want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestApplyAuth_NoAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should not have an Authorization header set by us
		if got := r.Header.Get("Authorization"); got != "" {
			t.Errorf("Unexpected Authorization header: %q", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req := newTestRequest("GET", server.URL, map[string]string{})

	resp, err := executeRequest(server.Client(), req)
	if err != nil {
		t.Fatalf("executeRequest failed: %v", err)
	}
	defer resp.Body.Close()
}
