package output

// ValidationError represents a single validation error or warning for formatting.
type ValidationError struct {
	FieldPath   string
	Expected    string
	Actual      string
	Message     string
	Remediation string
	Severity    string // "error" or "warning"
}

// ValidationResult represents the outcome of config validation for formatting.
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []ValidationError
	File     string
}

// Formatter defines the interface for formatting validation output.
type Formatter interface {
	FormatSuccess(result ValidationResult) string
	FormatErrors(result ValidationResult) string
	FormatWarnings(result ValidationResult) string
}
