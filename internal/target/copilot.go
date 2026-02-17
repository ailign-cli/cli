package target

// Copilot implements the Target interface for GitHub Copilot.
type Copilot struct{}

func (Copilot) Name() string            { return "copilot" }
func (Copilot) InstructionPath() string { return ".github/copilot-instructions.md" }
