package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	cfg "github.com/runtime-hq/runtime-agent/internal/config"
	rt "github.com/runtime-hq/runtime-agent/internal/runtime"
)

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

func handleExecuteRequest(config *cfg.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validate request came from us.
		executionRequest, err := ConstructExecutionRequest(config, w, r)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		rt.FulfillExecutionRequest(config.RuntimeScripts, executionRequest)
	}
}

func handleListRequest(config *cfg.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := VerifyRequestSignature(config, w, r)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		scriptsPayload := configToScriptsPayload(config)
		json.NewEncoder(w).Encode(*scriptsPayload)
	}
}

func Start(config *cfg.Config) error {
	r := mux.NewRouter()

	r.HandleFunc("/scripts/execute", handleExecuteRequest(config)).Methods(http.MethodPost)
	r.HandleFunc("/scripts/list", handleListRequest(config)).Methods(http.MethodGet)

	log.Println("Server running on port 8080...")
	return http.ListenAndServe(":8080", r)
}
