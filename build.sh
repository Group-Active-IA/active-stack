#!/usr/bin/env bash
# ============================================================================
#  build.sh - JR Stack
#  Cross-compila binarios estaticos (CGO desactivado) a la raiz del repo.
#  Uso:  ./build.sh        (chmod +x build.sh la primera vez)
#  Requiere: Go 1.26+ en el PATH.
#  Sirve en Linux, macOS, WSL y Termux.
# ============================================================================
set -euo pipefail
cd "$(dirname "$0")"

OUT=.
export CGO_ENABLED=0

# build <goos> <goarch> [ext]
build() {
  local goos=$1 goarch=$2 ext=${3:-}
  echo "== ${goos} ${goarch} =="
  GOOS="$goos" GOARCH="$goarch" \
    go build -trimpath -o "${OUT}/jr-stack_${goos}_${goarch}${ext}" ./cmd/jr-stack
}

# Matriz de targets. Agrega lineas para mas plataformas, p.ej.:
#   build darwin arm64    # macOS Apple Silicon
#   build linux  arm64    # Raspberry / ARM
build windows amd64 .exe
build linux   amd64

echo
echo "Listo. Binarios en la raiz del repo:"
ls -lh jr-stack_*

# ----------------------------------------------------------------------------
# Acceso directo en el escritorio (best-effort, idempotente).
# Solo Linux por ahora: macOS no usa .desktop y la matriz no compila un binario
# darwin nativo; Termux/WSL no tienen escritorio estandar. Falla blanda: nunca
# rompe el build.
# ----------------------------------------------------------------------------
desktop_shortcut() {
  local os; os="$(uname -s 2>/dev/null || echo unknown)"
  case "$os" in
    Linux)
      local bin="$PWD/jr-stack_linux_amd64"
      local desk="${XDG_DESKTOP_DIR:-$HOME/Desktop}"
      [ -x "$bin" ]  || { echo "  (sin binario Linux nativo; omito acceso directo)"; return 0; }
      [ -d "$desk" ] || { echo "  (sin carpeta de escritorio en $desk; omito acceso directo)"; return 0; }
      local lnk="$desk/jr-stack.desktop"
      printf '%s\n' \
        '[Desktop Entry]' \
        'Type=Application' \
        'Name=JR Stack' \
        "Exec=$bin" \
        "Path=$PWD" \
        'Terminal=true' \
        'Categories=Development;' > "$lnk"
      chmod +x "$lnk"
      echo "  -> $lnk"
      ;;
    *)
      echo "  (acceso directo automatico solo en Linux/Windows por ahora; SO: $os)"
      ;;
  esac
}

echo
echo "== Acceso directo en el escritorio =="
desktop_shortcut || true
