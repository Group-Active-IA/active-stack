// Package shadow es el material del piloto en sombra del change opsx-shadow-review (ficha 10):
// el paquete de un caso armado por lista de inclusión y con huella (tarea 3.1), el examen previo
// de todo el paquete antes de cualquier envío (3.2) y la muestra preparada (1.2).
//
// Igual que el resto de eval/, es externo al binario: no lo importan cmd/ ni internal/. No envía
// nada a ningún proveedor; el envío, el registro de envío y el arnés son tareas posteriores.
package shadow

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Group-Active-IA/active-stack/eval/baseline/record"
)

// Fingerprint es la huella {algorithm, value} de 02 y 06.
type Fingerprint = record.Fingerprint

// File es un archivo del paquete: ruta relativa con barras y bytes exactos.
type File struct {
	Path string
	Data []byte
}

// Package es lo que recibiría un juez para un caso. Sólo contiene lo que nombró la lista de
// inclusión (D5).
type Package struct {
	Case        string
	Files       []File
	Fingerprint Fingerprint
}

// Build arma el paquete de un caso leyendo, bajo root, exactamente los archivos de include. Una
// entrada absoluta, con "..", repetida o que no sea un archivo es un error: la lista es explícita.
func Build(caseID, root string, include []string) (Package, error) {
	files := make([]File, 0, len(include))
	for _, inc := range include {
		rel, err := cleanRel(inc)
		if err != nil {
			return Package{}, err
		}
		full := filepath.Join(root, filepath.FromSlash(rel))
		info, err := os.Stat(full)
		if err != nil {
			return Package{}, fmt.Errorf("lista de inclusión: %s: %w", inc, err)
		}
		if !info.Mode().IsRegular() {
			return Package{}, fmt.Errorf("lista de inclusión: %s no es un archivo", inc)
		}
		data, err := os.ReadFile(full)
		if err != nil {
			return Package{}, err
		}
		files = append(files, File{Path: rel, Data: data})
	}
	return FromFiles(caseID, files)
}

// FromFiles arma un paquete con archivos ya en memoria (material derivado, como un caso de 06 sin
// sus campos de respuesta). Valida las rutas como Build.
func FromFiles(caseID string, files []File) (Package, error) {
	if len(files) == 0 {
		return Package{}, fmt.Errorf("lista de inclusión vacía para %s", caseID)
	}
	seen := map[string]bool{}
	out := make([]File, 0, len(files))
	for _, f := range files {
		rel, err := cleanRel(f.Path)
		if err != nil {
			return Package{}, err
		}
		if seen[rel] {
			return Package{}, fmt.Errorf("lista de inclusión: %s repetido", rel)
		}
		seen[rel] = true
		out = append(out, File{Path: rel, Data: f.Data})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return Package{Case: caseID, Files: out, Fingerprint: PackageFingerprint(out)}, nil
}

// cleanRel normaliza una entrada a ruta relativa con barras y rechaza las que salen de la raíz.
func cleanRel(p string) (string, error) {
	s := strings.ReplaceAll(p, "\\", "/")
	if s == "" || strings.HasPrefix(s, "/") || filepath.IsAbs(p) || (len(s) > 1 && s[1] == ':') {
		return "", fmt.Errorf("lista de inclusión: %q no es una ruta relativa", p)
	}
	for _, part := range strings.Split(s, "/") {
		if part == ".." {
			return "", fmt.Errorf("lista de inclusión: %q sale de la raíz", p)
		}
	}
	return path.Clean(s), nil
}

// PackageFingerprint usa la convención de TreeHash de 02 (eval/baseline/fixture): sha256 sobre las
// líneas "<sha256 del archivo>  <ruta>" ordenadas por ruta.
func PackageFingerprint(files []File) Fingerprint {
	sorted := append([]File(nil), files...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Path < sorted[j].Path })
	h := sha256.New()
	for _, f := range sorted {
		sum := sha256.Sum256(f.Data)
		fmt.Fprintf(h, "%s  %s\n", hex.EncodeToString(sum[:]), f.Path)
	}
	return Fingerprint{Algorithm: "sha256", Value: hex.EncodeToString(h.Sum(nil))}
}
