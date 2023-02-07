package config_test

import (
	. "config-chg/config"
	"encoding/json"
	"fmt"
	"os"
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
)

func ExampleConfig_load() {
	c := New(os.DirFS("."), os.Stdout)

	c.Load("config.json", json.Unmarshal)
}

func ExampleConfig_getValue() {
	fs := fstest.MapFS{
		"config.json": {Data: []byte(config)},
	}
	c := New(fs, os.Stdout)

	c.Load("config.json", json.Unmarshal)
	value := c.Get("environment")
	fmt.Println(value)
	// Output: production
}

func ExampleConfig_getSection() {
	fs := fstest.MapFS{
		"config.json": {Data: []byte(config)},
	}
	c := New(fs, os.Stdout)

	c.Load("config.json", json.Unmarshal)
	value := c.Get("database")
	fmt.Println(value)
	// Output: map[host:mysql password:divido port:3306 username:divido]
}
