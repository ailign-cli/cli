---
description: Execute the implementation planning workflow using the plan template to generate design artifacts.
handoffs: 
  - label: Create Tasks
    agent: speckit.tasks
    prompt: Break the plan into tasks
    send: true
  - label: Create Checklist
    agent: speckit.checklist
    prompt: Create a checklist for the following domain...
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Outline

1. **Setup**: Run `.specify/scripts/bash/setup-plan.sh --json` from repo root and parse JSON for FEATURE_SPEC, IMPL_PLAN, SPECS_DIR, BRANCH. For single quotes in args like "I'm Groot", use escape syntax: e.g 'I'\''m Groot' (or double-quote if possible: "I'm Groot").

2. **Load context**: Read FEATURE_SPEC and `.specify/memory/constitution.md`. Load IMPL_PLAN template (already copied).

3. **Execute plan workflow**: Follow the structure in IMPL_PLAN template to:
   - Fill Technical Context (mark unknowns as "NEEDS CLARIFICATION")
   - Fill Constitution Check section from constitution
   - Evaluate gates (ERROR if violations unjustified)
   - Phase 0: Generate research.md (resolve all NEEDS CLARIFICATION)
   - Phase 1: Generate data-model.md, contracts/, quickstart.md
   - Phase 1: Update agent context by running the agent script
   - Re-evaluate Constitution Check post-design
   - Phase 2: Generate Gherkin feature files from spec user stories

4. **Stop and report**: Command ends after planning. Report branch, IMPL_PLAN path, generated artifacts, and generated feature files.

## Phases

### Phase 0: Outline & Research

1. **Extract unknowns from Technical Context** above:
   - For each NEEDS CLARIFICATION → research task
   - For each dependency → best practices task
   - For each integration → patterns task

2. **Generate and dispatch research agents**:

   ```text
   For each unknown in Technical Context:
     Task: "Research {unknown} for {feature context}"
   For each technology choice:
     Task: "Find best practices for {tech} in {domain}"
   ```

3. **Consolidate findings** in `research.md` using format:
   - Decision: [what was chosen]
   - Rationale: [why chosen]
   - Alternatives considered: [what else evaluated]

**Output**: research.md with all NEEDS CLARIFICATION resolved

### Phase 1: Design & Contracts

**Prerequisites:** `research.md` complete

1. **Extract entities from feature spec** → `data-model.md`:
   - Entity name, fields, relationships
   - Validation rules from requirements
   - State transitions if applicable

2. **Generate API contracts** from functional requirements:
   - For each user action → endpoint
   - Use standard REST/GraphQL patterns
   - Output OpenAPI/GraphQL schema to `/contracts/`

3. **Agent context update**:
   - Run `.specify/scripts/bash/update-agent-context.sh claude`
   - These scripts detect which AI agent is in use
   - Update the appropriate agent-specific context file
   - Add only new technology from current plan
   - Preserve manual additions between markers

**Output**: data-model.md, /contracts/*, quickstart.md, agent-specific file

### Phase 2: Generate Gherkin Feature Files

**Prerequisites:** Phase 1 complete (technical context available)

Feature files require technical context (data model, contracts, architecture) to write
concrete, executable scenarios. This is why they are generated here rather than during
`/speckit.specify`.

1. **For each user story in spec.md**, generate a `.feature` file at `features/[kebab-case-story-title].feature` (project root):
   ```gherkin
   # Source: specs/[###-feature-name]/spec.md - User Story N

   Feature: [User Story Title]
     In order to [benefit/goal from user story]
     As a [actor from user story]
     I want to [action from user story]

     Scenario: [Acceptance scenario 1 title]
       Given [concrete precondition with test data]
       When [concrete action]
       Then [concrete assertion]

     Scenario: [Acceptance scenario 2 title]
       Given [concrete precondition with test data]
       When [concrete action]
       Then [concrete assertion]
   ```

2. **Scenario quality requirements**:
   - Scenarios must use concrete test data (not vague prose)
   - Each scenario maps to one acceptance criterion from the user story summary table
   - Edge cases from the spec become additional scenarios (tagged @edge-case)
   - Leverage data model and contracts for realistic test data
   - Step phrasing should be reusable across scenarios where possible

3. **Update spec.md** acceptance scenario sections to reference the generated feature files

**Output**: features/*.feature, updated spec.md

## Key rules

- Use absolute paths
- ERROR on gate failures or unresolved clarifications
