package lib

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	chkr "github.com/tweithoener/checker"
)

func TestLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{} // Remove time for deterministic tests
		}
		return a
	}}))

	not := Logging(logger.With("test_prefix", "TEST"))
	cs := chkr.CheckState{Name: "mycheck", State: chkr.OK, Message: "Everything is OK"}
	not(context.Background(), cs)

	output := buf.String()
	expectedParts := []string{
		"level=INFO",
		"msg=\"check state changed\"",
		"test_prefix=TEST",
		"check_name=mycheck",
		"state=OK",
		"message=\"Everything is OK\"",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Expected part '%s' not found in output: %s", part, output)
		}
	}
}
