package git

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func GitCommand(out io.Writer, cmds []string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("git.exe", cmds...)
	case "linux", "darwin":
		cmd = exec.Command("git", cmds...)
	default:
		return fmt.Errorf("unsupported platform")
	}

	cmd.Stdin = os.Stdin
	if out != nil {
		cmd.Stdout = out
		cmd.Stderr = out
	}

	return cmd.Run()
}

func CreateBranch(name string) error {
	err := RepoCheck()
	if err != nil {
		return fmt.Errorf("the current directory is not a git repository")
	}
	return GitCommand(os.Stdout, []string{"checkout", "-b", name})
}

func RepoCheck() error {
	return GitCommand(ioutil.Discard, []string{"rev-parse", "--show-toplevel"})
}

func CurrentBranch() (string, error) {
	var buf bytes.Buffer
	err := GitCommand(&buf, []string{"symbolic-ref", "--short", "HEAD"})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}
