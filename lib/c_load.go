package lib

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shirou/gopsutil/v4/load"
	chkr "github.com/tweithoener/checker"
)

// LoadArgs defines the arguments for a Load Average check.
type LoadArgs struct {
	WarnLoad5 float64
	FailLoad5 float64
}

type loadMaker struct{}

var loadMkr = loadMaker{}

func (loadMaker) Maker() string {
	return "Load"
}

func (loadMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	args := LoadArgs{}
	if err := json.Unmarshal(j, &args); err != nil {
		return args, fmt.Errorf("can't unmarshal Load arguments: %v", err)
	}
	return args, nil
}

func (loadMaker) FromConfig(c chkr.CheckConfig) (chkr.Check, error) {
	args, ok := c.Args.(LoadArgs)
	if !ok {
		return nil, fmt.Errorf("configured arguments are not Load arguments")
	}
	return Load(args.WarnLoad5, args.FailLoad5), nil
}

// Load returns a check that verifies the 5-minute system load average.
func Load(warnLoad5, failLoad5 float64) chkr.Check {
	return func(ctx context.Context, cs chkr.CheckState) (chkr.State, string) {
		avg, err := load.AvgWithContext(ctx)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Failed to get load average: %v", err)
		}

		msg := fmt.Sprintf("Load average (1m, 5m, 15m): %.2f, %.2f, %.2f", avg.Load1, avg.Load5, avg.Load15)
		if failLoad5 > 0 && avg.Load5 >= failLoad5 {
			return chkr.Fail, msg
		}
		if warnLoad5 > 0 && avg.Load5 >= warnLoad5 {
			return chkr.Warn, msg
		}

		return chkr.OK, msg
	}
}
