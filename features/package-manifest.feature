@wip
# Source: specs/004-package-install/spec.md - User Story 2

Feature: Package manifest defines package identity
  In order to publish and discover instruction packages with clear metadata
  As a package author
  I want to define my package's identity, type, version, and content in a manifest file

  Scenario: Valid instruction package manifest
    Given a package manifest "ailign-pkg.yml" containing:
      """
      name: "company/security"
      type: "instructions"
      version: "1.3.0"
      description: "Company-wide security instructions for AI coding assistants"
      content:
        main: "instructions.md"
      """
    When the manifest is validated
    Then the manifest will be accepted
    And the parsed name will be "company/security"
    And the parsed type will be "instructions"
    And the parsed version will be "1.3.0"

  Scenario: Manifest missing required field
    Given a package manifest "ailign-pkg.yml" containing:
      """
      name: "company/security"
      type: "instructions"
      version: "1.3.0"
      content:
        main: "instructions.md"
      """
    When the manifest is validated
    Then the manifest will be rejected
    And the error will contain "description"
    And the error will contain "required"

  Scenario: Manifest with invalid version format
    Given a package manifest "ailign-pkg.yml" containing:
      """
      name: "company/security"
      type: "instructions"
      version: "1.3"
      description: "Security instructions"
      content:
        main: "instructions.md"
      """
    When the manifest is validated
    Then the manifest will be rejected
    And the error will contain "version"
    And the error will contain "MAJOR.MINOR.PATCH"

  Scenario: Manifest with unsupported type
    Given a package manifest "ailign-pkg.yml" containing:
      """
      name: "company/tools"
      type: "mcp"
      version: "1.0.0"
      description: "MCP tool definitions"
      content:
        main: "tools.yml"
      """
    When the manifest is validated
    Then the manifest will be rejected
    And the error will contain "mcp"
    And the error will contain "instructions"

  Scenario: Manifest name matches registry path
    Given a package manifest "ailign-pkg.yml" containing:
      """
      name: "company/security"
      type: "instructions"
      version: "1.3.0"
      description: "Security instructions"
      content:
        main: "instructions.md"
      """
    And the package reference is "instructions/company/security@1.3.0"
    When the manifest is validated against the package reference
    Then the manifest will be accepted
    And the manifest name "company/security" will match the reference path "instructions/company/security"
