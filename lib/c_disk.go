package lib

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shirou/gopsutil/v4/disk"
	chkr "github.com/tweithoener/checker"
)

// DiskArgs defines the arguments for a Disk usage check.
type DiskArgs struct {
	Path        string
	WarnPercent float64
	FailPercent float64
}

type diskMaker struct{}

var diskMkr = diskMaker{}

func (diskMaker) Maker() string {
	return "Disk"
}

func (diskMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	args := DiskArgs{}
	if err := json.Unmarshal(j, &args); err != nil {
		return args, fmt.Errorf("can't unmarshal Disk arguments: %v", err)
	}
	return args, nil
}

func (diskMaker) FromConfig(c chkr.CheckConfig) (chkr.Check, error) {
	args, ok := c.Args.(DiskArgs)
	if !ok {
		return nil, fmt.Errorf("configured arguments are not Disk arguments")
	}
	return Disk(args.Path, args.WarnPercent, args.FailPercent), nil
}

// Disk returns a check that verifies the disk usage percentage for a specific path.
func Disk(path string, warnPercent, failPercent float64) chkr.Check {
	if path == "" {
		path = "/"
	}
	return func(ctx context.Context, h chkr.History) (chkr.State, string) {
		u, err := disk.UsageWithContext(ctx, path)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Failed to get disk stats for %s: %v", path, err)
		}

		msg := fmt.Sprintf("Disk usage on %s is at %.2f%%", path, u.UsedPercent)
		if failPercent > 0 && u.UsedPercent >= failPercent {
			return chkr.Fail, msg
		}
		if warnPercent > 0 && u.UsedPercent >= warnPercent {
			return chkr.Warn, msg
		}

		return chkr.OK, msg
	}
}
