# Source: specs/003-install-distribution/spec.md - User Story 4

Feature: Installation Documentation
  In order to find the right installation method for my environment
  As a developer visiting the project for the first time
  I want to see clear, concise installation instructions in the README

  Scenario: README has install section
    Given the project README.md file
    When a developer reads the file
    Then there will be an "Installation" or "Install" section
    And it will appear before any usage instructions

  Scenario: All methods documented
    Given the Installation section in README.md
    Then it will contain instructions for Homebrew
    And it will contain instructions for go install
    And it will contain instructions for the install script
    And it will contain instructions for Scoop
    And it will contain instructions for npm
    And it will contain instructions for Docker
    And it will contain instructions for direct download
    And it will contain instructions for Linux packages

  Scenario: Commands are copy-pasteable
    Given the Installation section in README.md
    Then each installation method will have a code block
    And each code block will contain a single runnable command

  Scenario: Version verification documented
    Given the Installation section in README.md
    Then it will show how to verify with "ailign --version"
