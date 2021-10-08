package main

import (
	cfg "github.com/runtime-hq/runtime-agent/internal/config"
	"github.com/runtime-hq/runtime-agent/internal/server"
)

func main() {
	config := cfg.LoadConfig()

	server.Start(config)
}
