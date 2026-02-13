package config

// ValidationError represents a single validation error or warning.
type ValidationError struct {
	FieldPath   string // dot notation path (e.g., "targets", "targets[0]")
	Expected    string // what the schema requires
	Actual      string // what was found (empty when field is missing)
	Message     string // human-readable error description
	Remediation string // concrete action to fix the issue
	Severity    string // "error" or "warning"
}

// ValidationResult represents the outcome of validating a config file.
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []ValidationError
	Config   *Config // nil if validation failed
}
