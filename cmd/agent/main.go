package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"

	"gopkg.in/yaml.v2"
)

type RuntimeScriptParameter struct {
	Type        string
	Description string
}

type RuntimeScript struct {
	Description string
	Parameters  []RuntimeScriptParameter
	Script      []string
}

type RuntimeScripts map[string]RuntimeScript

// TODO add parameters
type RuntimeScriptPayload struct {
	RuntimeResponseURL string `json:"runtime_response_url"`
	ScriptID           string `json:"script_id"`
}

func loadScriptConfig() RuntimeScripts {
	yamlFile, err := ioutil.ReadFile("examples/example.yml")
	if err != nil {
		log.Fatalf("ReadFile: %v", err)
	}

	runtimeScripts := make(RuntimeScripts)

	err = yaml.Unmarshal(yamlFile, &runtimeScripts)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return runtimeScripts
}

func executeRuntimeScript(runtimeScript RuntimeScript, responseURL string) {
	for _, script := range runtimeScript.Script {
		fmt.Printf("Running `%s`\n", script)
		// TODO build options from RuntimeScript parameters and pass to command
		cmd := exec.Command(script)
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("RUNTIME_RESPONSE_URL=%s", responseURL),
		)
		cmd.Stdout = os.Stdout

		// TODO(next) run this command in a speciifc directory
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
	}
}

func fulfillRuntimeScriptRequest(runtimeScripts *RuntimeScripts, payload *RuntimeScriptPayload) {
	fmt.Printf("got %v\n", (*payload))
	runtimeScript := (*runtimeScripts)[(*payload).ScriptID]
	executeRuntimeScript(runtimeScript, (*payload).RuntimeResponseURL)
}

func handleRequest(runtimeScripts *RuntimeScripts) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload RuntimeScriptPayload
		// Try to decode the request body into the struct. If there is an error,
		// respond to the client with the error message and a 400 status code.
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fulfillRuntimeScriptRequest(runtimeScripts, &payload)

	}
}

func startServer(runtimeScripts *RuntimeScripts) {
	http.HandleFunc("/", handleRequest(runtimeScripts))
	fmt.Printf("Runtime server running on port 8080...\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	runtimeScripts := loadScriptConfig()

	startServer(&runtimeScripts)
}
