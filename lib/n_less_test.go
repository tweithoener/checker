package lib

import (
	"context"
	"testing"

	chkr "github.com/tweithoener/checker"
)

func TestLess(t *testing.T) {
	called := 0
	inner := func(ctx context.Context, cs chkr.CheckState) {
		called++
	}
	not := Less(inner)
	cs := chkr.CheckState{Name: "mycheck", State: chkr.Fail}

	// Streak <= 3 should always notify
	for i := 1; i <= 3; i++ {
		cs.Streak = i
		not(context.Background(), cs)
	}
	if called != 3 {
		t.Errorf("Less should notify 3 times for streaks 1, 2, 3 but got %d", called)
	}

	// Streak 4 should not notify (since it was just notified)
	cs.Streak = 4
	not(context.Background(), cs)
	if called != 3 {
		t.Errorf("Less should NOT notify for streak 4 but got %d", called)
	}
}
