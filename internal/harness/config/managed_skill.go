package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/filemerge"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// ManagedSkillAdapter resolves machine paths without requiring a client-specific layout.
type ManagedSkillAdapter interface {
	Agent() model.Agent
	InstructionsPath(string) string
	SkillsDir(string) string
}

var managedSkillWrite = filemerge.WriteFileAtomicPreservingParent

const managedInventory = ".active-stack-managed.json"

type managedSkillState struct {
	Owner          string            `json:"owner"`
	Harness        string            `json:"harness"`
	Files          map[string]string `json:"files"`
	RestoreBlocked map[string]bool   `json:"restore_blocked,omitempty"`
}

func managedRoot(a ManagedSkillAdapter, home, id string) (string, error) {
	if id != "sdd-orchestrator" || a.Agent() != model.AgentCodex {
		return "", fmt.Errorf("unsupported managed skill %q for %s", id, a.Agent())
	}
	if a.SkillsDir(home) == "" || a.InstructionsPath(home) == "" {
		return "", errors.New("managed skill adapter has empty paths")
	}
	return filepath.Join(a.SkillsDir(home), id), nil
}
func migrationRoot(a ManagedSkillAdapter, home, id string) string {
	return filepath.Join(home, ".active-stack", "backups", id+"-"+string(a.Agent())+"-native-migration")
}
func digest(content []byte) string { sum := sha256.Sum256(content); return hex.EncodeToString(sum[:]) }
func readOptional(path string) ([]byte, error) {
	b, e := os.ReadFile(path)
	if os.IsNotExist(e) {
		return nil, nil
	}
	return b, e
}
func removeOrchestrator(content string) string {
	return filemerge.RemoveMarkdownSectionsPreservingContent(content, "sdd-orchestrator")
}

func managedInstructions(existing, universal string) (string, error) {
	clean := removeOrchestrator(existing)
	for _, prefix := range []string{"active-stack:", "jr-stack:"} {
		if strings.Contains(clean, "<!-- "+prefix+"sdd-orchestrator -->") || strings.Contains(clean, "<!-- /"+prefix+"sdd-orchestrator -->") {
			return "", errors.New("malformed orchestrator markers")
		}
	}
	count := strings.Count(existing, "<!-- active-stack:sdd-orchestrator -->") + strings.Count(existing, "<!-- jr-stack:sdd-orchestrator -->")
	if count == 1 {
		return filemerge.InjectMarkdownSection(existing, "sdd-orchestrator", universal), nil
	}
	return filemerge.InjectMarkdownSection(clean, "sdd-orchestrator", universal), nil
}

func acquireManagedLock(a ManagedSkillAdapter, home, id string) (func(), error) {
	path := filepath.Join(home, ".active-stack", "locks", id+"-"+string(a.Agent()))
	if err := checkRegularParents(path, home); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	if err := os.Mkdir(path, 0700); err != nil {
		return nil, fmt.Errorf("managed skill operation busy or lock unavailable at %s: %w", path, err)
	}
	return func() { _ = os.Remove(path) }, nil
}

// ManagedSkillFiles returns the desired package files at adapter-resolved paths.
func ManagedSkillFiles(h model.Harness, a ManagedSkillAdapter, home string) (map[string]string, error) {
	root, err := managedRoot(a, home, h.ID)
	if err != nil {
		return nil, err
	}
	pkg, err := ComposeCodexPackage(h.Toggles)
	if err != nil {
		return nil, err
	}
	raw, err := readOptional(a.InstructionsPath(home))
	if err != nil {
		return nil, err
	}
	files := make(map[string]string, len(pkg.Files)+1)
	instructions, err := managedInstructions(string(raw), pkg.Universal)
	if err != nil {
		return nil, err
	}
	files[a.InstructionsPath(home)] = instructions
	for name, content := range pkg.Files {
		files[filepath.Join(root, filepath.FromSlash(name))] = content
	}
	return files, nil
}

