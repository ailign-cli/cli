package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"
)

// knownSchemaProperties lists the top-level fields defined in the schema.
var knownSchemaProperties = map[string]bool{
	"targets": true,
}

// Validate validates a Config against the embedded JSONSchema.
// All errors are collected and returned at once (never early-exit).
func Validate(cfg *Config) *ValidationResult {
	result := &ValidationResult{Valid: true}

	jsonData, err := json.Marshal(cfg)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			FieldPath:   "(internal)",
			Message:     "failed to marshal config to JSON",
			Remediation: "Check that the config file is well-formed YAML",
			Severity:    "error",
		})
		return result
	}

	schema, err := compileSchema()
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			FieldPath:   "(internal)",
			Message:     fmt.Sprintf("internal error: %v", err),
			Remediation: "This is a bug in AIlign. Please report it.",
			Severity:    "error",
		})
		return result
	}

	var doc interface{}
	if err := json.Unmarshal(jsonData, &doc); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			FieldPath:   "(internal)",
			Message:     "failed to parse config as JSON",
			Remediation: "Check that the config file is well-formed YAML",
			Severity:    "error",
		})
		return result
	}

	err = schema.Validate(doc)
	if err != nil {
		validationErr, ok := err.(*jsonschema.ValidationError)
		if ok {
			errs := transformErrors(validationErr)
			result.Errors = errs
			result.Valid = false
		} else {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				FieldPath:   "(internal)",
				Message:     err.Error(),
				Remediation: "Check the config file against the AIlign schema",
				Severity:    "error",
			})
		}
	}

	if result.Valid {
		result.Config = cfg
	}

	return result
}

// DetectUnknownFields parses raw YAML and returns warnings for any
// top-level fields not defined in the schema.
func DetectUnknownFields(rawYAML []byte) []ValidationError {
	var warnings []ValidationError

	if len(bytes.TrimSpace(rawYAML)) == 0 {
		return warnings
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(rawYAML, &raw); err != nil {
		return warnings
	}

	for key := range raw {
		if !knownSchemaProperties[key] {
			warnings = append(warnings, ValidationError{
				FieldPath:   key,
				Expected:    "",
				Actual:      "",
				Message:     "unrecognized field",
				Remediation: "Remove it or check for typos",
				Severity:    "warning",
			})
		}
	}

	return warnings
}

func compileSchema() (*jsonschema.Schema, error) {
	schemaDoc, err := jsonschema.UnmarshalJSON(bytes.NewReader(SchemaJSON))
	if err != nil {
		return nil, fmt.Errorf("parsing schema: %w", err)
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", schemaDoc); err != nil {
		return nil, fmt.Errorf("adding schema resource: %w", err)
	}

	return compiler.Compile("schema.json")
}

// transformErrors converts jsonschema validation errors into user-friendly
// ValidationError structs with remediation guidance.
func transformErrors(err *jsonschema.ValidationError) []ValidationError {
	var result []ValidationError
	collectErrors(err, &result)
	return result
}

func collectErrors(err *jsonschema.ValidationError, result *[]ValidationError) {
	if len(err.Causes) == 0 {
		ve := errorToValidationError(err)
		if ve != nil {
			*result = append(*result, *ve)
		}
		return
	}

	for _, cause := range err.Causes {
		collectErrors(cause, result)
	}
}

func errorToValidationError(err *jsonschema.ValidationError) *ValidationError {
	fieldPath := instanceLocationToFieldPath(err.InstanceLocation)

	ve := &ValidationError{
		FieldPath: fieldPath,
		Severity:  "error",
	}

	switch k := err.ErrorKind.(type) {
	case *kind.Required:
		missing := strings.Join(k.Missing, ", ")
		if len(k.Missing) > 0 {
			ve.FieldPath = k.Missing[0]
		}
		ve.Expected = "required"
		ve.Message = "required field missing"
		ve.Remediation = fmt.Sprintf("Add the required field(s): %s", missing)

	case *kind.Enum:
		ve.Expected = "one of claude, cursor, copilot, windsurf"
		ve.Actual = fmt.Sprintf("%v", k.Got)
		ve.Message = "invalid target name"
		ve.Remediation = "Use a supported target name: claude, cursor, copilot, windsurf"

	case *kind.MinItems:
		ve.Expected = "at least 1 target"
		ve.Actual = fmt.Sprintf("%d items", k.Got)
		ve.Message = "targets array is empty"
		ve.Remediation = "Add at least one target to the \"targets\" array"

	case *kind.UniqueItems:
		ve.Expected = "unique target names"
		ve.Actual = fmt.Sprintf("duplicate items at indices %d and %d", k.Duplicates[0], k.Duplicates[1])
		ve.Message = "duplicate targets found"
		ve.Remediation = "Remove duplicate target entries"

	case *kind.Type:
		ve.Expected = fmt.Sprintf("type %v", k.Want)
		ve.Actual = k.Got
		ve.Message = fmt.Sprintf("expected type %v", k.Want)
		ve.Remediation = "Check the config file against the AIlign schema"

	default:
		ve.Message = fmt.Sprintf("%v", err.ErrorKind)
		ve.Remediation = "Check the config file against the AIlign schema"
	}

	return ve
}

// instanceLocationToFieldPath converts a jsonschema v6 InstanceLocation
// ([]string like ["targets", "1"]) into a dot-notation field path
// (like "targets[1]").
func instanceLocationToFieldPath(parts []string) string {
	if len(parts) == 0 {
		return ""
	}

	var result strings.Builder
	for i, part := range parts {
		// Check if part is a numeric index
		isIndex := true
		for _, c := range part {
			if c < '0' || c > '9' {
				isIndex = false
				break
			}
		}

		if isIndex && i > 0 {
			result.WriteString("[" + part + "]")
		} else {
			if i > 0 {
				result.WriteString(".")
			}
			result.WriteString(part)
		}
	}

	return result.String()
}

