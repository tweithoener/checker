package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	chkr "github.com/tweithoener/checker"
)

// LessArgs defines the arguments for a Less notifier wrapper.
type LessArgs struct {
	Notifier chkr.NotifierConfig
}

type lessMaker struct {
	chkr.WithRecursion
}

var lessMkr = lessMaker{}

func (lessMaker) Maker() string {
	return "Less"
}

func (lessMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	args := LessArgs{}
	if err := json.Unmarshal(j, &args); err != nil {
		return args, fmt.Errorf("can't unmarshal LessArgs arguments: %v", err)
	}
	return args, nil
}

func (lessMaker) FromConfig(c chkr.NotifierConfig) (chkr.Notifier, error) {
	args, ok := c.Args.(LessArgs)
	if !ok {
		return nil, fmt.Errorf("configured arguments are not LessArgs arguments")
	}
	inner, err := lessMkr.NotifierRecursion(args.Notifier)
	if err != nil {
		return nil, err
	}
	return Less(inner), nil
}

// Less returns a notifier that wraps another notifier, rate-limiting the notifications.
func Less(n chkr.Notifier) chkr.Notifier {
	var last time.Time
	return func(ctx context.Context, name string, cs chkr.CheckState) {
		if time.Since(last) > 1*time.Hour || cs.Streak <= 3 {
			last = time.Now()
			n(ctx, name, cs)
		}
	}
}
