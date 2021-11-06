package runtime

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"

	"github.com/google/shlex"
)

type RuntimeScriptParameter struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Placeholder string `json:"placeholder"`
	Flag        string `json:"flag"`
}

type RuntimeScriptArgument struct {
	ParameterID string `json:"parameter_id"`
	Value       string `json:"value"`
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
	for i, arg := range arguments {

		// TODO append to ExecutionRequest earlier for efficiency
		var parameterFlag string
		for _, p := range runtimeScript.Parameters {
			if p.ID == arg.ParameterID {
				parameterFlag = p.Flag
			}
		}

		cmdArgs[i] = fmt.Sprintf("%s=%s", parameterFlag, arg.Value)
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

func executeScript(scriptName string, args *[]string, env []string) ([]byte, error) {
	cmd := exec.Command(scriptName, *args...)
	cmd.Env = env

	return cmd.CombinedOutput()
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

	for idx, script := range runtimeScript.Script {
		scriptTokens, err := shlex.Split(script)
		if err != nil {
			return err
		}
		scriptName := scriptTokens[0]
		argv := append(scriptTokens[1:], *args...)

		log.Printf("runtime::running script %d from script `%s`\n", idx, runtimeScript.Name)

		out, err := executeScript(scriptName, &argv, env)
		if err != nil {
			log.Printf("script::stderr::%s", err)
		}
		if out != nil {
			log.Printf("script::stdout::%s", string(out))
		}
	}

	return nil
}
