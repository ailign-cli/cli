---
name: gherkin-scenarios
description: 'Gherkin scenario and step writing patterns for AIlign CLI. Use when writing Given/When/Then steps, structuring scenarios, naming scenarios, or reviewing Gherkin quality. Enforces declarative style, one-behavior-per-scenario rule, mixed tense, and third-person perspective.'
---

# Gherkin Scenarios

Standards for writing Gherkin scenarios and steps in the AIlign CLI
project. These rules ensure scenarios are readable, maintainable, and
serve as living documentation.

## When This Skill Applies

- Writing Given/When/Then steps in `.feature` files
- Writing step definitions in `*_test.go` files
- Reviewing or refactoring existing scenarios
- Naming scenarios

## The Cardinal Rule: One Scenario, One Behavior

Every scenario MUST test exactly one behavior. This is non-negotiable.

**Test:** Show the scenario to someone unfamiliar with the feature.
They should immediately understand what single behavior is being
verified.

GOOD — one behavior per scenario:
```gherkin
Scenario: Valid configuration is loaded
  Given a repository contained a valid .ailign.yml with targets "claude,cursor"
  When the CLI parses the configuration
  Then the targets will be loaded successfully

Scenario: Missing configuration file
  Given a repository had no .ailign.yml file
  When the CLI attempts to load configuration
  Then it will report an error containing "not found"
```

BAD — multiple behaviors crammed into one:
```gherkin
Scenario: Config parsing and validation
  Given a .ailign.yml with targets "claude"
  When the CLI parses the configuration
  Then the targets are loaded
  When the user adds an invalid target "vscode"
  Then it reports a validation error
```

Multiple When-Then pairs signal multiple behaviors. Split them.

### Rare Exceptions to One-Behavior Rule

Multiple When-Then pairs are acceptable ONLY for:
- End-to-end audit scenarios in regulated domains
- Service call chains where responses feed subsequent requests

Even then, prefer splitting into separate scenarios first.

## Step Tense: Mixed (Past / Present / Future)

Each keyword maps to a temporal role:

| Keyword | Tense | Role | Example |
|---------|-------|------|---------|
| Given | Past | Setup that already happened | `Given the user logged in` |
| When | Present | Action happening now | `When the CLI parses the configuration` |
| Then | Future | Outcome that should follow | `Then the targets will be loaded successfully` |
| And/But | Same as preceding | Continues the preceding keyword | `And it will exit with code 2` |

### Given Steps — Past Tense

Givens establish context that exists before the action. Use past tense
to convey that the setup has already occurred.

```gherkin
# State descriptions (passive past)
Given a repository contained a valid .ailign.yml with targets "claude,cursor"
Given a .ailign.yml was configured with targets "vscode"
Given the user had logged in as an admin

# Completed setup actions
Given the CLI loaded the configuration
Given the developer created a .ailign.yml file
```

For pure state descriptions where past tense reads awkwardly, use
past participle or "had" constructions:

```gherkin
Given a repository had no .ailign.yml file
Given the configuration file had an empty targets array
```

### When Steps — Present Tense

Whens describe the action being performed right now.

```gherkin
When the CLI parses the configuration
When the developer runs ailign validate
When the CLI validates the configuration
When the user searches for "panda"
```

### Then Steps — Future Tense

Thens describe the expected outcome using "will" or "should".

```gherkin
Then the targets will be loaded successfully
Then it will report an error containing "not found" to stderr
Then it will exit with code 2
Then validation will succeed
Then results will be shown for "panda"
```

### And/But — Match the Preceding Keyword

```gherkin
Given a .ailign.yml was configured with targets "vscode"
  And the file contained an unknown field "custom_field"    # past (matches Given)
When the CLI validates the configuration                     # present
Then it will report a warning about "custom_field"           # future (matches Then)
  And validation will succeed                                # future (matches Then)
  But it will not report any errors                          # future (matches Then)
```

## Perspective: Third-Person Only

ALL steps use third-person perspective. Never use "I" or "my" in
steps (the Feature narrative uses first-person, but steps do not).

GOOD — third-person:
```gherkin
Given the developer created a .ailign.yml file
When the CLI parses the configuration
Then the system will report an error
```

BAD — first-person:
```gherkin
Given I have a .ailign.yml file
When I run the CLI
Then I see an error
```

