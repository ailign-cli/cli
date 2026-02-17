@wip
# Source: specs/002-local-instruction-sync/spec.md - User Story 2

Feature: Preview Sync Changes
  In order to review changes before they are applied
  As a developer
  I want to preview what ailign sync would do without modifying any
  files

  Background:
    Given a repository with a valid .ailign.yml

  Scenario: Dry-run shows files that would be created
    Given a .ailign.yml with targets "claude,cursor" and overlay "base.md"
    And an overlay file "base.md" containing "Instructions"
    When the developer runs ailign sync with "--dry-run"
    Then stdout contains "would"
    And stdout contains ".claude/instructions.md"
    And stdout contains ".cursorrules"
    And it exits with code 0

  Scenario: Dry-run does not modify any files
    Given a .ailign.yml with targets "claude" and overlay "base.md"
    And an overlay file "base.md" containing "Instructions"
    When the developer runs ailign sync with "--dry-run"
    Then the hub file ".ailign/instructions.md" does not exist
    And no symlinks are created

  Scenario: Dry-run with no changes needed
    Given a .ailign.yml with targets "cursor" and overlay "base.md"
    And an overlay file "base.md" containing "Instructions"
    And ailign sync has been run previously
    When the developer runs ailign sync with "--dry-run"
    Then stdout contains "up to date"

  Scenario: Dry-run output in JSON format
    Given a .ailign.yml with targets "claude" and overlay "base.md"
    And an overlay file "base.md" containing "Instructions"
    When the developer runs ailign sync with "--dry-run --format json"
    Then stdout is valid JSON
    And the JSON output has "dry_run" set to true
