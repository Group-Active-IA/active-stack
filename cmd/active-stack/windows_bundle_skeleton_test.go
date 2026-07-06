package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWindowsBundleSkeletonFilesExist(t *testing.T) {
	root := filepath.Join("..", "..")

	projectFile := filepath.Join(root, "installer", "windows", "ActiveStack.Bundle.wixproj")
	bundleFile := filepath.Join(root, "installer", "windows", "Bundle.wxs")
	themeFile := filepath.Join(root, "installer", "windows", "theme", "ActiveStackTheme.xml")
	readmeFile := filepath.Join(root, "installer", "windows", "README.md")

	assertFileContains(t, projectFile, "WixToolset.Sdk")
	assertFileContains(t, projectFile, "WixToolset.BootstrapperApplications.wixext")
	assertFileContains(t, bundleFile, "<Bundle")
	assertFileContains(t, bundleFile, "<BootstrapperApplication")
	assertFileContains(t, bundleFile, "<Payload SourceFile=\"$(var.PayloadDir)\\ba\\active-stack.exe\" />")
	assertFileContains(t, bundleFile, "ActiveStackEventsFile")
	assertFileContains(t, themeFile, "<Theme")
	assertFileContains(t, readmeFile, "active-stack windows")
}

func assertFileContains(t *testing.T, path string, want string) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(body), want) {
		t.Fatalf("%s does not contain %q", path, want)
	}
}
