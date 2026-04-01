package checker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// peer returns a check that returns the state of a remote checker instance.
func (chkr *Checker) peerCheck(address string) Check {
	if chkr.httpClient == nil {
		chkr.httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	return func(ctx context.Context, cs CheckState) (s State, message string) {
		var body []byte
		var err error
		snap := chkr.snapshot().PeerStates
		body, err = json.Marshal(snap)
		if err != nil {
			slog.Error("can't marshal json PeerStates", "error", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", "http://"+address, bytes.NewReader(body))
		if err != nil {
			return Fail, fmt.Sprintf("Failed to create request: %v", err)
		}
		req.Header.Set("Accept", "application/json")

		resp, err := chkr.httpClient.Do(req)
		if err != nil {
			return Fail, fmt.Sprintf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			io.Copy(io.Discard, resp.Body)
			return Fail, fmt.Sprintf("Unexpected status code: %d", resp.StatusCode)
		}

		var st CheckerState
		if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
			return Fail, fmt.Sprintf("Failed to decode response: %v", err)
		}

		st.Address = address
		chkr.integratePeerState(st.PeerState)
		for _, ps := range st.PeerStates {
			chkr.integratePeerState(ps)
		}

		return st.State, fmt.Sprintf("Remote checker %s: %s", st.Name, st.Summary)
	}
}
