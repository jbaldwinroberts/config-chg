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

	configLocal = `{
  "environment": "development",
  "database": {
    "host": "127.0.0.1",
    "port": 3306,
    "username": "divido",
    "password": "divido"
  },
  "cache": {
    "redis": {
      "host": "127.0.0.1",
      "port": 6379
    }
  }
}`
	configInvalid = `This is not a valid JSON file`
)

func TestLoadJson(t *testing.T) {
	t.Run("with a missing file", func(t *testing.T) {
		fs := fstest.MapFS{}

		c := New(fs)
		err := c.loadJson("missing.json")
		assertError(t, err)
	})

	t.Run("with an invalid json file", func(t *testing.T) {
		fs := fstest.MapFS{
			"configInvalid.json": {Data: []byte(configInvalid)},
		}

		c := New(fs)
		err := c.loadJson("configInvalid.json")
		assertError(t, err)
	})

	t.Run("with a single valid json file", func(t *testing.T) {
		fs := fstest.MapFS{
			"config.json": {Data: []byte(config)},
		}

		c := New(fs)
		err := c.loadJson("config.json")
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

		assertMap(t, c.config, want)
	})

	//t.Run("with multiple valid json files", func(t *testing.T) {
	//	fs := fstest.MapFS{
	//		"config.json": {Data: []byte(config)},
	//		"configLocal.json": {Data: []byte(configLocal)},
	//	}
	//
	//	c := New(fs)
	//	got, err := c.loadJson("config.json")
	//
	//
	//	assertNilError(t, err)
	//
	//	want := map[string]any{
	//		"environment": "production",
	//		"database": map[string]any{
	//			"host":     "mysql",
	//			"port":     float64(3306),
	//			"username": "divido",
	//			"password": "divido",
	//		},
	//		"cache": map[string]any{
	//			"redis": map[string]any{
	//				"host": "redis",
	//				"port": float64(6379),
	//			},
	//		},
	//	}
	//
	//	assertMap(t, got, want)
	//})
}

func assertError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatalf("did not get expected error")
	}
}

func assertNilError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("got an unexpected error: %v", err)
	}
}

func assertMap(t *testing.T, got, want map[string]any) {
	t.Helper()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("assert mismatch (-want +got):\n%s", diff)
	}
}
