#!/usr/bin/env bash
# Medición de la huella de bytes y del estado probado (tarea 1.3 del change opsx-verifiable-evidence).
# Uso: medir-huella.sh <directorio de trabajo> <core.autocrlf: true|false|input>
# Crea un repositorio descartable en <directorio de trabajo> (se borra antes de empezar).
# No toca ningún repositorio ni configuración del usuario: el autocrlf se fija solo en el repo descartable.
# Funciona en Git Bash (Windows) y en bash de Linux/WSL.
set -u
W="$1"
AC="$2"
rm -rf "$W"
mkdir -p "$W"
cd "$W" || exit 1

hash_of() { sha256sum --binary "$1" | cut -d' ' -f1; }

# Huella del estado alineada con TreeHash (eval/baseline/fixture/load.go): sha256 de las líneas
# "<sha256 del archivo>  <ruta>\n" ordenadas por ruta, sobre los bytes de la copia de trabajo
# (rastreados con cambios o sin ellos, y nuevos no ignorados; sin los archivos borrados).
state_treehash() {
  git ls-files -z --cached --others --exclude-standard | tr '\0' '\n' | LC_ALL=C sort -u |
    while IFS= read -r f; do
      [ -f "$f" ] && printf '%s  %s\n' "$(hash_of "$f")" "$f"
    done | sha256sum --binary | cut -d' ' -f1
}

# Alternativa: árbol de un índice temporal de Git (git add -A sobre una copia del índice).
state_index() {
  local idx
  idx="$(mktemp)"
  cp .git/index "$idx" 2>/dev/null || : >"$idx"
  GIT_INDEX_FILE="$idx" git add -A >/dev/null 2>&1
  GIT_INDEX_FILE="$idx" git write-tree
  rm -f "$idx"
}

row() { printf '%-46s treehash=%s index=%s\n' "$1" "$(state_treehash | cut -c1-16)" "$(state_index | cut -c1-16)"; }

LF='\n'
CRLF='\r\n'

echo "## entorno"
uname -srm
git --version
echo "sha256sum: $(sha256sum --version | head -1)"
echo "core.autocrlf (repo descartable) = $AC"
echo

mkdir repo && cd repo || exit 1
git init -q -b main
git config user.email t@example.invalid
git config user.name t
git config core.autocrlf "$AC"
printf "a${LF}b${LF}" >f.txt
printf "a${CRLF}b${CRLF}" >g.txt
git add -A 2>/dev/null
git commit -q -m base

echo "## M1. bytes de un mismo texto con LF y con CRLF (archivos aparte, fuera del repo)"
printf "linea${LF}" >../lf.txt
printf "linea${CRLF}" >../crlf.txt
echo "lf.txt   sha256sum --binary : $(sha256sum --binary ../lf.txt | cut -d' ' -f1)"
echo "lf.txt   sha256sum --text   : $(sha256sum --text ../lf.txt | cut -d' ' -f1)"
echo "crlf.txt sha256sum --binary : $(sha256sum --binary ../crlf.txt | cut -d' ' -f1)"
echo "crlf.txt sha256sum --text   : $(sha256sum --text ../crlf.txt | cut -d' ' -f1)"
echo "tamaños: lf=$(wc -c <../lf.txt) crlf=$(wc -c <../crlf.txt)"
echo "git hash-object (con filtros)    lf=$(git hash-object ../lf.txt) crlf=$(git hash-object ../crlf.txt)"
echo "git hash-object --no-filters     lf=$(git hash-object --no-filters ../lf.txt) crlf=$(git hash-object --no-filters ../crlf.txt)"
echo

echo "## M2. estado probado (primeros 16 hex): treehash = alineado con TreeHash; index = índice temporal de Git"
# Cada escenario parte de los mismos bytes en la copia de trabajo (se reescriben con printf, no con
# git checkout, que traduciría los fines de línea según core.autocrlf).
base() {
  git reset -q
  rm -f nuevo.txt ignorado.log .gitignore
  printf "a${LF}b${LF}" >f.txt
  printf "a${CRLF}b${CRLF}" >g.txt
}
base;                                                          row "S0 limpio (bytes del commit base)"
base; printf "a${LF}b${LF}c${LF}" >f.txt;                      row "S1 f.txt modificado sin preparar"
base; printf "a${CRLF}b${CRLF}c${CRLF}" >g.txt; git add g.txt 2>/dev/null; row "S2 mismo HEAD, g.txt modificado y preparado"
base; printf "nuevo${LF}" >nuevo.txt;                           row "S3 archivo nuevo sin trackear"
base; printf "a${CRLF}b${CRLF}" >f.txt;                         row "S4 f.txt: mismo texto pero CRLF"
base; rm g.txt;                                                row "S5 g.txt borrado"
base; echo "temporal" >ignorado.log; echo "*.log" >.gitignore; row "S6 archivo ignorado + .gitignore nuevo"
base; printf "a${LF}b${LF}" >g.txt;                             row "S7 g.txt con LF (el commit lo tenía con CRLF)"
echo
echo "## M3. estabilidad: dos corridas seguidas sin tocar nada"
base; row "S0 corrida 1"
row "S0 corrida 2"
