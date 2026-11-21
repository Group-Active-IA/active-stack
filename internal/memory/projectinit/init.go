package projectinit

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/filemerge"
)

var (
	snapshotCreate = func(dir string, paths []string) error {
		_, err := backup.NewSnapshotter().Create(dir, paths)
		return err
	}
	// writeFile is used for the forward writes only. The rollback always uses
	// defaultWriteFile, so an injected failure cannot also break the undo.
	writeFile = defaultWriteFile
)

// SetSnapshotCreate replaces the snapshot function (tests) and returns a
// function that restores the previous one.
func SetSnapshotCreate(fn func(dir string, paths []string) error) (restore func()) {
	old := snapshotCreate
	snapshotCreate = fn
	return func() { snapshotCreate = old }
}

// SetWriteFile replaces the file writer (tests) and returns a restore function.
func SetWriteFile(fn func(path string, content []byte) error) (restore func()) {
	old := writeFile
	writeFile = fn
	return func() { writeFile = old }
}

// defaultWriteFile writes atomically, keeps the permissions of an existing file
// and does not change the permissions of the directory.
func defaultWriteFile(path string, content []byte) error {
	perm := fs.FileMode(0o644)
	if fi, err := os.Stat(path); err == nil {
		perm = fi.Mode().Perm()
	}
	_, err := filemerge.WriteFileAtomicPreservingParent(path, content, perm)
	return err
}

// target is one file Init may write.
type target struct {
	rel     string
	abs     string
	before  []byte
	existed bool
	next    []byte
}

func (t target) changed() bool { return !t.existed || string(t.before) != string(t.next) }

// Init leaves the project ready for the team memory mode and proposes the
// commit. See the package documentation for the guarantees.
func Init(opts Options) (Result, error) {
	var res Result
	if opts.Home == "" {
		return res, errors.New("projectinit: Options.Home is required")
	}
	root, err := resolveRoot(opts.Dir)
	if err != nil {
		return res, err
	}
	res.Root = root
	if sameDir(root, opts.Home) {
		return res, ErrHomeDirectory
	}

	// 1. the name, without writing anything
	name, source := "", SourceFlag
	if opts.Name != "" {
		if name, err = sanitizeName(opts.Name); err != nil {
			return res, err
		}
	} else if name, source, err = DeriveName(root); err != nil {
		return res, err
	}
	res.ProjectName, res.NameSource = name, source

	// 2. read and plan the three files; any conflict or invalid file stops here
	cfg, err := loadTarget(root, ConfigPath)
	if err != nil {
		return res, err
	}
	if cfg.next, err = planJSON(ConfigPath, cfg.before, cfg.existed, "project_name", name); err != nil {
		return res, err
	}
	mk, err := loadTarget(root, MarkerPath)
	if err != nil {
		return res, err
	}
	if mk.next, err = planJSON(MarkerPath, mk.before, mk.existed, "memory_policy", Policy); err != nil {
		return res, err
	}
	gi, err := loadTarget(root, GitignorePath)
	if err != nil {
		return res, err
	}
	if gi.next, err = planGitignore(gi.before, gi.existed); err != nil {
		return res, err
	}
	targets := []*target{&cfg, &mk, &gi}

	// 3. the person confirms the name
	if source != SourceFlag && !opts.Yes {
		if opts.Confirm == nil {
			return res, &NeedsConfirmationError{Derived: name, Source: source}
		}
		ok, cerr := opts.Confirm(name)
		if cerr != nil {
			return res, cerr
		}
		if !ok {
			return res, &NeedsConfirmationError{Derived: name, Source: source, Declined: true}
		}
	}

	// 4. warnings (read only)
	all := []string{ConfigPath, MarkerPath, GitignorePath}
	if dirty, derr := dirtyFiles(root, all); derr == nil && len(dirty) > 0 {
		res.Warnings = append(res.Warnings, Warning{Code: WarnDirtyFiles, Files: dirty})
	}
	if other, oerr := trackedOther(root); oerr == nil && len(other) > 0 {
		res.Warnings = append(res.Warnings, Warning{Code: WarnTrackedOther, Files: other})
	}

	// 5. snapshot of what already exists and is about to change
	var toBackUp []string
	anyChange := false
	for _, t := range targets {
		if t.changed() {
			anyChange = true
			if t.existed {
				toBackUp = append(toBackUp, t.abs)
			}
		}
	}
	if len(toBackUp) > 0 {
		sum := sha256.Sum256([]byte(root))
		dir := filepath.Join(opts.Home, ".active-stack", "backups", "memory-init-"+hex.EncodeToString(sum[:4]))
		if err := snapshotCreate(dir, toBackUp); err != nil {
			return res, &BackupError{Err: err}
		}
		res.BackupDir = dir
	}

	// 6. write; if anything fails from here on, put everything back
	createdDir := false
	if anyChange {
		if createdDir, err = writeAll(root, targets); err != nil {
			return res, err
		}
	}
	for _, t := range targets {
		a := ActionUnchanged
		switch {
		case !t.existed && t.changed():
			a = ActionCreated
		case t.changed():
			a = ActionUpdated
		}
		res.Files = append(res.Files, FileResult{Path: t.rel, Action: a})
	}

	// 7. Git must see exactly the two files as versionable and the rest of
	// .engram/ as ignored. If not, undo what was written.
	if verr := verifyIgnore(root); verr != nil {
		if anyChange {
			if rerr := rollback(root, targets, createdDir); rerr != nil {
				return res, fmt.Errorf("%w (and the undo failed: %v)", verr, rerr)
			}
			res.Files = nil
		}
		return res, verr
	}

	// 8. propose, never execute
	pending, perr := pendingFiles(root, all)
	if perr != nil {
		return res, perr
	}
	if len(pending) > 0 {
		res.Proposal = proposal(pending)
	}
	return res, nil
}

