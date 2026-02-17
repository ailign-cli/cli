package target

// Claude implements the Target interface for Claude Code.
type Claude struct{}

func (Claude) Name() string            { return "claude" }
func (Claude) InstructionPath() string { return ".claude/instructions.md" }