func loadManagedState(root, id string) (managedSkillState, bool, error) {
	raw, err := readOptional(filepath.Join(root, managedInventory))
	if err != nil {
		return managedSkillState{}, false, err
	}
	if raw == nil {
		return managedSkillState{}, false, nil
	}
	var state managedSkillState
	if err = json.Unmarshal(raw, &state); err != nil {
		return state, false, fmt.Errorf("invalid managed inventory: %w", err)
	}
	if state.Owner != "active-stack" || state.Harness != id || state.Files == nil {
		return state, false, errors.New("foreign or incomplete managed inventory")
	}
	for name, hash := range state.Files {
		if name == "" || filepath.IsAbs(name) || name == ".." || strings.HasPrefix(filepath.Clean(name), ".."+string(filepath.Separator)) || filepath.Clean(name) != name || name == managedInventory {
			return state, false, fmt.Errorf("invalid managed file %q", name)
		}
		if b, err := hex.DecodeString(hash); err != nil || len(b) != sha256.Size {
			return state, false, fmt.Errorf("invalid hash for %q", name)
		}
	}
	return state, true, nil
}

func checkRegularParents(path, home string) error {
	relative, err := filepath.Rel(home, path)
	if err != nil || filepath.IsAbs(relative) || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("managed path %q outside home %q", path, home)
	}
	for current := path; current != filepath.Dir(home); current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing symlink path %q", current)
		}
		if current != path && !info.IsDir() {
			return fmt.Errorf("parent %q is not a directory", current)
		}
	}
	return nil
}

