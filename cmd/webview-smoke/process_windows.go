//go:build windows

package main

import (
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"golang.org/x/sys/windows"
)

func configureConsole() { _ = windows.SetConsoleOutputCP(65001) }

func killProcessTree(p *os.Process) {
	cmd := exec.Command("taskkill", "/PID", strconv.Itoa(p.Pid), "/T", "/F")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Run(); err != nil {
		_ = p.Kill()
	}
}
