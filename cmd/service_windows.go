//go:build windows

package cmd

import (
	"os"
	"syscall"
)

// sysProcAttr returns nil on Windows (Setpgid not supported)
func sysProcAttr() *syscall.SysProcAttr {
	return nil
}

// signalTerm kills the process on Windows (no SIGTERM equivalent)
func signalTerm(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}
