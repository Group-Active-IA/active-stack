package verify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Group-Active-IA/active-stack/internal/harness/autosync"
	"github.com/Group-Active-IA/active-stack/internal/harness/hooks"
	"github.com/Group-Active-IA/active-stack/internal/harness/proxyreg"
	"github.com/Group-Active-IA/active-stack/internal/i18n"
	"github.com/Group-Active-IA/active-stack/internal/memory/projectinit"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// engramCloudID is the harness these checks belong to.
const engramCloudID = "engram-cloud"

// Option adjusts how the checks are built. ChecksForHarness and BuildHook
// accept them variadically so existing callers keep working unchanged.
type Option func(*options)

type options struct {
	lang   i18n.Lang
	target model.InstallTarget
}

// WithLang sets the language of the engram-cloud messages (default
// i18n.Default).
func WithLang(lang i18n.Lang) Option { return func(o *options) { o.lang = lang } }

// WithTarget tells the checks where the harness was installed. The hooks, the
// save proxy and AUTOSYNC are user-level (decision D12): a Project install has
// none, so those checks are not built for it. The default is Machine.
func WithTarget(t model.InstallTarget) Option { return func(o *options) { o.target = t } }

func newOptions(opts []Option) options {
	o := options{lang: i18n.Default}
	for _, f := range opts {
		f(&o)
	}
	return o
}

func (o options) tc(key string, args ...any) string { return i18n.Tf(o.lang, key, args...) }

// Every check below only reads: no file is written, no process is started and
// no check prints a value from a configuration file (T-F10-8), only paths,
// agent names and key names.

// cloudChecks returns the checks of the engram-cloud harness for one adapter
// that the harness supports (Claude and Codex). The idempotent marker check
// (T-F10-1) is the generic one in checkConfig.
func cloudChecks(adapter Adapter, homeDir string, o options) []Check {
	if o.target == model.Project {
		return nil
	}
	agent := adapter.Agent()
	name := string(agent)

	// A single, read-only look at the installed hooks, shared by the checks.
	hookStatus := func() (hooks.AgentStatus, error) {
		sts, err := hooks.Status(homeDir, []hooks.Adapter{adapter})
		if err != nil {
			return hooks.AgentStatus{}, err
		}
		for _, s := range sts {
			if s.Agent == agent {
				return s, nil
			}
		}
		return hooks.AgentStatus{}, errors.New("no settings file for this agent")
	}

	checks := []Check{
		{
			ID:          "cloud:hooks:" + name,
			Description: o.tc("verify.cloud.hooks.desc", name),
			Run: func(context.Context) error {
				st, err := hookStatus()
				if err != nil {
					return errors.New(o.tc("verify.cloud.status_error", err))
				}
				if !st.Installed {
					return errors.New(o.tc("verify.cloud.hooks.missing", name, st.File, strings.Join(st.Missing, ", ")))
				}
				return nil
			},
		},
		{
			// T-F10-3 / R-16 D-C: the install proceeds without the binary on PATH, so
			// verify marks it as a warning, never a rollback.
			ID:          "cloud:path:" + name,
			Soft:        true,
			Description: o.tc("verify.cloud.path.desc", name),
			Run: func(context.Context) error {
				st, err := hookStatus()
				if err != nil {
					return errors.New(o.tc("verify.cloud.status_error", err))
				}
				if !st.BinaryOnPath {
					return errors.New(o.tc("verify.cloud.path.missing", name))
				}
				return nil
			},
		},
	}

	if agent == model.AgentCodex {
		// T-F10-4: trust cannot be read from the config, so it is always a
		// reminder, never a failure.
		checks = append(checks, Check{
			ID:          "cloud:trust:" + name,
			Description: o.tc("verify.cloud.trust.desc"),
			Soft:        true,
			Run: func(context.Context) error {
				return errors.New(o.tc("verify.cloud.trust.notice"))
			},
		})
	}

	checks = append(checks,
		Check{
			ID:          "cloud:proxy:" + name,
			Description: o.tc("verify.cloud.proxy.desc", name),
			Run: func(context.Context) error {
				sts, err := proxyreg.Status(homeDir, []proxyreg.Adapter{adapter})
				if err != nil {
					return errors.New(o.tc("verify.cloud.status_error", err))
				}
				for _, s := range sts {
					if s.Agent != agent {
						continue
					}
					if !s.HasEntry {
						return errors.New(o.tc("verify.cloud.proxy.noentry", name, s.File))
					}
					if !s.Registered {
						return errors.New(o.tc("verify.cloud.proxy.missing", name, s.File))
					}
					return nil
				}
				return errors.New(o.tc("verify.cloud.proxy.noentry", name, adapter.MCPConfigPath(homeDir, proxyreg.ServerName)))
			},
		},
		Check{
			// T-F10-7: soft.
			ID:          "cloud:autosync:" + name,
			Description: o.tc("verify.cloud.autosync.desc", name),
			Soft:        true,
			Run: func(context.Context) error {
				sts, err := autosync.Status(homeDir, []autosync.Adapter{adapter})
				if err != nil {
					return errors.New(o.tc("verify.cloud.status_error", err))
				}
				for _, s := range sts {
					if s.Agent == agent && s.Enabled {
						return nil
					}
				}
				return errors.New(o.tc("verify.cloud.autosync.off", name, adapter.MCPConfigPath(homeDir, autosync.ServerName)))
			},
		},
	)
	return checks
}

