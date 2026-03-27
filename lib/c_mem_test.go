package lib

import (
	"context"
	"errors"
	"testing"

	"github.com/shirou/gopsutil/v4/mem"
	chkr "github.com/tweithoener/checker"
)

func TestMemCheck(t *testing.T) {
	orig := getVirtualMemory
	defer func() { getVirtualMemory = orig }()

	tests := []struct {
		name        string
		warn        float64
		fail        float64
		mockPercent float64
		mockErr     error
		expected    chkr.State
	}{
		{"OK", 80, 90, 40, nil, chkr.OK},
		{"Warn", 80, 90, 85, nil, chkr.Warn},
		{"Fail", 80, 90, 95, nil, chkr.Fail},
		{"Error", 80, 90, 0, errors.New("err"), chkr.Fail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getVirtualMemory = func(ctx context.Context) (*mem.VirtualMemoryStat, error) {
				if tt.mockErr != nil {
					return nil, tt.mockErr
				}
				return &mem.VirtualMemoryStat{UsedPercent: tt.mockPercent}, nil
			}

			chk := Mem(tt.warn, tt.fail)
			s, _ := chk(context.Background(), chkr.CheckState{})
			if s != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, s)
			}
		})
	}
}
