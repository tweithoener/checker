package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	chkr "github.com/tweithoener/checker"
)

// PeerArgs defines the arguments for a Peer check.
type PeerArgs struct {
	Address string
}

type peerMaker struct{}

var peerMkr = peerMaker{}

func (peerMaker) Maker() string {
	return "Peer"
}

func (peerMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	args := PeerArgs{}
	if err := json.Unmarshal(j, &args); err != nil {
		return args, fmt.Errorf("can't unmarshal Peer arguments: %v", err)
	}
	return args, nil
}

func (peerMaker) FromConfig(c chkr.CheckConfig) (chkr.Check, error) {
	args, ok := c.Args.(PeerArgs)
	if !ok {
		return nil, fmt.Errorf("configured arguments are not Peer arguments")
	}
	return Peer(args.Address), nil
}

// Peer returns a check that returns the state of a remote checker instance.
func Peer(address string) chkr.Check {
	return func(ctx context.Context, h chkr.History) (s chkr.State, message string) {
		req, err := http.NewRequestWithContext(ctx, "GET", address, nil)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Failed to create request: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return chkr.Fail, fmt.Sprintf("Unexpected status code: %d", resp.StatusCode)
		}

		var st chkr.StateTransfer
		if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
			return chkr.Fail, fmt.Sprintf("Failed to decode response: %v", err)
		}

		if st.State != chkr.OK {
			count := 0
			var newest *chkr.CheckStateTransfer
			for i, c := range st.Checks {
				if c.State == st.State {
					count++
					if newest == nil || c.Since.After(newest.Since) {
						newest = &st.Checks[i]
					}
				}
			}
			if count > 0 && newest != nil {
				sinceStr := newest.Since.Local().Format("2006-01-02 15:04:05")
				return st.State, fmt.Sprintf("%d check(s) in state %s. Latest: '%s' since %s", count, st.State, newest.Name, sinceStr)
			}
			return st.State, fmt.Sprintf("Remote checker is in state: %s", st.State)
		}

		return chkr.OK, "Remote checker is OK"
	}
}
