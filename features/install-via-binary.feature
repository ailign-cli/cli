@wip
# Source: specs/003-install-distribution/spec.md - User Story 3

Feature: Download Pre-built Binary
  In order to install ailign on any system without a package manager
  As a developer or CI/CD pipeline
  I want to download a pre-built binary from GitHub Releases

  Scenario: Install script on macOS
    Given the developer is on macOS with arm64 architecture
    When the developer runs the install script via curl
    Then the ailign binary will be installed to the default location
    And running "ailign --version" will print a version string

  Scenario: Install script on Linux
    Given the developer is on Linux with amd64 architecture
    When the developer runs the install script via curl
    Then the ailign binary will be installed to the default location
    And running "ailign --version" will print a version string

  Scenario: Install script with custom directory
    Given the developer sets INSTALL_DIR to "/tmp/test-bin"
    When the developer runs the install script
    Then the ailign binary will be at "/tmp/test-bin/ailign"

  Scenario: Install script with specific version
    Given the developer sets AILIGN_VERSION to "v0.2.0"
    When the developer runs the install script
    Then running "ailign --version" will contain "0.2.0"

  Scenario: Checksum verification
    Given a release archive has been downloaded
    And the checksums.txt file has been downloaded
    When the checksum is verified against the archive
    Then the checksum will match

  Scenario: Install script warns if not in PATH
    Given the developer sets INSTALL_DIR to a directory not in PATH
    When the developer runs the install script
    Then the output will contain a warning about PATH

  @edge-case
  Scenario: Install script on unsupported platform
    Given the developer is on an unsupported OS
    When the developer runs the install script
    Then the script will exit with an error
    And the error message will list supported platforms
