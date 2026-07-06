package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWindowsBundlePayloadWiringFilesExist(t *testing.T) {
	root := filepath.Join("..", "..")

	projectFile := filepath.Join(root, "installer", "windows", "ActiveStack.Bundle.wixproj")
	bundleFile := filepath.Join(root, "installer", "windows", "Bundle.wxs")
	buildScript := filepath.Join(root, "installer", "windows", "build.ps1")
	gitIgnore := filepath.Join(root, "installer", "windows", ".gitignore")

	assertFileContains(t, projectFile, "InstallerVersion")
	assertFileContains(t, projectFile, "PayloadDir=$(MSBuildProjectDirectory)\\payload")
	assertFileContains(t, bundleFile, "SourceFile=\"$(var.PayloadDir)\\active-stack.exe\"")
	assertFileContains(t, bundleFile, "SourceFile=\"$(var.PayloadDir)\\ba\\active-stack.exe\"")
	assertFileContains(t, bundleFile, "SourceFile=\"$(var.PayloadDir)\\ba\\mbanative.dll\"")
	assertFileContains(t, bundleFile, "ActiveStackEventsFile")
	assertFileContains(t, bundleFile, "InstallArguments=\"windows install --agent [ActiveStackAssistantId] --mode [ActiveStackInstallModeId] --windows-events-file")
	assertFileContains(t, bundleFile, "UninstallArguments=\"windows uninstall --json-stream\"")
	assertFileContains(t, buildScript, "go build")
	assertFileContains(t, buildScript, "payload\\active-stack.exe")
	assertFileContains(t, buildScript, "dotnet build")
	assertFileContains(t, buildScript, "ActiveStack.Bundle.wixproj")
	assertFileContains(t, gitIgnore, "payload/")
	assertFileContains(t, gitIgnore, "dist/")
}

func TestWindowsBundleBuildScriptSupportsStagingOnly(t *testing.T) {
	root := filepath.Join("..", "..")
	buildScript := filepath.Join(root, "installer", "windows", "build.ps1")

	body, err := os.ReadFile(buildScript)
	if err != nil {
		t.Fatalf("read %s: %v", buildScript, err)
	}
	if !strings.Contains(string(body), "SkipBundleBuild") {
		t.Fatalf("%s does not contain %q", buildScript, "SkipBundleBuild")
	}
	if !strings.Contains(string(body), "dotnet --list-sdks") {
		t.Fatalf("%s does not contain %q", buildScript, "dotnet --list-sdks")
	}
	if !strings.Contains(string(body), "requires a .NET SDK") {
		t.Fatalf("%s does not contain %q", buildScript, "requires a .NET SDK")
	}
	if !strings.Contains(string(body), "$LASTEXITCODE") {
		t.Fatalf("%s does not contain %q", buildScript, "$LASTEXITCODE")
	}
	if !strings.Contains(string(body), "WiX bundle build failed") {
		t.Fatalf("%s does not contain %q", buildScript, "WiX bundle build failed")
	}
}
