package lib

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	chkr "github.com/tweithoener/checker"
)

// We reuse the mock pattern from c_cmd_test.go, but this time we specifically want
// to verify that the command passed to exec.CommandContext is actually "ssh" and
// that the arguments include the correct user/host combination.
func mockSshExecCommand(expectedArgs string) func(context.Context, string, ...string) *exec.Cmd {
	return func(ctx context.Context, command string, args ...string) *exec.Cmd {
		// Verify that the underlying command being executed is actually "ssh"
		if command != "ssh" {
			// If not, we trigger a panic to fail the test immediately
			panic(fmt.Sprintf("Expected command 'ssh', got '%s'", command))
		}

		joinedArgs := strings.Join(args, " ")
		if joinedArgs != expectedArgs {
			panic(fmt.Sprintf("Expected arguments '%s', got '%s'", expectedArgs, joinedArgs))
		}

		// Continue with the HelperProcess redirection as usual, returning a success
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.CommandContext(ctx, os.Args[0], cs...)

		cmd.Env = []string{
			"GO_WANT_HELPER_PROCESS=1",
			"MOCK_OUTPUT=ssh success",
			"MOCK_EXIT_CODE=0",
		}
		return cmd
	}
}

func TestSsh_WithUser(t *testing.T) {
	origExec := execCommandContext
	defer func() { execCommandContext = origExec }()

	// We expect the user "admin" and the host "example.com" to be joined correctly.
	expectedArgs := "-oBatchmode=yes admin@example.com uptime"
	execCommandContext = mockSshExecCommand(expectedArgs)

	chk := Ssh(nil, "example.com", "admin", "uptime")
	
	// If the arguments are wrong, the mockSshExecCommand will panic and fail the test.
	state, _ := chk(context.Background(), chkr.CheckState{})

	if state != chkr.OK {
		t.Errorf("Expected OK, got %v", state)
	}
}

func TestSsh_WithoutUser(t *testing.T) {
	origExec := execCommandContext
	defer func() { execCommandContext = origExec }()

	// We expect only the host "example.com" if no user is provided.
	expectedArgs := "-oBatchmode=yes example.com df -h"
	execCommandContext = mockSshExecCommand(expectedArgs)

	chk := Ssh(nil, "example.com", "", "df -h")
	
	state, _ := chk(context.Background(), chkr.CheckState{})

	if state != chkr.OK {
		t.Errorf("Expected OK, got %v", state)
	}
}
