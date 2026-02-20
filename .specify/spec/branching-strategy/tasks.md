# Tasks: Branching Strategy

**Plan**: `.specify/spec/branching-strategy/plan.md`
**Created**: 2026-02-20

## Phase 1: Script Changes

### 1.1 — Push feature branch to reserve number
- [ ] **T-1.1.1**: In `create-new-feature.sh`, after `git checkout -b "$BRANCH_NAME"` (line 275), add `git push -u origin "$BRANCH_NAME"` with error handling (warn on failure, don't exit)
- [ ] **T-1.1.2**: Guard the push behind `$HAS_GIT` check (matching the existing `git checkout -b` guard on line 274)
- [ ] **T-1.1.3**: Add warning message on push failure: `[specify] Warning: Could not push feature branch to remote. Push manually to reserve the feature number.`

**Files**: `.specify/scripts/bash/create-new-feature.sh`

### 1.2 — Create and checkout `/spec` sub-branch
- [ ] **T-1.2.1**: After pushing the feature branch, add `git checkout -b "${BRANCH_NAME}/spec"` and `git push -u origin "${BRANCH_NAME}/spec"` (also guarded by `$HAS_GIT`, warn on push failure)
- [ ] **T-1.2.2**: Update JSON output to include `SPEC_BRANCH`: `printf '{"BRANCH_NAME":"%s","SPEC_BRANCH":"%s/spec","SPEC_FILE":"%s","FEATURE_NUM":"%s"}\n'`
- [ ] **T-1.2.3**: Update plain-text output to include `SPEC_BRANCH: ${BRANCH_NAME}/spec`

**Files**: `.specify/scripts/bash/create-new-feature.sh`

---

## Phase 2: Speckit Prompt Changes

### 2.1 — Update `speckit.specify.md` branch creation flow
- [ ] **T-2.1.1**: Remove step 2e (the `git branch -m $BRANCH_NAME/spec` rename block, lines 62–66)
- [ ] **T-2.1.2**: Remove the duplicate branch-checking logic in steps 2a–2c (lines 39–55: `git fetch`, `git ls-remote`, `git branch`, specs directory scan) — this is already handled by `create-new-feature.sh`'s `check_existing_branches` function
- [ ] **T-2.1.3**: Simplify step 2 to: generate short name → call `create-new-feature.sh` with `--number` and `--short-name` → parse JSON output (which now includes `SPEC_BRANCH`)
- [ ] **T-2.1.4**: Update the explanation text to describe the new two-branch creation: the script creates the feature integration branch (`NNN-NAME`), pushes it to reserve the number, then creates and checks out the `/spec` sub-branch
- [ ] **T-2.1.5**: Update IMPORTANT notes: remove "only run this script once" caveat about branch renaming; add note that the script now pushes both branches to remote

**Files**: `.claude/commands/speckit.specify.md`

### 2.2 — Update `speckit.implement.md` phase branching
- [ ] **T-2.2.1**: Step 2 — Replace "Each implementation phase [...] gets its own branch off `main`" with "gets its own branch off the feature integration branch (`NNN-FEATURE`)"
- [ ] **T-2.2.2**: Step 2 — Change branch creation from `git checkout main && git pull` / `git checkout -b <feature>/<phase-slug>` to: `git checkout <feature> && git pull origin <feature>` / `git merge main` / `git checkout -b <feature>/<phase-slug>`
- [ ] **T-2.2.3**: Step 2 — Add explicit note that phase PRs target the feature integration branch (`NNN-FEATURE`), not `main`
- [ ] **T-2.2.4**: Step 2 — Add completion note: after all phases are merged into the feature branch, create a final PR from `NNN-FEATURE` → `main` to complete the feature

### 2.3 — Update `speckit.implement.md` feature detection
- [ ] **T-2.3.1**: Step 1 — Rewrite the starting-point detection to prioritize: (1) from `NNN-FEATURE`: auto-detect feature, create phase branch; (2) from `NNN-FEATURE/phase`: already on a phase branch, continue; (3) from `main`: detect feature from context (less common)

**Files**: `.claude/commands/speckit.implement.md`

---

## Phase 3: CI Workflow Changes

### 3.1 — PR-size conditional for feature integration PRs
- [ ] **T-3.1.1**: Add `if` condition to the `pr-size` job in `pull-request.yml` that skips the size check when the PR base is `main` AND the head branch has no `/` (feature integration branch pattern)
- [ ] **T-3.1.2**: Verify the `allow-large-pr` label exclusion remains as a manual fallback

**Files**: `.github/workflows/pull-request.yml`

---

## Phase 4: Release-Drafter Categories

### 4.1 — Add missing release-drafter categories
- [ ] **T-4.1.1**: Add "Specification" category with label `type:specification`
- [ ] **T-4.1.2**: Add "Testing" category with labels `type:bdd` and `type:test`
- [ ] **T-4.1.3**: Verify category ordering: version labels (Breaking Changes, Features, Bug Fixes) first, then type labels (Specification, Testing, Documentation, Build & CI)

**Files**: `.github/release-drafter.yml`

---

## Phase 5: Documentation

### 5.1 — CLAUDE.md branching strategy section
- [ ] **T-5.1.1**: Add "Branching Strategy" subsection under "Development Workflow" in `CLAUDE.md` describing hierarchical branches, PR targets, and linking to the spec

**Files**: `CLAUDE.md`

---

## Summary

| Phase | Tasks | Files |
|-------|-------|-------|
| 1. Script Changes | 6 | `create-new-feature.sh` |
| 2. Speckit Prompts | 9 | `speckit.specify.md`, `speckit.implement.md` |
| 3. CI Workflows | 2 | `pull-request.yml` |
| 4. Release-Drafter | 3 | `release-drafter.yml` |
| 5. Documentation | 1 | `CLAUDE.md` |
| **Total** | **21** | **6 files** |

All phases fit in a single PR (~80 lines estimated).
Phase 6 (speckit command validation) is deferred.
