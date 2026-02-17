package target

// Cursor implements the Target interface for Cursor.
type Cursor struct{}

func (Cursor) Name() string            { return "cursor" }
func (Cursor) InstructionPath() string { return ".cursorrules" }
