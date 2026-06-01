#!/usr/bin/env bash
# ============================================================================
#  build.sh - JR Stack
#  Cross-compila binarios estaticos (CGO desactivado) a dist/.
#  Uso:  ./build.sh        (chmod +x build.sh la primera vez)
#  Requiere: Go 1.26+ en el PATH.
#  Sirve en Linux, macOS, WSL y Termux.
# ============================================================================
set -euo pipefail
cd "$(dirname "$0")"

OUT=dist
mkdir -p "$OUT"
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
echo "Listo. Binarios en ${OUT}/:"
ls -lh "$OUT"
