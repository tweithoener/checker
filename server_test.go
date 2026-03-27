package checker

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestChecker_ServeHTTP_JSON(t *testing.T) {
	c := New()
	c.SetName("test-server")
	_ = c.AddCheck("test-check", func(ctx context.Context, cs CheckState) (State, string) {
		return OK, "Everything fine"
	})

	// Run check to update state
	c.runCheck(c.checks[0])
	time.Sleep(50 * time.Millisecond) // Wait for async execution

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "application/json")
	rr := httptest.NewRecorder()

	c.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ServeHTTP() returned status: %v, want %v", rr.Code, http.StatusOK)
	}

	var state CheckerState
	if err := json.NewDecoder(rr.Body).Decode(&state); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if state.Name != "test-server" {
		t.Errorf("State name mismatch: got %s, want %s", state.Name, "test-server")
	}

	if len(state.Checks) != 1 {
		t.Errorf("Checks count mismatch: got %d, want 1", len(state.Checks))
	}
}

func TestChecker_ServeHTTP_HTML(t *testing.T) {
	c := New()
	c.SetName("test-server")

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "text/html")
	rr := httptest.NewRecorder()

	c.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ServeHTTP() returned status: %v, want %v", rr.Code, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("Content-Type mismatch: got %s, want %s", contentType, "text/html")
	}

	if !bytes.Contains(rr.Body.Bytes(), []byte("test-server")) {
		t.Error("HTML response did not contain checker name")
	}
}

func TestChecker_ServeHTTP_POST(t *testing.T) {
	c := New()
	c.SetName("receiving-checker")

	peerData := map[string]PeerState{
		"peer-address": {
			Name:    "remote-peer",
			Address: "peer-address",
			State:   OK,
			Checks: map[string]CheckState{
				"remote-check": {Name: "remote-check", State: OK},
			},
		},
	}
	body, _ := json.Marshal(peerData)

	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Accept", "application/json")
	rr := httptest.NewRecorder()

	c.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ServeHTTP() POST returned status: %v, want %v", rr.Code, http.StatusOK)
	}

	// Verify integration
	c.mu.RLock()
	ps, ok := c.peerStates["peer-address"]
	c.mu.RUnlock()

	if !ok {
		t.Fatal("Peer state was not integrated")
	}
	if ps.Name != "remote-peer" {
		t.Errorf("Integrated peer name mismatch: got %s, want %s", ps.Name, "remote-peer")
	}
}
