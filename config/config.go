package config

import (
	"fmt"
	"github.com/imdario/mergo"
	"io"
	"io/fs"
	"strings"
)

// Parser defines the function signature of a function that can
// be passed into Load. This enables the user to specify what
// parser to use, enabling support for different file types.
type Parser func(data []byte, v any) error

type Config struct {
	fileSystem fs.FS
	config     map[string]any
	writer     io.Writer
}

// New creates an instance of config
// The file system and writer are passed in to improve
// testability by removing the dependency on a real file system, or stdout
func New(fs fs.FS, writer io.Writer) Config {
	return Config{
		fileSystem: fs,
		config:     map[string]any{},
		writer:     writer,
	}
}

// Load loads the config in the specified filename, using the specified parser function
// If the config already exists it will merge the new config
// into the existing config, overwriting any values that already exist
func (c *Config) Load(filename string, parser Parser) {
	file, err := fs.ReadFile(c.fileSystem, filename)
	if err != nil {
		// Added this to handle the requirement in note 2
		// I would prefer to return the error
		_, _ = fmt.Fprintf(c.writer, err.Error())
		return
	}

	var data map[string]any
	if err = parser(file, &data); err != nil {
		// Added this to handle the requirement in note 2
		// I would prefer to return the error
		_, _ = fmt.Fprintf(c.writer, err.Error())
		return
	}

	// Merge data into any existing config
	// Overwrites any values that exist in both existing config and data
	_ = mergo.Merge(&c.config, data, mergo.WithOverride)
}

// Get retrieves the config specified by the path
// It will return an empty string if the config specified
// by the path does not exist
func (c *Config) Get(path string) any {
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
