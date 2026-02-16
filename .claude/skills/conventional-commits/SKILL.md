---
name: conventional-commits
description: 'Format commit messages and PR titles following Conventional Commits specification for automated versioning and release notes. Use when asked about "commit message format", "PR title", "conventional commits", "semantic versioning", "commit types", or when creating commits and pull requests.'
---

# Conventional Commits

Format commit messages and pull request titles following the Conventional Commits specification for automated semantic versioning and release note generation.

## When to Use This Skill

- Creating commit messages
- Writing PR titles
- Asked about "commit format", "conventional commits", "semantic versioning"
- Need to determine commit type (feat, fix, chore, etc.)
- Determining version impact of changes
- Writing release notes

## Quick Reference

### Format

```
<type>[!]: <description>
```

**Components:**
- `<type>` - One of: feat, fix, chore (primary), or docs, style, refactor, perf, test, build, ci, revert
- `!` - Optional suffix for breaking changes
- `<description>` - Brief, lowercase, imperative mood, no period

### Version Impact

| Commit Type | Without `!` | With `!` |
|-------------|-------------|----------|
| `feat:` | Minor (`0.1.0` → `0.2.0`) | Major (`0.1.0` → `1.0.0`) |
| `fix:` | Patch (`0.1.0` → `0.1.1`) | Major (`0.1.0` → `1.0.0`) |
| `chore:` | Patch (`0.1.0` → `0.1.1`) | Major (`0.1.0` → `1.0.0`) |

**Breaking changes (`!`) ALWAYS trigger major version bump.**

## Commit Types

### Primary Types

**feat:** - New features or functionality
- Adds value to the module/product
- User-facing enhancements
- **Impact:** Minor version bump

**Examples:**
```
feat: add support for custom tags
feat: implement automatic scaling configuration
feat: add encryption at rest support
```

**fix:** - Bug fixes
- Corrects module behavior
- Resolves issues without adding features
- **Impact:** Patch version bump

**Examples:**
```
fix: correct validation logic for security group rules
fix: resolve resource naming conflict
fix: correct IAM policy document syntax
```

**chore:** - Maintenance and internal changes
- Tests, CI/CD, refactoring
- No functional impact on module
- Developer-facing changes
- **Impact:** Patch version bump

**Examples:**
```
chore: update pre-commit hooks
chore: refactor internal helper functions
chore: update terraform-docs to v0.18.0
```

### Secondary Types

**docs:** - Documentation only
```
docs: update README examples
docs: add migration guide
```

**style:** - Code style/formatting
```
style: format with terraform fmt
style: fix indentation
```

**refactor:** - Code refactoring (no behavior change)
```
refactor: extract common logic to module
refactor: simplify conditional expressions
```

**perf:** - Performance improvements
```
perf: optimize resource creation
perf: reduce API calls
```

**test:** - Adding or updating tests
```
test: add integration test for scaling
test: improve test coverage for edge cases
```

**build:** - Build system or dependency changes
```
build: update terraform provider versions
build: add dependency constraints
```

**ci:** - CI/CD configuration changes
```
ci: add automated testing workflow
ci: update pre-commit configuration
```

**revert:** - Reverting previous changes
```
revert: revert "feat: add experimental feature"
```

## Breaking Changes

**Any type can indicate a breaking change with `!`:**

```
feat!: remove deprecated variable
fix!: change default region to eu-west-1
chore!: require Terraform 1.6+
docs!: restructure module interface
```

**What constitutes a breaking change:**
- Removing or renaming variables
- Changing variable defaults that affect behavior
- Changing output structure
- Requiring newer Terraform/provider versions
- Removing deprecated features
- Changing resource naming patterns

## Description Guidelines

### Style Rules

✅ **DO:**
- Use lowercase
- Use imperative mood ("add" not "adds" or "added")
- Be concise (under 72 characters)
- Be specific about what changed
- Omit period at the end

❌ **DON'T:**
- Use uppercase (except acronyms)
- Add period at end
- Use past tense
- Be vague
- Explain why (use PR description for context)

### Good Examples

