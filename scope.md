# Scope

## MVP Goal

**Stop drift and make org-wide updates simple, explainable, and traceable.**

This document defines what's IN and OUT for the first version, based on delivering minimal but complete value.

## In Scope (MVP Features)

### 1. Package Management + Versioning

**What:**
- Versioned instruction packages (e.g., `company/security@1.3.0`, `team/platform@0.4.0`)
- Simple storage backend (S3 or similar)
- REST API for package retrieval

**Why:**
- Foundation for all other features
- Enables version locking and rollback

### 2. Repository Configuration

**What:**
- Config file (`.ailign.yml`) in each repo specifying:
  - Which packages to use
  - Which versions
  - Which target tools (Cursor, Claude, Copilot, etc.)
  - Local overlay files

**Example:**
```yaml
packages:
  - company/security@1.3.0
  - company/typescript@2.1.0
  - team/platform@0.4.0

targets:
  - claude
  - cursor

local_overlays:
  - .ai-instructions/project-context.md
```

**Why:**
- Declarative, version-controlled configuration
- Clear what each repo uses

### 3. CLI Tool

**Commands:**
- `ailign init` - Bootstrap config + generate initial output files
- `ailign pull` - Fetch packages + render to tool formats
- `ailign status` - Show installed versions + drift detection
- `ailign diff` - Preview changes before update
- `ailign explain` - Show origin of each instruction section

**Distribution:**
- Single binary (no runtime dependencies)
- Fast, predictable, no interactive prompts (for automation)

**Why:**
- Developer workflow integration
- Builds trust through transparency (diff, explain)
- Enables CI/CD integration

### 4. Renderers (Start with 2-3 Tools)

**Targets for v1:**
- GitHub Copilot (`.github/copilot-instructions.md`)
- Claude Code (`.claude/instructions.md`)
- OR Cursor (`.cursorrules`) instead of one above

**Rendering logic:**
- Deterministic composition (no "intelligent merge" yet)
- Size-aware output (respect tool limits like Cursor's 8KB)
- Content tiers: `critical`, `recommended`, `extra`

**Why:**
- Proves cross-tool value immediately
- Two tools = sufficient for validation
- Deterministic = predictable, debuggable

### 5. Merge Strategy (Simple)

**Rules:**
- Central packages = **baseline**
- Local overlays = **additions**
- Hard precedence rules (no ambiguity):
  1. Critical content from central always included
  2. Repo-specific overlays appended
  3. Size budget enforced per tool

**Why:**
- Predictable behavior = trust
- Avoids "support hell" of unclear merge logic
- Can be explained with `ailign explain`

### 6. Content Tiers + Size Budgets

**Tiers:**
- `critical` - Security, compliance (always included)
- `recommended` - Best practices, conventions
- `extra` - Nice-to-have context

**Size budgets:**
- Each renderer knows tool limits (e.g., Cursor 8KB)
- Fills budget in tier order
- Warns if content truncated

**Why:**
- Respects tool constraints
- Gives control over priority
- Prevents silent failures

## Out of Scope (Explicitly NOT in v1)

### 1. Intelligent Merge
**What:** AI-powered conflict resolution, semantic merging
**Why not:** Adds complexity, hard to debug, can do later
**Instead:** Deterministic composition with clear precedence

### 2. Background Auto-Updates
**What:** Automatic syncing without developer action
**Why not:** Risky (unexpected changes), reduces trust
**Instead:** Optional CI drift-check + manual `pull`

### 3. Executable Skills/Scripts
**What:** Downloading and running executable code
**Why not:** Supply-chain security risk
**Instead:** Text-only instructions for v1, scripts later with signing

### 4. RBAC and Advanced Audit
**What:** Role-based access control, detailed audit logs
**Why not:** Overkill for MVP
**Instead:** Basic API logging, add RBAC post-validation

### 5. Staged Rollouts
**What:** Gradual rollout to subset of repos
**Why not:** Useful but not critical for initial adoption
**Instead:** All-at-once updates, add staging later

### 6. Web UI
**What:** Browser-based management interface
**Why not:** CLI-first for developer workflow
**Instead:** Pure CLI for v1, UI later if needed

### 7. Multi-Registry Support
**What:** Federated registries, private + public packages
**Why not:** Premature, adds architectural complexity
**Instead:** Single org registry, one source of truth

## Key Decisions

### Decision 1: Deterministic over Intelligent

**Choice:** Simple, rule-based composition instead of AI/heuristic merging

**Rationale:**
- Predictable = debuggable = trust
- Can always add intelligence later
- "What wins" must be instantly explainable

### Decision 2: CLI-first, Automation-friendly

**Choice:** No interactive prompts, designed for CI/CD

**Rationale:**
- Adoption dies if manual friction exists
- Git hooks + CI checks drive consistency
- Fast, silent execution enables automation

### Decision 3: Text-only for v1

**Choice:** No executable skills/scripts initially

**Rationale:**
- Security risk without proper signing/validation
- Text instructions cover 80% of use cases
- Can add scripts later with checksums/allowlisting

### Decision 4: Start with 2 Tools

**Choice:** Render to 2-3 AI tools, not all 4+

**Rationale:**
- Proves cross-tool value
- Limits initial renderer complexity
- Can add more tools incrementally

## Success Metrics for MVP

### Adoption
- **Target:** 50% of repos using AI Sync within 3 months
- **Measure:** Number of repos with `.ailign.yml`

### Drift Reduction
- **Target:** <10% of repos with drift after 1 month
- **Measure:** `ailign status` showing "outdated" count

### Time to Rollout
- **Target:** Security update to all repos in <4 hours
- **Measure:** Time from package publish to all repos updated

### Developer Experience
- **Target:** <30 seconds for `ailign pull` command
- **Target:** Zero manual conflicts requiring human resolution
- **Measure:** CLI performance metrics, support tickets

## Post-MVP Roadmap

### Phase 2 (after validation)
- Executable skills with signing
- RBAC and enhanced audit
- Staged rollout capabilities
- Additional tool renderers

### Phase 3 (if needed)
- Intelligent merge heuristics
- Web UI for package management
- Analytics dashboard (usage, adoption, drift)
- Multi-registry federation

## Risk Mitigation

### Risk 1: Adoption Friction
**Mitigation:** Make `ailign pull` frictionless, add optional CI drift-check

### Risk 2: Merge Conflicts
**Mitigation:** Hard precedence rules + excellent `diff`/`explain` tooling

### Risk 3: Tool Limits
**Mitigation:** Content tiers + size budgets per renderer

### Risk 4: Supply Chain
**Mitigation:** Text-only for v1, defer scripts until signing ready