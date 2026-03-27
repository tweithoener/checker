package checker

import (
	"fmt"
	"maps"
	"time"
)

// PeerState is used to transfer json encoded state information between
// checker instances using the checker server.
type PeerState struct {
	Name    string                `json:"name"`
	Address string                `json:"address"`
	Summary string                `json:"summary"`
	State   State                 `json:"state"`
	Checks  map[string]CheckState `json:"checks"`
}

func (ps PeerState) deepCopy() PeerState {
	nps := ps
	nps.Checks = make(map[string]CheckState, len(ps.Checks))
	for k, v := range ps.Checks {
		nps.Checks[k] = v.deepCopy()
	}
	return nps
}

// CheckerState is used to transfer json encoded state information between
// checker instances using the checker server.
type CheckerState struct {
	PeerState
	PeerStates map[string]PeerState `json:"peerStates"`
}

// CheckState represents the state and historical data of a single check
type CheckState struct {
	Name    string              `json:"name"`
	State   State               `json:"state"`
	Message string              `json:"message"`
	Streak  int                 `json:"streak"`
	Last    map[State]time.Time `json:"last"`
	Since   time.Time           `json:"since"`
}

func (cs CheckState) deepCopy() CheckState {
	ncs := cs
	ncs.Last = maps.Clone(cs.Last)
	return ncs
}

func (cs CheckState) String() string {
	return fmt.Sprintf("%s: %s: %s (%dx since %s)", cs.State, cs.Name, cs.Message, cs.Streak, cs.Since.Local().Format("2006-01-02 15:04:05"))
}
