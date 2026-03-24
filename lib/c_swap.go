package lib

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shirou/gopsutil/v4/mem"
	chkr "github.com/tweithoener/checker"
)

// SwapArgs defines the arguments for a Swap usage check.
type SwapArgs struct {
	WarnPercent float64
	FailPercent float64
}

type swapMaker struct{}

var swapMkr = swapMaker{}

func (swapMaker) Maker() string {
	return "Swap"
}

func (swapMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	args := SwapArgs{}
	if err := json.Unmarshal(j, &args); err != nil {
		return args, fmt.Errorf("can't unmarshal Swap arguments: %v", err)
	}
	return args, nil
}

func (swapMaker) FromConfig(c chkr.CheckConfig) (chkr.Check, error) {
	args, ok := c.Args.(SwapArgs)
	if !ok {
		return nil, fmt.Errorf("configured arguments are not Swap arguments")
	}
	return Swap(args.WarnPercent, args.FailPercent), nil
}

// Swap returns a check that verifies the system's swap memory usage percentage.
func Swap(warnPercent, failPercent float64) chkr.Check {
	return func(ctx context.Context, h chkr.History) (chkr.State, string) {
		v, err := mem.SwapMemoryWithContext(ctx)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Failed to get swap stats: %v", err)
		}

		msg := fmt.Sprintf("Swap usage is at %.2f%%", v.UsedPercent)
		if failPercent > 0 && v.UsedPercent >= failPercent {
			return chkr.Fail, msg
		}
		if warnPercent > 0 && v.UsedPercent >= warnPercent {
			return chkr.Warn, msg
		}

		return chkr.OK, msg
	}
}
