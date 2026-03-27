package lib

import (
	"context"
	"testing"

	chkr "github.com/tweithoener/checker"
)

func TestFail(t *testing.T) {
	mockOK := func(ctx context.Context, cs chkr.CheckState) (chkr.State, string) {
		return chkr.OK, "is ok"
	}
	mockFail := func(ctx context.Context, cs chkr.CheckState) (chkr.State, string) {
		return chkr.Fail, "is fail"
	}

	chk := Fail(mockOK)
	s, _ := chk(context.Background(), chkr.CheckState{})
	if s != chkr.Fail {
		t.Error("Fail(OK) should return Fail")
	}

	chk = Fail(mockFail)
	s, _ = chk(context.Background(), chkr.CheckState{})
	if s != chkr.OK {
		t.Error("Fail(Fail) should return OK")
	}
}
