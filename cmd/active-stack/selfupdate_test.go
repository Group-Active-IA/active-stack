package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/selfupdate"
)

// withRelease serves a fake "latest release" tag and points the package at it.
func withRelease(t *testing.T, tag, version string) *int {
	t.Helper()
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		fmt.Fprintf(w, `{"tag_name":%q,"assets":[]}`, tag)
	}))
	oldAPI, oldVer := selfupdate.APIBase, buildVersion
	selfupdate.APIBase, buildVersion = srv.URL, version
	t.Cleanup(func() { srv.Close(); selfupdate.APIBase, buildVersion = oldAPI, oldVer })
	return &hits
}

func isolateHome(t *testing.T) {
	t.Helper()
	d := t.TempDir()
	for _, k := range []string{"USERPROFILE", "HOME", "LOCALAPPDATA", "APPDATA"} {
		t.Setenv(k, d)
	}
	t.Setenv("CI", "")
	t.Setenv("ACTIVE_STACK_NO_UPDATE_CHECK", "")
}

func TestStartupCheckWarnsOnInteractiveCommands(t *testing.T) {
	isolateHome(t)
	withRelease(t, "v0.9.0", "v0.8.1")
	var out bytes.Buffer
	maybeNotifyUpdate([]string{"install"}, &out)
	if !strings.Contains(out.String(), "v0.9.0") || !strings.Contains(out.String(), "self-update") {
		t.Fatalf("no warning: %q", out.String())
	}
}

func TestStartupCheckNeverRunsForHooksOrProxy(t *testing.T) {
	isolateHome(t)
	hits := withRelease(t, "v0.9.0", "v0.8.1")
	for _, cmd := range []string{"memory", "version", "self-update", "evidence-state"} {
		var out bytes.Buffer
		maybeNotifyUpdate([]string{cmd, "hook"}, &out)
		if out.Len() != 0 {
			t.Errorf("%s printed %q", cmd, out.String())
		}
	}
	if *hits != 0 {
		t.Fatalf("the network was consulted %d times for non-interactive commands", *hits)
	}
}

func TestStartupCheckCanBeTurnedOff(t *testing.T) {
	for _, env := range []string{"ACTIVE_STACK_NO_UPDATE_CHECK", "CI"} {
		isolateHome(t)
		t.Setenv(env, "1")
		hits := withRelease(t, "v0.9.0", "v0.8.1")
		var out bytes.Buffer
		maybeNotifyUpdate(nil, &out)
		if out.Len() != 0 || *hits != 0 {
			t.Errorf("%s=1 still checked: out=%q hits=%d", env, out.String(), *hits)
		}
	}
}

func TestSelfUpdateAskedAndDeclinedChangesNothing(t *testing.T) {
	isolateHome(t)
	withRelease(t, "v0.9.0", "v0.8.1")
	var out, errb bytes.Buffer
	code := runSelfUpdate(nil, strings.NewReader("n\n"), &out, &errb)
	if code != 0 || !strings.Contains(out.String(), "v0.9.0") || errb.Len() != 0 {
		t.Fatalf("code=%d out=%q err=%q", code, out.String(), errb.String())
	}
}

func TestSelfUpdateUpToDate(t *testing.T) {
	isolateHome(t)
	withRelease(t, "v0.8.1", "v0.8.1")
	var out bytes.Buffer
	if code := runSelfUpdate(nil, strings.NewReader(""), &out, &bytes.Buffer{}); code != 0 || out.Len() == 0 {
		t.Fatalf("code=%d out=%q", code, out.String())
	}
}

func TestSelfUpdateRejectsUnknownFlag(t *testing.T) {
	isolateHome(t)
	if code := runSelfUpdate([]string{"--force"}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}); code != 2 {
		t.Fatalf("code=%d", code)
	}
}
