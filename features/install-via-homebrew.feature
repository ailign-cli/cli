@ci
# Source: specs/003-install-distribution/spec.md - User Story 1

Feature: Install via Homebrew
  In order to install ailign using my existing package manager workflow
  As a macOS or Linux developer
  I want to run brew install and have a working binary

  Scenario: Fresh install from tap
    Given the ailign Homebrew tap has been added
    When the developer runs "brew install ailign-cli/distribution/ailign"
    Then the ailign binary will be available on PATH
    And running "ailign --version" will print a version string

  Scenario: Version matches release
    Given ailign was installed via Homebrew from a release tagged "v0.3.0"
    When the developer runs "ailign --version"
    Then the output will contain "0.3.0"

  Scenario: Upgrade to newer version
    Given ailign v0.2.0 is installed via Homebrew
    And a new release v0.3.0 has been published to the tap
    When the developer runs "brew upgrade ailign"
    Then running "ailign --version" will contain "0.3.0"

  Scenario: Uninstall cleanly
    Given ailign is installed via Homebrew
    When the developer runs "brew uninstall ailign"
    Then the ailign binary will no longer be on PATH
