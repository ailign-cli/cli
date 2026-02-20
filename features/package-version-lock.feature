@wip
# Source: specs/004-package-install/spec.md - User Story 4

Feature: Version resolution and lock file
  In order to have reproducible builds and controlled updates
  As a developer or CI/CD pipeline
  I want to lock installed package versions and verify integrity through checksums

  Scenario: Lock file created on first install
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
    And no lock file exists
    When the developer runs "ailign install"
    Then the exit code will be 0
    And the file "ailign-lock.yml" will exist
    And the file "ailign-lock.yml" will contain "lockfileVersion: 1"
    And the file "ailign-lock.yml" will contain "instructions/company/security"
    And the file "ailign-lock.yml" will contain "1.3.0"
    And the file "ailign-lock.yml" will contain "sha256-"

  Scenario: Subsequent install uses locked versions
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
    And the lock file records version "1.3.0" for "instructions/company/security"
    When the developer runs "ailign install"
    Then the exit code will be 0
    And the file "ailign-lock.yml" will be unchanged

  Scenario: Config version change updates lock
    Given the developer created a repository with .ailign.yml containing:
      """
      packages:
        - instructions/company/security@1.4.0
      targets:
        - claude
      """
    And the lock file records version "1.3.0" for "instructions/company/security"
    And the registry has package "instructions/company/security@1.4.0" with content:
      """
      # Security v1.4
      """
    When the developer runs "ailign install"
    Then the exit code will be 0
    And the file "ailign-lock.yml" will contain "1.4.0"
    And the file "ailign-lock.yml" will not contain "1.3.0"

  Scenario: Lock file checksum mismatch detected
    Given the developer created a repository with .ailign.yml containing:
      """
      packages:
        - instructions/company/security@1.3.0
      targets:
        - claude
      """
    And the lock file records integrity "sha256-AAAA" for "instructions/company/security"
    And the registry has package "instructions/company/security@1.3.0" with content:
      """
      # Modified content
      """
    When the developer runs "ailign install"
    Then the exit code will be 2
    And stderr will contain "integrity check failed"
    And stderr will contain "instructions/company/security@1.3.0"

  Scenario: Lock file is human-readable
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
    And the file "ailign-lock.yml" will be valid YAML
    And the lock file packages will be sorted by reference

  @edge-case
  Scenario: Corrupt lock file triggers clear error
    Given the developer created a repository with .ailign.yml containing:
      """
      packages:
        - instructions/company/security@1.3.0
      targets:
        - claude
      """
    And the file "ailign-lock.yml" contains invalid YAML:
      """
      lockfileVersion: 1
      packages:
        - reference: [invalid yaml
      """
    When the developer runs "ailign install"
    Then the exit code will be 2
    And stderr will contain "ailign-lock.yml"
    And stderr will contain "delete"
