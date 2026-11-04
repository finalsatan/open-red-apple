//go:build windows

package gateway

import (
	"os/exec"
	"syscall"
)

func hideProcessWindow(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
