package main

import "testing"

func TestWindowsBundleThemeContainsMinimalInstallerPages(t *testing.T) {
	root := "..\\.."
	themeFile := root + "\\installer\\windows\\theme\\ActiveStackTheme.xml"

	assertFileContains(t, themeFile, "<Page Name=\"Install\">")
	assertFileContains(t, themeFile, "<Page Name=\"Progress\">")
	assertFileContains(t, themeFile, "<Page Name=\"Success\">")
	assertFileContains(t, themeFile, "<Page Name=\"Modify\">")
	assertFileContains(t, themeFile, "Set up your AI coding workspace")
	assertFileContains(t, themeFile, "Install Active Stack")
	assertFileContains(t, themeFile, "Preparing your setup")
	assertFileContains(t, themeFile, "Active Stack is ready")
}
