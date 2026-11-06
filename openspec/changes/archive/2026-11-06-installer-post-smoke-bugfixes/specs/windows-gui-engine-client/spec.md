## ADDED Requirements

### Requirement: Engine client isolates the GUI from child-process decoding and stream faults

`ProcessInstallerEngineClient` SHALL decode the engine child process's standard output and standard error as UTF-8 by setting `StandardOutputEncoding` and `StandardErrorEncoding` to UTF-8 on the `ProcessStartInfo` used to launch the engine, so accented and non-ASCII text (e.g. Spanish copy) is read without corruption. The progress-stream parser (`InstallProgressSnapshotParser.Parse`) SHALL tolerate a stdout line that is not valid JSON by treating it as an ignorable/log-type snapshot rather than throwing an uncaught exception, so a single malformed line cannot abort the whole stream. This tolerance applies to individual malformed lines only: a genuine non-zero engine exit MUST still be surfaced as a thrown failure (so the shell can present it as a terminal Error state), never silently swallowed.

#### Scenario: Child-process stdout and stderr are decoded as UTF-8

- **WHEN** the client builds the `ProcessStartInfo` to launch the engine
- **THEN** both `StandardOutputEncoding` and `StandardErrorEncoding` are UTF-8

#### Scenario: A non-JSON stdout line does not throw

- **WHEN** `InstallProgressSnapshotParser.Parse` receives a stdout line that is not valid JSON
- **THEN** it returns a snapshot (an ignorable/log-type entry carrying the raw line) instead of throwing an uncaught `JsonException`

#### Scenario: A non-zero engine exit is still surfaced

- **WHEN** the engine process exits with a non-zero code
- **THEN** the streaming/one-shot client still raises a failure (it is not swallowed), so the caller can present it as an error rather than the failure passing unnoticed
