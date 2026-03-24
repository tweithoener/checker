package lib

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shirou/gopsutil/v4/mem"
	chkr "github.com/tweithoener/checker"
)

// MemArgs defines the arguments for a memory usage check.
type MemArgs struct {
	WarnPercent float64
	FailPercent float64
}

type memMaker struct{}

var memMkr = memMaker{}

func (memMaker) Maker() string {
	return "Mem"
}

func (memMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	args := MemArgs{}
	if err := json.Unmarshal(j, &args); err != nil {
		return args, fmt.Errorf("can't unmarshal Mem arguments: %v", err)
	}
	return args, nil
}

func (memMaker) FromConfig(c chkr.CheckConfig) (chkr.Check, error) {
	args, ok := c.Args.(MemArgs)
	if !ok {
		return nil, fmt.Errorf("configured arguments are not Mem arguments")
	}
	return Mem(args.WarnPercent, args.FailPercent), nil
}

// Mem returns a check that verifies the system's virtual memory usage percentage.
func Mem(warnPercent, failPercent float64) chkr.Check {
	return func(ctx context.Context, h chkr.History) (chkr.State, string) {
		v, err := mem.VirtualMemoryWithContext(ctx)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Failed to get memory stats: %v", err)
		}

		msg := fmt.Sprintf("Memory usage is at %.2f%%", v.UsedPercent)
		if failPercent > 0 && v.UsedPercent >= failPercent {
			return chkr.Fail, msg
		}
		if warnPercent > 0 && v.UsedPercent >= warnPercent {
			return chkr.Warn, msg
		}

		return chkr.OK, msg
	}
}
