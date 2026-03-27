package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	chkr "github.com/tweithoener/checker"
)

// LoggingArgs defines the arguments for a Logging notifier.
type LoggingArgs struct {
	Prefix string
}

type loggingMaker struct{}

var logginMkr = loggingMaker{}

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
	return Logging(args.Prefix), nil
}

// Logging returns a notifier that outputs check results to the standard log.
func Logging(prefix string) chkr.Notifier {
	return func(_ context.Context, name string, cs chkr.CheckState) {
		log.Printf("%s%s", prefix, cs)
	}
}
