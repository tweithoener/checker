package checker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// peer returns a check that returns the state of a remote checker instance.
func (chkr *Checker) peerCheck(address string) Check {
	cl := http.Client{Timeout: 20 * time.Second}
	return func(ctx context.Context, cs CheckState) (s State, message string) {
		body := []byte{}
		var err error
		snap := chkr.snapshot().PeerStates
		body, err = json.Marshal(snap)
		if err != nil {
			log.Printf("can't marshal json PeerStates: %v", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", "http://"+address, bytes.NewReader(body))
		if err != nil {
			return Fail, fmt.Sprintf("Failed to create request: %v", err)
		}
		req.Header.Set("Accept", "application/json")

		resp, err := cl.Do(req)
		if err != nil {
			return Fail, fmt.Sprintf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
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
