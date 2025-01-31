package cmd

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// this code is based on this example:
// https://github.com/george-e-shaw-iv/integration-tests-example/blob/master/cmd/listd/tests/main_test.go
// referenced in this article:
// https://www.ardanlabs.com/blog/2019/10/integration-testing-in-go-set-up-and-writing-tests.html

func TestMain(m *testing.M) {
	config = filepath.Join(".", "testsupport", "config.yaml")
	env = "test"
	os.Exit(m.Run())
}

func CaptureCmdOutput(f func(*cobra.Command, []string), cmd *cobra.Command, args []string) string {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f(cmd, args)

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = rescueStdout
	return string(out)
}

func CaptureCmdOutputE(f func(*cobra.Command, []string) error, cmd *cobra.Command, args []string) (string, error) {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := f(cmd, args)

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = rescueStdout
	return string(out), err
}

func CaptureCmdStdoutStderrE(f func(*cobra.Command, []string) error, cmd *cobra.Command, args []string) (string, string, error) {
	rescueStdout := os.Stdout
	rescueStderr := os.Stderr

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		log.Fatal(err)
	}

	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		log.Fatal(err)
	}

	os.Stdout = stdoutW
	os.Stderr = stderrW

	err = f(cmd, args)

	stdoutW.Close()
	stderrW.Close()

	stdout, _ := io.ReadAll(stdoutR)
	stderr, _ := io.ReadAll(stderrR)
	os.Stdout = rescueStdout
	os.Stderr = rescueStderr

	return string(stdout), string(stderr), err
}

func setCmdFlag(f *cobra.Command, flag string, val string) {
	f.Flags().Set(flag, val)
}

// copied from Cobra test code:
// https://github.com/spf13/cobra/blob/40d34bca1bffe2f5e84b18d7fd94d5b3c02275a6/command_test.go#L49
func checkStringContains(t *testing.T, got, expected string) {
	if !strings.Contains(got, expected) {
		t.Errorf("Expected to contain: \n %v\nGot:\n %v\n", expected, got)
	}
}

func getSubcommandsNames(c *cobra.Command) []string {
	var names []string
	for _, cmd := range c.Commands() {
		names = append(names, cmd.Name())
	}
	return names
}

func checkStringSlices(t *testing.T, got, want []string) {
	if len(got) != len(want) {
		t.Errorf("Expected %d elements, got %d", len(want), len(got))
		t.Errorf("Expected %v", want)
		t.Errorf("Got      %v", got)
		return
	}

	for i, v := range want {
		if got[i] != v {
			t.Errorf("Expected %s, got %s", v, got[i])
		}
	}
}
