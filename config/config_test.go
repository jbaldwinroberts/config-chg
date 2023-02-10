package config

import (
	"bytes"
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	toml "github.com/pelletier/go-toml"
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

func TestLoadJson(t *testing.T) {
	t.Run("with a missing file", func(t *testing.T) {
		fs := fstest.MapFS{}
		buffer := &bytes.Buffer{}
		c := New(fs, buffer)

		c.Load("missing.json", json.Unmarshal)
		assertError(t, buffer)
	})

	t.Run("with an invalid json file", func(t *testing.T) {
		fs := fstest.MapFS{
			"configInvalid.json": {Data: []byte(configInvalid)},
		}
		buffer := &bytes.Buffer{}
		c := New(fs, buffer)

		c.Load("configInvalid.json", json.Unmarshal)
		assertError(t, buffer)
	})

	t.Run("with a single valid json file", func(t *testing.T) {
		fs := fstest.MapFS{
			"config.json": {Data: []byte(configJson)},
		}
		buffer := &bytes.Buffer{}
		c := New(fs, buffer)

		c.Load("config.json", json.Unmarshal)
		assertNilError(t, buffer)

		want := map[string]any{
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
		}

		assertValue(t, c.config, want)
	})

	t.Run("with multiple valid json files", func(t *testing.T) {
		fs := fstest.MapFS{
			"config.json":      {Data: []byte(configJson)},
			"configLocal.json": {Data: []byte(configLocalJson)},
		}
		buffer := &bytes.Buffer{}
		c := New(fs, buffer)

		c.Load("config.json", json.Unmarshal)
		assertNilError(t, buffer)
		c.Load("configLocal.json", json.Unmarshal)
		assertNilError(t, buffer)

		want := map[string]any{
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
		}

		assertValue(t, c.config, want)
	})

	t.Run("with a valid json file and a valid yaml file", func(t *testing.T) {
		fs := fstest.MapFS{
			"config.json":      {Data: []byte(configJson)},
			"configLocal.yaml": {Data: []byte(configLocalYaml)},
		}
		buffer := &bytes.Buffer{}
		c := New(fs, buffer)

		c.Load("config.json", json.Unmarshal)
		assertNilError(t, buffer)
		c.Load("configLocal.yaml", yaml.Unmarshal)
		assertNilError(t, buffer)

		want := map[string]any{
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
		}

		assertValue(t, c.config, want)
	})

	t.Run("with a valid json file and a valid toml file", func(t *testing.T) {
		fs := fstest.MapFS{
			"config.json":      {Data: []byte(configJson)},
			"configLocal.toml": {Data: []byte(configLocalToml)},
		}
		buffer := &bytes.Buffer{}
		c := New(fs, buffer)

		c.Load("config.json", json.Unmarshal)
		assertNilError(t, buffer)
		c.Load("configLocal.toml", toml.Unmarshal)
		assertNilError(t, buffer)

		want := map[string]any{
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
		}

		assertValue(t, c.config, want)
	})
}

func TestGet(t *testing.T) {
	fs := fstest.MapFS{
		"config.json": {Data: []byte(configJson)},
	}
	buffer := &bytes.Buffer{}
	c := New(fs, buffer)

	c.Load("config.json", json.Unmarshal)
	assertNilError(t, buffer)

	t.Run("Get a non-existent value", func(t *testing.T) {
		got := c.Get("protocol")
		assertValue(t, got, "")
	})

	t.Run("Get an outer value", func(t *testing.T) {
		got := c.Get("environment")
		assertValue(t, got, "production")
	})

	t.Run("Get an inner value", func(t *testing.T) {
		got := c.Get("cache.redis.port")
		assertValue(t, got, float64(6379))
	})

	t.Run("Get an outer section", func(t *testing.T) {
		got := c.Get("database")
		assertValue(t, got, map[string]any{
			"host":     "mysql",
			"port":     float64(3306),
			"username": "divido",
			"password": "divido",
		})
	})

	t.Run("Get an inner section", func(t *testing.T) {
		got := c.Get("cache.redis")
		assertValue(t, got, map[string]any{
			"host": "redis",
			"port": float64(6379),
		})
	})
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
