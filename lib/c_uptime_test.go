package lib

import (
	"context"
	"errors"
	"testing"
	"time"

	chkr "github.com/tweithoener/checker"
)

func TestUptimeCheck(t *testing.T) {
	orig := getHostUptime
	defer func() { getHostUptime = orig }()

	tests := []struct {
		name       string
		minMinutes uint64
		mockUptime uint64
		mockErr    error
		expected   chkr.State
	}{
		{"OK", 10, 1000, nil, chkr.OK}, // 1000s > 10m (600s)
		{"Warn", 10, 300, nil, chkr.Warn}, // 300s < 10m (600s)
		{"Error", 10, 0, errors.New("err"), chkr.Fail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getHostUptime = func(ctx context.Context) (uint64, error) {
				if tt.mockErr != nil {
					return 0, tt.mockErr
				}
				return tt.mockUptime, nil
			}

			chk := Uptime(time.Duration(tt.minMinutes) * time.Minute)
			s, _ := chk(context.Background(), chkr.CheckState{})
			if s != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, s)
			}
		})
	}
}
