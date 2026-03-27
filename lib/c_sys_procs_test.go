package lib

import (
	"context"
	"errors"
	"testing"

	chkr "github.com/tweithoener/checker"
)

func TestSysProcsCheck(t *testing.T) {
	orig := getPids
	defer func() { getPids = orig }()

	tests := []struct {
		name      string
		warn      int
		fail      int
		mockPids  []int32
		mockErr   error
		expected  chkr.State
	}{
		{"OK", 100, 200, make([]int32, 50), nil, chkr.OK},
		{"Warn", 100, 200, make([]int32, 150), nil, chkr.Warn},
		{"Fail", 100, 200, make([]int32, 250), nil, chkr.Fail},
		{"Error", 100, 200, nil, errors.New("err"), chkr.Fail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getPids = func(ctx context.Context) ([]int32, error) {
				if tt.mockErr != nil {
					return nil, tt.mockErr
				}
				return tt.mockPids, nil
			}

			chk := SysProcs(tt.warn, tt.fail)
			s, _ := chk(context.Background(), chkr.CheckState{})
			if s != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, s)
			}
		})
	}
}
