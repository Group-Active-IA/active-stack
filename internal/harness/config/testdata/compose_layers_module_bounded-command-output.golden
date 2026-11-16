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
- bash / sh:
  `cmd > "$f" 2>&1; echo "exit=$?" > "$f.exit"`
- PowerShell 5.1 and 7 (keeps the command's bytes; avoids UTF-16 and `NativeCommandError`):
  `cmd /c "<command> > ""$f"" 2>&1"; "exit=$LASTEXITCODE" | Set-Content -Encoding ascii "$f.exit"`
- Put `$f` under `<user cache dir>/active-stack/runs/<change>/<run-id>/`, or under
  `_sin-change` when no change is active. Never inside the repository.
- Redact before you keep it. Replace every value that matches the sensitive list (`.env`
  assignments, tokens and API keys, cloud credentials, SSH keys and `*.pem`, environment variable
  values, home directory paths, account ids, emails and account UUIDs) with an omission mark,
  keeping the line in its place. Write the redacted copy to a second file and delete the raw file
  before you report.
- Read the captured copy with directed reads (failure patterns, exit code). Do not summarize
  from what the client showed you.
- A command that already ran without capture and turned out verbose: if it has no side effects,
  run it again with capture; if it has side effects, do not repeat it and mark the summary
  partial.

### What the summary keeps

- Every error, every failed test and the exit code. A non-zero exit code is always a failure.
- Failure patterns: `go test`: `--- FAIL`, `FAIL\t`, `panic:`; `go build` and `go vet`:
  `<file>:<line>:<column>:`. For a tool with no declared patterns, write
  `failures identified by exit code only`.
- Only failures the output records. Lines that only look like errors are not failures.
- Repetition without failures is dropped.

### Fixed lines

The harness counts these lines. Write them exactly:

```text
exit: <code>
failure: <line as captured>
ref: cmd=<command> run=<run-id> path=<redacted copy> sha256:<hex> expires=<YYYY-MM-DD>
capture: <harness | client:<source> | agent> · redactor: <same set> · producer: <who wrote this summary>
status: unverified — <what is missing>
partial — unverified: <reason>
```

- No captured original: write `partial — unverified: <reason>`, with no path and no hash. Say
  which case it is: `never captured`, `copy missing` (the record names a file that is not there),
  `expired on <date>`, or `no persistent disk` (surface without the user's disk).
- A result that you or a worker declare stays `unverified` until another party corroborates
  the execution and the code state with independent capture or a re-run. A matching hash proves
  the copy is intact; it does not prove who ran the command.
- If verifying needs a detail that redaction removed, the result stays `unverified`; say which
  detail is missing.
- Never cite the client's own copy (session transcript, client output file) as the reference:
  it is not redacted. Name it as a limit when it holds a sensitive value.
