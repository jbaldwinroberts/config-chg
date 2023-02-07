package config

import (
	"encoding/json"
	"io/fs"
)

type Config struct {
	fileSystem fs.FS
}

func New(fs fs.FS) Config {
	return Config{fs}
}

func (c *Config) loadJson(filename string) (map[string]any, error) {
	file, err := fs.ReadFile(c.fileSystem, filename)
	if err != nil {
		return nil, err
	}

	var data map[string]any
	if err := json.Unmarshal(file, &data); err != nil {
		return nil, err
	}

	return data, nil
}
