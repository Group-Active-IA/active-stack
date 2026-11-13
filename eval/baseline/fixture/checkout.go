package fixture

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Checkout es un checkout descartable de una repeticion. Dir vive dentro de
// un directorio temporal propio; nunca es el arbol de trabajo del repo.
type Checkout struct {
	Dir  string
	root string
	f    *Fixture
}

// Prepare materializa un checkout nuevo desde el fixture inmutable y
// comprueba que su huella sea la congelada en task.json.
func (f *Fixture) Prepare() (*Checkout, error) {
	root, err := os.MkdirTemp("", "opsx-fixture-")
	if err != nil {
		return nil, err
	}
	c := &Checkout{Dir: filepath.Join(root, "checkout"), root: root, f: f}
	if err := c.Reset(); err != nil {
		os.RemoveAll(root)
		return nil, err
	}
	return c, nil
}

// Reset borra el checkout y lo vuelve a materializar desde la fuente
// inmutable (restauracion entre repeticiones o intentos).
func (c *Checkout) Reset() error {
	if err := os.RemoveAll(c.Dir); err != nil {
		return err
	}
	if err := c.f.materialize(c.Dir); err != nil {
		return err
	}
	got, err := TreeHash(c.Dir)
	if err != nil {
		return err
	}
	if got != c.f.Task.Source.SHA256 {
		return fmt.Errorf("checkout materializado %s no coincide con la huella congelada %s", got, c.f.Task.Source.SHA256)
	}
	return nil
}

// Close borra el checkout y su directorio temporal.
func (c *Checkout) Close() error { return os.RemoveAll(c.root) }

// materialize escribe el estado inicial en dst (que no debe existir).
func (f *Fixture) materialize(dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	s := f.Task.Source
	switch s.Kind {
	case "embedded":
		return copyTree(f.path(s.Dir), dst)
	case "git":
		repo, err := f.repoRoot()
		if err != nil {
			return err
		}
		// core.autocrlf=false: el contenido depende solo de la revision y de
		// sus .gitattributes, no de la configuracion local.
		cmd := exec.Command("git", "-C", repo, "-c", "core.autocrlf=false", "archive", "--format=tar", s.Revision)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("git archive %s: %v: %s", s.Revision, err, stderr.String())
		}
		if err := untar(bytes.NewReader(out), dst); err != nil {
			return err
		}
		if s.Patch != "" {
			return ApplyPatch(dst, f.path(s.Patch))
		}
		return nil
	}
	return fmt.Errorf("tipo de fuente desconocido %q", s.Kind)
}

// repoRoot es el repositorio que contiene el fixture o, si el fixture fue
// copiado fuera, el del directorio actual.
func (f *Fixture) repoRoot() (string, error) {
	for _, dir := range []string{f.Dir, "."} {
		out, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").Output()
		if err == nil {
			return strings.TrimSpace(string(out)), nil
		}
	}
	return "", fmt.Errorf("no se encontró el repositorio para la revisión del fixture")
}

func untar(r io.Reader, dst string) error {
	tr := tar.NewReader(r)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		name := filepath.Clean(filepath.FromSlash(h.Name))
		if filepath.IsAbs(name) || name == ".." || strings.HasPrefix(name, ".."+string(filepath.Separator)) {
			return fmt.Errorf("ruta insegura en el archivo: %q", h.Name)
		}
		target := filepath.Join(dst, name)
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			b, err := io.ReadAll(tr)
			if err != nil {
				return err
			}
			if err := os.WriteFile(target, b, 0o644); err != nil {
				return err
			}
		case tar.TypeXGlobalHeader:
		default:
			return fmt.Errorf("entrada no soportada en el archivo: %q", h.Name)
		}
	}
}

// ApplyPatch aplica patch (ruta) en dir con git apply. GIT_CEILING_DIRECTORIES
// impide que git descubra un repositorio por encima de dir y aplique alli.
func ApplyPatch(dir, patch string) error {
	abs, err := filepath.Abs(patch)
	if err != nil {
		return err
	}
	cmd := exec.Command("git", "apply", "--whitespace=nowarn", abs)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CEILING_DIRECTORIES="+filepath.Dir(dir))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git apply %s: %v: %s", filepath.Base(patch), err, out)
	}
	return nil
}
