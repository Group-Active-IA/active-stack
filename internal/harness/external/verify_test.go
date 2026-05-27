package external

import (
	"context"
	"errors"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

func TestVerify_NPM_BinaryFound(t *testing.T) {
	defer withFakeLookPath(func(name string) (string, error) {
		return "/usr/local/bin/" + name, nil
	})()

	h := harnessWithMethod("npm", "@fission-ai/openspec", "")
	err := Verify(context.Background(), h, Result{BinaryPath: "/usr/local/bin/openspec"})
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestVerify_NPM_BinaryMissing(t *testing.T) {
	defer withFakeLookPath(func(name string) (string, error) {
		return "", errors.New("not found")
	})()

	h := harnessWithMethod("npm", "@fission-ai/openspec", "")
	err := Verify(context.Background(), h, Result{})
	if err == nil {
		t.Error("expected error when binary not in PATH")
	}
}

func TestVerify_Homebrew_BinaryFound(t *testing.T) {
	defer withFakeLookPath(func(name string) (string, error) {
		return "/usr/local/bin/" + name, nil
	})()

	h := harnessWithMethod("homebrew", "engram", "")
	err := Verify(context.Background(), h, Result{BinaryPath: "/usr/local/bin/engram"})
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestVerify_Homebrew_BinaryMissing(t *testing.T) {
	defer withFakeLookPath(func(name string) (string, error) {
		return "", errors.New("not found")
	})()

	h := harnessWithMethod("homebrew", "engram", "")
	err := Verify(context.Background(), h, Result{})
	if err == nil {
		t.Error("expected error when binary not in PATH")
	}
}

func TestVerify_MCP_HTTPS_SkipsCheck(t *testing.T) {
	h := harnessWithMethod("mcp", "", "https://mcp.context7.com")
	err := Verify(context.Background(), h, Result{ConfigFiles: []string{"/home/user/.config/mcp.json"}})
	if err != nil {
		t.Errorf("HTTPS mcp should skip local verify and return nil, got: %v", err)
	}
}

func TestVerify_NilExternal(t *testing.T) {
	h := model.Harness{ID: "bad"}
	err := Verify(context.Background(), h, Result{})
	if err == nil {
		t.Error("expected error for nil External")
	}
}

func TestVerify_UnknownMethod(t *testing.T) {
	h := harnessWithMethod("ftp", "thing", "")
	err := Verify(context.Background(), h, Result{})
	if err == nil {
		t.Error("expected error for unknown method")
	}
}
