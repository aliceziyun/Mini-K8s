package iptable

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

type Error struct {
	exec.ExitError
	cmd        exec.Cmd
	msg        string
	exitStatus *int //for overriding
}

func (e *Error) ExitStatus() int {
	if e.exitStatus != nil {
		return *e.exitStatus
	}
	return e.Sys().(syscall.WaitStatus).ExitStatus()
}

func (e *Error) Error() string {
	return fmt.Sprintf("running %v: exit status %v: %v", e.cmd.Args, e.ExitStatus(), e.msg)
}

func (e *Error) IsNotExist() bool {
	if e.ExitStatus() != 1 {
		return false
	}
	msgNoRuleExist := "Bad rule (does a matching rule exist in that chain?).\n"
	msgNoChainExist := "No chain/target/match by that name.\n"
	return strings.Contains(e.msg, msgNoRuleExist) || strings.Contains(e.msg, msgNoChainExist)
}
