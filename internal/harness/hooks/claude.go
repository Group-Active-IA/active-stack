package hooks

import (
	"bytes"
	"encoding/json"
	"strings"
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
	root      span
	rootMs    []member
	hooksIdx  int // index in rootMs, -1 when there is no "hooks"
	hooks     span
	hooksMs   []member
	eventIdx  int // index in hooksMs, -1 when the event is absent
	event     span
	eventElem []span
}

// layout locates "hooks" and one event. It refuses an unexpected structure:
// the caller leaves the file intact.
func layout(src []byte, event string) (claudeLayout, error) {
	l := claudeLayout{hooksIdx: -1, eventIdx: -1}
	var err error
	if l.root, err = rootSpan(src); err != nil {
		return l, err
	}
	if l.rootMs, err = parseObject(src, l.root); err != nil {
		return l, err
	}
	i, ok := lastMember(l.rootMs, "hooks")
	if !ok {
		return l, nil
	}
	l.hooksIdx, l.hooks = i, l.rootMs[i].val
	if src[l.hooks.start] != '{' {
		return l, shapeErr(`"hooks" is not an object`)
	}
	if l.hooksMs, err = parseObject(src, l.hooks); err != nil {
		return l, err
	}
	j, ok := lastMember(l.hooksMs, event)
	if !ok {
		return l, nil
	}
	l.eventIdx, l.event = j, l.hooksMs[j].val
	if src[l.event.start] != '[' {
		return l, shapeErr(`"hooks.%s" is not an array`, event)
	}
	l.eventElem, err = parseArray(src, l.event)
	return l, err
}

// claudeHasOwn reports whether the event already holds our handler for s.
func claudeHasOwn(src []byte, s Spec) (bool, error) {
	l, err := layout(src, s.Event)
	if err != nil {
		return false, err
	}
	for _, e := range l.eventElem {
		if groupHasOwn(src[e.start:e.end], s.Subcommand) {
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
	st := detectStyle(src)
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
			src, err = addMember(src, l.root, "hooks", map[string][]claudeGroup{s.Event: {group}}, st)
		case l.eventIdx < 0:
			src, err = addMember(src, l.hooks, s.Event, []claudeGroup{group}, st)
		default:
			src, err = addElem(src, l.event, group, st)
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
	if _, err := rootSpan(src); err != nil {
		return nil, err
	}
	st := detectStyle(src)
	for _, s := range Specs() {
		for {
			l, err := layout(src, s.Event)
			if err != nil {
				return nil, err
			}
			gi := -1
			for i, e := range l.eventElem {
				if groupHasOwn(src[e.start:e.end], s.Subcommand) {
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

func removeOwn(src []byte, l claudeLayout, gi int, s Spec, st style) ([]byte, error) {
	group := l.eventElem[gi]
	handlers, err := handlerSpans(src, group)
	if err != nil {
		return nil, err
	}
	if len(handlers) > 1 {
		for hi, h := range handlers {
			if isOwnHandler(src[h.start:h.end], s.Subcommand) {
				arr, err := handlersArray(src, group)
				if err != nil {
					return nil, err
				}
				return removeElem(src, arr, hi)
			}
		}
	}
	// the group holds only our handler: the whole group goes
	src, err = removeElem(src, l.event, gi)
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
	if src, err = removeMember(src, l2.hooks, l2.eventIdx); err != nil {
		return nil, err
	}
	l3, err := layout(src, s.Event)
	if err != nil {
		return nil, err
	}
	if len(l3.hooksMs) > 0 {
		return src, nil
	}
	return removeMember(src, l3.root, l3.hooksIdx)
}

// handlersArray is the span of the "hooks" array of a group.
func handlersArray(src []byte, group span) (span, error) {
	ms, err := parseObject(src, group)
	if err != nil {
		return span{}, err
	}
	i, ok := lastMember(ms, "hooks")
	if !ok {
		return span{}, shapeErr("group without hooks")
	}
	return ms[i].val, nil
}

func handlerSpans(src []byte, group span) ([]span, error) {
	arr, err := handlersArray(src, group)
	if err != nil {
		return nil, err
	}
	return parseArray(src, arr)
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
