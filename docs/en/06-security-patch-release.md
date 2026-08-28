# v0.2.2 Security Patch Release Checklist

## Status and user impact

`v0.2.1` is affected by `SEC-001`: its Scrape.do client could send the API
token and target URL over plaintext HTTP. The planned fixed version is
`v0.2.2`. It is not released and must not be described as fixed until every P0
release blocker is closed and the checklist below passes on the tagged commit.

Users of `v0.2.1` should stop using active tokens with that version. After
installing `v0.2.2`, revoke and rotate the Scrape.do token and any other
provider token passed through `v0.2.1` command arguments, generated agent
configuration, or debug output. Confirm that the revoked token is rejected and
that the replacement succeeds only over HTTPS.

## Current remediation evidence

- Commit `50d057f` introduced HTTPS-only Scrape.do transport; the focused
  regression is `TestClient_HTTPSPolicy` in `pkg/scrapedo/security_test.go`.
- Commit `23c1132` applied the transport policy to built-in providers.
- Commits `c6f44c7` and `1a85d5f` removed command-line token values and added a
  cross-surface secret-leak regression.
- Commit `bca7612` made `go run ./cmd/devtool audit` reject scanner version
  drift and empty or mismatched evidence.

These commits are implementation evidence, not release evidence. Repeat every
check below against the final candidate.

## Blocking release checks

- [ ] `bd list --priority 0 --status open` reports no unfinished P0 child.
- [ ] `bd list --priority 0 --status in_progress` reports no unfinished P0
  child.
- [ ] `go run ./cmd/devtool release-gate` passes Testcontainers-Go verification,
  lint, and all four pinned scanner modes on Podman while preserving
  `.ai/audits/p0/versions.json`, scanner logs, and JSON evidence.
- [ ] The release commit is clean, contains this advisory, and is the exact
  commit selected for `v0.2.2`.
- [ ] Release artifacts report version `0.2.2`; `devtool release-manifest`
  verifies every archive/package checksum and its SPDX SBOM before the draft
  release is published.
- [ ] The published release notes identify affected `v0.2.1`, fixed `v0.2.2`,
  token rotation, and the verification evidence above.

Do not create or publish the `v0.2.2` tag while any item remains unchecked.

## Pinned release inputs

The release workflow uses `ubuntu-24.04`, Go `1.27.0`, GoReleaser `v2.18.0`,
and `goreleaser-action` commit
`f06c13b6b1a9625abc9e6e439d9c05a8f2190e94`. The Go-only
`devtool release-tools` command verifies these upstream archives before
extraction:

- UPX `5.2.1`:
  `402162aad30af47e60dbd767fb2e64ca394ace9727ba1f40283641f1d1b91657`.
- PowerShell `7.6.5`:
  `b34ab3b19acac1d3d4d0d3cfdb02acf62f457b0b6a962ff008132033f7566844`.
- Syft `1.51.1`:
  `8fcb33017a0dc1058298c923c436d19dfa68ae93968e0b423248542e3afb9fc3`.

GoReleaser stages assets as a draft. The workflow publishes that draft only
after `dist/checksums.txt` maps each archive and package to a checksum-verified,
metadata-complete SPDX 2.x document.

The local Testcontainers toolchain remains the Go `1.27.0-trixie` arm64 image
at digest
`sha256:6c83d163c89d1dbe66f8fc466870d66fe6325d15bb5ad1730cdc0b482fab1eec`.
