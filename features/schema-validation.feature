# Source: specs/001-config-parsing/spec.md - User Story 2
Feature: Schema Validation with Actionable Errors
  In order to catch mistakes before running ailign pull
  As a developer
  I want to validate my config file against the schema

  Scenario: Invalid target name
    Given a .ailign.yml was configured with targets "vscode"
    When the CLI validates the configuration
    Then it will report an error at field path "targets[0]"
    And the error will suggest valid targets "claude, cursor, copilot, windsurf"

  Scenario: Multiple errors reported at once
    Given a .ailign.yml was configured with targets "vscode,vscode"
    When the CLI validates the configuration
    Then it will report at least 2 errors
    And all errors will include field paths and remediation

  Scenario: Unknown fields produce warnings
    Given a .ailign.yml was configured with targets "claude" and unknown field "custom_field"
    When the CLI validates the configuration
    Then it will report a warning about "custom_field"
    And validation will succeed

  Scenario: Valid config via ailign validate
    Given a repository contained a valid .ailign.yml with targets "claude,cursor"
    When the developer runs ailign validate
    Then it will report success to stdout
    And it will exit with code 0

  Scenario: Invalid config via ailign validate
    Given a .ailign.yml was configured with no targets field
    When the developer runs ailign validate
    Then it will report errors to stderr
    And it will exit with code 2

  @edge-case
  Scenario: Unicode BOM in config file
    Given a .ailign.yml was configured with UTF-8 BOM and targets "claude"
    When the CLI parses the configuration
    Then the targets will be loaded successfully

  @edge-case
  Scenario: Duplicate targets
    Given a .ailign.yml was configured with targets "claude,claude"
    When the CLI validates the configuration
    Then it will report an error about duplicate targets
