# Strict TDD Module — Apply Phase

> **This module is loaded ONLY when Strict TDD Mode is enabled AND a test runner is available.**
> If you are reading this, the orchestrator already verified both conditions. Follow every instruction.

## TDD Philosophy

TDD is not testing. TDD is **software design driven by tests**. You write a test that describes what the code SHOULD do, then write the minimum code to make it real. The tests design the API, the contracts, the behavior. Code is a side effect of tests.

### The Three Laws

1. **Do NOT write production code** until you have a failing test
2. **Do NOT write more test** than is necessary to fail
3. **Do NOT write more code** than is necessary to pass the test

## TDD Implementation Cycle

For EVERY task assigned to you, follow this cycle strictly:

```
FOR EACH TASK:
├── 0. SAFETY NET (only if modifying existing files)
│   ├── Run existing tests for files being modified
│   ├── Capture baseline: "{N} tests passing"
│   ├── If any FAIL → STOP, report as "pre-existing failure"
│   │   (do NOT fix pre-existing failures — report to orchestrator)
│   └── This baseline proves you did not break what already worked
│
├── 1. UNDERSTAND
│   ├── Read the task description
│   ├── Read relevant spec scenarios (these ARE your acceptance criteria)
│   ├── Read the design decisions (these CONSTRAIN your approach)
│   ├── Read existing code and test patterns (match the style)
│   └── Determine test layer (see "Choosing Test Layer" below)
│
├── 2. RED — Write a failing test FIRST
│   ├── Write test(s) that describe the expected behavior from the spec
│   ├── Prefer pure functions where possible (no side effects = easy to test)
│   ├── The test MUST reference production code that does NOT exist yet
│   └── GATE: Do NOT proceed to GREEN until the test is written
│
├── 3. GREEN — Write the MINIMUM code to pass
│   ├── Implement ONLY what the failing test needs
│   ├── Fake It is VALID here (hardcoded return values are OK)
│   ├── EXECUTE tests → must PASS
│   └── GATE: Do NOT proceed until GREEN is confirmed by execution
│
├── 4. TRIANGULATE (MANDATORY for most tasks)
│   ├── Add a second test case with DIFFERENT inputs/expected outputs
│   ├── EXECUTE tests → if Fake It breaks, generalize to real logic
│   ├── Repeat until ALL spec scenarios for this task are covered
│   ├── MINIMUM: at least 2 test cases per behavior (happy path + one edge case)
│   └── GATE: All spec scenarios for this task must have tests before REFACTOR
│
├── 5. REFACTOR — Improve without changing behavior
│   ├── Extract constants, functions; improve naming; remove duplication
│   ├── EXECUTE tests after EACH refactoring step → must STILL PASS
│   └── GATE: Tests green after EVERY refactoring change
│
├── 6. Mark task complete [x]
└── 7. Note any deviations or issues discovered
```

## Return Summary Extension

When Strict TDD Mode is active, your return summary MUST include:

```markdown
### TDD Cycle Evidence
| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| 1.1 | `path/test.ext` | Unit | ✅ 5/5 | ✅ Written | ✅ Passed | ✅ 3 cases | ✅ Clean |
```

## Rules

The cycle above is binding (its GATEs, the Three Laws, the Safety Net, the Evidence table). One rule it does NOT make obvious:

- NEVER write trivial assertions (tautologies, type-only checks, ghost loops) — they pass without exercising behavior.
