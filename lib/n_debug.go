package lib

import (
	"context"
	"encoding/json"
	"fmt"

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

// Debug returns a notifier that prints the check state directly to stdout.
// It is primarily intended for debugging and development.
func Debug(prefix string) chkr.Notifier {
	return func(_ context.Context, name string, cs chkr.CheckState) {
		fmt.Printf("%s%s\n", prefix, cs)
	}
}
