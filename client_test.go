package checker

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChecker_PeerCheck(t *testing.T) {
	// Remote checker setup
	remoteChecker := New()
	remoteChecker.SetName("remote-checker")
	_ = remoteChecker.AddCheck("remote-check", func(ctx context.Context, cs CheckState) (State, string) {
		return OK, "Everything OK remotely"
	})

	// Perform a check so it has state
	remoteChecker.runCheck(remoteChecker.checks[0])

	ts := httptest.NewServer(remoteChecker)
	defer ts.Close()

	// Local checker setup
	localChecker := New()
	localChecker.SetName("local-checker")

	// The address for the peerCheck is the test server's address
	// httptest address includes http://, but peerCheck adds it too if not present?
	// Actually, client.go says: "http://"+address
	// So we need to strip http:// from ts.URL
	address := ts.URL[7:]

	chk := localChecker.peerCheck(address)

	// cs is dummy
	s, msg := chk(context.Background(), CheckState{})

	if s != OK {
		t.Errorf("PeerCheck returned state: %s, want %s (message: %s)", s, OK, msg)
	}

	expectedPrefix := "Remote checker remote-checker: OK (1 OK, 0 Warning, 0 Failed). Newest OK: remote-check: Everything OK remotely"
	if !strings.Contains(msg, expectedPrefix) {
		t.Errorf("Unexpected message:\ngot:  %s\nwant prefix: %s", msg, expectedPrefix)
	}

	// Verify that remote peer state was integrated into localChecker
	localChecker.mu.RLock()
	ps, ok := localChecker.peerStates[address]
	localChecker.mu.RUnlock()

	if !ok {
		t.Fatalf("Peer %s not integrated into local checker", address)
	}
	if ps.Name != "remote-checker" {
		t.Errorf("Integrated peer name mismatch: got %s, want %s", ps.Name, "remote-checker")
	}
}
