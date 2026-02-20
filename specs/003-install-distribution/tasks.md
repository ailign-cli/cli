# Tasks: Install & Distribution

**Input**: Design documents from `/specs/003-install-distribution/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, quickstart.md

**Tests**: BDD feature files exist at `features/install-*.feature`. Scenarios are tagged to indicate where they execute: untagged scenarios run locally via godog step definitions; `@ci` tagged scenarios run in CI/CD smoke test workflows post-release. The godog test runner will be configured in T005 to exclude `@ci` scenarios from local runs. All scenarios remain in the feature files as the single source of truth. Shell script validation uses shellcheck.

**Organization**: Tasks are organized by PR (from plan.md decomposition) to enable independent, deployable increments per distribution tier.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Verification)

**Purpose**: Verify existing release infrastructure is in place before extending it

- [x] T001 Verify existing GoReleaser configuration at .goreleaser.yml (builds, archives, checksum sections present)
- [x] T002 [P] Verify existing release workflow at .github/workflows/release.yml triggers on release publish
- [x] T003 [P] Verify version ldflags are set in .goreleaser.yml and cmd/ailign/main.go embeds version/commit
- [x] T004 [P] Validate GoReleaser config with `goreleaser check` (or `goreleaser build --snapshot --clean` for dry-run)
- [x] T005 Configure godog test runner in features/steps/suite_test.go to exclude `@ci` tagged scenarios from local test runs by updating the Tags expression to `~@wip && ~@ci` (godog uses `&&` to combine tag exclusions)

**Checkpoint**: Existing release infrastructure confirmed working, godog configured to exclude `@ci` — extension can begin

---

## Phase 2: Version + Install Script (PR 1) → US2, US3

**Goal**: Ensure `go install` works with correct version embedding, and create a universal install script for CI/CD and quick installs

**Independent Test**: Run `go install github.com/ailign/cli/cmd/ailign@latest` and verify `ailign --version` prints the release version. Run `install.sh` on macOS/Linux and verify binary is installed.

### BDD Scenarios

> **NOTE**: Tag integration scenarios (real `go install` from published module) with `@ci`. "Version output includes tag" runs locally via godog. Install script logic (custom dir, PATH warning, unsupported platform, checksum) runs locally; curl-based scenarios tagged `@ci`.

- [x] T006 [P] [US2] Tag integration scenarios in features/install-via-go.feature with `@ci` (Install latest, Install specific version), keep "Version output includes tag" untagged for local godog, remove `@wip` tag
- [x] T007 [P] [US3] Tag curl-based scenarios in features/install-via-binary.feature with `@ci` (Install script on macOS, Install script on Linux), keep locally-testable scenarios untagged (custom dir, specific version, checksum, PATH warning, unsupported platform), remove `@wip` tag
- [x] T008 [P] [US2] Write step definitions for version output scenario in features/steps/install_go_steps_test.go — build binary with test ldflags, verify `--version` output contains version string (expect RED)
- [x] T009 [US3] Write step definitions for install script scenarios in features/steps/install_binary_steps_test.go — test custom INSTALL_DIR, AILIGN_VERSION override, checksum verification, PATH warning, unsupported platform error by invoking install.sh with mocked environment (expect RED)

### Implementation

- [x] T010 [US2] Verify `go install` module path resolves correctly (check go.mod module path matches expected `github.com/ailign/cli`) and ensure version is reported correctly: GoReleaser-built binaries use ldflags, `go install`-built binaries fall back to Go module build info via `runtime/debug.ReadBuildInfo` when version is "dev"
- [x] T011 [US3] Create universal install script at install.sh with: OS/arch detection, GitHub API latest version fetch, AILIGN_VERSION env var override, correct archive download from GitHub Releases, checksum verification via checksums.txt, configurable install directory (INSTALL_DIR > ~/.local/bin > /usr/local/bin), PATH warning, unsupported platform error handling, edge cases (wrong platform archive, older version via different method)
- [x] T012 [US3] Validate install.sh with shellcheck (install shellcheck if needed, run `shellcheck install.sh`)
- [x] T013 [US3] Verify step definitions pass for locally-testable scenarios (GREEN) — run `go test ./features/steps/... -v` excluding `@ci` tagged scenarios

**Checkpoint**: `go install` path verified, install.sh ready for use, BDD scenarios GREEN. PR 1 can be merged independently.

---

## Phase 3: GoReleaser Tier 1 — Homebrew, Scoop, nFPM (PR 2) → US1, US3

**Goal**: Extend GoReleaser to publish to Homebrew tap, Scoop bucket, and generate deb/rpm/apk packages

**Independent Test**: Run `goreleaser check` to validate config. After release: `brew install ailign-cli/distribution/ailign` works, Scoop manifest is pushed to ailign-cli/distribution, nFPM packages appear as release assets.

### BDD Scenarios

> **NOTE**: All Homebrew scenarios require a real tap and `brew` command — tagged `@ci`, verified by post-release smoke test workflow.

- [x] T014 [P] [US1] Tag all scenarios in features/install-via-homebrew.feature with `@ci`, remove `@wip` tag — all scenarios verified by CI/CD smoke tests post-release

### Implementation

- [x] T015 [US1] Add Homebrew cask configuration (homebrew_casks section) to .goreleaser.yml — tap repo ailign-cli/distribution (directory: Casks/), cask name ailign, caveats for Gatekeeper workaround (brews is deprecated in favor of homebrew_casks; casks don't support depends_on or test blocks). Use `pull_request.enabled: true` in repository config for audit trail.
- [x] T016 [P] [US3] Add Scoop manifest configuration (scoops section) to .goreleaser.yml — bucket repo ailign-cli/distribution (root directory — subdirectories break `scoop bucket list`), project name, license. Use `pull_request.enabled: true` in repository config for audit trail.
- [x] T017 [P] [US3] Add nFPM configuration (nfpms section) to .goreleaser.yml — generate deb, rpm, apk packages with package name, description, maintainer, license
- [x] T018 Update release workflow at .github/workflows/release.yml to pass DISTRIBUTION_REPO_TOKEN as environment variable to GoReleaser step (single token for unified ailign-cli/distribution repo). Configure GoReleaser publishers with `skip_upload: auto` so missing tokens skip gracefully rather than failing the release.
- [x] T019 Validate extended GoReleaser config with `goreleaser check`

**Checkpoint**: GoReleaser Tier 1 configured. Next release publishes to Homebrew, Scoop, and generates Linux packages. PR 2 can be merged independently.

---

## Phase 4: GoReleaser Tier 2 — Nix, Docker, WinGet (PR 3) → US3

**Goal**: Extend GoReleaser to publish to Tier 2 channels that have credentials available (DISTRIBUTION_REPO_TOKEN or GITHUB_TOKEN)

**Independent Test**: Run `goreleaser check` to validate config. After release: packages published to configured Tier 2 channels.

### Implementation

- [x] T020 [P] [US3] Add Nix/NUR configuration (nix section) to .goreleaser.yml — NUR repo ailign-cli/distribution (directory: nix/), package name, homepage, license. Use `pull_request.enabled: true` in repository config for audit trail.
- [x] T021 [US3] Add Docker configuration (dockers section) to .goreleaser.yml — image ghcr.io/ailign-cli/ailign, tags (version + latest), build context. Create Dockerfile at repository root (FROM scratch or alpine, COPY binary, ENTRYPOINT)
- [x] T022 [P] [US3] Add WinGet configuration (winget section) to .goreleaser.yml — repo ailign-cli/distribution (directory: winget/), package identifier, publisher, short description, license. Use `pull_request.enabled: true` in repository config for audit trail.
- [x] T023 [US3] Update release workflow at .github/workflows/release.yml — add docker login step (ghcr.io). DISTRIBUTION_REPO_TOKEN already configured in T018 for NUR/WinGet.
- [x] T024 Validate extended GoReleaser config with `goreleaser check`

**Checkpoint**: GoReleaser Tier 2 configured. Nix, Docker, and WinGet channels ready. PR 3 can be merged independently.

---

## Phase 5: NPM Wrapper Package (PR 4) → US3

**Goal**: Create an NPM wrapper package that distributes the ailign binary through the npm ecosystem, following the platform-specific optionalDependencies pattern (esbuild/Biome/SWC pattern)

**Independent Test**: `npx @ailign/cli --version` works. `npm install -g @ailign/cli` installs the binary.

### Setup (done)

- [x] T025 [US3] Publish placeholder packages (v0.0.0) to npm for all 6 packages (@ailign/cli + 5 platform packages) to enable OIDC configuration
- [x] T026 [US3] Configure DISTRIBUTION_REPO_TOKEN (fine-grained PAT: Contents R/W, Pull requests R/W, Metadata RO) as GitHub Actions secret on ailign-cli/cli

### Implementation

- [x] T027 [US3] Update main wrapper package at npm/ailign/package.json — bump version, finalize bin entry, optionalDependencies for all 5 platform packages
- [x] T028 [US3] Create CLI shim at npm/ailign/bin/ailign — Node.js script that finds and executes the platform-specific binary
- [x] T029 [US3] Create platform detection module at npm/ailign/lib/platform.js — maps process.platform + process.arch to package name and binary path
- [x] T030 [US3] Create postinstall fallback downloader at npm/ailign/scripts/install.js — downloads binary from GitHub Releases if optionalDependencies failed (handles --ignore-scripts, custom registries)
- [x] T031 [P] [US3] Update platform-specific packages at npm/ailign-darwin-arm64/package.json, npm/ailign-darwin-x64/package.json, npm/ailign-linux-x64/package.json, npm/ailign-linux-arm64/package.json, npm/ailign-win32-x64/package.json — bump version, add placeholder for binary
- [x] T032 [US3] Add npm publish job to .github/workflows/release.yml — runs after GoReleaser job, copies binaries into platform packages, publishes all 6 packages to npm registry. <!-- TODO: add `environment: npm` for OIDC trust once npm OIDC is configured (requires security key) -->
- [x] T033 [US3] Validate npm packages with `npm pack --dry-run` for main and platform packages

**Checkpoint**: NPM wrapper package complete. After next release + npm publish: `npx @ailign/cli` works. PR 4 can be merged independently.

---

## Phase 6: README Documentation + CI/CD Smoke Tests (PR 5) → US4, All

**Goal**: Add comprehensive installation instructions to README.md and create CI/CD smoke test workflow for post-release integration verification

**Independent Test**: A first-time visitor can find and follow installation instructions within 60 seconds. All documented commands are correct. Smoke test workflow validates real installs after release.

### BDD Scenarios

> **NOTE**: All documentation scenarios are testable locally by parsing README.md — no `@ci` tag needed.

- [x] T034 [P] [US4] Remove `@wip` tag from features/install-documentation.feature — all scenarios run locally via godog (no `@ci` needed)
- [x] T035 [US4] Write step definitions for documentation scenarios in features/steps/install_docs_steps_test.go — parse README.md, verify Installation section exists before usage, verify all methods documented, verify code blocks present, verify version verification shown (expect RED)

### Implementation

- [x] T036 [US4] Add Installation section to README.md — position before any usage instructions, with sub-sections for: Homebrew, go install, install script, Scoop, NPM/npx, Docker, direct download, Linux packages (deb/rpm/apk)
- [x] T037 [US4] Add version verification instructions to README.md — show `ailign --version` expected output
- [x] T038 [US4] Verify all installation commands in README.md match actual package names, module paths, and URLs from GoReleaser config and npm packages
- [x] T039 [US4] Verify step definitions pass for documentation scenarios (GREEN)
- [x] T040 Create CI/CD post-release smoke test workflow at .github/workflows/smoke-test.yml — reusable workflow (`workflow_call`) called by release.yml after GoReleaser completes. Uses GitHub Actions matrix strategy: matrix of {os: [ubuntu-latest, macos-latest], method: [install-script, go-install, brew, npx, docker]} with exclude rules (brew only on macos, docker only on ubuntu). Each matrix entry: installs via that method, runs `ailign --version`, verifies output contains the release version. Maps to `@ci` tagged feature file scenarios.

**Checkpoint**: Documentation complete, CI/CD smoke tests ready. All installation methods documented and will be automatically verified on each release. PR 5 can be merged independently.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup across all PRs

- [x] T041 Run `goreleaser check` to validate final .goreleaser.yml configuration
- [x] T042 [P] Run full test suite `go test ./...` and verify all tests pass (unit + BDD)
- [x] T043 [P] Validate quickstart.md at specs/003-install-distribution/quickstart.md against implemented commands and package names
- [x] T044 [P] Verify all secrets are documented in research.md and referenced in release workflow
- [x] T045 Mark all tasks complete in specs/003-install-distribution/tasks.md

---

## Phase 8: DEFERRED — External Account Channels

**Status**: DEFERRED — requires external account creation. Pick up when accounts are available.

**Goal**: Enable distribution channels that require external service accounts (AUR, Chocolatey, Snapcraft) and configure npm OIDC.

> **NOTE**: These tasks are independent of each other. Each can be enabled individually by creating the account, adding the secret, and uncommenting the GoReleaser config block.

### External accounts

- [ ] T046 [US3] Create AUR account, generate SSH key (passphrase-free), add AUR_KEY secret to ailign-cli/cli. Add AUR configuration (aurs section) to .goreleaser.yml — package name, description, maintainer, git URL.
- [ ] T047 [P] [US3] Create Chocolatey account, generate API key, add CHOCOLATEY_API_KEY secret to ailign-cli/cli. Add Chocolatey configuration (chocolateys section) to .goreleaser.yml — package name, owners, title, project URL, license.
- [ ] T048 [P] [US3] Create Snapcraft/Ubuntu One account, run `snapcraft export-login`, add SNAPCRAFT_STORE_CREDENTIALS secret to ailign-cli/cli. Add Snapcraft configuration (snapcrafts section) to .goreleaser.yml — snap name, summary, description, grade, confinement, apps.

### npm OIDC

- [x] T049 [US3] Configure npm OIDC on all 6 @ailign packages (requires security key). Create GitHub Actions `npm` environment on ailign-cli/cli. Update T032 npm publish job to use `environment: npm` for OIDC trust. <!-- Done: OIDC trusted publishing configured, environment gate deferred -->

### Docker multi-arch

- [ ] T050 [US3] Add multi-arch Docker images (amd64 + arm64) — migrate from `dockers` to `dockers_v2` with buildx, add `docker_manifests` for unified tags. Include both GHCR and Docker Hub image templates.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — verification only
- **Version + Install Script (Phase 2)**: Depends on Setup — creates first PR
- **GoReleaser Tier 1 (Phase 3)**: Depends on Setup — can run in parallel with Phase 2
- **GoReleaser Tier 2 (Phase 4)**: Depends on Phase 3 (extends same .goreleaser.yml)
- **NPM Wrapper (Phase 5)**: Depends on Setup — can run in parallel with Phases 2-4 (separate files)
- **Documentation + Smoke Tests (Phase 6)**: Depends on Phases 2-5 (needs all methods implemented to document and test accurately)
- **Polish (Phase 7)**: Depends on all phases complete
- **DEFERRED (Phase 8)**: Independent — can be picked up any time after Phase 3

### User Story Dependencies

- **US1 (Homebrew)**: Phase 3 — depends on Phase 1 only
- **US2 (Go install)**: Phase 2 — depends on Phase 1 only
- **US3 (Pre-built binary)**: Phases 2-5 — install script (Phase 2), GoReleaser packages (Phases 3-4), NPM (Phase 5)
- **US4 (Documentation)**: Phase 6 — depends on all other stories

### Within Each Phase (PR)

1. BDD scenario verification/step definitions (if applicable) → RED
2. Implementation tasks in order
3. BDD step definitions pass → GREEN
4. Validation (goreleaser check / shellcheck / npm pack)
5. Phase checkpoint

### Parallel Opportunities

- Phase 2 (install script) and Phase 3 (GoReleaser Tier 1) can run in parallel
- Phase 5 (NPM wrapper) can run in parallel with Phases 2-4 (entirely separate files)
- Within Phase 3: Scoop (T016) and nFPM (T017) are parallel with each other
- Within Phase 4: Nix (T020) and WinGet (T022) are parallel with each other
- Within Phase 5: Platform-specific packages (T031) are parallel with each other
- Within Phase 8: All tasks are independent — enable channels as accounts are created

---

## Implementation Strategy

### MVP First (Phases 1-2 Only)

1. Complete Phase 1: Setup verification
2. Complete Phase 2: Version + install script
3. **STOP and VALIDATE**: `go install` works, install.sh works, BDD scenarios GREEN
4. Deploy/demo if ready — basic install path available

### Incremental Delivery

1. Phase 1 + Phase 2 → Install script + go install ready (PR 1)
2. Phase 3 → Homebrew + Scoop + Linux packages (PR 2)
3. Phase 4 → Nix + Docker + WinGet (PR 3)
4. Phase 5 → NPM ecosystem (PR 4)
5. Phase 6 → Documentation + CI/CD smoke tests (PR 5)
6. Each PR adds independently deployable distribution channels
7. Phase 8 → DEFERRED: AUR, Chocolatey, Snapcraft, npm OIDC (pick up when accounts/security key available)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- `@ci` tag on scenarios = executed by CI/CD smoke test workflow, excluded from local godog runs via `~@ci`
- No tag on scenarios = executed locally by godog step definitions
- CI/CD smoke test workflow (T040) uses GitHub Actions matrix strategy for multi-OS, multi-method verification
- DEFERRED channels can be enabled individually by creating the account, adding the secret, and uncommenting the GoReleaser config block
- Commit after each task or logical group
- Stop at any checkpoint to validate the PR independently
