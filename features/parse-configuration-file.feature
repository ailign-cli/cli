# Source: specs/001-config-parsing/spec.md - User Story 1
Feature: Parse Configuration File
  In order to declare which AI tools my repository targets
  As a developer working in a repository
  I want to parse a .ailign.yml config file

  Scenario: Valid configuration is loaded
    Given a repository contained a valid .ailign.yml with targets "claude,cursor"
    When the CLI parses the configuration
    Then the targets will be loaded successfully
    And the loaded targets will be "claude,cursor"

  Scenario: Missing configuration file
    Given a repository had no .ailign.yml file
    When the CLI attempts to load configuration
    Then it will report an error containing "not found" to stderr
    And it will exit with code 2

  Scenario: Empty configuration file
    Given a repository contained an empty .ailign.yml file
    When the CLI attempts to parse it
    Then it will report a validation error about missing "targets" field
