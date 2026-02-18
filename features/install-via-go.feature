@wip
# Source: specs/003-install-distribution/spec.md - User Story 2

Feature: Install via Go Toolchain
  In order to install ailign using the Go toolchain I already have
  As a Go developer
  I want to run go install and have a working binary

  Scenario: Install latest version
    Given the developer has Go 1.24 or later installed
    When the developer runs "go install github.com/ailign/cli/cmd/ailign@latest"
    Then the ailign binary will be available in GOPATH/bin
    And running "ailign --version" will print a version string

  Scenario: Install specific version
    Given the developer has Go 1.24 or later installed
    When the developer runs "go install github.com/ailign/cli/cmd/ailign@v0.2.0"
    Then running "ailign --version" will contain "0.2.0"

  Scenario: Version output includes tag
    Given ailign was installed via "go install" from tag "v0.3.0"
    When the developer runs "ailign --version"
    Then the output will contain "0.3.0"
