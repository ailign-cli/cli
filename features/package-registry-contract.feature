@wip
# Source: specs/004-package-install/spec.md - User Story 5

Feature: Registry contract for testing
  In order to develop the CLI and registry independently
  As a developer of either the CLI or the registry
  I want a well-defined contract between CLI and registry that both sides test against

  Scenario: Fetch package by type, name, and version
    Given the stub registry is running
    And the stub registry has package "instructions/company/security@1.3.0"
    When the CLI fetches "instructions/company/security@1.3.0" from the registry
    Then the response will contain a valid manifest
    And the manifest name will be "company/security"
    And the manifest type will be "instructions"
    And the manifest version will be "1.3.0"
    And the response will include a content URL
    And the response will include an integrity checksum

  Scenario: Package not found
    Given the stub registry is running
    And the stub registry has no package "instructions/company/nonexistent@1.0.0"
    When the CLI fetches "instructions/company/nonexistent@1.0.0" from the registry
    Then the response will be a 404 error
    And the error will contain "package_not_found"

  Scenario: Version not found
    Given the stub registry is running
    And the stub registry has package "instructions/company/security@1.3.0"
    When the CLI fetches "instructions/company/security@9.9.9" from the registry
    Then the response will be a 404 error
    And the error will contain "version_not_found"
    And the error will suggest available versions

  @edge-case
  Scenario: Registry unreachable
    Given the registry URL points to an unreachable host
    When the CLI fetches "instructions/company/security@1.3.0" from the registry
    Then the error will indicate a connection failure
    And the error will suggest checking network connectivity

  @edge-case
  Scenario: Registry returns invalid manifest
    Given the stub registry is running
    And the stub registry returns an invalid manifest for "instructions/company/broken@1.0.0"
    When the CLI fetches "instructions/company/broken@1.0.0" from the registry
    Then the error will indicate an invalid manifest
    And the error will contain the specific validation failure
