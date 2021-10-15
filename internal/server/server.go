package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	cfg "github.com/runtime-hq/runtime-agent/internal/config"
	rt "github.com/runtime-hq/runtime-agent/internal/runtime"
)

func handleRequest(config *cfg.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var runtimeReq rt.RuntimeScriptRequest
		// Try to decode the request body into the struct. If there is an error,
		// respond to the client with the error message and a 400 status code.
		err := json.NewDecoder(r.Body).Decode(&runtimeReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		rt.FulfillRuntimeScriptRequest(config.RuntimeScripts, &runtimeReq)
	}
}

type RuntimeScriptDef struct {
	ID          string                      `json:"id"`
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	Parameters  []rt.RuntimeScriptParameter `json:"parameters"`
}

type ScriptsPayloadData struct {
	Scripts []RuntimeScriptDef `json:"scripts"`
}

type ScriptsPayload struct {
	Data ScriptsPayloadData `json:"data"`
}

func configToScriptsPayload(config *cfg.Config) *ScriptsPayload {
	var scripts []RuntimeScriptDef
	for scriptID, script := range *config.RuntimeScripts {
		var scriptDef = RuntimeScriptDef{
			ID:          scriptID,
			Name:        script.Name,
			Description: script.Description,
			Parameters:  script.Parameters,
		}
		scripts = append(scripts, scriptDef)
	}

	payload := ScriptsPayload{
		Data: ScriptsPayloadData{
			Scripts: scripts,
		},
	}

	return &payload
}

func handleScriptsRequest(config *cfg.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		scriptsPayload := configToScriptsPayload(config)
		json.NewEncoder(w).Encode(*scriptsPayload)
	}
}

func Start(config *cfg.Config) {
	http.HandleFunc("/", handleRequest(config))
	http.HandleFunc("/scripts", handleScriptsRequest(config))

	fmt.Printf("Runtime server running on port 8080...\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
