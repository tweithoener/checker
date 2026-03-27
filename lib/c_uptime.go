package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/host"
	chkr "github.com/tweithoener/checker"
)

// UptimeArgs defines the arguments for an Uptime check.
type UptimeArgs struct {
	MinMinutes uint64 // Warn if uptime is less than this value (e.g. recent reboot)
}

type uptimeMaker struct{}

var uptimeMkr = uptimeMaker{}

func (uptimeMaker) Maker() string {
	return "Uptime"
}

func (uptimeMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	args := UptimeArgs{}
	if err := json.Unmarshal(j, &args); err != nil {
		return args, fmt.Errorf("can't unmarshal Uptime arguments: %v", err)
	}
	return args, nil
}

func (uptimeMaker) FromConfig(c chkr.CheckConfig) (chkr.Check, error) {
	args, ok := c.Args.(UptimeArgs)
	if !ok {
		return nil, fmt.Errorf("configured arguments are not Uptime arguments")
	}
	return Uptime(time.Duration(args.MinMinutes) * time.Minute), nil
}

// Uptime returns a check that verifies the system's uptime and warns if it's lower than a minimum (e.g. after a reboot).
func Uptime(minUptime time.Duration) chkr.Check {
	return func(ctx context.Context, cs chkr.CheckState) (chkr.State, string) {
		uptimeSecs, err := host.UptimeWithContext(ctx)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Failed to get uptime: %v", err)
		}
		uptime := time.Duration(uptimeSecs) * time.Second

		msg := fmt.Sprintf("System uptime is %v", uptime)
		if uptime < minUptime {
			return chkr.Warn, fmt.Sprintf("%s (less than expected %v)", msg, minUptime)
		}

		return chkr.OK, msg
	}
}
