package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ecnepsnai/configsync"
)

func main() {
	args := os.Args
	if len(args) == 1 {
		fmt.Fprintf(os.Stderr, "Usage %s <Config Path>\n", os.Args[0])
		os.Exit(1)
	}

	f, err := os.OpenFile(args[1], os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	config := configsync.Config{}
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		panic(err)
	}

	configsync.Sync(&config)
}
