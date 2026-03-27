package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	chkr "github.com/tweithoener/checker"
)

// CmdArgs defines the arguments for a command execution check.
type CmdArgs struct {
	Command string
	Args    []string
}

type cmdMaker struct{}

var cmdMkr = cmdMaker{}

func (cmdMaker) Maker() string {
	return "Cmd"
}

func (cmdMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	args := CmdArgs{}
	if err := json.Unmarshal(j, &args); err != nil {
		return args, fmt.Errorf("can't unmarshal Cmd arguments: %v", err)
	}
	return args, nil
}

func (cmdMaker) FromConfig(c chkr.CheckConfig) (chkr.Check, error) {
	args, ok := c.Args.(CmdArgs)
	if !ok {
		return nil, fmt.Errorf("configured arguments are not Cmd arguments")
	}
	return Cmd(defaultAnalyzer, args.Command, args.Args...), nil
}

func defaultAnalyzer(exitCode int, output string) (chkr.State, string) {
	if exitCode == 0 {
		return chkr.OK, ""
	}
	return chkr.Fail, fmt.Sprintf("exit code %d (%s)", exitCode, output)
}

var execCommandContext = exec.CommandContext

// Cmd returns a check that executes the given command and analyzes the results using the given analyze function.
// If analyze is nil a default analyzer function is used. This default function only considers the exit code of the
// called command (non zero exit code results in a failed check).
// Note: The default analyzer is also used when configuring this check using a json config file with checker.ReadConfig.
func Cmd(analyze func(exitCode int, output string) (chkr.State, string), name string, args ...string) chkr.Check {
	return func(ctx context.Context, cs chkr.CheckState) (s chkr.State, message string) {
		if analyze == nil {
			analyze = defaultAnalyzer
		}
		cmd := execCommandContext(ctx, name, args...)
		outbs, err := cmd.CombinedOutput()
		out := string(outbs)
		out = strings.TrimSpace(out)
		if err != nil {
			if exerr, ok := err.(*exec.ExitError); ok {
				ec := exerr.ProcessState.ExitCode()
				return analyze(ec, out)
			}
			return chkr.Fail, fmt.Sprintf("Command execution failed: %v", err)
		}
		return analyze(0, string(out))
	}
}
