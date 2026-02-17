package target

// Target defines the interface for an AI tool target.
type Target interface {
	Name() string
	InstructionPath() string
}

// Registry holds all available target implementations.
type Registry struct {
	targets map[string]Target
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry {
	return &Registry{targets: make(map[string]Target)}
}

// Register adds a target to the registry.
func (r *Registry) Register(t Target) {
	r.targets[t.Name()] = t
}

// Get looks up a target by name.
func (r *Registry) Get(name string) (Target, bool) {
	t, ok := r.targets[name]
	return t, ok
}

// IsValid returns true if the given name is registered.
func (r *Registry) IsValid(name string) bool {
	_, ok := r.targets[name]
	return ok
}

// KnownTargets returns a sorted list of all registered target names.
func (r *Registry) KnownTargets() []string {
	names := make([]string, 0, len(r.targets))
	for name := range r.targets {
		names = append(names, name)
	}
	// Sort for deterministic output
	sortStrings(names)
	return names
}

// NewDefaultRegistry creates a Registry with all built-in targets.
func NewDefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register(Claude{})
	r.Register(Cursor{})
	r.Register(Copilot{})
	r.Register(Windsurf{})
	return r
}

// sortStrings sorts a slice of strings in place (insertion sort, no import needed for small slices).
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// Package-level convenience functions using a default registry.
var defaultRegistry = NewDefaultRegistry()

// IsValid returns true if the given name is a known target.
func IsValid(name string) bool {
	return defaultRegistry.IsValid(name)
}

// KnownTargets returns a copy of the list of known target names.
func KnownTargets() []string {
	return defaultRegistry.KnownTargets()
}
