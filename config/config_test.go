package config

import (
	"github.com/google/go-cmp/cmp"
	"testing"
	"testing/fstest"
)

const (
	config = `{
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

	// I've removed the unchanged fields from configLocal to test that
	// existing fields in the config above don't Get changed or removed
	configLocal = `{
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
	configInvalid = `This is not a valid JSON file`
)

func TestLoadJson(t *testing.T) {
	t.Run("with a missing file", func(t *testing.T) {
		fs := fstest.MapFS{}

		c := New(fs)
		err := c.LoadJson("missing.json")
		assertError(t, err)
	})

	t.Run("with an invalid json file", func(t *testing.T) {
		fs := fstest.MapFS{
			"configInvalid.json": {Data: []byte(configInvalid)},
		}

		c := New(fs)
		err := c.LoadJson("configInvalid.json")
		assertError(t, err)
	})

	t.Run("with a single valid json file", func(t *testing.T) {
		fs := fstest.MapFS{
			"config.json": {Data: []byte(config)},
		}

		c := New(fs)
		err := c.LoadJson("config.json")
		assertNilError(t, err)

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
			"config.json":      {Data: []byte(config)},
			"configLocal.json": {Data: []byte(configLocal)},
		}

		c := New(fs)
		err := c.LoadJson("config.json")
		assertNilError(t, err)
		err = c.LoadJson("configLocal.json")
		assertNilError(t, err)

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
}

func TestGet(t *testing.T) {
	fs := fstest.MapFS{
		"config.json": {Data: []byte(config)},
	}

	c := New(fs)
	err := c.LoadJson("config.json")
	assertNilError(t, err)

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

func assertError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatalf("did not Get expected error")
	}
}

func assertNilError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("got an unexpected error: %v", err)
	}
}

//func assertMap(t *testing.T, got, want map[string]any) {
//	t.Helper()
//
//	if diff := cmp.Diff(want, got); diff != "" {
//		t.Errorf("assert mismatch (-want +got):\n%s", diff)
//	}
//}

func assertValue(t *testing.T, got, want any) {
	t.Helper()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("assert mismatch (-want +got):\n%s", diff)
	}
}
