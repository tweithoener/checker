package lib

import (
	"context"
	"errors"
	"testing"

	"github.com/shirou/gopsutil/v4/disk"
	chkr "github.com/tweithoener/checker"
)

func TestDiskCheck(t *testing.T) {
	orig := getDiskUsage
	defer func() { getDiskUsage = orig }()

	tests := []struct {
		name        string
		path        string
		warn        float64
		fail        float64
		mockPercent float64
		mockErr     error
		expected    chkr.State
	}{
		{"OK", "/", 80, 90, 40, nil, chkr.OK},
		{"Warn", "/", 80, 90, 85, nil, chkr.Warn},
		{"Fail", "/", 80, 90, 95, nil, chkr.Fail},
		{"Error", "/", 80, 90, 0, errors.New("err"), chkr.Fail},
		{"EmptyPath", "", 80, 90, 40, nil, chkr.OK}, // Should default to "/"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getDiskUsage = func(ctx context.Context, path string) (*disk.UsageStat, error) {
				if tt.mockErr != nil {
					return nil, tt.mockErr
				}
				expectedPath := tt.path
				if expectedPath == "" {
					expectedPath = "/"
				}
				if path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, path)
				}
				return &disk.UsageStat{UsedPercent: tt.mockPercent}, nil
			}

			chk := Disk(tt.path, tt.warn, tt.fail)
			s, _ := chk(context.Background(), chkr.CheckState{})
			if s != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, s)
			}
		})
	}
}
