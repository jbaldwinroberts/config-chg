package config

import (
	"encoding/json"
	"github.com/imdario/mergo"
	"io/fs"
	"strings"
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

func (c *Config) get(path string) any {
	return retrieve(c.config, path)
}

func retrieve(config map[string]any, path string) any {
	// Get the parts of the path before the first '.' and after the first '.'
	before, after, _ := strings.Cut(path, ".")

	// Find the value associated with the key stored in before
	// if it's not found then the requested path does not exist
	found, ok := config[before]
	if !ok {
		// It could be better to return an error here depending on the requirements
		// I'm returning an empty string to keep things simple
		return ""
	}

	// If after is empty then we have finished traversing
	// the path, and we can return the value
	if len(after) == 0 {
		return found
	}

	// Assert the type to check that found is the expected type
	config, ok = found.(map[string]any)
	if !ok {
		// It could be better to return an error here depending on the requirements
		// I'm returning an empty string to keep things simple
		return ""
	}

	// If we haven't traversed the path then call retrieve again
	// to traverse the next level down
	return retrieve(config, after)
}
