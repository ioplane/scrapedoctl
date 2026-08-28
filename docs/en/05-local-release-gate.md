# Local Release Security Gate

This policy is blocking for a local release candidate. Semgrep and Snyk remain
local CLIs; their reports are stored under ignored `.ai/audits/p0/` and are not
uploaded by GitHub workflows.

## Fixed scanner policy

Verified on 2026-08-28 against the Homebrew delivery channel and upstream
release notes:

| Scanner | Exact version | Mode | Blocking threshold |
|---|---:|---|---|
| Semgrep | `1.175.0` | `scan --error --config auto` | Every finding |
| Snyk Open Source | `1.1307.0` | `test --all-projects` | High and critical |
| Snyk Code | `1.1307.0` | `code test` | High and critical |
| Snyk Container | `1.1307.0` | `container test --app-vulns` | High and critical |

Version sources: [Semgrep v1.175.0 release notes](https://github.com/semgrep/semgrep/releases/tag/v1.175.0)
and [Snyk CLI v1.1307.0 release notes](https://github.com/snyk/cli/releases/tag/v1.1307.0).
The gate rejects any other installed version; upgrades require updating this
policy and reading every intervening release note first.

## Required corpus evidence

A zero-finding result passes only when the scanner proves it received input:

- Semgrep JSON must report version `1.175.0`, a non-empty `paths.scanned`, and
  an empty `errors` list. `.git`, `.ai`, `.beads`, and `vendor` remain excluded.
- Snyk Open Source JSON must identify the project and report at least one
  dependency.
- Snyk Code must identify the repository path in its scan log, while the gate
  independently confirms at least one production Go source file. Snyk does not
  create JSON output for a zero-finding SAST scan.
- Snyk Container JSON must identify the requested tagged image through `path`
  and report its package manager. A zero dependency count is valid for a
  minimal image and is not used as an empty-corpus signal.

Missing, stale, empty, malformed, or mismatched evidence fails the gate.

## Exit policy

- Exit `0` passes only after version and corpus evidence validation.
- Semgrep exit `1` and Snyk exit `1` mean findings and fail the gate.
- Snyk exits `2`, `3`, `44`, `69`, `75`, `77`, every other non-zero exit, and
  failure to start a scanner fail the gate.
- Authentication or service unavailability never downgrades to a warning.

These meanings follow the [Semgrep scan exit-status reference](https://github.com/semgrep/semgrep-docs/blob/main/docs/snippets/cli-reference/help-scan.mdx)
and [Snyk CLI exit-code definitions](https://github.com/snyk/cli/blob/main/src/cli/exit-codes.ts).
