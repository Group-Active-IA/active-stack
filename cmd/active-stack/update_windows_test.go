package main

import (
	"path/filepath"
	"syscall"
	"testing"
)

func TestSamePathRecognizesWindowsShortDirectoryAlias(t *testing.T) {
	long := t.TempDir()
	input, err := syscall.UTF16PtrFromString(long)
	if err != nil {
		t.Fatal(err)
	}
	buffer := make([]uint16, 32768)
	n, err := syscall.GetShortPathName(input, &buffer[0], uint32(len(buffer)))
	if err != nil {
		t.Fatal(err)
	}
	short := syscall.UTF16ToString(buffer[:n])
	if filepath.Clean(short) == filepath.Clean(long) {
		t.Skip("filesystem does not expose a short alias")
	}
	if !samePath(long, short) {
		t.Fatalf("same directory rejected: long=%q short=%q", long, short)
	}
}
