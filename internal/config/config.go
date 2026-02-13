package config

import _ "embed"

//go:embed schema.json
var SchemaJSON []byte

// Config represents the parsed .ailign.yml configuration file.
type Config struct {
	Targets []string `yaml:"targets"`
}
