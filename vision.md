# Vision

## Problem Statement

### Context Fragmentation Across AI Tools

Companies use multiple AI coding assistants simultaneously (Claude, Cursor, GitHub Copilot, Windsurf), each requiring its own format and location:

- **Cursor:** `.cursorrules` (8KB limit)
- **Claude:** `.claude/instructions.md` + `.claude/skills/`
- **Windsurf:** `.windsurfrules`
- **GitHub Copilot:** `.github/copilot-instructions.md`

### The Core Problems

**1. No Single Source of Truth**
- Company coding standards, security guidelines, and best practices get duplicated across dozens/hundreds of repositories
- Each repo has its own copy, leading to outdated or conflicting instructions

**2. Drift and Inconsistency**
- When standards update (e.g., new security policy), teams must manually update files in every repo
- No visibility into which instructions are actually being used

**3. Mixed Concerns**
- Repos need both company-wide standards AND project-specific context
- No good way to compose global + local instructions

**4. Specific Pain Points**
- Security team updates authentication standards → must update 50+ repos manually
- New developer joins → each repo has different/outdated instructions
- Cursor has 8KB limit → must manually prioritize what fits
- Skills with executable scripts can't be shared (each repo duplicates logic)

### Current "Solutions" Are Inadequate

- **Manual copy-paste** - Error-prone, forgotten
- **Git submodules** - Clunky, version conflicts
- **Documentation wikis** - Context not in AI tools
- **Acceptance** - Teams just live with fragmentation

## Our Solution

### One-Liner
**AIlign is instruction governance & distribution for engineering orgs: one source of truth for AI coding assistants across tools and repos.**

### How It Works

```
Central Registry (API)
  ├─ company/security@1.3.0
  ├─ company/typescript@2.1.0
  └─ team/platform@0.4.0
       ↓
  CLI: ailign pull
       ↓
  Composition (baseline + overlay)
       ↓
  Rendered formats
  ├─ .claude/instructions.md
  ├─ .cursorrules
  └─ .github/copilot-instructions.md
```

**Key capabilities:**
1. **Compose** - Global baseline + repo overlays
2. **Render** - One format → multiple tool outputs
3. **Track** - Version, audit, detect drift
4. **Update** - Rollout changes safely across organization

## Business Value

### Time Saved
- **2-5 hours/week per repo** - No more copying/updating AI instructions manually
- **Instant onboarding** - New developers get correct context immediately

### Quality Improvements
- **Consistency** - All teams use same coding standards, reducing review cycles
- **Up-to-date security** - Updates propagate instantly (weeks → hours)
- **Better AI output** - Proper context = fewer hallucinations, more relevant suggestions

### Risk Reduction
- **Compliance** - Ensure all repos use current security/compliance standards
- **Audit trail** - Track when/where standards deployed via API logs
- **No outdated practices** - Eliminate drift and stale instructions

### Measurable Proof Points
- Time-to-rollout for security baseline (hours vs. weeks)
- Percentage of repos up-to-date (drift = 0%)
- Reduction in convention/security review comments
- Developer time saved (less searching/manual syncing)

## Target Market

### Ideal Customer Profile (ICP)
- Engineering organizations with **20+ repositories**
- **Multiple AI tools** in use (Copilot + Cursor/Claude/Windsurf)
- **Security/compliance pressure** (standards, traceability requirements)

### User Personas

**Primary: Developers**
- Need: Consistent, current instructions without manual work
- Pain: Context fragmentation, outdated instructions, manual syncing

**Secondary: Security/Compliance Teams**
- Need: Enforce standards, track compliance, audit trail
- Pain: No visibility, slow rollout, drift

## Differentiation

### What Makes AIlign Unique

1. **Cross-tool parity** - One baseline → multiple formats (vs. tool-specific solutions)
2. **Composition model** - Global baseline + repo overlays without chaos
3. **Governance built-in** - Audit, drift detection, staged rollout, versioning
4. **Developer-first** - CLI workflow, no manual maintenance

### vs. Alternatives

| Approach | Limitation | AIlign Advantage |
|----------|-----------|-------------------|
| Manual copy-paste | Error-prone, forgotten | Automated, versioned |
| Git submodules | Version conflicts, clunky | Dedicated composition |
| Wiki/docs | Context not in tools | Rendered directly to tools |
| Per-tool solutions | Lock-in, duplication | Cross-tool from one source |

## Success Criteria

AIlign succeeds when:
1. **Adoption:** >80% of repos use AIlign within 6 months
2. **Drift elimination:** <5% repos with outdated instructions
3. **Time savings:** Security updates rolled out in <1 day (vs. weeks)
4. **Developer satisfaction:** NPS >40 from developers
5. **Governance:** Complete audit trail for compliance requirements