**Why:** Third-person is generic, reusable, and unambiguous about who
the actor is. First-person creates confusion in multi-user scenarios
and limits step reusability.

Identify actors explicitly:
- `the developer`, `the user`, `the admin`
- `the CLI`, `the system`, `the validator`
- `the configuration file`, `the schema`

## Declarative Over Imperative

Describe WHAT happens, not HOW it happens. Implementation details
belong in step definitions, not in Gherkin.

GOOD — declarative:
```gherkin
Given the developer was logged in with valid credentials
When the developer creates a new project
Then the project will appear in the dashboard
```

BAD — imperative (UI mechanics):
```gherkin
Given the developer navigated to https://app.example.com/login
And the developer typed "user@example.com" in the email field
And the developer typed "password123" in the password field
And the developer clicked the "Login" button
When the developer clicked "New Project" in the sidebar
And the developer typed "My Project" in the name field
And the developer clicked "Create"
Then the developer will see "My Project" in the project list
```

**The test:** If the wording needs to change when the implementation
changes (UI redesign, API change), it is too imperative.

## Scenario Titles

Titles appear in test logs and reports. Write them well.

### Rules for Good Titles

1. **One concise line** describing the behavior
2. **No conjunctions** (and, or, but) — they signal multiple behaviors
3. **No assertion language** (verify, assert, should, check, test)
4. **No "because/since/so"** — titles describe WHAT, not WHY
5. **Describe the outcome or behavior**, not the steps

GOOD:
```gherkin
Scenario: Valid configuration is loaded
Scenario: Missing configuration file produces clear error
Scenario: Duplicate targets are rejected
Scenario: Unknown fields produce warnings
```

BAD:
```gherkin
Scenario: Verify the CLI can parse a config and validate targets     # conjunction + assertion
Scenario: Test that missing file returns error                        # assertion language
Scenario: The user creates a config because they need to set targets  # "because" = why
Scenario: Parse and validate                                          # too vague + conjunction
```

## Step Count

Keep scenarios short. Aim for 3-5 steps. Rarely exceed 7.

| Steps | Assessment |
|-------|------------|
| 1-3 | Ideal for focused unit behaviors |
| 4-5 | Good for typical scenarios |
| 6-7 | Acceptable for complex flows |
| 8+ | Too long — split or use Background |

If a scenario has many Given steps, extract shared ones into Background.
If it has many Then steps, consider whether you are testing multiple
behaviors.

## Concrete Test Data

Scenarios MUST use concrete, specific values — not vague descriptions.

GOOD — concrete:
```gherkin
Given a .ailign.yml was configured with targets "claude,cursor"
When the CLI validates the configuration
Then it will report an error at field path "targets[0]"
And the error will suggest valid targets "claude, cursor, copilot, windsurf"
```

BAD — vague:
```gherkin
Given a config file with some targets
When the CLI validates it
Then it reports an error with helpful information
```

**However**, do not over-specify brittle data:

BAD — brittle:
```gherkin
Then the error message will be "Error: targets[0]: invalid value 'vscode', expected one of: claude, cursor, copilot, windsurf"
```

GOOD — resilient:
```gherkin
Then it will report an error at field path "targets[0]"
And the error will suggest valid targets "claude, cursor, copilot, windsurf"
```

Check that the important parts are present without locking into exact
formatting.

## Conjunctive Steps — One Action Per Step

Each step does ONE thing. If a step contains "and" connecting two
separate actions, split it.

BAD:
```gherkin
Then it will report an error and exit with code 2
```

GOOD:
```gherkin
Then it will report an error containing "not found" to stderr
And it will exit with code 2
```

## Independent and Deterministic Scenarios

Scenarios MUST NOT depend on each other. No shared state between
scenarios. No assumed execution order. Every scenario sets up its
own world from scratch.

In our godog step definitions, this is enforced by the `Before` hook
that resets the `testWorld` struct before each scenario:

```go
ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
    w.dir, _ = os.MkdirTemp("", "ailign-bdd-*")
    w.cfg = nil
    w.loadErr = nil
    w.result = nil
    w.stdout = ""
    w.stderr = ""
    w.exitCode = -1
    return ctx, nil
})
```

## Step Definition Reusability

Write step definitions to be reusable across features. The step text
is a pattern — the same phrasing in different feature files reuses
the same step definition.

**Build a step definition library over time.** Common patterns:

