package sync

// ComposeResult holds the outcome of composing overlay files.
type ComposeResult struct {
	Content  []byte
	Warnings []string
}

// SyncResult holds the outcome of a sync operation.
type SyncResult struct {
	HubPath   string
	HubStatus string // "written" or "unchanged"
	Links     []LinkResult
	DryRun    bool
}

// LinkResult holds the per-target symlink outcome.
type LinkResult struct {
	Target   string
	LinkPath string
	Status   string // "created", "exists", "replaced", "error"
	Error    string
}

// SyncOptions configures the sync operation.
type SyncOptions struct {
	DryRun bool
}
