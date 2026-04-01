package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	chkr "github.com/tweithoener/checker"
)

// LoggingArgs defines the arguments for a Logging notifier configured via JSON.
type LoggingArgs struct {
	// Optional: Static attributes that will be added to every log event.
	Attributes map[string]string `json:"attributes,omitempty"`
}

type loggingMaker struct{}

var loggingMkr = loggingMaker{}

func (loggingMaker) Maker() string {
	return "Logging"
}

func (loggingMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	args := LoggingArgs{}
	if err := json.Unmarshal(j, &args); err != nil {
		return args, fmt.Errorf("can't unmarshal LoggingArgs arguments: %v", err)
	}
	return args, nil
}

func (loggingMaker) FromConfig(c chkr.NotifierConfig) (chkr.Notifier, error) {
	args, ok := c.Args.(LoggingArgs)
	if !ok {
		return nil, fmt.Errorf("configured arguments are not LoggingArgs arguments")
	}

	logger := slog.Default()
	if len(args.Attributes) > 0 {
		var attrs []any
		for k, v := range args.Attributes {
			attrs = append(attrs, slog.String(k, v))
		}
		logger = logger.With(attrs...)
	}

	return Logging(logger), nil
}

// Logging returns a notifier that outputs check results using structured logging.
// If logger is nil, slog.Default() is used.
func Logging(logger *slog.Logger) chkr.Notifier {
	if logger == nil {
		logger = slog.Default()
	}

	return func(_ context.Context, cs chkr.CheckState) {
		level := slog.LevelInfo
		switch cs.State {
		case chkr.Fail:
			level = slog.LevelError
		case chkr.Warn:
			level = slog.LevelWarn
		}

		logger.Log(
			context.Background(),
			level,
			"check state changed",
			"check_name", cs.Name,
			"state", string(cs.State),
			"message", cs.Message,
			"streak", cs.Streak,
			"since", cs.Since.Format(time.RFC3339),
		)
	}
}
