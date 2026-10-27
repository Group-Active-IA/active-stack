package main

import (
	"fmt"
	"time"

	"golang.org/x/sys/windows"
)

func waitForParent(pid int) error {
	handle, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
	if err == windows.ERROR_INVALID_PARAMETER {
		return nil // The parent exited before the helper opened its handle.
	}
	if err != nil {
		return fmt.Errorf("open parent process %d: %w", pid, err)
	}
	defer windows.CloseHandle(handle)
	result, err := windows.WaitForSingleObject(handle, uint32((60 * time.Second).Milliseconds()))
	if err != nil {
		return fmt.Errorf("wait for parent process %d: %w", pid, err)
	}
	if result != windows.WAIT_OBJECT_0 {
		return fmt.Errorf("parent process %d did not exit within 60 seconds", pid)
	}
	return nil
}
