package lib

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	chkr "github.com/tweithoener/checker"
)

func TestDebug(t *testing.T) {
	// Intercept stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	not := Debug("DEBUG-")
	since := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	cs := chkr.CheckState{
		Name:    "testcheck",
		State:   chkr.OK,
		Message: "all good",
		Streak:  1,
		Since:   since,
	}
	not(context.Background(), cs)

	// Close writer and restore stdout
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// The timestamp in the output should be replaced by "2001-01-01 01:01:01"
	expected := "DEBUG-OK: testcheck: all good (1x since 2001-01-01 01:01:01)\n"
	if output != expected {
		t.Errorf("Debug output mismatch\ngot:  %q\nwant: %q", output, expected)
	}
}
