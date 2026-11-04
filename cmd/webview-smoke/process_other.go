//go:build !windows

package main

import "os"

func configureConsole() {}

func killProcessTree(p *os.Process) { _ = p.Kill() }
