package steps

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
)

// installDocsState holds state specific to documentation scenarios.
type installDocsState struct {
	readmeContent      string
	installSection     string
	installSectionLine int
}

func registerInstallDocsSteps(ctx *godog.ScenarioContext, w *testWorld) {
	ds := &installDocsState{}

	ctx.Given(`^the project README\.md file$`, ds.theProjectReadmeFile)
	ctx.Given(`^the Installation section in README\.md$`, ds.theInstallationSection)

	ctx.When(`^a developer reads the file$`, func() error { return nil })

	ctx.Then(`^there will be an "Installation" or "Install" section$`, ds.hasInstallSection)
	ctx.Then(`^it will appear before any usage instructions$`, ds.installBeforeUsage)
	ctx.Then(`^it will contain instructions for (.+)$`, ds.containsInstructionsFor)
	ctx.Then(`^each installation method will have a code block$`, ds.eachMethodHasCodeBlock)
	ctx.Then(`^each code block will contain a single runnable command$`, ds.eachCodeBlockHasCommand)
	ctx.Then(`^it will show how to verify with "([^"]*)"$`, ds.showsVerifyCommand)
}

func (ds *installDocsState) theProjectReadmeFile() error {
	readmePath := filepath.Join(findRepoRoot(), "README.md")
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("failed to read README.md: %w", err)
	}
	ds.readmeContent = string(content)
	return nil
}

func (ds *installDocsState) theInstallationSection() error {
	if err := ds.theProjectReadmeFile(); err != nil {
		return err
	}
	if err := ds.hasInstallSection(); err != nil {
		return err
	}
	return nil
}

func (ds *installDocsState) hasInstallSection() error {
	lines := strings.Split(ds.readmeContent, "\n")
	re := regexp.MustCompile(`(?i)^#{1,3}\s+(installation|install)\b`)
	for i, line := range lines {
		if re.MatchString(line) {
			ds.installSectionLine = i
			// Extract from this heading to the next same-level or higher heading
			level := strings.Count(strings.TrimLeft(line, " "), "#")
			level = len(strings.TrimRight(strings.Split(line, " ")[0], " "))
			var sb strings.Builder
			sb.WriteString(line + "\n")
			for j := i + 1; j < len(lines); j++ {
				if isHeading(lines[j]) && headingLevel(lines[j]) <= level {
					break
				}
				sb.WriteString(lines[j] + "\n")
			}
			ds.installSection = sb.String()
			return nil
		}
	}
	return fmt.Errorf("no 'Installation' or 'Install' section found in README.md")
}

func (ds *installDocsState) installBeforeUsage() error {
	lines := strings.Split(ds.readmeContent, "\n")
	usageRe := regexp.MustCompile(`(?i)^#{1,3}\s+(usage|getting started|quick start|commands)\b`)
	for i, line := range lines {
		if usageRe.MatchString(line) {
			if i < ds.installSectionLine {
				return fmt.Errorf("usage section (line %d) appears before installation section (line %d)", i+1, ds.installSectionLine+1)
			}
			return nil
		}
	}
	// No usage section found — that's fine, install is before nothing
	return nil
}

func (ds *installDocsState) containsInstructionsFor(method string) error {
	method = strings.TrimSpace(method)
	lower := strings.ToLower(ds.installSection)
	methodLower := strings.ToLower(method)

	// Map feature file method names to what we'd expect in the README
	checks := map[string][]string{
		"homebrew":        {"homebrew", "brew install"},
		"go install":      {"go install"},
		"the install script": {"install.sh", "curl"},
		"install script":  {"install.sh", "curl"},
		"scoop":           {"scoop"},
		"npm":             {"npm", "npx"},
		"docker":          {"docker"},
		"direct download": {"github.com/ailign-cli/cli/releases", "direct download", "download"},
		"linux packages":  {"deb", "rpm", "apk", "dpkg", "linux package"},
	}

	patterns, ok := checks[methodLower]
	if !ok {
		// Fallback: just check if the method name appears
		if strings.Contains(lower, methodLower) {
			return nil
		}
		return fmt.Errorf("no instructions found for %q in Installation section", method)
	}

	for _, p := range patterns {
		if strings.Contains(lower, strings.ToLower(p)) {
			return nil
		}
	}
	return fmt.Errorf("no instructions found for %q in Installation section (looked for: %v)", method, patterns)
}

func (ds *installDocsState) eachMethodHasCodeBlock() error {
	// Check that there's at least one code block in the install section
	if !strings.Contains(ds.installSection, "```") {
		return fmt.Errorf("no code blocks found in Installation section")
	}

	// Check that each sub-heading has at least one code block after it
	lines := strings.Split(ds.installSection, "\n")
	inSubSection := false
	hasCodeBlock := false
	subSectionName := ""

	for _, line := range lines {
		if isHeading(line) && headingLevel(line) >= 3 {
			if inSubSection && !hasCodeBlock && subSectionName != "" {
				return fmt.Errorf("installation method %q has no code block", subSectionName)
			}
			inSubSection = true
			hasCodeBlock = false
			subSectionName = strings.TrimSpace(strings.TrimLeft(line, "#"))
		}
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			hasCodeBlock = true
		}
	}
	return nil
}

func (ds *installDocsState) eachCodeBlockHasCommand() error {
	// Extract code blocks and verify they contain commands
	blocks := extractCodeBlocks(ds.installSection)
	if len(blocks) == 0 {
		return fmt.Errorf("no code blocks found")
	}
	for _, block := range blocks {
		trimmed := strings.TrimSpace(block)
		if trimmed == "" {
			return fmt.Errorf("found empty code block in Installation section")
		}
	}
	return nil
}

func (ds *installDocsState) showsVerifyCommand(cmd string) error {
	if strings.Contains(ds.installSection, cmd) {
		return nil
	}
	return fmt.Errorf("Installation section does not contain %q", cmd)
}

// --- helpers ---

func isHeading(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "#")
}

func headingLevel(line string) int {
	trimmed := strings.TrimSpace(line)
	level := 0
	for _, c := range trimmed {
		if c == '#' {
			level++
		} else {
			break
		}
	}
	return level
}

func extractCodeBlocks(content string) []string {
	var blocks []string
	lines := strings.Split(content, "\n")
	inBlock := false
	var current strings.Builder
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			if inBlock {
				blocks = append(blocks, current.String())
				current.Reset()
				inBlock = false
			} else {
				inBlock = true
			}
			continue
		}
		if inBlock {
			current.WriteString(line + "\n")
		}
	}
	return blocks
}
