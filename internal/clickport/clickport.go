package clickport

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"

	"github.com/google/shlex"
)

type ClickportScriptParameter struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Placeholder string `json:"placeholder"`
	Flag        string `json:"flag"`
}

type ClickportScriptArgument struct {
	ParameterID string `json:"parameter_id"`
	Value       string `json:"value"`
}

type ClickportScript struct {
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Parameters  []ClickportScriptParameter `json:"parameters"`
	Script      []string                   `json:"script"`
}

type ClickportScripts map[string]ClickportScript

type ExecutionRequest struct {
	ScriptID      string                    `json:"script_id"`
	Arguments     []ClickportScriptArgument `json:"arguments"`
	ResponseToken string                    `json:"response_token"`
}

const responseTokenEnvKey = "CLICKPORT_RESPONSE_TOKEN"

var (
	argumentValueRegex = regexp.MustCompile("^[a-zA-Z0-9 ]{1,255}$")
	scriptIDRegex      = regexp.MustCompile("^[a-z_]{1,50}$")

	errInvalidRequestArgument      = errors.New("invalid argument")
	errInvalidRequestScriptID      = errors.New("invalid script_id")
	errInvalidRequestResponseToken = errors.New("invalid response_token")
)

func validateScriptID(clickportScripts *ClickportScripts, scriptID string) (*ClickportScript, error) {
	scriptIDValid := scriptIDRegex.MatchString(scriptID)
	if !scriptIDValid {
		return nil, errInvalidRequestScriptID
	}

	clickportScript, ok := (*clickportScripts)[scriptID]
	if !ok {
		return nil, errInvalidRequestScriptID
	}

	return &clickportScript, nil
}

func validateArgument(clickportScript *ClickportScript, arg *ClickportScriptArgument) (*ClickportScriptArgument, error) {
	// Check ParameterID is valid.
	parameterValid := false
	for _, p := range clickportScript.Parameters {
		if p.ID == (*arg).ParameterID {
			parameterValid = true
		}
	}
	if !parameterValid {
		return nil, errInvalidRequestArgument
	}

	// Check Value matches regex
	valueValid := argumentValueRegex.MatchString((*arg).Value)
	if !valueValid {
		return nil, errInvalidRequestArgument
	}

	return arg, nil
}

func validateArguments(clickportScript *ClickportScript, req *ExecutionRequest) ([]ClickportScriptArgument, error) {
	for _, arg := range (*req).Arguments {
		if _, err := validateArgument(clickportScript, &arg); err != nil {
			return nil, err
		}
	}

	return (*req).Arguments, nil
}

func buildArguments(clickportScript *ClickportScript, req *ExecutionRequest) (*[]string, error) {
	arguments := (*req).Arguments

	var cmdArgs = make([]string, len(arguments))
	for i, arg := range arguments {

		// TODO append to ExecutionRequest earlier for efficiency
		var parameterFlag string
		for _, p := range clickportScript.Parameters {
			if p.ID == arg.ParameterID {
				parameterFlag = p.Flag
			}
		}

		cmdArgs[i] = fmt.Sprintf("%s=%s", parameterFlag, arg.Value)
	}

	return &cmdArgs, nil
}

func validateExecutionRequest(clickportScripts *ClickportScripts, req *ExecutionRequest) (*ClickportScript, error) {
	scriptId := (*req).ScriptID
	clickportScript, err := validateScriptID(clickportScripts, scriptId)
	if err != nil {
		return nil, err
	}

	if _, err = validateArguments(clickportScript, req); err != nil {
		return nil, err
	}

	if (*req).ResponseToken == "" {
		return nil, errInvalidRequestResponseToken
	}

	return clickportScript, nil
}

func executeScript(scriptName string, args *[]string, env []string) ([]byte, error) {
	cmd := exec.Command(scriptName, *args...)
	cmd.Env = env

	return cmd.CombinedOutput()
}

func FulfillExecutionRequest(clickportScripts *ClickportScripts, req *ExecutionRequest) error {
	clickportScript, err := validateExecutionRequest(clickportScripts, req)
	if err != nil {
		return err
	}

	args, err := buildArguments(clickportScript, req)
	if err != nil {
		return err
	}

	env := append(os.Environ(),
		fmt.Sprintf("%s=%s", responseTokenEnvKey, (*req).ResponseToken),
	)

	for idx, script := range clickportScript.Script {
		scriptTokens, err := shlex.Split(script)
		if err != nil {
			return err
		}
		scriptName := scriptTokens[0]
		argv := append(scriptTokens[1:], *args...)

		log.Printf("clickport::running script %d from script `%s`\n", idx, clickportScript.Name)

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
