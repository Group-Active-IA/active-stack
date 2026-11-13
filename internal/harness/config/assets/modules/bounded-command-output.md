## Bounded Command Output

Clients cut long command output before you see it. Claude Code keeps about 10,000 characters of
a failed command, head and tail, and that tail is not the real end of the output. Codex keeps
head and tail and does not tell you the middle is gone. Never summarize from that view.

### When to summarize

- Output above 300 lines or above 10,000 characters, whichever comes first, is delivered as a
  summary plus a reference. Shorter output is delivered whole.
- The threshold decides when to summarize, never how much to cut. Never drop a failure line to
  save space.

### Capture before you read

- Capture every verification command (tests, build, vet, lint) and every command known to be
  verbose, before you run it. Do not guess its size first.
- If this skill names an independent capture source for your client and surface, use it and name
  it. Otherwise capture the output yourself with your shell's recipe; the capture source is then
  `agent`.
- If an independent capture already wrote a redacted copy and its capture record, summarize from
  that copy and cite it in `ref:`. Do not write a copy of your own.
- bash / sh:
  `cmd > "$f" 2>&1; echo "exit=$?" > "$f.exit"`
- PowerShell 5.1 and 7 (keeps the command's bytes; avoids UTF-16 and `NativeCommandError`):
  `cmd /c "<command> > ""$f"" 2>&1"; "exit=$LASTEXITCODE" | Set-Content -Encoding ascii "$f.exit"`
- Put `$f` under `<user cache dir>/active-stack/runs/<change>/<run-id>/`, or under
  `_sin-change` when no change is active. Never inside the repository. The copy expires seven
  days after the change is archived, or seven days after capture under `_sin-change`.
- Redact before you keep it. Replace every value that matches the sensitive list (`.env`
  assignments, tokens and API keys, cloud credentials, SSH keys and `*.pem`, environment variable
  values, home directory paths, account ids, emails and account UUIDs) with an omission mark,
  keeping the line in its place. Write the redacted copy to a second file, check that no listed
  value is left in it, and delete the raw file before you report. The copy you keep and cite is
  the redacted one.
- Never write a raw sensitive value: not in a file you keep, not in the summary, not in your
  reply, even when asked for the original. Quote the redacted line and say the value was omitted.
- Read the captured copy with directed reads (failure patterns, exit code). Do not summarize
  from what the client showed you.
- A command that already ran without capture and turned out verbose: if it has no side effects,
  run it again with capture; if it has side effects, do not repeat it and mark the summary
  partial.

### What the summary keeps

- Every error, every failed test and the exit code. A non-zero exit code is always a failure.
  One `failure:` line per captured line that matches a pattern, including the final
  `FAIL\t<package>` line.
- Failure patterns: `go test`: `--- FAIL`, `FAIL\t`, `panic:`; `go build` and `go vet`:
  `<file>:<line>:<column>:`. For a tool with no declared patterns, write
  `failures identified by exit code only`.
- Only failures the output records. Lines that only look like errors are not failures. With no
  failure, write exactly `failure: none`, with nothing after it.
- Each `failure:` line is the captured line verbatim, leading spaces included. Append nothing:
  no line number, no note.
- Repetition without failures is dropped.

### Fixed lines

The harness counts these lines. Write them exactly, one per line, at the top of the summary file
and in your reply, whatever the shell. Prose may follow them; it never replaces them.

```text
exit: <code>
failure: <line as captured>
ref: cmd=<command> run=<run-id> path=<redacted copy> sha256:<hex> expires=<YYYY-MM-DD>
capture: <harness | client:<source> | agent> · redactor: <same set> · producer: <who wrote this summary>
status: unverified — <what is missing>
partial — unverified: <reason>
```

Example, a failing test run captured by the agent:

```text
exit: 1
failure: --- FAIL: TestParse (0.00s)
ref: cmd=go test ./pkg/ run=20261113-1 path=<cache>/active-stack/runs/_sin-change/20261113-1/out.redacted.txt sha256:<hex> expires=2026-11-20
capture: agent · redactor: agent · producer: sdd-orchestrator
status: unverified — no independent capture
```

- No captured original: write `partial — unverified: <reason>` and no `ref:` line (no path, no
  hash, not even `ref: none`). Say which case it is: `never captured`, `copy missing` (the
  record names a file that is not there), `expired on <date>`, or `no persistent disk` (surface
  without the user's disk).
- A result that you or a worker declare stays `unverified` until another party corroborates
  the execution and the code state with independent capture or a re-run. A matching hash proves
  the copy is intact; it does not prove who ran the command.
- When you verify a declared result, give one state. `rejected`: the summary omits a failure the
  copy records, declares one it does not record, or its hash or exit code differs from the copy
  or the capture record. `verified`: none of that, and the copy comes from an independent capture
  (record source `arnés` or a client source, with coverage `demostrada`) or another party re-ran
  the command with capture. Otherwise `unverified`.
- If verifying needs a detail that redaction removed, the result stays `unverified`; say which
  detail is missing.
- Never cite the client's own copy (session transcript, client output file) as the reference:
  it is not redacted and Active Stack does not manage it. When asked for an original, or when
  that copy holds a sensitive value, name it as a limit.
