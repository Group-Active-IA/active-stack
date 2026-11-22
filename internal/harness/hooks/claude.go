package hooks

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/Group-Active-IA/active-stack/internal/jsonedit"
)

// Claude Code: ~/.claude/settings.json, key "hooks" ->
// <Event> -> [ { "matcher"?, "hooks": [ { "type", "command", "args", "timeout" } ] } ].
//
// The handler runs in exec form (command + args, "with no shell involved" per
// the Claude Code documentation), so it does not depend on the user's shell on
// Windows. Our handler is identified by command and args, never by position.

type claudeHandler struct {
	Type    string   `json:"type"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Timeout int      `json:"timeout"`
}

type claudeGroup struct {
	Matcher string          `json:"matcher,omitempty"`
	Hooks   []claudeHandler `json:"hooks"`
}

func claudeGroupFor(s Spec) claudeGroup {
	return claudeGroup{
		Matcher: s.Matcher,
		Hooks: []claudeHandler{{
			Type:    "command",
			Command: Command,
			Args:    []string{"memory", "hook", s.Subcommand},
			Timeout: s.TimeoutSeconds,
		}},
	}
}

// isOwnHandler reports whether a handler is the one installed for sub.
func isOwnHandler(raw json.RawMessage, sub string) bool {
	var h struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}
	if json.Unmarshal(raw, &h) != nil {
		return false
	}
	return h.Command == Command && strings.Join(h.Args, " ") == "memory hook "+sub
}

// groupHandlers returns the raw handlers of one group, or nil if it has none
// in the expected shape (a foreign group is never an error).
func groupHandlers(group []byte) []json.RawMessage {
	var g struct {
		Hooks []json.RawMessage `json:"hooks"`
	}
	if json.Unmarshal(group, &g) != nil {
		return nil
	}
	return g.Hooks
}

func groupHasOwn(group []byte, sub string) bool {
	for _, h := range groupHandlers(group) {
		if isOwnHandler(h, sub) {
			return true
		}
	}
	return false
}

// claudeLayout is where the pieces are in the current source.
type claudeLayout struct {
	root      jsonedit.Span
	rootMs    []jsonedit.Member
	hooksIdx  int // index in rootMs, -1 when there is no "hooks"
	hooks     jsonedit.Span
	hooksMs   []jsonedit.Member
	eventIdx  int // index in hooksMs, -1 when the event is absent
	event     jsonedit.Span
	eventElem []jsonedit.Span
}

// layout locates "hooks" and one event. It refuses an unexpected structure:
// the caller leaves the file intact.
func layout(src []byte, event string) (claudeLayout, error) {
	l := claudeLayout{hooksIdx: -1, eventIdx: -1}
	var err error
	if l.root, err = jsonedit.RootSpan(src); err != nil {
		return l, err
	}
	if l.rootMs, err = jsonedit.ParseObject(src, l.root); err != nil {
		return l, err
	}
	i, ok := jsonedit.LastMember(l.rootMs, "hooks")
	if !ok {
		return l, nil
	}
	l.hooksIdx, l.hooks = i, l.rootMs[i].Val
	if src[l.hooks.Start] != '{' {
		return l, jsonedit.ShapeErr(`"hooks" is not an object`)
	}
	if l.hooksMs, err = jsonedit.ParseObject(src, l.hooks); err != nil {
		return l, err
	}
	j, ok := jsonedit.LastMember(l.hooksMs, event)
	if !ok {
		return l, nil
	}
	l.eventIdx, l.event = j, l.hooksMs[j].Val
	if src[l.event.Start] != '[' {
		return l, jsonedit.ShapeErr(`"hooks.%s" is not an array`, event)
	}
	l.eventElem, err = jsonedit.ParseArray(src, l.event)
	return l, err
}

// claudeHasOwn reports whether the event already holds our handler for s.
func claudeHasOwn(src []byte, s Spec) (bool, error) {
	l, err := layout(src, s.Event)
	if err != nil {
		return false, err
	}
	for _, e := range l.eventElem {
		if groupHasOwn(src[e.Start:e.End], s.Subcommand) {
			return true, nil
		}
	}
	return false, nil
}

// claudeInstall returns the source with our handlers added. existed is false
// for a file that is not there yet.
func claudeInstall(src []byte, existed bool) ([]byte, error) {
	if !existed || len(bytes.TrimSpace(src)) == 0 {
		src = []byte("{}\n")
	}
	st := jsonedit.DetectStyle(src)
	// validate the whole structure before the first edit
	for _, s := range Specs() {
		if _, err := layout(src, s.Event); err != nil {
			return nil, err
		}
	}
	for _, s := range Specs() {
		has, err := claudeHasOwn(src, s)
		if err != nil {
			return nil, err
		}
		if has {
			continue
		}
		l, err := layout(src, s.Event)
		if err != nil {
			return nil, err
		}
		group := claudeGroupFor(s)
		switch {
		case l.hooksIdx < 0:
			src, err = jsonedit.AddMember(src, l.root, "hooks", map[string][]claudeGroup{s.Event: {group}}, st)
		case l.eventIdx < 0:
			src, err = jsonedit.AddMember(src, l.hooks, s.Event, []claudeGroup{group}, st)
		default:
			src, err = jsonedit.AddElem(src, l.event, group, st)
		}
		if err != nil {
			return nil, err
		}
	}
	return src, nil
}

// claudeRemove deletes only our handlers, and the containers that this left
// empty. Nothing else is touched.
func claudeRemove(src []byte) ([]byte, error) {
	if _, err := jsonedit.RootSpan(src); err != nil {
		return nil, err
	}
	st := jsonedit.DetectStyle(src)
	for _, s := range Specs() {
		for {
			l, err := layout(src, s.Event)
			if err != nil {
				return nil, err
			}
			gi := -1
			for i, e := range l.eventElem {
				if groupHasOwn(src[e.Start:e.End], s.Subcommand) {
					gi = i
					break
				}
			}
			if gi < 0 {
				break
			}
			if src, err = removeOwn(src, l, gi, s, st); err != nil {
				return nil, err
			}
		}
	}
	return src, nil
}

func removeOwn(src []byte, l claudeLayout, gi int, s Spec, st jsonedit.Style) ([]byte, error) {
	group := l.eventElem[gi]
	handlers, err := handlerSpans(src, group)
	if err != nil {
		return nil, err
	}
	if len(handlers) > 1 {
		for hi, h := range handlers {
			if isOwnHandler(src[h.Start:h.End], s.Subcommand) {
				arr, err := handlersArray(src, group)
				if err != nil {
					return nil, err
				}
				return jsonedit.RemoveElem(src, arr, hi)
			}
		}
	}
	// the group holds only our handler: the whole group goes
	src, err = jsonedit.RemoveElem(src, l.event, gi)
	if err != nil {
		return nil, err
	}
	// containers we emptied go too
	l2, err := layout(src, s.Event)
	if err != nil {
		return nil, err
	}
	if len(l2.eventElem) > 0 {
		return src, nil
	}
	if src, err = jsonedit.RemoveMember(src, l2.hooks, l2.eventIdx); err != nil {
		return nil, err
	}
	l3, err := layout(src, s.Event)
	if err != nil {
		return nil, err
	}
	if len(l3.hooksMs) > 0 {
		return src, nil
	}
	return jsonedit.RemoveMember(src, l3.root, l3.hooksIdx)
}

// handlersArray is the jsonedit.Span of the "hooks" array of a group.
func handlersArray(src []byte, group jsonedit.Span) (jsonedit.Span, error) {
	ms, err := jsonedit.ParseObject(src, group)
	if err != nil {
		return jsonedit.Span{}, err
	}
	i, ok := jsonedit.LastMember(ms, "hooks")
	if !ok {
		return jsonedit.Span{}, jsonedit.ShapeErr("group without hooks")
	}
	return ms[i].Val, nil
}

func handlerSpans(src []byte, group jsonedit.Span) ([]jsonedit.Span, error) {
	arr, err := handlersArray(src, group)
	if err != nil {
		return nil, err
	}
	return jsonedit.ParseArray(src, arr)
}

// claudeMissing lists the subcommands whose handler is not installed.
func claudeMissing(src []byte, existed bool) ([]string, error) {
	var missing []string
	if !existed {
		for _, s := range Specs() {
			missing = append(missing, s.Subcommand)
		}
		return missing, nil
	}
	for _, s := range Specs() {
		has, err := claudeHasOwn(src, s)
		if err != nil {
			return nil, err
		}
		if !has {
			missing = append(missing, s.Subcommand)
		}
	}
	return missing, nil
}
