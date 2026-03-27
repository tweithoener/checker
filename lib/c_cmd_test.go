package lib

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	chkr "github.com/tweithoener/checker"
)

// TestHelperProcess is the target of the mock execCommandContext.
// It is not meant to be run by the test framework as a normal test.
// Instead, when the test framework runs it, it checks the GO_WANT_HELPER_PROCESS
// environment variable. If it's "1", it means this process was spawned by our mock
// to act as the "fake" command. It then prints the requested output and exits
// with the requested status code, simulating a real process execution perfectly.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return // Normal test run, skip the helper logic.
	}

	// We print the mock output specified by the parent test.
	fmt.Print(os.Getenv("MOCK_OUTPUT"))

	// We exit with the mock exit code specified by the parent test.
	code, _ := strconv.Atoi(os.Getenv("MOCK_EXIT_CODE"))
	os.Exit(code)
}

// mockExecCommand returns a mock execCommandContext function that intercepts
// the command execution and redirects it to the TestHelperProcess function above.
// It also verifies that the expected command and arguments were passed.
func mockExecCommand(expectedCommand string, expectedArgs []string, mockOutput string, mockExitCode int) func(context.Context, string, ...string) *exec.Cmd {
	return func(ctx context.Context, command string, args ...string) *exec.Cmd {
		// Verify that Cmd() actually passed the correct command name
		if command != expectedCommand {
			panic(fmt.Sprintf("Expected command '%s', got '%s'", expectedCommand, command))
		}

		// Verify that Cmd() actually passed all expected arguments
		if len(args) != len(expectedArgs) {
			panic(fmt.Sprintf("Expected %d arguments, got %d", len(expectedArgs), len(args)))
		}
		for i, arg := range args {
			if arg != expectedArgs[i] {
				panic(fmt.Sprintf("Expected argument at index %d to be '%s', got '%s'", i, expectedArgs[i], arg))
			}
		}

		// os.Args[0] is the path to the currently running compiled test binary.
		// We execute ourselves, but we specifically ask the test framework to ONLY
		// run the "TestHelperProcess" function.
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.CommandContext(ctx, os.Args[0], cs...)

		// Set the environment variables that our TestHelperProcess will read.
		cmd.Env = []string{
			"GO_WANT_HELPER_PROCESS=1",
			fmt.Sprintf("MOCK_OUTPUT=%s", mockOutput),
			fmt.Sprintf("MOCK_EXIT_CODE=%d", mockExitCode),
		}
		return cmd
	}
}

func TestCmd_Success(t *testing.T) {
	// Temporarily replace the package-level execCommandContext with our mock.
	origExec := execCommandContext
	defer func() { execCommandContext = origExec }()

	// We want to simulate a successful command (exit code 0) and verify args are passed.
	execCommandContext = mockExecCommand("some-command", []string{"--flag", "value"}, "success output", 0)

	// We use the default analyzer implicitly by passing nil.
	chk := Cmd(nil, "some-command", "--flag", "value")
	state, msg := chk(context.Background(), chkr.CheckState{})

	if state != chkr.OK {
		t.Errorf("Expected OK, got %v", state)
	}
	if msg != "" {
		t.Errorf("Expected empty message, got '%s'", msg)
	}
}

func TestCmd_Fail(t *testing.T) {
	origExec := execCommandContext
	defer func() { execCommandContext = origExec }()

	// We want to simulate a failed command (exit code 1) with an error message.
	execCommandContext = mockExecCommand("some-command", []string{}, "error happened", 1)

	chk := Cmd(nil, "some-command")
	state, msg := chk(context.Background(), chkr.CheckState{})

	if state != chkr.Fail {
		t.Errorf("Expected Fail, got %v", state)
	}
	if msg != "exit code 1 (error happened)" {
		t.Errorf("Expected 'exit code 1 (error happened)', got '%s'", msg)
	}
}

func TestCmd_CustomAnalyzer(t *testing.T) {
	origExec := execCommandContext
	defer func() { execCommandContext = origExec }()

	// We simulate an exit code 0, but with a specific output.
	execCommandContext = mockExecCommand("some-command", []string{}, "WARNING: disk low", 0)

	// A custom analyzer that parses the output to determine the state.
	customAnalyzer := func(exitCode int, output string) (chkr.State, string) {
		if exitCode == 0 && strings.Contains(output, "WARNING") {
			return chkr.Warn, "Found a warning"
		}
		return chkr.OK, ""
	}

	chk := Cmd(customAnalyzer, "some-command")
	state, msg := chk(context.Background(), chkr.CheckState{})

	if state != chkr.Warn {
		t.Errorf("Expected Warn, got %v", state)
	}
	if msg != "Found a warning" {
		t.Errorf("Expected custom message, got '%s'", msg)
	}
}
