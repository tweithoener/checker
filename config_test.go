package checker

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

type mockMaker struct {
	name string
}

func (m *mockMaker) Maker() string {
	return m.name
}

func (m *mockMaker) FromConfig(c CheckConfig) (Check, error) {
	return func(ctx context.Context, cs CheckState) (State, string) {
		return OK, "mock check"
	}, nil
}

func (m *mockMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	return nil, nil
}

type mockNotifierMaker struct {
	name string
}

func (m *mockNotifierMaker) Maker() string {
	return m.name
}

func (m *mockNotifierMaker) FromConfig(c NotifierConfig) (Notifier, error) {
	return func(ctx context.Context, cs CheckState) {}, nil
}

func (m *mockNotifierMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	return nil, nil
}

func TestReadConfig(t *testing.T) {
	// Register mock makers
	_ = AddCheckMaker(&mockMaker{name: "mock_check"})
	_ = AddNotifierMaker(&mockNotifierMaker{name: "mock_notifier"})

	configJSON := `{
		"Interval": 60,
		"Server": {
			"Enabled": true,
			"Listen": ":8888"
		},
		"Peers": ["1.2.3.4:8080"],
		"Checks": [
			{
				"Maker": "mock_check",
				"Name": "my-check",
				"Args": {}
			}
		],
		"Notifiers": [
			{
				"Maker": "mock_notifier",
				"Args": {}
			}
		]
	}`

	chkr := New()
	err := chkr.ReadConfig(strings.NewReader(configJSON))
	if err != nil {
		t.Fatalf("ReadConfig() failed: %v", err)
	}

	if chkr.interval != 60*time.Second {
		t.Errorf("Interval mismatch: got %v, want %v", chkr.interval, 60*time.Second)
	}

	if !chkr.serverConfig.Enabled || chkr.serverConfig.Listen != ":8888" {
		t.Errorf("Server config mismatch: %+v", chkr.serverConfig)
	}

	if len(chkr.checks) != 2 { // 1 check + 1 peer
		t.Errorf("Checks count mismatch: got %d, want 2", len(chkr.checks))
	}

	foundCheck := false
	foundPeer := false
	for _, c := range chkr.checks {
		if c.Name == "my-check" {
			foundCheck = true
		}
		if c.Name == "Peer 1.2.3.4:8080" {
			foundPeer = true
		}
	}

	if !foundCheck {
		t.Error("Registered check 'my-check' not found")
	}
	if !foundPeer {
		t.Error("Peer check not found")
	}

	if len(chkr.notifiers) != 1 {
		t.Errorf("Notifiers count mismatch: got %d, want 1", len(chkr.notifiers))
	}
}

func TestAddCheckMaker_Duplicate(t *testing.T) {
	maker := &mockMaker{name: "duplicate_check"}
	err := AddCheckMaker(maker)
	if err != nil {
		t.Errorf("First AddCheckMaker failed: %v", err)
	}
	err = AddCheckMaker(maker)
	if err == nil {
		t.Error("Second AddCheckMaker should have failed for duplicate")
	}
}
