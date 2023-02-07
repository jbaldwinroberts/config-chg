package config_test

import (
	. "config-chg/config"
	"fmt"
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

func ExampleGet_value() {
	fs := fstest.MapFS{
		"config.json": {Data: []byte(config)},
	}

	c := New(fs)
	c.LoadJson("config.json")

	value := c.Get("environment")
	fmt.Println(value)
	// Output: production
}

func ExampleGet_section() {
	fs := fstest.MapFS{
		"config.json": {Data: []byte(config)},
	}

	c := New(fs)
	c.LoadJson("config.json")

	value := c.Get("database")
	fmt.Println(value)
	// Output: map[host:mysql password:divido port:3306 username:divido]
}
