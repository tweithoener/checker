package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	chkr "github.com/tweithoener/checker"
)

// DebugArgs defines the arguments for a Debug notifier.
type DebugArgs struct {
	Prefix string
}

type debugMaker struct{}

var debugMkr = debugMaker{}

func (debugMaker) Maker() string {
	return "Debug"
}

func (debugMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	args := DebugArgs{}
	if err := json.Unmarshal(j, &args); err != nil {
		return args, fmt.Errorf("can't unmarshal DebugArgs arguments: %v", err)
	}
	return args, nil
}

func (debugMaker) FromConfig(c chkr.NotifierConfig) (chkr.Notifier, error) {
	args, ok := c.Args.(DebugArgs)
	if !ok {
		return nil, fmt.Errorf("configured arguments are not DebugArgs arguments")
	}
	return Debug(args.Prefix), nil
}

var tsRegex = regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`)

// Debug returns a notifier that prints the check state directly to stdout.
// It is primarily intended for debugging and development.
// It replaces any timestamp in the output with '2001-01-01 01:01:01' to make it deterministic for testing.
func Debug(prefix string) chkr.Notifier {
	return func(_ context.Context, cs chkr.CheckState) {
		out := fmt.Sprintf("%s%s", prefix, cs)
		out = tsRegex.ReplaceAllString(out, "2001-01-01 01:01:01")
		fmt.Println(out)
	}
}
