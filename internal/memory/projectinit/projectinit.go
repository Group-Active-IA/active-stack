// Package projectinit implements "active-stack memory init" (R-17, [EN] Fase 7):
// it leaves a project ready for the team memory mode by writing
// .engram/config.json (project_name), .engram/active-stack.json (the
// cloud-branch marker) and the .gitignore rules, and it only PROPOSES the
// commit: it never runs git add or git commit.
//
// Governance ALTO: it writes inside the user's repository. Everything is
// computed before the first write; a conflicting or invalid file stops the run
// untouched; a snapshot precedes the writes; and a failure after the first
// write puts every file back as it was.
package projectinit

import "errors"

// Files this package may write, relative to the repository root, with forward
// slashes (what Git expects).
const (
	ConfigPath    = ".engram/config.json"
	MarkerPath    = ".engram/active-stack.json"
	GitignorePath = ".gitignore"
)

// Policy is the only memory policy this command writes.
const Policy = "cloud-branch"

// CommitMessage is the proposed commit message.
const CommitMessage = "chore(engram): pin project and team memory policy"

// Block markers of the .gitignore rules.
const (
	BlockBegin = "# active-stack:memory begin"
	BlockEnd   = "# active-stack:memory end"
)

// Warning codes.
const (
	WarnDirtyFiles   = "dirty_files"   // files we write already had uncommitted changes
	WarnTrackedOther = "tracked_other" // other files under .engram/ are already versioned
)

// Name sources.
const (
	SourceFlag      = "flag"
	SourceRemote    = "remote"
	SourceDirectory = "directory"
)

// Options of Init.
type Options struct {
	// Dir is any directory inside the repository.
	Dir string
	// Name is the explicit project name (--name). It counts as confirmed.
	Name string
	// Yes accepts the derived name (--yes).
	Yes bool
	// Confirm asks the person to accept the derived name. Nil means there is
	// no one to ask (not a terminal).
	Confirm func(derived string) (bool, error)
	// Home is the person's home directory: where backups go and a place
	// Init refuses to write into.
	Home string
}

// Action is what happened to a file.
type Action string

// File actions.
const (
	ActionCreated   Action = "created"
	ActionUpdated   Action = "updated"
	ActionUnchanged Action = "unchanged"
)

// FileResult is the outcome for one file.
type FileResult struct {
	Path   string
	Action Action
}

// Warning is something worth telling that does not stop the run.
type Warning struct {
	Code  string
	Files []string
}

// Proposal is the commit that is proposed and never executed.
type Proposal struct {
	Files    []string
	Message  string
	Commands []string
}

// Result is the outcome of Init. On error it still carries what was known.
type Result struct {
	Root        string
	ProjectName string
	NameSource  string
	Files       []FileResult
	BackupDir   string
	Warnings    []Warning
	Proposal    *Proposal
}

// Errors.
var (
	// ErrNotGitRepo means the directory is not inside a Git repository.
	ErrNotGitRepo = errors.New("not inside a Git repository")
	// ErrHomeDirectory means the repository root is the person's home.
	ErrHomeDirectory = errors.New("the repository root is the home directory")
	// ErrEmptyName means no usable project name could be built.
	ErrEmptyName = errors.New("the project name has no usable characters")
	// ErrGitMissing means the git executable is not available.
	ErrGitMissing = errors.New("git is not installed or not in PATH")
)

// Conflict reasons.
const (
	ReasonValue        = "value"         // a key already has another value
	ReasonIgnoredDir   = "ignored_dir"   // a rule of the person ignores .engram/ as a whole
	ReasonStillIgnored = "still_ignored" // Git would still ignore a file we need versioned
	ReasonNotIgnored   = "not_ignored"   // Git would not ignore the rest of .engram/
)

// ConflictError means something of the person's would have to be overwritten
// or contradicted. Nothing was changed.
type ConflictError struct {
	File   string
	Reason string
	Key    string
	Have   string
	Want   string
}

func (e *ConflictError) Error() string { return "conflict in " + e.File + ": " + e.Reason }

// NeedsConfirmationError means the project name was not confirmed. Nothing was
// changed.
type NeedsConfirmationError struct {
	Derived  string
	Source   string
	Declined bool
}

func (e *NeedsConfirmationError) Error() string {
	return "the project name " + e.Derived + " was not confirmed"
}

// InvalidFileError means a file could not be read as what it should be.
type InvalidFileError struct {
	File string
	Err  error
}

func (e *InvalidFileError) Error() string { return e.File + ": " + e.Err.Error() }
func (e *InvalidFileError) Unwrap() error { return e.Err }

// BackupError means the snapshot failed, so nothing was written.
type BackupError struct{ Err error }

func (e *BackupError) Error() string { return "backup failed, nothing was touched: " + e.Err.Error() }
func (e *BackupError) Unwrap() error { return e.Err }
