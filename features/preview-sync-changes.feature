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
    Then stdout will contain "would"
    And stdout will contain ".claude/instructions.md"
    And stdout will contain ".cursorrules"
    And it will exit with code 0

  Scenario: Dry-run does not modify any files
    Given a .ailign.yml with targets "claude" and overlay "base.md"
    And an overlay file "base.md" containing "Instructions"
    When the developer runs ailign sync with "--dry-run"
    Then the hub file ".ailign/instructions.md" will not exist
    And no symlinks will be created

  Scenario: Dry-run with no changes needed
    Given a .ailign.yml with targets "cursor" and overlay "base.md"
    And an overlay file "base.md" containing "Instructions"
    And ailign sync has been run previously
    When the developer runs ailign sync with "--dry-run"
    Then stdout will contain "up to date"

  Scenario: Dry-run output in JSON format
    Given a .ailign.yml with targets "claude" and overlay "base.md"
    And an overlay file "base.md" containing "Instructions"
    When the developer runs ailign sync with "--dry-run --format json"
    Then stdout will be valid JSON
    And the JSON output will have "dry_run" set to true
