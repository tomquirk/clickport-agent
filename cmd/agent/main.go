package main

import (
	"log"

	cfg "gitlab.com/runtime-hq/runtime-agent/internal/config"
	"gitlab.com/runtime-hq/runtime-agent/internal/server"
)

func main() {
	config, err := cfg.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	if err = server.Start(config); err != nil {
		log.Fatal(err)
	}
}
