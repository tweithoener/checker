package lib

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	chkr "github.com/tweithoener/checker"
)

func TestDebug(t *testing.T) {
	// Intercept stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	not := Debug("DEBUG-")
	cs := chkr.CheckState{Name: "testcheck", State: chkr.OK, Message: "all good"}
	not(context.Background(), "testcheck", cs)

	// Close writer and restore stdout
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "DEBUG-OK: testcheck: all good") {
		t.Errorf("Debug output mismatch, got: %s", output)
	}
}
