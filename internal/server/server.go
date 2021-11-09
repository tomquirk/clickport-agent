package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	rt "gitlab.com/clickport/clickport-agent/internal/clickport"
	cfg "gitlab.com/clickport/clickport-agent/internal/config"
)

type ClickportScriptDef struct {
	ID          string                        `json:"id"`
	Name        string                        `json:"name"`
	Description string                        `json:"description"`
	Parameters  []rt.ClickportScriptParameter `json:"parameters"`
}

type ScriptsPayloadData struct {
	Scripts []ClickportScriptDef `json:"scripts"`
}

type ScriptsPayload struct {
	Data ScriptsPayloadData `json:"data"`
}

func configToScriptsPayload(config *cfg.Config) *ScriptsPayload {
	var scripts []ClickportScriptDef
	for scriptID, script := range *config.ClickportScripts {
		var scriptDef = ClickportScriptDef{
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
		log.Printf("clickport::GET /scripts/execute")

		// Validate request came from us.
		executionRequest, err := ConstructExecutionRequest(config, w, r)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		rt.FulfillExecutionRequest(config.ClickportScripts, executionRequest)
	}
}

func handleListRequest(config *cfg.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("clickport::GET /scripts/list")

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
	r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong")
	}).Methods(http.MethodGet)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Clickport server running on port %s...", port)
	return http.ListenAndServe(fmt.Sprintf(":%s", port), r)
}
