package lib

import (
	"bytes"
	"context"
	"log"
	"os"
	"strings"
	"testing"

	chkr "github.com/tweithoener/checker"
)

func TestLogging(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	not := Logging("TEST-")
	cs := chkr.CheckState{Name: "mycheck", State: chkr.OK, Message: "Everything is OK"}
	not(context.Background(), "mycheck", cs)

	output := buf.String()
	if !strings.Contains(output, "TEST-OK: mycheck: Everything is OK") {
		t.Errorf("Logging output mismatch, got: %s", output)
	}
}
