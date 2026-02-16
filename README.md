# AIlign CLI Specifications

**Instruction governance & distribution for engineering organizations**

One source of truth for AI coding assistant instructions across tools and repositories.

## Quick Navigation

- **[Vision](vision.md)** - Problem statement, solution approach, and business value
- **[Scope](scope.md)** - What's in/out for MVP, key decisions
- **[Constitution](constitution.md)** - Design principles and values
- **[Features](features/)** - Individual feature specifications

## What is AIlign CLI?

AIlign manages org-wide baselines and repo-specific overlays for AI assistant instructions, rendering them to different tool formats (GitHub Copilot, Cursor, Claude Code, Windsurf).

**Core concept:** `central baseline + repo overlay → rendered formats`

## Target Users

- **Primary:** Developers working in organizations with 20+ repositories
- **Secondary:** Security/Compliance teams enforcing standards

## MVP Features

1. **Package Management** - Versioned instruction packages (`company/security@1.3.0`)
2. **CLI Tool** - `init`, `pull`, `status`, `diff`, `explain` commands
3. **Multi-Tool Rendering** - Output to 2-3 AI tool formats
4. **Deterministic Composition** - Central baseline + local overlays

## Development Approach

This project uses [spec-kit](https://github.com/github/spec-kit) for specification-driven development:

1. Define features in `/features/[feature-name]/`
2. Each feature has: spec, tasks, implementation notes
3. Constitution provides guiding principles
4. Small, incremental changes based on feature boundaries

## Status

🚧 **In Design Phase** - Specifications being written