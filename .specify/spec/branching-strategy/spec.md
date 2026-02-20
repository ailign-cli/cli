# Process Specification: Branching Strategy

**Created**: 2026-02-20
**Status**: Draft
**Type**: Process (development workflow, not CLI feature)

## Problem Statement

When multiple features are developed in parallel, phase branches
(spec, implementation phases) from different features can collide
on `main`. The current workflow merges phase PRs directly into
`main`, meaning incomplete feature work blocks or conflicts with
other feature work.

Additionally, the feature numbering scheme (`NNN-feature-name`)
is susceptible to race conditions when `speckit.specify` is called
for a new feature while an existing feature branch has not yet
merged to `main`.

## Solution: Hierarchical Branch Structure

### Branch Hierarchy

```
main
  └─ NNN-FEATURENAME                    (feature integration branch)
       ├─ NNN-FEATURENAME/spec          (specification phase)
       ├─ NNN-FEATURENAME/PHASENAME     (implementation phase 1)
       ├─ NNN-FEATURENAME/PHASENAME     (implementation phase 2)
       └─ NNN-FEATURENAME/PHASENAME     (implementation phase N)
```

### Branch Lifecycle

#### Feature Integration Branch (`NNN-FEATURENAME`)

- Created from `main` by `speckit.specify` (via
  `create-new-feature.sh`)
- MUST be pushed to remote immediately after creation to
  reserve the feature number
- Serves as the integration point for all phase branches
- Merges into `main` only when the entire feature is complete
  (all phases merged)
- SHOULD be kept up to date with `main` via periodic merge
  commits (not rebases) to preserve traceability (Constitution
  Principle VI: Governance as Foundation)

#### Specification Branch (`NNN-FEATURENAME/spec`)

- Created from `NNN-FEATURENAME`
- Used for: `speckit.specify`, `speckit.plan`, `speckit.tasks`,
  `speckit.analyze`, `speckit.clarify`, `speckit.checklist`
- PR merges into `NNN-FEATURENAME` (not `main`)
- Contains: `specs/NNN-FEATURENAME/spec.md`, `plan.md`,
  `tasks.md`, `research.md`, `data-model.md`, `contracts/`,
  feature files, and step definition scaffolds

#### Implementation Phase Branches (`NNN-FEATURENAME/PHASENAME`)

- Created from `NNN-FEATURENAME` after spec branch is merged
- Used for: `speckit.implement` (one branch per phase from
  `tasks.md`)
- PR merges into `NNN-FEATURENAME` (not `main`)
- Phase names correspond to phases in `tasks.md` (e.g.,
  `setup`, `foundational`, `user-story-1`, `user-story-2`,
  `polish`)
- Each phase branch MUST be independently buildable and
  testable (Constitution Principle IX: Working Software)

### Speckit Command Mapping

| Command | Branch | Merges Into |
|---------|--------|-------------|
| `speckit.specify` | Creates `NNN-FEATURENAME` + `NNN-FEATURENAME/spec` | — |
| `speckit.plan` | `NNN-FEATURENAME/spec` | — |
| `speckit.tasks` | `NNN-FEATURENAME/spec` | — |
| `speckit.analyze` | `NNN-FEATURENAME/spec` | — |
| `speckit.clarify` | `NNN-FEATURENAME/spec` | — |
| `speckit.checklist` | `NNN-FEATURENAME/spec` | — |
| Spec PR | `NNN-FEATURENAME/spec` | `NNN-FEATURENAME` |
| `speckit.implement` | `NNN-FEATURENAME/PHASENAME` | `NNN-FEATURENAME` |
| Feature PR | `NNN-FEATURENAME` | `main` |

### Feature Number Race Condition

**Problem**: When `speckit.specify` is called for a new feature
while `NNN-FEATURENAME` has not yet merged to `main`, the next
number might collide if only local `specs/` directories are
scanned.

**Risk control**: The `create-new-feature.sh` script MUST:

1. Run `git fetch --all --prune` before determining the next
   number
2. Scan both remote branches and local `specs/` directories
   for existing `NNN-*` patterns
3. Take `max(highest_remote_branch, highest_local_spec) + 1`
4. Push the feature branch immediately after creation to
   reserve the number

**Current state**: This risk control is already implemented in
`create-new-feature.sh` (`check_existing_branches` function).
The only missing piece is the immediate push after branch
creation.

### Keeping Feature Branches Current

Feature integration branches (`NNN-FEATURENAME`) that live for
multiple development cycles MUST be kept current with `main`:

- Use `git merge main` (merge commits), not `git rebase`
- Merge commits preserve history and traceability (Constitution
  Principle VI)
- The developer SHOULD merge `main` into the feature branch
  before creating each new phase branch
- CI SHOULD validate that the feature branch is not
  significantly behind `main`

### PR Flow

```
NNN-FEATURENAME/spec  ──PR──>  NNN-FEATURENAME
NNN-FEATURENAME/phase-1  ──PR──>  NNN-FEATURENAME
NNN-FEATURENAME/phase-2  ──PR──>  NNN-FEATURENAME
NNN-FEATURENAME  ──PR──>  main
```

Each phase PR stays within CI size limits (500/750 lines).
The final feature-to-main PR may be large but represents a
complete, tested feature.

## Requirements

### Functional Requirements

- **FR-001**: `create-new-feature.sh` MUST push the feature
  branch to remote immediately after creation
- **FR-002**: `check_feature_branch()` in `common.sh` MUST
  accept the hierarchical branch pattern:
  `NNN-name`, `NNN-name/spec`, `NNN-name/slug`
  (already implemented)
- **FR-003**: `find_feature_dir_by_prefix()` MUST resolve
  spec directories from any branch in the hierarchy
  (already implemented)
- **FR-004**: Speckit commands MUST validate they are on
  the correct branch type for their operation (e.g.,
  `speckit.plan` requires a `/spec` branch or base feature
  branch)
- **FR-005**: CI workflows MUST support PRs targeting
  feature branches (not just `main`)

### Process Requirements

- **PR-001**: Phase PRs MUST target the feature integration
  branch, not `main`
- **PR-002**: Only the feature integration branch MUST target
  `main`
- **PR-003**: The spec branch MUST be merged before
  implementation phase branches are created
- **PR-004**: Feature branches SHOULD be merged from `main`
  periodically to avoid divergence

## Changes Required

### Already Implemented

- `check_feature_branch()` accepts hierarchical names ✅
- `find_feature_dir_by_prefix()` strips `/suffix` from
  branch names ✅
- `check_existing_branches()` scans remote branches +
  local specs ✅

### Needs Implementation

1. **`create-new-feature.sh`**: Add `git push -u origin
   $BRANCH_NAME` after branch creation
2. **`create-new-feature.sh`**: Also create and checkout
   the `/spec` sub-branch after pushing the feature branch
3. **CI (`pull-request.yml`)**: Update to support PRs
   targeting feature branches (currently may only trigger
   for PRs to `main`)
4. **Speckit commands**: Add branch-type validation (e.g.,
   `speckit.implement` warns if on a `/spec` branch)
5. **CLAUDE.md / constitution**: Document the branching
   strategy in the development workflow section

## Success Criteria

- **SC-001**: Two features can be developed in parallel
  without branch collisions on `main`
- **SC-002**: Feature numbers are unique across all
  developers and worktrees
- **SC-003**: Each phase PR targets the correct feature
  branch
- **SC-004**: CI validates all PRs regardless of target
  branch