```
feat: add support for custom tags
fix: correct validation logic for security group rules
chore: update terraform-docs to v0.18.0
feat!: change variable naming convention
fix!: change default value for enable_monitoring
```

### Bad Examples

```
Add encryption                              # Missing type
feat:Add encryption                         # Missing space after colon
feat: Added encryption support.             # Past tense, has period
feat: ENCRYPTION SUPPORT                    # Uppercase
Fixed a bug                                 # Missing type, not imperative
chore: stuff                                # Too vague
feat: add feature because users requested   # Explaining why
```

## Workflow for Creating Commits

1. **Identify the type of change:**
   - New feature? → `feat:`
   - Bug fix? → `fix:`
   - Internal/maintenance? → `chore:`
   - Documentation only? → `docs:`

2. **Determine if breaking:**
   - Does it break existing users? → Add `!`

3. **Write description:**
   - Start with verb (add, fix, update, remove)
   - Keep lowercase
   - Be specific
   - Under 72 characters

4. **Format:**
   ```
   <type>[!]: <description>
   ```

5. **Verify:**
   - [ ] Type is correct
   - [ ] Breaking change marked with `!` if needed
   - [ ] Description is lowercase, imperative, no period
   - [ ] Description is specific and concise

## Workflow for PR Titles

Same format as commit messages:

1. **PR title MUST follow conventional commits**
2. **Type reflects the overall change:**
   - Multiple features → `feat:`
   - Multiple fixes → `fix:`
   - Mixed changes → Choose primary type

3. **Breaking changes in PR:**
   - If ANY commit is breaking → Add `!` to PR title

4. **Examples:**
   ```
   feat: add multi-region support
   fix: resolve intermittent connection issues
   chore: update dependencies and improve tests
   feat!: redesign authentication flow
   ```

## GitHub Integration

### Automatic Labeling

PR titles trigger automatic labels:

| Title Pattern | Version Label | Type Labels |
|--------------|---------------|-------------|
| `feat:` | `version: minor` | `feature` |
| `feat!:` | `version: major` | `feature` |
| `fix:` | `version: patch` | `bug` |
| `fix!:` | `version: major` | `bug` |
| `chore:` | `version: patch` | `chore`, `maintenance` |
| `chore!:` | `version: major` | `chore`, `maintenance` |

### Release Notes

Type prefix is automatically removed from release notes:

**PR Title:** `feat: add support for custom tags`
**In Release Notes:** `add support for custom tags`

## Quick Decision Matrix

| What Changed | Type | Breaking? | Version Impact |
|--------------|------|-----------|----------------|
| New feature | `feat:` | No | Minor (0.1.0 → 0.2.0) |
| New feature, breaks API | `feat!:` | Yes | Major (0.1.0 → 1.0.0) |
| Bug fix | `fix:` | No | Patch (0.1.0 → 0.1.1) |
| Bug fix, changes default | `fix!:` | Yes | Major (0.1.0 → 1.0.0) |
| Refactor, no behavior change | `chore:` | No | Patch (0.1.0 → 0.1.1) |
| Update Terraform version req | `chore!:` | Yes | Major (0.1.0 → 1.0.0) |
| Documentation only | `docs:` | No | No release |
| Test changes only | `test:` | No | No release |

## Common Scenarios

### Adding a new variable

```
feat: add custom_tags variable for user-defined tags
```

### Changing default value (breaking)

```
fix!: change default enable_monitoring to true
```

### Removing deprecated feature

```
feat!: remove deprecated enable_legacy_mode variable
```

### Updating dependencies

```
chore: update terraform-docs to v0.18.0
```

### Fixing validation logic

```
fix: correct validation for security group rules
```

### Refactoring without behavior change

```
chore: refactor internal helper functions
```

### Documentation improvements

```
docs: add examples for advanced configuration
```

### CI/CD updates

```
ci: add automated release workflow
```

## References

- [Conventional Commits Specification](https://www.conventionalcommits.org/en/v1.0.0/)
- [Semantic Versioning](https://semver.org/)
- Repository labeler: `.github/labeler.yml`
- Release drafter: `.github/release-drafter.yml`
