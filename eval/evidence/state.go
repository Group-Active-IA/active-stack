package evidence

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

// StateFingerprint calcula la huella del estado probado de un worktree de Git (design.md D2).
//
// El conjunto de archivos es `git ls-files -z --cached --others --exclude-standard`: versionados,
// modificados o no, más los nuevos no ignorados, sin los borrados y sin `.git`. Con él se arma el
// mismo texto que TreeHash de 02 (eval/baseline/fixture): las líneas "<sha256 del archivo>  <ruta>\n"
// ordenadas por ruta, y se hace el sha256 de ese texto. Se leen los bytes del disco, así que el
// valor no depende de core.autocrlf (medido en informes/medicion-huella-2026-11-13.md).
//
// Límites: un archivo ignorado por Git no entra; un archivo que no es regular (enlace simbólico,
// submódulo) es error, como en TreeHash.
func StateFingerprint(dir string) (Fingerprint, error) {
	cmd := exec.Command("git", "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return Fingerprint{}, fmt.Errorf("git ls-files en %s: %w: %s", dir, err, bytes.TrimSpace(stderr.Bytes()))
	}
	seen := map[string]bool{}
	var paths []string
	for _, p := range bytes.Split(out, []byte{0}) {
		if len(p) == 0 || seen[string(p)] {
			continue
		}
		seen[string(p)] = true
		paths = append(paths, string(p))
	}
	sort.Strings(paths) // orden de bytes, el mismo que LC_ALL=C sort

	h := sha256.New()
	for _, p := range paths {
		full := filepath.Join(dir, filepath.FromSlash(p))
		info, err := os.Lstat(full)
		if os.IsNotExist(err) {
			continue // borrado en la copia de trabajo
		}
		if err != nil {
			return Fingerprint{}, err
		}
		if !info.Mode().IsRegular() {
			return Fingerprint{}, fmt.Errorf("%s: no es un archivo regular", p)
		}
		sum, err := fileSHA256(full)
		if err != nil {
			return Fingerprint{}, err
		}
		fmt.Fprintf(h, "%s  %s\n", sum, p)
	}
	return Fingerprint{Algorithm: "sha256", Value: hex.EncodeToString(h.Sum(nil))}, nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
