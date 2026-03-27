package checker

import (
	"testing"
	"time"
)

func TestCheckState_DeepCopy(t *testing.T) {
	cs := CheckState{
		Name:    "test",
		State:   OK,
		Message: "msg",
		Streak:  1,
		Last:    map[State]time.Time{OK: time.Now()},
		Since:   time.Now(),
	}

	ncs := cs.deepCopy()

	if ncs.Name != cs.Name || ncs.State != cs.State || ncs.Message != cs.Message || ncs.Streak != cs.Streak || !ncs.Since.Equal(cs.Since) {
		t.Errorf("deepCopy failed, field mismatch: %+v vs %+v", cs, ncs)
	}

	if len(ncs.Last) != len(cs.Last) || !ncs.Last[OK].Equal(cs.Last[OK]) {
		t.Errorf("deepCopy failed, Last map mismatch: %+v vs %+v", cs.Last, ncs.Last)
	}

	// Ensure it's a deep copy of the map
	ncs.Last[Warn] = time.Now()
	if _, ok := cs.Last[Warn]; ok {
		t.Error("deepCopy failed, Last map is not deeply copied")
	}
}

func TestPeerState_DeepCopy(t *testing.T) {
	ps := PeerState{
		Name:    "peer1",
		Address: "addr1",
		Summary: "summ1",
		State:   OK,
		Checks: map[string]CheckState{
			"check1": {Name: "check1", State: OK},
		},
	}

	nps := ps.deepCopy()

	if nps.Name != ps.Name || nps.Address != ps.Address || nps.Summary != ps.Summary || nps.State != ps.State {
		t.Errorf("deepCopy failed, field mismatch: %+v vs %+v", ps, nps)
	}

	if len(nps.Checks) != len(ps.Checks) || nps.Checks["check1"].Name != ps.Checks["check1"].Name {
		t.Errorf("deepCopy failed, Checks map mismatch: %+v vs %+v", ps.Checks, nps.Checks)
	}

	// Ensure it's a deep copy of the map
	nps.Checks["check2"] = CheckState{Name: "check2"}
	if _, ok := ps.Checks["check2"]; ok {
		t.Error("deepCopy failed, Checks map is not deeply copied")
	}
}

func TestCheckState_String(t *testing.T) {
	since := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	cs := CheckState{
		Name:    "test-check",
		State:   Warn,
		Message: "something is wrong",
		Streak:  5,
		Since:   since,
	}

	expected := "Warning: test-check: something is wrong (5x since " + since.Local().Format("2006-01-02 15:04:05") + ")"
	if cs.String() != expected {
		t.Errorf("String() mismatch:\ngot:  %s\nwant: %s", cs.String(), expected)
	}
}
