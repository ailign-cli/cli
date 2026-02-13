package target

// Target defines the interface for an AI tool target.
// Per-target implementation is out of scope for this feature.
type Target interface {
	Name() string
}

var knownTargets = []string{"claude", "cursor", "copilot", "windsurf"}

// IsValid returns true if the given name is a known target.
func IsValid(name string) bool {
	for _, t := range knownTargets {
		if t == name {
			return true
		}
	}
	return false
}

// KnownTargets returns a copy of the list of known target names.
func KnownTargets() []string {
	result := make([]string, len(knownTargets))
	copy(result, knownTargets)
	return result
}
