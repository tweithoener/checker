package lib

import (
	"context"
	"errors"
	"testing"

	chkr "github.com/tweithoener/checker"
)

func TestProcExistsCheck(t *testing.T) {
	orig := getProcessNames
	defer func() { getProcessNames = orig }()

	tests := []struct {
		name      string
		procName  string
		mockNames []string
		mockErr   error
		expected  chkr.State
	}{
		{"FoundSingle", "nginx", []string{"bash", "nginx", "sshd"}, nil, chkr.OK},
		{"FoundMultiple", "nginx", []string{"nginx", "nginx"}, nil, chkr.OK},
		{"NotFound", "apache2", []string{"bash", "nginx", "sshd"}, nil, chkr.Fail},
		{"EmptyList", "nginx", []string{}, nil, chkr.Fail},
		{"Error", "nginx", nil, errors.New("err"), chkr.Fail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getProcessNames = func(ctx context.Context) ([]string, error) {
				if tt.mockErr != nil {
					return nil, tt.mockErr
				}
				return tt.mockNames, nil
			}

			chk := ProcExists(tt.procName)
			s, _ := chk(context.Background(), chkr.CheckState{})
			if s != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, s)
			}
		})
	}
}
