package config

import (
	"bytes"
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
	"sync"
	"testing"
	"testing/fstest"
)

const (
	configJson = `{
	 "environment": "production",
	 "database": {
	   "host": "mysql",
	   "port": 3306,
	   "username": "divido",
	   "password": "divido"
	 },
	 "cache": {
	   "redis": {
	     "host": "redis",
	     "port": 6379
	   }
	 }
	}`

	// I've removed the unchanged fields from configLocalJson to test that
	// existing fields in the configJson above don't get changed or removed
	configLocalJson = `{
  "environment": "development",
  "database": {
    "host": "127.0.0.1",
    "port": 3306
  },
  "cache": {
    "redis": {
      "host": "127.0.0.1"
    }
  }
}`
	configLocalYaml = `---
environment: development
database:
  host: 127.0.0.1
  port: 3306
cache:
  redis:
    host: 127.0.0.1`

	configLocalToml = `environment = "development"

[database]
host = "127.0.0.1"
port = 3_306

[cache.redis]
host = "127.0.0.1"`

	configInvalid = `This is not a valid JSON file`
)

func TestLoad_SingleConfig(t *testing.T) {
	type test struct {
		name     string
		filename string
		fs       fstest.MapFS
		want     map[string]any
		err      bool
	}

	tests := []test{
		{
			name:     "with a missing file",
			filename: "missing.json",
			fs:       fstest.MapFS{},
			want:     map[string]any{},
			err:      true,
		},
		{
			name:     "with an invalid json file",
			filename: "configInvalid.json",
			fs: fstest.MapFS{
				"configInvalid.json": {Data: []byte(configInvalid)},
			},
			want: map[string]any{},
			err:  true,
		},
		{
			name:     "with a single valid json file",
			filename: "config.json",
			fs: fstest.MapFS{
				"config.json": {Data: []byte(configJson)},
			},
			want: map[string]any{
				"environment": "production",
				"database": map[string]any{
					"host":     "mysql",
					"port":     float64(3306),
					"username": "divido",
					"password": "divido",
				},
				"cache": map[string]any{
					"redis": map[string]any{
						"host": "redis",
						"port": float64(6379),
					},
				},
			},
			err: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			c := New(tc.fs, buffer)

			c.Load(tc.filename, json.Unmarshal)

			if tc.err {
				assertError(t, buffer)
			} else {
				assertNilError(t, buffer)
			}

			assertValue(t, c.config, tc.want)
		})
	}
}

func TestLoad_MultipleConfig(t *testing.T) {
	type file struct {
		name   string
		parser Parser
	}

	type test struct {
		name  string
		files []file
		fs    fstest.MapFS
		want  map[string]any
	}

	tests := []test{
		{
			name: "with multiple valid json files",
			fs: fstest.MapFS{
				"config.json":      {Data: []byte(configJson)},
				"configLocal.json": {Data: []byte(configLocalJson)},
			},
			files: []file{
				{name: "config.json", parser: json.Unmarshal},
				{name: "configLocal.json", parser: json.Unmarshal},
			},
			want: map[string]any{
				"environment": "development",
				"database": map[string]any{
					"host":     "127.0.0.1",
					"port":     float64(3306),
					"username": "divido",
					"password": "divido",
				},
				"cache": map[string]any{
					"redis": map[string]any{
						"host": "127.0.0.1",
						"port": float64(6379),
					},
				},
			},
		},
		{
			name: "with a valid json file and a valid yaml file",
			fs: fstest.MapFS{
				"config.json":      {Data: []byte(configJson)},
				"configLocal.yaml": {Data: []byte(configLocalYaml)},
			},
			files: []file{
				{name: "config.json", parser: json.Unmarshal},
				{name: "configLocal.yaml", parser: yaml.Unmarshal},
			},
			want: map[string]any{
				"environment": "development",
				"database": map[string]any{
					"host":     "127.0.0.1",
					"port":     int(3306),
					"username": "divido",
					"password": "divido",
				},
				"cache": map[string]any{
					"redis": map[string]any{
						"host": "127.0.0.1",
						"port": float64(6379),
					},
				},
			},
		},
		{
			name: "with a valid json file and a valid toml file",
			fs: fstest.MapFS{
				"config.json":      {Data: []byte(configJson)},
				"configLocal.toml": {Data: []byte(configLocalToml)},
			},
			files: []file{
				{name: "config.json", parser: json.Unmarshal},
				{name: "configLocal.toml", parser: toml.Unmarshal},
			},
			want: map[string]any{
				"environment": "development",
				"database": map[string]any{
					"host":     "127.0.0.1",
					"port":     int64(3306),
					"username": "divido",
					"password": "divido",
				},
				"cache": map[string]any{
					"redis": map[string]any{
						"host": "127.0.0.1",
						"port": float64(6379),
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			c := New(tc.fs, buffer)

			for _, f := range tc.files {
				c.Load(f.name, f.parser)
				assertNilError(t, buffer)
			}

			assertValue(t, c.config, tc.want)
		})
	}
}

func TestGet(t *testing.T) {
	fs := fstest.MapFS{
		"config.json": {Data: []byte(configJson)},
	}
	buffer := &bytes.Buffer{}
	c := New(fs, buffer)

	c.Load("config.json", json.Unmarshal)
	assertNilError(t, buffer)

	type test struct {
		name string
		path string
		want any
	}

	tests := []test{
		{
			name: "get a non-existent value",
			path: "protocol",
			want: "",
		},
		{
			name: "get an outer value",
			path: "environment",
			want: "production",
		},
		{
			name: "get an inner value",
			path: "cache.redis.port",
			want: float64(6379),
		},
		{
			name: "get an outer section",
			path: "database",
			want: map[string]any{
				"host":     "mysql",
				"port":     float64(3306),
				"username": "divido",
				"password": "divido",
			},
		},
		{
			name: "get an inner section",
			path: "cache.redis",
			want: map[string]any{
				"host": "redis",
				"port": float64(6379),
			},
		},
	}

	t.Run("Get an inner section", func(t *testing.T) {
		got := c.Get("cache.redis")
		assertValue(t, got, map[string]any{
			"host": "redis",
			"port": float64(6379),
		})
	})

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := c.Get(tc.path)
			assertValue(t, got, tc.want)
		})
	}
}

func TestConcurrency(t *testing.T) {
	fs := fstest.MapFS{
		"config.json":      {Data: []byte(configJson)},
		"configLocal.json": {Data: []byte(configLocalJson)},
	}
	buffer := &bytes.Buffer{}
	c := New(fs, buffer)

	c.Load("config.json", json.Unmarshal)

	wg := sync.WaitGroup{}
	wg.Add(2)

	// Attempt to read and write from the config in parallel
	go func() {
		got := c.Get("environment")
		assertValue(t, got, "production")
		wg.Done()
	}()

	go func() {
		c.Load("configLocal.json", json.Unmarshal)
		wg.Done()
	}()

	wg.Wait()

	got := c.Get("environment")
	assertValue(t, got, "development")
}

func assertError(t *testing.T, buffer *bytes.Buffer) {
	t.Helper()

	if buffer.Len() == 0 {
		t.Fatalf("did not get expected error")
	}
}

func assertNilError(t *testing.T, buffer *bytes.Buffer) {
	t.Helper()

	if buffer.Len() != 0 {
		t.Fatalf("got an unexpected error: %v", buffer.String())
	}
}

func assertValue(t *testing.T, got, want any) {
	t.Helper()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("assert mismatch (-want +got):\n%s", diff)
	}
}