// InstallManagedSkill installs a complete machine package after validating ownership.
func InstallManagedSkill(h model.Harness, a ManagedSkillAdapter, home string) (Result, error) {
	root, err := managedRoot(a, home, h.ID)
	if err != nil {
		return Result{}, err
	}
	release, err := acquireManagedLock(a, home, h.ID)
	if err != nil {
		return Result{}, err
	}
	defer release()
	if err = checkRegularParents(root, home); err != nil {
		return Result{}, err
	}
	if err = checkRegularParents(migrationRoot(a, home, h.ID), home); err != nil {
		return Result{}, err
	}

	files, err := ManagedSkillFiles(h, a, home)
	if err != nil {
		return Result{}, err
	}
	state, exists, err := loadManagedState(root, h.ID)
	if err != nil {
		return Result{}, err
	}
	if !exists {
		if entries, e := os.ReadDir(root); e == nil && len(entries) > 0 {
			return Result{}, fmt.Errorf("foreign skill directory %q", root)
		} else if e != nil && !os.IsNotExist(e) {
			return Result{}, e
		}
	}
	for name, hash := range state.Files {
		raw, e := os.ReadFile(filepath.Join(root, name))
		if e != nil || digest(raw) != hash {
			return Result{}, fmt.Errorf("managed file changed: %s", filepath.Join(root, name))
		}
	}
	paths := make([]string, 0, len(files)+1)
	for path := range files {
		paths = append(paths, path)
	}
	for name := range state.Files {
		path := filepath.Join(root, name)
		if _, ok := files[path]; !ok {
			files[path] = ""
			paths = append(paths, path)
		}
	}
	inventoryPath := filepath.Join(root, managedInventory)
	paths = append(paths, inventoryPath)
	sort.Strings(paths)
	for _, path := range paths {
		if err = checkRegularParents(path, home); err != nil {
			return Result{}, err
		}
		if info, e := os.Lstat(path); e == nil && !info.Mode().IsRegular() {
			return Result{}, fmt.Errorf("not a regular file: %s", path)
		}
	}
	next := managedSkillState{Owner: "active-stack", Harness: h.ID, Files: map[string]string{}, RestoreBlocked: state.RestoreBlocked}
	if next.RestoreBlocked == nil {
		next.RestoreBlocked = map[string]bool{}
	}
	for path, content := range files {
		if path != a.InstructionsPath(home) && content != "" {
			relative, _ := filepath.Rel(root, path)
			next.Files[relative] = digest([]byte(content))
		}
	}
	firstPath := filepath.Join(migrationRoot(a, home, h.ID), backup.ManifestFilename)
	first, err := backup.ReadManifest(firstPath)
	if errors.Is(err, os.ErrNotExist) {
		if exists {
			return Result{}, fmt.Errorf("original migration backup missing: %s", firstPath)
		}
		snapshotPaths, dirs := managedSnapshotPaths(paths, home)
		first, err = backup.NewSnapshotter().CreateWithDirHints(migrationRoot(a, home, h.ID), snapshotPaths, dirs)
		first.Pinned = true
		first.Source = backup.BackupSourceInstall
		first.Description = "Codex OPSX native migration (original configuration)"
	}
	if err != nil {
		return Result{}, err
	}
	// Extend the original snapshot only with newly introduced, previously absent managed paths.
	known := map[string]bool{}
	for _, entry := range first.Entries {
		known[entry.OriginalPath] = true
	}
	for _, path := range paths {
		if !known[path] {
			if _, e := os.Stat(path); !os.IsNotExist(e) {
				return Result{}, fmt.Errorf("new migration path already exists: %s", path)
			}
			first.Entries = append(first.Entries, backup.ManifestEntry{OriginalPath: path})
		}
	}
	snapshotPaths, dirHints := managedSnapshotPaths(paths, home)
	for _, path := range snapshotPaths {
		if !known[path] && dirHints[path] {
			info, e := os.Stat(path)
			if e != nil && !os.IsNotExist(e) {
				return Result{}, e
			}
			first.Entries = append(first.Entries, backup.ManifestEntry{OriginalPath: path, IsDir: true, Existed: e == nil && info.IsDir()})
			known[path] = true
		}
	}
	if expected, ok := first.RestoreGuards[a.InstructionsPath(home)]; ok {
		raw, e := readOptional(a.InstructionsPath(home))
		if e != nil {
			return Result{}, e
		}
		if sha256.Sum256(raw) != expected {
			next.RestoreBlocked[a.InstructionsPath(home)] = true
		}
	}
	rawState, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return Result{}, err
	}
	rawState = append(rawState, '\n')
	files[inventoryPath] = string(rawState)
	unchanged := true
	for _, path := range paths {
		raw, e := readOptional(path)
		if e != nil {
			return Result{}, e
		}
		if string(raw) != files[path] {
			unchanged = false
		}
	}
	if unchanged {
		return Result{AllAlready: true}, nil
	}
	// A published migration point is always guarded, including an interrupted first install.
	for _, entry := range first.Entries {
		if !entry.IsDir && !next.RestoreBlocked[entry.OriginalPath] {
			first.GuardPersistentRestore(entry.OriginalPath, sha256.Sum256([]byte(files[entry.OriginalPath])))
		}
	}
	if !exists {
		if err = writeManagedManifest(firstPath, first); err != nil {
			return Result{}, err
		}
	}
	operationPaths, _ := managedSnapshotPaths(paths, home)
	var createdDirs []string
	for _, path := range operationPaths {
		if _, ok := files[path]; ok {
			continue
		}
		if _, e := os.Lstat(path); os.IsNotExist(e) {
			createdDirs = append(createdDirs, path)
		}
	}
	beforeStates := map[string]ManagedWriteState{}
	operationDir := filepath.Join(home, ".active-stack", "backups", h.ID+"-"+string(a.Agent())+"-native-operation")
	if err = checkRegularParents(operationDir, home); err != nil {
		return Result{}, err
	}
	operation, err := backup.NewSnapshotter().Create(operationDir, paths)
	if err != nil {
		return Result{}, err
	}
	for _, path := range paths {
		raw, e := readOptional(path)
		if e != nil {
			return Result{}, e
		}
		beforeStates[path] = ManagedWriteState{Exists: raw != nil, Hash: sha256.Sum256(raw)}
		if _, e := os.Stat(path); os.IsNotExist(e) {
			operation.GuardRollbackAbsent(path)
		} else {
			operation.GuardRollback(path, sha256.Sum256(raw))
		}
	}
	var written []string
	sort.SliceStable(paths, func(i, j int) bool { return paths[i] != inventoryPath && paths[j] == inventoryPath })
	for _, path := range paths {
		content := files[path]
		if content == "" {
			if e := os.Remove(path); e != nil && !os.IsNotExist(e) {
				err = e
				break
			}
		} else {
			mode := os.FileMode(0600)
			if info, e := os.Stat(path); e == nil {
				mode = info.Mode().Perm()
			}
			_, err = managedSkillWrite(path, []byte(content), mode)
			if err != nil {
				break
			}
		}
		if content == "" {
			operation.GuardRollbackAbsent(path)
		} else {
			operation.GuardRollback(path, sha256.Sum256([]byte(content)))
		}
		written = append(written, path)
		if !next.RestoreBlocked[path] {
			first.GuardPersistentRestore(path, sha256.Sum256([]byte(content)))
		}
	}
	if err == nil {
		err = writeManagedManifest(firstPath, first)
	}
	if err != nil {
		rollbackErr := (backup.RestoreService{}).Restore(operation)
		sort.Slice(createdDirs, func(i, j int) bool { return len(createdDirs[i]) > len(createdDirs[j]) })
		for _, dir := range createdDirs {
			if info, e := os.Lstat(dir); e == nil && info.IsDir() {
				_ = os.Remove(dir)
			}
		}
		restoredStates := map[string]ManagedWriteState{}
		for _, path := range written {
			restoredStates[path] = beforeStates[path]
		}
		return Result{Files: written, ManagedWrites: restoredStates}, errors.Join(err, rollbackErr)
	}
	states := map[string]ManagedWriteState{}
	for _, path := range written {
		content := files[path]
		states[path] = ManagedWriteState{Exists: content != "", Hash: sha256.Sum256([]byte(content))}
	}
	manifestBytes, e := json.MarshalIndent(first, "", "  ")
	if e != nil {
		return Result{}, e
	}
	states[firstPath] = ManagedWriteState{Exists: true, Hash: sha256.Sum256(append(manifestBytes, '\n'))}
	written = append(written, firstPath)
	return Result{Files: written, ManagedWrites: states}, nil
}

