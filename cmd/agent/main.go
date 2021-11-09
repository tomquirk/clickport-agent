package main

import (
	"log"

	cfg "gitlab.com/clickport/clickport-agent/internal/config"
	"gitlab.com/clickport/clickport-agent/internal/server"
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
