package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

// LoadFromFile reads and parses a YAML config file at the given path.
// Returns a Config struct on success. Returns an error for missing files,
// permission errors, or YAML parse errors.
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("config not found: %s", path)
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	// Strip UTF-8 BOM if present
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

// LoadAndValidate loads a config file, validates it against the schema,
// and detects unknown fields. Returns the full validation result.
func LoadAndValidate(path string) *ValidationResult {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &ValidationResult{
				Valid: false,
				Errors: []ValidationError{{
					FieldPath:   "",
					Message:     fmt.Sprintf("config not found: %s", path),
					Remediation: "Run \"ailign init\" to create a configuration file",
					Severity:    "error",
				}},
			}
		}
		return &ValidationResult{
			Valid: false,
			Errors: []ValidationError{{
				FieldPath:   "",
				Message:     fmt.Sprintf("reading config: %v", err),
				Remediation: "Check file permissions",
				Severity:    "error",
			}},
		}
	}

	// Strip UTF-8 BOM if present
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return &ValidationResult{
			Valid: false,
			Errors: []ValidationError{{
				FieldPath:   "",
				Message:     fmt.Sprintf("parsing config: %v", err),
				Remediation: "Check YAML syntax",
				Severity:    "error",
			}},
		}
	}

	// Schema validation
	result := Validate(&cfg)

	// Unknown field detection
	warnings := DetectUnknownFields(data)
	result.Warnings = append(result.Warnings, warnings...)

	return result
}