// VerifyManagedSkill checks structural content and ownership of the installed package.
func VerifyManagedSkill(h model.Harness, a ManagedSkillAdapter, home string) error {
	root, err := managedRoot(a, home, h.ID)
	if err != nil {
		return err
	}
	state, exists, err := loadManagedState(root, h.ID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("managed skill inventory missing")
	}
	files, err := ManagedSkillFiles(h, a, home)
	if err != nil {
		return err
	}
	for path, content := range files {
		raw, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		if string(raw) != content {
			return fmt.Errorf("managed skill content differs: %s", path)
		}
		if path != a.InstructionsPath(home) {
			relative, _ := filepath.Rel(root, path)
			if state.Files[relative] != digest(raw) {
				return fmt.Errorf("managed skill ownership differs: %s", path)
			}
		}
	}
	if len(state.Files) != len(files)-1 {
		return errors.New("managed inventory contains stale files")
	}
	return nil
}

// ManagedSkillOwnedPaths enumerates only recorded files and instructions for snapshots.
func ManagedSkillOwnedPaths(a ManagedSkillAdapter, home, id string) ([]string, error) {
	root, err := managedRoot(a, home, id)
	if err != nil {
		return nil, err
	}
	state, exists, err := loadManagedState(root, id)
	if err != nil {
		return nil, err
	}
	paths := []string{a.InstructionsPath(home)}
	if !exists {
		if _, e := os.Stat(filepath.Join(root, "SKILL.md")); e == nil {
			return nil, fmt.Errorf("unowned skill at %s", root)
		}
		return paths, nil
	}
	paths = append(paths, filepath.Join(root, managedInventory))
	for name := range state.Files {
		paths = append(paths, filepath.Join(root, name))
	}
	sort.Strings(paths)
	return paths, nil
}

