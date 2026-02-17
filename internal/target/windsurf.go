package target

// Windsurf implements the Target interface for Windsurf.
type Windsurf struct{}

func (Windsurf) Name() string            { return "windsurf" }
func (Windsurf) InstructionPath() string { return ".windsurfrules" }
