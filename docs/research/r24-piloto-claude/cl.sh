#!/bin/bash
# uso: cl.sh <cwd> <salida.json> <prompt> [args extra de claude...]
cwd="$1"; out="$2"; prompt="$3"; shift 3
H='C:\r24b\home'
cd "$cwd" && env -i PATH="$PATH" SYSTEMROOT="$SYSTEMROOT" TEMP='C:\r24b\tmp' TMP='C:\r24b\tmp' HOME=$H USERPROFILE=$H LOCALAPPDATA="$H\AppData\Local" APPDATA="$H\AppData\Roaming" \
 "/c/Users/Juan Cruz Robledo/.local/bin/claude.exe" -p "$prompt" "$@" > "$out" 2> "$out.err" < /dev/null