// RemoveManagedSkill removes the managed package while retaining unrelated files.
func RemoveManagedSkill(a ManagedSkillAdapter, home, id string) ([]string, error) {
	root, err := managedRoot(a, home, id)
	if err != nil {
		return nil, err
	}
	release, err := acquireManagedLock(a, home, id)
	if err != nil {
		return nil, err
	}
	defer release()

	paths, err := ManagedSkillOwnedPaths(a, home, id)
	if err != nil {
		return nil, err
	}
	for _, path := range paths {
		if e := checkRegularParents(path, home); e != nil {
			return nil, e
		}
		if info, e := os.Lstat(path); e == nil && !info.Mode().IsRegular() {
			return nil, fmt.Errorf("managed path is not regular: %s", path)
		}
	}
	state, exists, err := loadManagedState(root, id)
	if err != nil {
		return nil, err
	}
	for name, hash := range state.Files {
		raw, e := os.ReadFile(filepath.Join(root, name))
		if e != nil || digest(raw) != hash {
			return nil, fmt.Errorf("preserved edited managed file: %s", name)
		}
	}
	raw, err := readOptional(a.InstructionsPath(home))
	if err != nil {
		return nil, err
	}
	var changed []string
	updated := removeOrchestrator(string(raw))
	if updated != string(raw) {
		mode := os.FileMode(0600)
		if info, e := os.Stat(a.InstructionsPath(home)); e == nil {
			mode = info.Mode().Perm()
		}
		if _, err = managedSkillWrite(a.InstructionsPath(home), []byte(updated), mode); err != nil {
			return nil, err
		}
	}
	if updated != string(raw) {
		changed = append(changed, a.InstructionsPath(home))
	}
	if !exists {
		return changed, nil
	}
	var dirs []string
	// Inventory is removed last, so failures retain ownership information.
	sort.SliceStable(paths, func(i, j int) bool {
		return paths[i] != filepath.Join(root, managedInventory) && paths[j] == filepath.Join(root, managedInventory)
	})
	for _, path := range paths {
		if path == a.InstructionsPath(home) {
			continue
		}
		if err = os.Remove(path); err != nil && !os.IsNotExist(err) {
			return changed, err
		}
		changed = append(changed, path)
		for dir := filepath.Dir(path); dir != filepath.Dir(root); dir = filepath.Dir(dir) {
			dirs = append(dirs, dir)
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return len(dirs[i]) > len(dirs[j]) })
	for _, dir := range dirs {
		info, e := os.Lstat(dir)
		if e == nil && info.IsDir() {
			_ = os.Remove(dir)
		}
	}
	return changed, nil
}

func managedSnapshotPaths(paths []string, home string) ([]string, map[string]bool) {
	all := append([]string(nil), paths...)
	dirs := map[string]bool{}
	for _, path := range paths {
		for dir := filepath.Dir(path); dir != home && dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
			if !dirs[dir] {
				dirs[dir] = true
				all = append(all, dir)
			}
		}
	}
	sort.Strings(all)
	return all, dirs
}

func writeManagedManifest(path string, manifest backup.Manifest) error {
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	_, err = managedSkillWrite(path, append(raw, '\n'), 0600)
	return err
}

// ManagedSkillOperationPaths returns file paths the machine installation may write.
func ManagedSkillOperationPaths(h model.Harness, a ManagedSkillAdapter, home string) ([]string, error) {
	files, err := ManagedSkillFiles(h, a, home)
	if err != nil {
		return nil, err
	}
	owned, err := ManagedSkillOwnedPaths(a, home, h.ID)
	if err != nil {
		return nil, err
	}
	set := map[string]bool{}
	for path := range files {
		set[path] = true
	}
	for _, path := range owned {
		set[path] = true
	}
	root, err := managedRoot(a, home, h.ID)
	if err != nil {
		return nil, err
	}
	set[filepath.Join(root, managedInventory)] = true
	set[filepath.Join(migrationRoot(a, home, h.ID), backup.ManifestFilename)] = true
	paths := make([]string, 0, len(set))
	for path := range set {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, nil
}
