#!/bin/bash
# sc.sh <cwd> <nombre> <prompt> [extra]
cwd=$1; n=$2; p=$3; shift 3
/c/r24b/cl.sh "$cwd" "C:/r24b/out/sc-$n.json" "$p" --model sonnet --output-format stream-json --verbose --setting-sources user,project --max-budget-usd 1.2 --no-session-persistence "$@"
python /c/r24b/show.py "C:/r24b/out/sc-$n.json"
