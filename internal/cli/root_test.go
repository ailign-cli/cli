package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func executeRootWithSubcommand(args []string, dir string) (stdout string, stderr string, err error) {
	rootCmd := NewRootCommand()

	// Add a no-op subcommand to trigger PersistentPreRunE
	checkCmd := &cobra.Command{
		Use:  "check",
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	rootCmd.AddCommand(checkCmd)

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	rootCmd.SetOut(stdoutBuf)
	rootCmd.SetErr(stderrBuf)
	rootCmd.SetArgs(args)

	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	execErr := rootCmd.Execute()
	return stdoutBuf.String(), stderrBuf.String(), execErr
}

func TestRootCommand_ValidConfig_Proceeds(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, ".ailign.yml"),
		[]byte("targets:\n  - claude\n  - cursor\n"), 0644)

	_, stderr, err := executeRootWithSubcommand([]string{"check"}, dir)
	assert.NoError(t, err)
	assert.Empty(t, stderr)
}

func TestRootCommand_WarningsDoNotBlock(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, ".ailign.yml"),
		[]byte("targets:\n  - claude\ncustom_field: value\n"), 0644)

	_, stderr, err := executeRootWithSubcommand([]string{"check"}, dir)
	assert.NoError(t, err)
	assert.Contains(t, stderr, "warning")
	assert.Contains(t, stderr, "custom_field")
}

func TestRootCommand_HelpDoesNotRequireConfig(t *testing.T) {
	dir := t.TempDir() // No .ailign.yml

	rootCmd := NewRootCommand()
	stdoutBuf := new(bytes.Buffer)
	rootCmd.SetOut(stdoutBuf)
	rootCmd.SetArgs([]string{"--help"})

	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	err := rootCmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, stdoutBuf.String(), "ailign")
}

func TestRootCommand_FormatFlag_DefaultIsHuman(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.SetArgs([]string{"--help"})
	_ = rootCmd.Execute()

	flag := rootCmd.PersistentFlags().Lookup("format")
	assert.NotNil(t, flag)
	assert.Equal(t, "human", flag.DefValue)
}

func TestRootCommand_FormatFlag_ShortFlag(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.SetArgs([]string{"--help"})
	_ = rootCmd.Execute()

	flag := rootCmd.PersistentFlags().ShorthandLookup("f")
	assert.NotNil(t, flag)
	assert.Equal(t, "format", flag.Name)
}

func TestRootCommand_InvalidFormat_Rejected(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, ".ailign.yml"),
		[]byte("targets:\n  - claude\n"), 0644)

	_, _, err := executeRootWithSubcommand([]string{"--format", "yaml", "check"}, dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown output format")
	assert.Contains(t, err.Error(), "yaml")
}

func TestRootCommand_VersionDoesNotRequireConfig(t *testing.T) {
	dir := t.TempDir() // No .ailign.yml

	rootCmd := NewRootCommand()
	rootCmd.Version = "1.2.3 (abc1234)"

	stdoutBuf := new(bytes.Buffer)
	rootCmd.SetOut(stdoutBuf)
	rootCmd.SetArgs([]string{"--version"})

	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	err := rootCmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, stdoutBuf.String(), "1.2.3")
	assert.Contains(t, stdoutBuf.String(), "abc1234")
}

func TestRootCommand_InvalidConfig_ReturnsError(t *testing.T) {
	dir := t.TempDir() // No .ailign.yml

	_, stderr, err := executeRootWithSubcommand([]string{"check"}, dir)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAlreadyReported)
	assert.Contains(t, stderr, "not found")
}
