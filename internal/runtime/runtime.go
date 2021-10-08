package runtime

import (
	"fmt"
	"log"
	"os"
	"os/exec"
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

func FulfillRuntimeScriptRequest(runtimeScripts *RuntimeScripts, payload *RuntimeScriptPayload) {
	fmt.Printf("got %v\n", (*payload))

	runtimeScript := (*runtimeScripts)[(*payload).ScriptID]
	executeRuntimeScript(runtimeScript, (*payload).RuntimeResponseURL)
}
