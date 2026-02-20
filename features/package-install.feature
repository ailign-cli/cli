@wip
# Source: specs/004-package-install/spec.md - User Story 1

Feature: Install packages from registry
  In order to use org-wide instruction standards in my repository without manually copying files
  As a developer working in an organization with multiple repositories
  I want to run ailign install and have the declared instruction packages fetched and rendered to my configured tool formats

  Scenario: Install single instruction package
    Given the developer created a repository with .ailign.yml containing:
      """
      packages:
        - instructions/company/security@1.3.0
      targets:
        - claude
      local_overlays:
        - .ai-instructions/project-context.md
      """
    And a local overlay file ".ai-instructions/project-context.md" containing:
      """
      # Project Context
      This is a Go project.
      """
    And the registry has package "instructions/company/security@1.3.0" with content:
      """
      # Security Instructions
      Always validate input.
      """
    When the developer runs "ailign install"
    Then the exit code will be 0
    And the file ".ailign/instructions.md" will contain "Security Instructions"
    And the file ".ailign/instructions.md" will contain "Project Context"
    And the symlink ".claude/instructions.md" will point to "../.ailign/instructions.md"
    And stdout will contain "instructions/company/security@1.3.0"
    And stdout will contain "fetched"

  Scenario: Install multiple packages
    Given the developer created a repository with .ailign.yml containing:
      """
      packages:
        - instructions/company/security@1.3.0
        - instructions/company/typescript@2.1.0
      targets:
        - claude
        - cursor
      """
    And the registry has package "instructions/company/security@1.3.0" with content:
      """
      # Security
      Validate input.
      """
    And the registry has package "instructions/company/typescript@2.1.0" with content:
      """
      # TypeScript
      Use strict mode.
      """
    When the developer runs "ailign install"
    Then the exit code will be 0
    And the file ".ailign/instructions.md" will contain "Security"
    And the file ".ailign/instructions.md" will contain "TypeScript"
    And stdout will contain "Installed 2 packages"

  Scenario: Install with local overlays
    Given the developer created a repository with .ailign.yml containing:
      """
      packages:
        - instructions/company/security@1.3.0
      targets:
        - claude
      local_overlays:
        - .ai-instructions/local.md
      """
    And a local overlay file ".ai-instructions/local.md" containing:
      """
      # Local Override
      Project-specific rules.
      """
    And the registry has package "instructions/company/security@1.3.0" with content:
      """
      # Security
      Base rules.
      """
    When the developer runs "ailign install"
    Then the exit code will be 0
    And the file ".ailign/instructions.md" will contain "Security"
    And the file ".ailign/instructions.md" will contain "Local Override"

  Scenario: Install shows summary
    Given the developer created a repository with .ailign.yml containing:
      """
      packages:
        - instructions/company/security@1.3.0
        - instructions/company/typescript@2.1.0
      targets:
        - claude
      """
    And the registry has package "instructions/company/security@1.3.0" with content:
      """
      # Security
      """
    And the registry has package "instructions/company/typescript@2.1.0" with content:
      """
      # TypeScript
      """
    When the developer runs "ailign install"
    Then the exit code will be 0
    And stdout will contain "instructions/company/security@1.3.0"
    And stdout will contain "instructions/company/typescript@2.1.0"

  Scenario: Install with JSON output
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
    When the developer runs "ailign install --format json"
    Then the exit code will be 0
    And stdout will be valid JSON
    And the JSON at "packages[0].reference" will be "instructions/company/security"
    And the JSON at "packages[0].version" will be "1.3.0"
    And the JSON at "packages[0].status" will be "fetched"

  Scenario: Install is idempotent
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
    And the developer has previously run "ailign install"
    When the developer runs "ailign install"
    Then the exit code will be 0
    And the file ".ailign/instructions.md" will be unchanged
    And the file "ailign-lock.yml" will be unchanged

  Scenario: Install creates lock file
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
    And the file "ailign-lock.yml" will exist
    And the file "ailign-lock.yml" will contain "instructions/company/security"
    And the file "ailign-lock.yml" will contain "1.3.0"
    And the file "ailign-lock.yml" will contain "sha256-"

  @edge-case
  Scenario: Install with no packages declared
    Given the developer created a repository with .ailign.yml containing:
      """
      targets:
        - claude
      """
    When the developer runs "ailign install"
    Then the exit code will be 0
    And stdout will contain "No packages declared"

  @edge-case
  Scenario: Install with duplicate package declared
    Given the developer created a repository with .ailign.yml containing:
      """
      packages:
        - instructions/company/security@1.3.0
        - instructions/company/security@1.3.0
      targets:
        - claude
      """
    When the developer runs "ailign install"
    Then the exit code will be 2
    And stderr will contain "duplicate"
    And stderr will contain "instructions/company/security"
