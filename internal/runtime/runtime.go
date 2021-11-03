package runtime

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
)

type RuntimeScriptParameter struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Placeholder string `json:"placeholder"`
}

type RuntimeScriptArgument struct {
	ParameterID string `json:"parameter_id"`
	Value       string `json:"value"` // TODO consider supporting multiple types, interface{}
}

type RuntimeScript struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Parameters  []RuntimeScriptParameter `json:"parameters"`
	Script      []string                 `json:"script"`
}

type RuntimeScripts map[string]RuntimeScript

type ExecutionRequest struct {
	ScriptID      string                  `json:"script_id"`
	Arguments     []RuntimeScriptArgument `json:"arguments"`
	ResponseToken string                  `json:"response_token"`
}

const responseTokenEnvKey = "RESPONSE_TOKEN"

var argumentValueRegex = regexp.MustCompile("^[a-zA-Z0-9 ]{1,255}$")
var scriptIDRegex = regexp.MustCompile("^[a-z_]{1,50}$")

func validateScriptID(runtimeScripts *RuntimeScripts, scriptID string) (*RuntimeScript, error) {
	scriptIDValid := scriptIDRegex.MatchString(scriptID)
	if !scriptIDValid {
		return nil, errors.New("invalid script_id")
	}

	runtimeScript, ok := (*runtimeScripts)[scriptID]
	if !ok {
		return nil, errors.New("invalid script_id")
	}

	return &runtimeScript, nil
}

func validateArgument(runtimeScript *RuntimeScript, arg *RuntimeScriptArgument) (*RuntimeScriptArgument, error) {
	// Check ParameterID is valid.
	parameterValid := false
	for _, p := range runtimeScript.Parameters {
		if p.ID == (*arg).ParameterID {
			parameterValid = true
		}
	}
	if !parameterValid {
		return nil, errors.New("invalid argument")
	}

	// Check Value matches regex
	valueValid := argumentValueRegex.MatchString((*arg).Value)
	if !valueValid {
		return nil, errors.New("invalid argument")
	}

	return arg, nil
}

func validateArguments(runtimeScript *RuntimeScript, req *ExecutionRequest) ([]RuntimeScriptArgument, error) {
	for _, arg := range (*req).Arguments {
		if _, err := validateArgument(runtimeScript, &arg); err != nil {
			return nil, err
		}
	}

	return (*req).Arguments, nil
}

func buildArguments(runtimeScript *RuntimeScript, req *ExecutionRequest) (*[]string, error) {
	arguments := (*req).Arguments
	var cmdArgs = make([]string, len(arguments))
	for i, validArg := range arguments {
		cmdArgs[i] = fmt.Sprintf("-%s=%s", validArg.ParameterID, validArg.Value)
	}

	return &cmdArgs, nil
}

func validateExecutionRequest(runtimeScripts *RuntimeScripts, req *ExecutionRequest) (*RuntimeScript, error) {
	scriptId := (*req).ScriptID
	runtimeScript, err := validateScriptID(runtimeScripts, scriptId)
	if err != nil {
		return nil, err
	}

	if _, err = validateArguments(runtimeScript, req); err != nil {
		return nil, err
	}

	if (*req).ResponseToken == "" {
		return nil, errors.New("invalid response_token")
	}

	return runtimeScript, nil
}

func executeScript(script string, args *[]string, env []string) error {
	log.Printf("Running `%s`\n", script)

	cmd := exec.Command(script, *args...)
	cmd.Env = env
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

func FulfillExecutionRequest(runtimeScripts *RuntimeScripts, req *ExecutionRequest) error {
	runtimeScript, err := validateExecutionRequest(runtimeScripts, req)
	if err != nil {
		return err
	}

	args, err := buildArguments(runtimeScript, req)
	if err != nil {
		return err
	}

	env := append(os.Environ(),
		fmt.Sprintf("%s=%s", responseTokenEnvKey, (*req).ResponseToken),
	)

	for _, script := range runtimeScript.Script {
		if err := executeScript(script, args, env); err != nil {
			return err
		}
	}

	return nil
}
