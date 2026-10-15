package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gookit/greq/ext/httpfile"
	"golang.org/x/oauth2/clientcredentials"
)

// applyAuth sets the Authorization header based on X-Auth-Type.
//
// Supported values (read from the request headers):
//   - "bearer"  -> uses X-Auth-Token header value
//   - "basic"   -> uses X-Auth-User + X-Auth-Pass header values
//   - "oauth2"  -> uses OAUTH_CLIENT_ID / OAUTH_CLIENT_SECRET / OAUTH_TOKEN_URL env vars
//   - (omitted) -> sends request as-is (whatever Authorization header is already set)
//
// Returns (client, nil) for oauth2, (nil, nil) otherwise.
// The caller uses the returned client if non-nil.
func applyAuth(req *http.Request) (*http.Client, error) {
	authType := strings.ToLower(req.Header.Get("X-Auth-Type"))

	// Always strip our custom helper headers before sending
	req.Header.Del("X-Auth-Type")

	switch authType {
	case "bearer":
		token := req.Header.Get("X-Auth-Token")
		req.Header.Del("X-Auth-Token")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

	case "basic":
		user := req.Header.Get("X-Auth-User")
		pass := req.Header.Get("X-Auth-Pass")
		req.Header.Del("X-Auth-User")
		req.Header.Del("X-Auth-Pass")
		if user != "" {
			creds := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
			req.Header.Set("Authorization", "Basic "+creds)
		}

	case "oauth2":
		cfg := clientcredentials.Config{
			ClientID:     os.Getenv("OAUTH_CLIENT_ID"),
			ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
			TokenURL:     os.Getenv("OAUTH_TOKEN_URL"),
			Scopes:       []string{"api.read"},
		}
		return cfg.Client(context.Background()), nil
	}

	return nil, nil
}

// executeRequest builds, authenticates, and sends a single HTTP request.
func executeRequest(req *httpfile.HTTPRequest) (*http.Response, error) {
	vars := map[string]string{} // add substitutions here if needed

	httpReq, err := http.NewRequest(
		req.Method,
		req.URLString(vars),
		strings.NewReader(req.BodyString(vars)),
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for key, val := range req.HeadersMap(vars) {
		httpReq.Header.Set(key, val)
	}

	client, err := applyAuth(httpReq)
	if err != nil {
		return nil, fmt.Errorf("apply auth: %w", err)
	}
	if client == nil {
		client = http.DefaultClient
	}

	return client.Do(httpReq)
}

func main() {
	httpFile := os.Getenv("HTTP_FILE")
	if httpFile == "" {
		httpFile = "api_requests.http"
	}

	file, err := httpfile.ParseHTTPFile(httpFile)
	if err != nil {
		fmt.Printf("❌ Error parsing %s: %v\n", httpFile, err)
		os.Exit(1)
	}

	for _, req := range file.Requests {
		name := req.Name
		if name == "" {
			name = req.Method + " " + req.URLString(nil)
		}
		fmt.Printf("\n=== %s ===\n", name)

		resp, err := executeRequest(req)
		if err != nil {
			fmt.Printf("❌ Request failed: %v\n", err)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		fmt.Printf("Status: %s\n", resp.Status)
		fmt.Printf("Body:   %s\n", string(body))
	}
}
