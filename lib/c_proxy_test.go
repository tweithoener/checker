package lib

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	chkr "github.com/tweithoener/checker"
)

func TestProxyCheck(t *testing.T) {
	orig := proxyDoRequest
	defer func() { proxyDoRequest = orig }()

	tests := []struct {
		name     string
		proxy    string
		expected int
		mockCode int
		mockErr  error
		result   chkr.State
	}{
		{"OK", "http://myproxy:8080", 200, 200, nil, chkr.OK},
		{"UnexpectedStatus", "http://myproxy:8080", 200, 404, nil, chkr.Fail},
		{"ProxyFailed", "http://myproxy:8080", 200, 0, http.ErrHandlerTimeout, chkr.Fail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxyDoRequest = func(cl *http.Client, req *http.Request) (*http.Response, error) {
				// Verify proxy configuration
				if cl.Transport == nil {
					t.Fatal("Transport is nil")
				}
				transport, ok := cl.Transport.(*http.Transport)
				if !ok {
					t.Fatal("Transport is not *http.Transport")
				}
				
				// Get the proxy URL from the transport
				proxyURL, err := transport.Proxy(req)
				if err != nil {
					t.Fatalf("Failed to get proxy URL: %v", err)
				}
				if proxyURL == nil || proxyURL.String() != tt.proxy {
					t.Errorf("expected proxy %s, got %v", tt.proxy, proxyURL)
				}

				if tt.mockErr != nil {
					return nil, tt.mockErr
				}
				return &http.Response{
					StatusCode: tt.mockCode,
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			}

			chk := Proxy("GET", "http://example.com", tt.proxy, tt.expected)
			s, _ := chk(context.Background(), chkr.CheckState{})
			if s != tt.result {
				t.Errorf("expected %v, got %v", tt.result, s)
			}
		})
	}
}
