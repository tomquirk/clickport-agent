package runtime

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

type RuntimeScriptParameter struct {
	ID          string
	Name        string
	Type        string
	Description string
}

type RuntimeScriptArgument struct {
	ParameterID string `json:"parameter_id"`
	Value       string `json:"value"` // TODO consider supporting multiple types, interface{}
}

type RuntimeScript struct {
	Description string
	Parameters  []RuntimeScriptParameter
	Script      []string
}

type RuntimeScripts map[string]RuntimeScript

type RuntimeScriptRequest struct {
	RuntimeResponseURL string                  `json:"runtime_response_url"`
	ScriptID           string                  `json:"script_id"`
	Arguments          []RuntimeScriptArgument `json:"arguments"`
}

// TODO(implement)
func validateArguments(runtimeScript RuntimeScript, req *RuntimeScriptRequest) ([]RuntimeScriptArgument, error) {
	return (*req).Arguments, nil
}

func buildArguments(runtimeScript RuntimeScript, req *RuntimeScriptRequest) (*[]string, error) {
	validArgs, err := validateArguments(runtimeScript, req)
	if err != nil {
		return nil, err
	}

	var args = make([]string, len(validArgs))
	for i, validArg := range validArgs {

		// TODO(sec) allows for arbitrary command injection via Value :/
		args[i] = fmt.Sprintf("-%s=%s", validArg.ParameterID, validArg.Value)
	}

	fmt.Printf("Args: %v", args)
	return &args, nil
}

func executeRuntimeScript(runtimeScript RuntimeScript, req *RuntimeScriptRequest) {
	for _, script := range runtimeScript.Script {
		fmt.Printf("Running `%s`\n", script)

		args, err := buildArguments(runtimeScript, req)
		if err != nil {
			log.Fatal(err)
		}

		cmd := exec.Command(script, *args...)
		cmd.Env = append(os.Environ(),
			// TODO(sec) allows for arbitrary command injection :/
			fmt.Sprintf("RUNTIME_RESPONSE_URL=%s", (*req).RuntimeResponseURL),
		)
		cmd.Stdout = os.Stdout

		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
	}
}

func FulfillRuntimeScriptRequest(runtimeScripts *RuntimeScripts, req *RuntimeScriptRequest) {
	fmt.Printf("got %v\n", *req)

	runtimeScript := (*runtimeScripts)[(*req).ScriptID]
	executeRuntimeScript(runtimeScript, req)
}
