package lib

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shirou/gopsutil/v4/process"
	chkr "github.com/tweithoener/checker"
)

// ProcExistsArgs defines the arguments for a Process Exists check.
type ProcExistsArgs struct {
	Name string
}

type procExistsMaker struct{}

var procExistsMkr = procExistsMaker{}

func (procExistsMaker) Maker() string {
	return "ProcExists"
}

func (procExistsMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	args := ProcExistsArgs{}
	if err := json.Unmarshal(j, &args); err != nil {
		return args, fmt.Errorf("can't unmarshal ProcExists arguments: %v", err)
	}
	return args, nil
}

func (procExistsMaker) FromConfig(c chkr.CheckConfig) (chkr.Check, error) {
	args, ok := c.Args.(ProcExistsArgs)
	if !ok {
		return nil, fmt.Errorf("configured arguments are not ProcExists arguments")
	}
	return ProcExists(args.Name), nil
}

// ProcExists returns a check that verifies if at least one process with the exact given name is running.
func ProcExists(name string) chkr.Check {
	return func(ctx context.Context, cs chkr.CheckState) (chkr.State, string) {
		procs, err := process.ProcessesWithContext(ctx)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Failed to list processes: %v", err)
		}

		count := 0
		for _, p := range procs {
			n, err := p.NameWithContext(ctx)
			if err == nil && n == name {
				count++
			}
		}

		if count > 0 {
			return chkr.OK, fmt.Sprintf("Found %d process(es) matching '%s'", count, name)
		}
		return chkr.Fail, fmt.Sprintf("Process '%s' is not running", name)
	}
}
