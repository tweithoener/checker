package lib

import (
	"context"
	"errors"
	"testing"
	"time"

	chkr "github.com/tweithoener/checker"
)

func TestCpuCheck(t *testing.T) {
	orig := getCpuPercent
	defer func() { getCpuPercent = orig }()

	tests := []struct {
		name        string
		warn        float64
		fail        float64
		mockPercent float64
		mockErr     error
		mockEmpty   bool
		expected    chkr.State
	}{
		{"OK", 80, 90, 40, nil, false, chkr.OK},
		{"Warn", 80, 90, 85, nil, false, chkr.Warn},
		{"Fail", 80, 90, 95, nil, false, chkr.Fail},
		{"Error", 80, 90, 0, errors.New("err"), false, chkr.Fail},
		{"Empty", 80, 90, 0, nil, true, chkr.Fail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getCpuPercent = func(ctx context.Context, interval time.Duration, percpu bool) ([]float64, error) {
				if tt.mockErr != nil {
					return nil, tt.mockErr
				}
				if tt.mockEmpty {
					return []float64{}, nil
				}
				return []float64{tt.mockPercent}, nil
			}

			chk := Cpu(tt.warn, tt.fail)
			s, _ := chk(context.Background(), chkr.CheckState{})
			if s != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, s)
			}
		})
	}
}
