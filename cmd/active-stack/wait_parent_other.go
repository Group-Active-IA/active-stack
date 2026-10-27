//go:build !windows

package main

import "fmt"

func waitForParent(pid int) error {
	return fmt.Errorf("waiting for process %d is only supported on Windows", pid)
}
