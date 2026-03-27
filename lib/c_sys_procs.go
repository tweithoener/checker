package lib

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shirou/gopsutil/v4/process"
	chkr "github.com/tweithoener/checker"
)

// SysProcsArgs defines the arguments for a Total Process Count check.
type SysProcsArgs struct {
	WarnCount int
	FailCount int
}

type sysProcsMaker struct{}

var sysProcsMkr = sysProcsMaker{}

func (sysProcsMaker) Maker() string {
	return "SysProcs"
}

func (sysProcsMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	args := SysProcsArgs{}
	if err := json.Unmarshal(j, &args); err != nil {
		return args, fmt.Errorf("can't unmarshal SysProcs arguments: %v", err)
	}
	return args, nil
}

func (sysProcsMaker) FromConfig(c chkr.CheckConfig) (chkr.Check, error) {
	args, ok := c.Args.(SysProcsArgs)
	if !ok {
		return nil, fmt.Errorf("configured arguments are not SysProcs arguments")
	}
	return SysProcs(args.WarnCount, args.FailCount), nil
}

// SysProcs returns a check that verifies the total number of running processes on the system.
func SysProcs(warnCount, failCount int) chkr.Check {
	return func(ctx context.Context, cs chkr.CheckState) (chkr.State, string) {
		pids, err := process.PidsWithContext(ctx)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Failed to list pids: %v", err)
		}

		count := len(pids)
		msg := fmt.Sprintf("System has %d running processes", count)
		if failCount > 0 && count >= failCount {
			return chkr.Fail, msg
		}
		if warnCount > 0 && count >= warnCount {
			return chkr.Warn, msg
		}

		return chkr.OK, msg
	}
}
