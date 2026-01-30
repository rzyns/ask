//go:build !windows

package cmd

import (
	"os"
	"syscall"
)

// sysProcAttr returns the SysProcAttr for Unix systems
func sysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// signalTerm sends SIGTERM to the process
func signalTerm(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Signal(syscall.SIGTERM)
}