```gherkin
# Reusable Given steps
Given a repository contained a valid .ailign.yml with targets "<targets>"
Given a .ailign.yml was configured with targets "<targets>"
Given a repository had no .ailign.yml file

# Reusable When steps
When the CLI parses the configuration
When the CLI validates the configuration
When the developer runs ailign <command>

# Reusable Then steps
Then it will exit with code <code>
Then it will report an error containing "<text>" to stderr
Then validation will succeed
```

Parameterize with regex capture groups. Keep step patterns general
enough to reuse but specific enough to be unambiguous.

## Writing Step Definitions in Go (godog)

Step definitions live in `features/*_test.go` files as methods on
a shared `testWorld` struct.

### Step Registration Pattern

```go
func registerSteps(ctx *godog.ScenarioContext) {
    w := &testWorld{}

    // Lifecycle hooks
    ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
        // Reset state before each scenario
        return ctx, nil
    })
    ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
        // Cleanup after each scenario
        return ctx, nil
    })

    // Register steps with regex patterns
    ctx.Given(`^a repository contained a valid \.ailign\.yml with targets "([^"]*)"$`, w.aRepoWithValidConfig)
    ctx.When(`^the CLI parses the configuration$`, w.theCLIParsesConfig)
    ctx.Then(`^the targets will be loaded successfully$`, w.targetsLoadedSuccessfully)
}
```

### Step Function Pattern

Step functions return `error`. Return `nil` for pass, `fmt.Errorf`
for fail. Do NOT use `t.Fatal` or `assert` — godog manages failures
through returned errors.

```go
func (w *testWorld) targetsLoadedSuccessfully() error {
    if w.loadErr != nil {
        return fmt.Errorf("expected no error but got: %v", w.loadErr)
    }
    if w.cfg == nil {
        return fmt.Errorf("expected config to be loaded but it is nil")
    }
    return nil
}
```

### Parameterized Steps

Use regex capture groups for parameters. godog converts matched
strings to the function parameter types.

```go
// String parameter
ctx.Then(`^it will report an error containing "([^"]*)"$`, w.itReportsErrorContaining)

func (w *testWorld) itReportsErrorContaining(substring string) error {
    if !strings.Contains(w.loadErr.Error(), substring) {
        return fmt.Errorf("expected error to contain %q, got: %v", substring, w.loadErr)
    }
    return nil
}

// Integer parameter
ctx.Then(`^it will exit with code (\d+)$`, w.itExitsWithCode)

func (w *testWorld) itExitsWithCode(code int) error {
    if w.exitCode != code {
        return fmt.Errorf("expected exit code %d, got %d", code, w.exitCode)
    }
    return nil
}
```

## Anti-Patterns to Avoid

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| Multiple When-Then pairs | Tests multiple behaviors | Split into separate scenarios |
| First-person steps ("I click...") | Ambiguous actor, poor reuse | Use third-person ("the user clicks...") |
| Imperative UI steps | Brittle, implementation-coupled | Use declarative behavior steps |
| Vague assertions ("it works") | Not testable | Use concrete expected values |
| 10+ steps per scenario | Too complex, hard to debug | Split or use Background |
| Conjunctions in step text | Multiple actions in one step | Split with And/But |
| Assertion language in titles | Testing mindset, not behavior | Describe the behavior |
| Shared state between scenarios | Order-dependent, flaky | Reset in Before hook |
| Scenario Outline with 20+ rows | Data-driven testing, not BDD | Keep 2-6 rows per table |
| Hard-coded exact error messages | Brittle to formatting changes | Assert key substrings |

## Checklist for New Scenarios

Before committing new Gherkin scenarios, verify:

- [ ] Each scenario covers exactly one behavior
- [ ] Title is a concise noun-phrase or behavior description
- [ ] Title has no conjunctions, assertion verbs, or "because"
- [ ] Given steps use past tense
- [ ] When steps use present tense
- [ ] Then steps use future tense ("will")
- [ ] All steps use third-person perspective
- [ ] Steps are declarative (no UI mechanics or implementation details)
- [ ] Concrete test data is used (not vague descriptions)
- [ ] Step count is 3-7 (not exceeding 7)
- [ ] Each step does one thing (no conjunctive "and" within a step)
- [ ] Scenario is independent (no dependency on other scenarios)
- [ ] Edge cases are tagged with `@edge-case`
- [ ] Step definitions return `error`, not using testify assertions
