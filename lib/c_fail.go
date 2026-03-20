package lib

import (
	"context"
	"encoding/json"
	"fmt"

	chkr "github.com/tweithoener/checker"
)

// FailArgs defines the arguments for a Fail check wrapper.
type FailArgs struct {
	chkr.WithRecursion
	Check chkr.CheckConfig
}

type failMaker struct {
	chkr.WithRecursion
}

var failMkr = failMaker{}

func (failMaker) Maker() string {
	return "Fail"
}

func (failMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	args := FailArgs{}
	if err := json.Unmarshal(j, &args); err != nil {
		return args, fmt.Errorf("can't unmarshal Fail arguments: %v", err)
	}
	return args, nil
}

func (failMaker) FromConfig(c chkr.CheckConfig) (chkr.Check, error) {
	args, ok := c.Args.(FailArgs)
	if !ok {
		return nil, fmt.Errorf("configured arguments are not Fail arguments")
	}
	inner, err := failMkr.CheckRecursion(args.Check)
	if err != nil {
		return nil, fmt.Errorf("can't make inner check for Fail check: %v", err)
	}
	return Fail(inner), nil
}

// Fail returns a check that wraps another check, inverting its success/fail result.
func Fail(chk chkr.Check) chkr.Check {
	return func(ctx context.Context, h chkr.History) (chkr.State, string) {
		s, msg := chk(ctx, h)
		if s != chkr.Fail {
			return chkr.Fail, fmt.Sprintf("Check was supposed to fail but did not: %s %s", s, msg)
		}
		return chkr.OK, fmt.Sprintf("Check failed as expected: %s %s", s, msg)
	}
}
