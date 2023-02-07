package main

import (
	"config-chg/config"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	config := config.New(os.DirFS("."), os.Stdout)

	config.Load("fixtures/config.json", json.Unmarshal)
	fmt.Println(config.Get("environment"))
	fmt.Println(config.Get("database"))
	fmt.Println(config.Get("cache"))

	config.Load("fixtures/config.local.json", json.Unmarshal)
	fmt.Println(config.Get("environment"))
	fmt.Println(config.Get("database"))
	fmt.Println(config.Get("cache"))
}
