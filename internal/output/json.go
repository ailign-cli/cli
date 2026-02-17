package output

import "encoding/json"

// JSONFormatter formats validation results as JSON.
type JSONFormatter struct{}

// jsonValidationError is the JSON wire representation of a single error or warning.
// Actual is a *string so that an empty/missing value serializes as JSON null.
type jsonValidationError struct {
	FieldPath   string  `json:"field_path"`
	Expected    string  `json:"expected"`
	Actual      *string `json:"actual"`
	Message     string  `json:"message"`
	Remediation string  `json:"remediation"`
}

// jsonValidationResult is the JSON wire representation of a full validation result.
type jsonValidationResult struct {
	Valid    bool                  `json:"valid"`
	Errors   []jsonValidationError `json:"errors"`
	Warnings []jsonValidationError `json:"warnings"`
	File     string                `json:"file"`
}

// FormatSuccess returns the full JSON result object for a successful validation.
func (f *JSONFormatter) FormatSuccess(result ValidationResult) string {
	return marshalResult(result)
}

// FormatErrors returns the full JSON result object for a failed validation.
func (f *JSONFormatter) FormatErrors(result ValidationResult) string {
	return marshalResult(result)
}

// FormatWarnings returns the full JSON result object with warnings embedded.
func (f *JSONFormatter) FormatWarnings(result ValidationResult) string {
	return marshalResult(result)
}

// marshalResult converts an internal ValidationResult to its JSON representation
// and returns the pretty-printed JSON string.
func marshalResult(result ValidationResult) string {
	jr := jsonValidationResult{
		Valid:    result.Valid,
		Errors:   convertErrors(result.Errors),
		Warnings: convertErrors(result.Warnings),
		File:     result.File,
	}

	data, err := json.MarshalIndent(jr, "", "  ")
	if err != nil {
		// This should never happen with the types we control, but return a
		// minimal valid JSON object rather than panicking.
		return `{"valid":false,"errors":[],"warnings":[],"file":""}`
	}
	return string(data)
}

// jsonSyncResult is the JSON wire representation of a sync result.
type jsonSyncResult struct {
	DryRun  bool            `json:"dry_run"`
	Hub     jsonHub         `json:"hub"`
	Links   []jsonLink      `json:"links"`
	Summary jsonSyncSummary `json:"summary"`
}

type jsonHub struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}

type jsonLink struct {
	Target   string `json:"target"`
	LinkPath string `json:"link_path"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}

type jsonSyncSummary struct {
	Total    int `json:"total"`
	Created  int `json:"created"`
	Existing int `json:"existing"`
	Errors   int `json:"errors"`
}

// FormatSyncResult returns the JSON representation of a sync result.
func (f *JSONFormatter) FormatSyncResult(result SyncResult) string {
	var created, existing, errCount int
	links := make([]jsonLink, 0, len(result.Links))
	for _, l := range result.Links {
		jl := jsonLink{
			Target:   l.Target,
			LinkPath: l.LinkPath,
			Status:   l.Status,
			Error:    l.Error,
		}
		links = append(links, jl)
		switch l.Status {
		case "created", "replaced":
			created++
		case "exists":
			existing++
		case "error":
			errCount++
		}
	}

	jr := jsonSyncResult{
		DryRun: result.DryRun,
		Hub: jsonHub{
			Path:   result.HubPath,
			Status: result.HubStatus,
		},
		Links: links,
		Summary: jsonSyncSummary{
			Total:    len(result.Links),
			Created:  created,
			Existing: existing,
			Errors:   errCount,
		},
	}

	data, err := json.MarshalIndent(jr, "", "  ")
	if err != nil {
		return `{"dry_run":false,"hub":{},"links":[],"summary":{}}`
	}
	return string(data)
}

// convertErrors maps a slice of internal ValidationError values to the JSON wire
// format. An empty or nil input slice produces a non-nil empty slice so that
// json.Marshal emits [] rather than null.
func convertErrors(errs []ValidationError) []jsonValidationError {
	out := make([]jsonValidationError, 0, len(errs))
	for _, e := range errs {
		je := jsonValidationError{
			FieldPath:   e.FieldPath,
			Expected:    e.Expected,
			Message:     e.Message,
			Remediation: e.Remediation,
		}
		if e.Actual != "" {
			a := e.Actual
			je.Actual = &a
		}
		out = append(out, je)
	}
	return out
}
