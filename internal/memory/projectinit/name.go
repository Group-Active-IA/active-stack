package projectinit

import (
	"net/url"
	"path/filepath"
	"strings"
	"unicode"
)

// DeriveName computes the project name for the repository that contains dir,
// and where it came from. The origin remote wins; without one the name is the folder
// of the MAIN checkout, never of a linked worktree, so a clone, the checkout
// and its worktrees all get the same name.
func DeriveName(dir string) (name, source string, err error) {
	root, err := resolveRoot(dir)
	if err != nil {
		return "", "", err
	}
	if rawURL := remoteURL(root); rawURL != "" {
		if n, nerr := nameFromRemoteURL(rawURL); nerr == nil {
			return n, SourceRemote, nil
		}
	}
	main := root
	if out, code, gerr := gitOut(root, "rev-parse", "--git-common-dir"); gerr == nil && code == 0 {
		common := filepath.FromSlash(strings.TrimSpace(out))
		if !filepath.IsAbs(common) {
			common = filepath.Join(root, common)
		}
		common = filepath.Clean(common)
		if filepath.Base(common) == ".git" {
			main = filepath.Dir(common)
		}
	}
	n, err := sanitizeName(filepath.Base(main))
	if err != nil {
		return "", "", err
	}
	return n, SourceDirectory, nil
}

// remoteURL is the URL of origin, or "" when there is none. Other remotes are
// not used: R17-C approved origin, then the folder of the main checkout.
func remoteURL(root string) string {
	out, code, err := gitOut(root, "config", "--get", "remote.origin.url")
	if err != nil || code != 0 {
		return ""
	}
	return strings.TrimSpace(out)
}

// nameFromRemoteURL takes the last path segment of a remote URL: https, ssh,
// scp-like (git@host:org/repo.git), file:// and local paths, with or without
// ".git" or a trailing slash.
func nameFromRemoteURL(raw string) (string, error) {
	s := strings.ReplaceAll(strings.TrimSpace(raw), `\`, "/")
	if i := strings.Index(s, "://"); i >= 0 {
		rest := s[i+3:]
		j := strings.Index(rest, "/")
		if j < 0 {
			return "", ErrEmptyName
		}
		s = rest[j:] // drop scheme and authority
	}
	s = strings.TrimRight(s, "/")
	s = strings.TrimSuffix(s, ".git")
	s = strings.TrimRight(s, "/")
	seg := s
	if k := strings.LastIndex(s, "/"); k >= 0 {
		seg = s[k+1:]
	} else if k := strings.LastIndex(s, ":"); k >= 0 {
		seg = s[k+1:] // scp-like without a slash: git@host:repo
	}
	if d, err := url.PathUnescape(seg); err == nil {
		seg = d
	}
	return sanitizeName(seg)
}

// sanitizeName changes what would break the name: whitespace and path
// separators become "-", and the name is lower-cased, because Engram lower-cases
// the project name itself (R17-D, owner decision 2026-11-21). Unicode is kept.
func sanitizeName(s string) (string, error) {
	isSep := func(r rune) bool { return unicode.IsSpace(r) || r == '/' || r == '\\' }
	s = strings.TrimFunc(s, isSep)
	var b strings.Builder
	inRun := false
	hasWord := false
	for _, r := range s {
		if isSep(r) {
			if !inRun {
				b.WriteByte('-')
			}
			inRun = true
			continue
		}
		inRun = false
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			hasWord = true
		}
		b.WriteRune(r)
	}
	if !hasWord {
		return "", ErrEmptyName
	}
	return strings.ToLower(b.String()), nil
}
