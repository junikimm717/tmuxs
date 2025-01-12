package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func runTmux(args []string, silent bool) error {
	tmux, err := exec.LookPath("tmux")
	if err != nil {
		return err
	}
	cmd := exec.Command(tmux, args...)
	if silent {
		cmd.Stdout = nil
		cmd.Stdin = nil
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
	}
	return cmd.Run()
}

func execTmux(args []string) error {
	tmux, err := exec.LookPath("tmux")
	if err != nil {
		return err
	}
	args = append([]string{tmux}, args...)

	return syscall.Exec(tmux, args, os.Environ())
}

func insideTmux() bool {
	return os.Getenv("TMUX") != ""
}

func sessionExists(name string) bool {
	return runTmux([]string{"has-session", "-t=" + name}, true) == nil
}

func OpenSession(path string) error {
	if len(path) == 0 {
		return errors.New("No path selected!")
	}
	name := strings.ReplaceAll(path, ".", "_")
	if !sessionExists(name) {
		err := runTmux([]string{
			"new-session", "-d", "-c", path, "-s", name,
		}, true)
		if err != nil {
			return err
		}
	}
	if insideTmux() {
		return execTmux(
			[]string{"switch-client", "-t=" + name},
		)
	} else {
		return execTmux(
			[]string{"attach", "-t", name},
		)
	}
}