// cloudGlobalChecks returns the checks that do not depend on one supported
// adapter: the warning for agents outside the scope (T-F10-5) and the project
// files (T-F10-6). They only appear when the project actually uses the mode
// (its marker file exists), so a user-level install never fails over a project.
func cloudGlobalChecks(adapters []Adapter, homeDir string, o options) []Check {
	var checks []Check
	root := homeDir
	marker := filepath.Join(root, filepath.FromSlash(projectinit.MarkerPath))
	config := filepath.Join(root, filepath.FromSlash(projectinit.ConfigPath))
	usesMode := fileExists(marker)

	if usesMode {
		for _, a := range adapters {
			agent := a.Agent()
			if agent != model.AgentOpenCode && agent != model.AgentAntigravity {
				continue
			}
			name := string(agent)
			checks = append(checks, Check{
				ID:          "cloud:unsupported:" + name,
				Description: o.tc("verify.cloud.unsupported.desc", name),
				Soft:        true,
				Run: func(context.Context) error {
					return errors.New(o.tc("verify.cloud.unsupported.warn", name, name))
				},
			})
		}
	}

	if usesMode || fileExists(config) {
		checks = append(checks, Check{
			ID:          "cloud:project-files",
			Description: o.tc("verify.cloud.project.desc"),
			Run: func(context.Context) error {
				if err := requireJSONString(o, config, projectinit.ConfigPath, "project_name", ""); err != nil {
					return err
				}
				return requireJSONString(o, marker, projectinit.MarkerPath, "memory_policy", projectinit.Policy)
			},
		})
	}
	return checks
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

// requireJSONString checks that the file is a JSON object whose key holds a
// non-empty string (equal to want when want is set). It never reports the
// value, only the file and the key.
func requireJSONString(o options, path, rel, key, want string) error {
	invalid := func(reason string) error {
		return errors.New(o.tc("verify.cloud.project.invalid", rel, reason))
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return invalid(fmt.Sprintf("%v", err))
	}
	raw = bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF}) // Notepad adds a BOM
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return invalid("JSON")
	}
	var got string
	if v, ok := obj[key]; !ok || json.Unmarshal(v, &got) != nil || strings.TrimSpace(got) == "" {
		return invalid(fmt.Sprintf("%q", key))
	}
	if want != "" && got != want {
		return invalid(fmt.Sprintf("%q", key))
	}
	return nil
}
