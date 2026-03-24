package lib

import (
	"encoding/json"
	"fmt"

	chkr "github.com/tweithoener/checker"
)

type SshArgs struct {
	Host    string
	User    string
	Command string
}

type sshMaker struct{}

var sshMkr = sshMaker{}

func (sshMaker) Maker() string {
	return "Ssh"
}

func (sshMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	args := SshArgs{}
	if err := json.Unmarshal(j, &args); err != nil {
		return args, fmt.Errorf("can't unmarshal Ssh arguments: %v", err)
	}
	return args, nil
}

func (sshMaker) FromConfig(c chkr.CheckConfig) (chkr.Check, error) {
	args, ok := c.Args.(SshArgs)
	if !ok {
		return nil, fmt.Errorf("configured arguments are not Ssh arguments")
	}
	return Ssh(defaultAnalyzer, args.Host, args.User, args.Command), nil
}

// Ssh returns a check that executes the given command on the remote host using the
// system's ssh command and then analyzing the command output using the analyze function.
// If analyze is nil a default analyzer function is used. This default function only
// considers the exit code of the called command (non zero exit code results in a failed
// check).
//
// Note: The default analyzer is also used when configuring this check using a json config
// file with checker.ReadConfig.
//
// Note: Make sure public key authentication and host key authorization is configured correctly
// for the user running checker for the remote machine.

func Ssh(analyze func(exitCode int, output string) (chkr.State, string), host, user, command string) chkr.Check {
	uh := host
	if user != "" {
		uh = user + "@" + host
	}
	return Cmd(analyze, "ssh", "-oBatchmode=yes", uh, command)
}
