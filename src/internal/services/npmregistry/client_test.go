package npmregistry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetLatestVersion(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		serverHandler http.HandlerFunc
		expectError   bool
		expectedVer   string
	}{
		{
			name:        "successful version fetch",
			packageName: "@claudex/cli",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				// Verify headers
				if r.Header.Get("User-Agent") != userAgent {
					t.Errorf("expected User-Agent %s, got %s", userAgent, r.Header.Get("User-Agent"))
				}
				if r.Header.Get("Accept") != "application/json" {
					t.Errorf("expected Accept application/json, got %s", r.Header.Get("Accept"))
				}

				// Return mock npm registry response
				response := PackageInfo{
					DistTags: DistTags{
						Latest: "1.2.3",
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			},
			expectError: false,
			expectedVer: "1.2.3",
		},
		{
			name:        "non-200 status code",
			packageName: "nonexistent-package",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			expectError: true,
		},
		{
			name:        "invalid JSON response",
			packageName: "malformed-package",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte("invalid json"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(tt.serverHandler)
			defer server.Close()

			// Create client with custom transport that redirects to test server
			client := &Client{
				httpClient: &http.Client{Timeout: timeout},
			}
			client.httpClient.Transport = &redirectTransport{
				target: server.URL,
			}

			version, err := client.GetLatestVersion(tt.packageName)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if version != tt.expectedVer {
					t.Errorf("expected version %s, got %s", tt.expectedVer, version)
				}
			}
		})
	}
}

func TestHTTPTimeout(t *testing.T) {
	// Create a server that delays response beyond timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second) // Longer than 3s timeout
		json.NewEncoder(w).Encode(PackageInfo{
			DistTags: DistTags{Latest: "1.0.0"},
		})
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{Timeout: timeout},
	}
	client.httpClient.Transport = &redirectTransport{
		target: server.URL,
	}

	_, err := client.GetLatestVersion("test-package")
	if err == nil {
		t.Error("expected timeout error, got nil")
	}

	// Check if it's a timeout error
	if err != nil && !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline") {
		// Note: Go's timeout errors may vary, they might contain "timeout" or "deadline exceeded"
		// We should still verify it's a network error of some kind
		if !strings.Contains(err.Error(), "context deadline exceeded") {
			t.Logf("got error (acceptable): %v", err)
		}
	}
}

// redirectTransport is a custom RoundTripper that redirects all requests to a test server
type redirectTransport struct {
	target string
}

func (rt *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Replace the URL with our test server URL, keeping the path
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(rt.target, "http://")
	return http.DefaultTransport.RoundTrip(req)
}
