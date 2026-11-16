package capture

import (
	"os"

	"golang.org/x/sys/windows"
)

func lockFile(file *os.File) error {
	var position windows.Overlapped
	return windows.LockFileEx(windows.Handle(file.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0, &position)
}
