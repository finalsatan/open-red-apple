//go:build !windows

package gateway

import "os/exec"

func hideProcessWindow(cmd *exec.Cmd) {}
