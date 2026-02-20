@wip
# Source: specs/004-package-install/spec.md - User Story 3

Feature: Type-scoped registry paths
  In order to distinguish between different kinds of packages in the registry
  As a developer declaring dependencies in .ailign.yml
  I want to use type-prefixed package references that make the content type explicit

  Scenario: Instructions type prefix resolves
    Given the developer created a repository with .ailign.yml containing:
      """
      packages:
        - instructions/company/security@1.3.0
      targets:
        - claude
      """
    And the registry has package "instructions/company/security@1.3.0" with content:
      """
      # Security
      """
    When the developer runs "ailign install"
    Then the exit code will be 0
    And stdout will contain "instructions/company/security@1.3.0"

  Scenario: Unsupported type prefix rejected
    Given the developer created a repository with .ailign.yml containing:
      """
      packages:
        - mcp/company/tools@1.0.0
      targets:
        - claude
      """
    When the developer runs "ailign install"
    Then the exit code will be 2
    And stderr will contain "unsupported package type"
    And stderr will contain "mcp"
    And stderr will contain "instructions"

  Scenario: Missing type prefix rejected
    Given the developer created a repository with .ailign.yml containing:
      """
      packages:
        - company/security@1.3.0
      targets:
        - claude
      """
    When the developer runs "ailign install"
    Then the exit code will be 2
    And stderr will contain "missing type prefix"
    And stderr will contain "company/security@1.3.0"

  Scenario: Type mismatch between config and manifest
    Given the developer created a repository with .ailign.yml containing:
      """
      packages:
        - instructions/company/security@1.3.0
      targets:
        - claude
      """
    And the registry has package "instructions/company/security@1.3.0" with manifest type "mcp"
    When the developer runs "ailign install"
    Then the exit code will be 2
    And stderr will contain "type mismatch"
