package config

import (
	"encoding/json"
	"github.com/imdario/mergo"
	"io/fs"
)

type Config struct {
	fileSystem fs.FS
	config     map[string]any
}

// New creates an instance of config
// The file system is passed in to improve testability by removing
// the dependency on a real file system
func New(fs fs.FS) Config {
	return Config{
		fileSystem: fs,
		config:     map[string]any{},
	}
}

func (c *Config) loadJson(filename string) error {
	file, err := fs.ReadFile(c.fileSystem, filename)
	if err != nil {
		return err
	}

	var data map[string]any
	if err = json.Unmarshal(file, &data); err != nil {
		return err
	}

	// Merge data into any existing config
	// Overwrites any values that exist in both existing config and data
	return mergo.Merge(&c.config, data, mergo.WithOverride)
}
