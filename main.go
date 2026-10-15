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
// Returns a custom *http.Client for OAuth2, or nil for other cases.
func applyAuth(req *http.Request) (*http.Client, error) {
	authType := strings.ToLower(req.Header.Get("X-Auth-Type"))
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
func executeRequest(client *http.Client, req *httpfile.HTTPRequest) (*http.Response, error) {
	vars := map[string]string{}

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

	authClient, err := applyAuth(httpReq)
	if err != nil {
		return nil, fmt.Errorf("apply auth: %w", err)
	}
	if authClient != nil {
		client = authClient
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

		resp, err := executeRequest(nil, req) // ← pass nil here
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
