---
name: gherkin-feature-files
description: 'Gherkin feature file creation and organization for AIlign CLI. Use when creating .feature files, organizing BDD scenarios into features, or structuring the features/ directory. Enforces Constitution Principle IV (BDD+TDD dual-loop) with proper file structure, naming, and feature-to-user-story mapping.'
---

# Gherkin Feature Files

Standards for creating and organizing `.feature` files in the AIlign CLI project.
Enforces Constitution Principle IV (BDD+TDD dual-loop).

## When This Skill Applies

- Creating a new `.feature` file during `/speckit.plan`
- Reorganizing or splitting feature files
- Adding scenarios to an existing feature file
- Reviewing feature file structure

## Feature File Location and Naming

All feature files live at `features/` in the project root.
One feature file per user story from spec.md.

```
features/
├── parse-configuration-file.feature    # US1: config parsing
├── schema-validation.feature           # US2: schema validation
└── steps/                              # Go step definitions
    ├── suite_test.go                   # godog test runner (one per project)
    ├── world_test.go                   # shared testWorld struct and helpers
    ├── config_parsing_steps_test.go    # step definitions for config parsing
    └── schema_validation_steps_test.go # step definitions for schema validation
```

**Naming rules:**

- Use `kebab-case` derived from the user story title
- Name describes the behavior, not the implementation
- Consistent with branch naming (`001-config-parsing`) and tag naming (`@edge-case`)
- GOOD: `parse-configuration-file.feature`, `schema-validation.feature`
- BAD: `test_yaml_parser.feature`, `us1.feature`, `config.feature`

## Feature File Structure

Every feature file follows this exact structure:

```gherkin
# Source: specs/[###-feature-name]/spec.md - User Story N

Feature: [User Story Title — a noun phrase describing the capability]
  In order to [benefit/goal — why this matters]
  As a [actor/role — who needs this]
  I want to [action/capability — what they need]

  Background:
    Given [shared precondition across all scenarios, if any]

  Scenario: [Descriptive behavior title]
    Given [precondition in past tense]
    When [action in present tense]
    Then [expected outcome in future tense]

  @edge-case
  Scenario: [Edge case title]
    Given [precondition]
    When [action]
    Then [expected outcome]
```

### Required Elements

| Element | Required | Notes |
|---------|----------|-------|
| `# Source:` comment | Yes | Links back to spec.md user story |
| `Feature:` with narrative | Yes | Connextra format (In order to / As a / I want to) |
| `Background:` | Only if shared setup | Keep short (1-3 steps max) |
| At least one `Scenario:` | Yes | One behavior per scenario |

### The Source Comment

Every feature file MUST start with a source comment linking to its
user story. This maintains traceability between spec and executable
scenarios.

```gherkin
# Source: specs/001-config-parsing/spec.md - User Story 1
```

### The Feature Narrative (Connextra Format)

The narrative uses benefit-first ordering. It is NOT a description of
the feature — it answers WHY the feature exists.

```gherkin
Feature: Parse Configuration File
  In order to declare which AI tools my repository targets
  As a developer working in a repository
  I want to parse a .ailign.yml config file
```

- **In order to**: The business value or goal (comes FIRST)
- **As a**: The actor or role who benefits
- **I want to**: The capability they need

The narrative uses first-person ("I want to") because it represents
the actor's voice from the Connextra user story format. Steps within
scenarios use third-person (see gherkin-scenarios skill).

## Scope: What Goes in a Feature File

A feature is a collection of related scenarios that illustrate one
coherent capability. In this project, each feature maps to one user
story from spec.md.

### Feature scope rules

- One feature per file. Never combine multiple features.
- A feature groups scenarios for ONE user story.
- All scenarios in a feature should relate to the same capability.
- If a feature grows beyond ~10-12 scenarios, consider splitting
  along sub-capabilities.
- Keep the total file under ~100 lines. Shorter is better.

### What belongs together

GOOD — same feature (all about config parsing):
```
Scenario: Valid configuration is loaded
Scenario: Missing configuration file
Scenario: Empty configuration file
```

BAD — different features mixed together:
```
Scenario: Valid configuration is loaded      # parsing
Scenario: Invalid target name validation     # validation (different story)
Scenario: CLI renders to .cursorrules        # rendering (different feature entirely)
```

## Background Section

Use `Background:` to extract Given steps shared by ALL scenarios in
the feature. Keep it minimal.

```gherkin
Background:
  Given a repository with a valid .ailign.yml

Scenario: Parse targets
  When the CLI parsed the configuration
  Then ...

Scenario: Validate schema
  When the CLI validated the configuration
  Then ...
```

**Rules:**

- Only one Background per feature file
- Place it before the first Scenario
- Keep to 1-3 steps maximum — long backgrounds make scenarios hard
  to understand in isolation
- Only include steps that EVERY scenario truly shares
- If only some scenarios share setup, duplicate the Given steps in
  those scenarios instead

## Tags

Use tags to categorize scenarios for selective execution.

```gherkin
@edge-case
Scenario: Unicode BOM in config file
  ...

@slow
Scenario: Large config file with 100 targets
  ...
```

**Tag conventions:**

- Lowercase with hyphens: `@edge-case`, `@slow`, `@smoke`
- Tags on a Feature apply to ALL its scenarios automatically
- Common tags: `@edge-case`, `@slow`, `@smoke`, `@wip`
- Do NOT use tags to encode test data or parameters

## Scenario Outline (Examples Tables)

Use Scenario Outline when the same behavior applies to multiple
inputs with different expected outcomes.

```gherkin
Scenario Outline: Valid target names are accepted
  Given a .ailign.yml contained targets "<target>"
  When the CLI validates the configuration
  Then validation will succeed

  Examples:
    | target   |
    | claude   |
    | cursor   |
    | copilot  |
    | windsurf |
```

**Rules:**

- Use only when variations represent distinct equivalence classes
- Keep Examples tables small (2-6 rows, 1-3 columns)
- Don't use Scenario Outline to test implementation variations —
  only behavior variations
- Multiple Scenario Outlines per feature are fine if each covers
  a different behavior
- Do NOT create massive data tables — that is data-driven testing,
  not BDD

## Updating spec.md

When a feature file is created, the corresponding user story in
spec.md MUST reference it instead of containing inline scenarios:

```markdown
**Acceptance Scenarios**: See [`features/parse_configuration_file.feature`](../../features/parse_configuration_file.feature)

| Scenario | Description |
|----------|-------------|
| Valid configuration is loaded | Targets are parsed and accessible |
| Missing configuration file | Clear error on stderr, exit code 2 |
```

## Checklist for New Feature Files

Before committing a new `.feature` file, verify:

- [ ] File is in `features/` at project root
- [ ] Named with `kebab-case` matching the user story title
- [ ] Has `# Source:` comment linking to spec.md
- [ ] Has Feature narrative in Connextra format
- [ ] Each scenario covers exactly one behavior
- [ ] Scenarios use concrete test data, not vague descriptions
- [ ] Edge cases are tagged with `@edge-case`
- [ ] Corresponding spec.md updated with reference and summary table
- [ ] File is under ~100 lines / ~12 scenarios
- [ ] Background (if present) has 3 or fewer steps
