package lib

import (
	"context"
	"errors"
	"testing"
	"time"

	chkr "github.com/tweithoener/checker"
)

func TestPingCheck(t *testing.T) {
	orig := runPing
	defer func() { runPing = orig }()

	tests := []struct {
		name         string
		warn         int
		fail         int
		mockDuration time.Duration
		mockErr      error
		expected     chkr.State
	}{
		{"OK", 100, 200, 50 * time.Millisecond, nil, chkr.OK},
		{"Warn", 100, 200, 150 * time.Millisecond, nil, chkr.Warn},
		{"Fail", 100, 200, 250 * time.Millisecond, nil, chkr.Fail},
		{"Error", 100, 200, 0, errors.New("network unreachable"), chkr.Fail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runPing = func(ctx context.Context, address string, timeout time.Duration) (time.Duration, error) {
				if tt.mockErr != nil {
					return 0, tt.mockErr
				}
				return tt.mockDuration, nil
			}

			chk := Ping("1.2.3.4", tt.warn, tt.fail)
			s, _ := chk(context.Background(), chkr.CheckState{})
			if s != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, s)
			}
		})
	}
}
