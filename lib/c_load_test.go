package lib

import (
	"context"
	"errors"
	"testing"

	"github.com/shirou/gopsutil/v4/load"
	chkr "github.com/tweithoener/checker"
)

func TestLoadCheck(t *testing.T) {
	orig := getLoadAvg
	defer func() { getLoadAvg = orig }()

	tests := []struct {
		name      string
		warn      float64
		fail      float64
		mockLoad5 float64
		mockErr   error
		expected  chkr.State
	}{
		{"OK", 2.0, 4.0, 1.0, nil, chkr.OK},
		{"Warn", 2.0, 4.0, 2.5, nil, chkr.Warn},
		{"Fail", 2.0, 4.0, 4.5, nil, chkr.Fail},
		{"Error", 2.0, 4.0, 0, errors.New("err"), chkr.Fail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getLoadAvg = func(ctx context.Context) (*load.AvgStat, error) {
				if tt.mockErr != nil {
					return nil, tt.mockErr
				}
				return &load.AvgStat{Load1: 1.0, Load5: tt.mockLoad5, Load15: 1.0}, nil
			}

			chk := Load(tt.warn, tt.fail)
			s, _ := chk(context.Background(), chkr.CheckState{})
			if s != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, s)
			}
		})
	}
}