func loadTarget(root, rel string) (target, error) {
	t := target{rel: rel, abs: filepath.Join(root, filepath.FromSlash(rel))}
	b, err := os.ReadFile(t.abs)
	switch {
	case err == nil:
		t.before, t.existed = b, true
	case os.IsNotExist(err):
	default:
		return t, fmt.Errorf("%s: %w", rel, err)
	}
	return t, nil
}

func writeAll(root string, targets []*target) (createdDir bool, err error) {
	engramDir := filepath.Join(root, ".engram")
	if _, serr := os.Stat(engramDir); os.IsNotExist(serr) {
		createdDir = true
	}
	var written []*target
	defer func() {
		if err == nil {
			return
		}
		if rerr := rollbackWritten(written, createdDir, engramDir); rerr != nil {
			err = fmt.Errorf("%w (and the undo failed: %v)", err, rerr)
		}
	}()
	for _, t := range targets {
		if !t.changed() {
			continue
		}
		if mkerr := os.MkdirAll(filepath.Dir(t.abs), 0o755); mkerr != nil {
			return createdDir, fmt.Errorf("%s: %w", t.rel, mkerr)
		}
		if werr := writeFile(t.abs, t.next); werr != nil {
			return createdDir, fmt.Errorf("%s: %w", t.rel, werr)
		}
		written = append(written, t)
	}
	return createdDir, nil
}

// rollback undoes a completed write phase.
func rollback(root string, targets []*target, removeDir bool) error {
	var written []*target
	for _, t := range targets {
		if t.changed() {
			written = append(written, t)
		}
	}
	return rollbackWritten(written, removeDir, filepath.Join(root, ".engram"))
}

func rollbackWritten(written []*target, removeDir bool, engramDir string) error {
	var firstErr error
	for i := len(written) - 1; i >= 0; i-- {
		t := written[i]
		var err error
		if t.existed {
			err = defaultWriteFile(t.abs, t.before)
		} else if rerr := os.Remove(t.abs); rerr != nil && !os.IsNotExist(rerr) {
			err = rerr
		}
		if err != nil && firstErr == nil {
			firstErr = fmt.Errorf("%s: %w", t.rel, err)
		}
	}
	if removeDir {
		_ = os.Remove(engramDir) // only if empty
	}
	return firstErr
}

// verifyIgnore asks Git, ignoring the index, what the rules say now.
func verifyIgnore(root string) error {
	for _, rel := range []string{ConfigPath, MarkerPath} {
		ig, err := isIgnored(root, rel)
		if err != nil {
			return err
		}
		if ig {
			return &ConflictError{File: GitignorePath, Reason: ReasonStillIgnored, Have: rel}
		}
	}
	const probe = ".engram/active-stack-probe"
	ig, err := isIgnored(root, probe)
	if err != nil {
		return err
	}
	if !ig {
		return &ConflictError{File: GitignorePath, Reason: ReasonNotIgnored, Have: probe}
	}
	return nil
}

func proposal(files []string) *Proposal {
	list := ""
	for _, f := range files {
		list += " " + f
	}
	return &Proposal{
		Files:   files,
		Message: CommitMessage,
		Commands: []string{
			"git add --" + list,
			`git commit -m "` + CommitMessage + `" --` + list,
		},
	}
}

func sameDir(a, b string) bool {
	fa, errA := os.Stat(a)
	fb, errB := os.Stat(b)
	return errA == nil && errB == nil && os.SameFile(fa, fb)
}
