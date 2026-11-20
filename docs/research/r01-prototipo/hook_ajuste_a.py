"""Ajuste A del piloto. La selección de campos se evalúa en el runtime."""
import json
import subprocess
import sys
from hook import rewrite


if __name__ == "__main__":
    try:
        result = rewrite(json.load(sys.stdin))
        if result:
            output = result["hookSpecificOutput"]
            fields = output["updatedInput"]
            output["updatedInput"] = {key: fields[key] for key in ("title", "topic_key", "content") if key in fields}
            print(json.dumps(result))
    except (OSError, ValueError, KeyError, TypeError, subprocess.SubprocessError) as error:
        print(f"Prototipo sin reescritura ({type(error).__name__})", file=sys.stderr)
