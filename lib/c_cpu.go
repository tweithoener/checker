package lib

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shirou/gopsutil/v4/cpu"
	chkr "github.com/tweithoener/checker"
)

// CpuArgs defines the arguments for a CPU usage check.
type CpuArgs struct {
	WarnPercent float64
	FailPercent float64
}

type cpuMaker struct{}

var cpuMkr = cpuMaker{}

func (cpuMaker) Maker() string {
	return "Cpu"
}

func (cpuMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	args := CpuArgs{}
	if err := json.Unmarshal(j, &args); err != nil {
		return args, fmt.Errorf("can't unmarshal Cpu arguments: %v", err)
	}
	return args, nil
}

func (cpuMaker) FromConfig(c chkr.CheckConfig) (chkr.Check, error) {
	args, ok := c.Args.(CpuArgs)
	if !ok {
		return nil, fmt.Errorf("configured arguments are not Cpu arguments")
	}
	return Cpu(args.WarnPercent, args.FailPercent), nil
}

// Cpu returns a check that verifies the system's total CPU usage percentage.
func Cpu(warnPercent, failPercent float64) chkr.Check {
	return func(ctx context.Context, h chkr.History) (chkr.State, string) {
		// Interval 0 gets the CPU usage since the last call.
		// Since checks are executed periodically, this works well without blocking.
		percentages, err := cpu.PercentWithContext(ctx, 0, false)
		if err != nil || len(percentages) == 0 {
			return chkr.Fail, fmt.Sprintf("Failed to get CPU stats: %v", err)
		}

		usedPercent := percentages[0]
		msg := fmt.Sprintf("CPU usage is at %.2f%%", usedPercent)
		if failPercent > 0 && usedPercent >= failPercent {
			return chkr.Fail, msg
		}
		if warnPercent > 0 && usedPercent >= warnPercent {
			return chkr.Warn, msg
		}

		return chkr.OK, msg
	}
}
