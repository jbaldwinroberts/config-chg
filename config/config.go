package config

import (
	"encoding/json"
	"io/fs"
)

type Config struct {
	fileSystem fs.FS
	config     map[string]any
}

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

	c.config = data

	return nil
}
