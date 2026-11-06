# Spec Delta

## Purpose

Defines how a cloned skill harness resolves its declared source reference (including an exact commit SHA) and which agents it is installed for, so pinned third-party skills install reproducibly and only where they apply.

## ADDED Requirements

### Requirement: A commit-SHA ref installs that exact commit

When a cloned skill harness declares a `ref` that is a full 40-character hexadecimal commit SHA, the installer SHALL install the skill content exactly as it exists at that commit, fetching only that commit (shallow) rather than the full history. The SHA MUST NOT be passed to git as a branch name. If the commit cannot be fetched, installation of that harness MUST fail with an error that names the harness, the repository and the ref; for a best-effort harness this failure MUST NOT abort the overall install.

#### Scenario: SHA ref is fetched by commit

- **WHEN** a skill harness with ref `b482f7a970abb98c4108d4a9f761e458c64cefc8` is installed
- **THEN** git is asked to fetch that commit with depth 1 and check it out detached, no `--branch` argument carries the SHA, and the skill directory is copied from the checked-out tree

#### Scenario: SHA fetch failure is reported

- **WHEN** fetching the pinned commit fails
- **THEN** the install of that harness returns an error naming the harness, the repository and the ref, and nothing is written to the agent's skills directory

### Requirement: Non-SHA refs keep the existing clone behaviour

For a `ref` that is a branch or tag name, the installer SHALL keep cloning with depth 1 and `--branch <ref>`; for an empty ref or `latest` it SHALL clone the default branch with depth 1 and no `--branch`. A ref that is hexadecimal but shorter than 40 characters MUST be treated as a branch or tag name.

#### Scenario: Tag ref

- **WHEN** a skill harness with ref `v1.2` is installed
- **THEN** git clone is invoked with `--depth 1` and `--branch v1.2`

#### Scenario: Latest ref

- **WHEN** a skill harness with ref `latest` is installed
- **THEN** git clone is invoked with `--depth 1` and without `--branch`

#### Scenario: Short hex ref

- **WHEN** a skill harness with ref `e8a175d` is installed
- **THEN** git clone is invoked with `--branch e8a175d` (no fetch-by-commit)

### Requirement: Skills install only for supported agents

The skill installer SHALL install a skill harness only into the skills directories of agents the harness supports. A harness with no declared agents MUST keep installing for every selected agent that has a skills directory. Agents the harness does not support MUST NOT get the skill directory, even when selected in the same install.

#### Scenario: Claude-only skill with Claude and OpenCode selected

- **WHEN** a skill harness that supports only `claude` is installed with agents `claude` and `opencode`
- **THEN** the skill is installed in Claude's skills directory and OpenCode's skills directory does not receive it

#### Scenario: Unscoped skill installs everywhere

- **WHEN** a skill harness with no declared agents is installed with agents `claude` and `opencode`
- **THEN** the skill is installed in both agents' skills directories
