package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func getMockScript() *RuntimeScript {
	return &RuntimeScript{
		Name:        "test",
		Description: "test",
		Parameters: []RuntimeScriptParameter{
			{
				ID: "valid_param",
			},
			{
				ID: "valid_param_2",
			},
		},
	}
}

func TestValidateResponseURL(t *testing.T) {
	tables := []struct {
		testURL string
		valid   bool
	}{
		{"https://runtimehq.com/valid/url", true},
		{"; rm -Rf .", false},
		{"runtimehq.com/valid/url", false},
	}

	for _, table := range tables {
		_, err := validateResponseURL(table.testURL)
		assert.Equal(t, err == nil, table.valid)
	}
}

func TestValidateArgument(t *testing.T) {
	tables := []struct {
		testParameterID string
		testValue       string
		valid           bool
	}{
		{"valid_param", "arg", true},
		{"valid_param", "123", true},
		{"valid_param", "arg123", true},
		{"valid_param", "arg 123", true},
		{"valid_param", "arg; rm -rf .", false},
		{"valid_param", "arg\\;", false},
		{"valid_param", "arg_abc", false},
		{"valid_param", "111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111", true},
		{"valid_param", "1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111", false},
		{"invalid_param", "arg", false},
	}

	mockScript := getMockScript()

	for _, table := range tables {
		arg := RuntimeScriptArgument{
			ParameterID: table.testParameterID,
			Value:       table.testValue,
		}
		_, err := validateArgument(mockScript, &arg)
		assert.Equal(t, err == nil, table.valid)
	}
}

func TestBuildArguments(t *testing.T) {
	mockScript := getMockScript()
	req := ExecutionRequest{
		Arguments: []RuntimeScriptArgument{
			{Value: "asd", ParameterID: "valid_param"},
			{Value: "asd2", ParameterID: "valid_param_2"},
		},
	}

	cmdArgs, err := buildArguments(mockScript, &req)
	assert.Equal(t, *cmdArgs, []string{"-valid_param=asd", "-valid_param_2=asd2"})
	assert.Nil(t, err)
}

func TestFulfillExecutionRequestInvalidArgumentValue(t *testing.T) {
	mockScript := getMockScript()
	mockScripts := RuntimeScripts{
		"test": *mockScript,
	}

	req := ExecutionRequest{
		ScriptID:           "test",
		RuntimeResponseURL: "https://runtimehq.com/path",
		Arguments: []RuntimeScriptArgument{
			{Value: "asd;", ParameterID: "valid_param"}, // bad arg
			{Value: "asd2", ParameterID: "valid_param_2"},
		},
	}

	err := FulfillExecutionRequest(&mockScripts, &req)
	assert.EqualError(t, err, "invalid argument")
}

func TestFulfillExecutionRequestInvalidArgumentParameterID(t *testing.T) {
	mockScript := getMockScript()
	mockScripts := RuntimeScripts{
		"test": *mockScript,
	}
	req := ExecutionRequest{
		ScriptID:           "test",
		RuntimeResponseURL: "https://runtimehq.com/path",
		Arguments: []RuntimeScriptArgument{
			{Value: "asd", ParameterID: "invalid_param"}, // bad arg
			{Value: "asd2", ParameterID: "valid_param_2"},
		},
	}

	err := FulfillExecutionRequest(&mockScripts, &req)
	assert.EqualError(t, err, "invalid argument")
}

func TestFulfillExecutionRequestInvalidScriptID(t *testing.T) {
	mockScript := getMockScript()
	mockScripts := RuntimeScripts{
		"test": *mockScript,
	}
	req := ExecutionRequest{
		ScriptID:           "baddybad",
		RuntimeResponseURL: "https://runtimehq.com/path",
		Arguments: []RuntimeScriptArgument{
			{Value: "asd", ParameterID: "valid_param"}, // bad arg
			{Value: "asd2", ParameterID: "valid_param_2"},
		},
	}

	err := FulfillExecutionRequest(&mockScripts, &req)
	assert.EqualError(t, err, "invalid script_id")
}

func TestFulfillExecutionRequestInvalidResponseURL(t *testing.T) {
	mockScript := getMockScript()
	mockScripts := RuntimeScripts{
		"test": *mockScript,
	}
	req := ExecutionRequest{
		ScriptID:           "test",
		RuntimeResponseURL: "bad url",
		Arguments: []RuntimeScriptArgument{
			{Value: "asd", ParameterID: "valid_param"}, // bad arg
			{Value: "asd2", ParameterID: "valid_param_2"},
		},
	}

	err := FulfillExecutionRequest(&mockScripts, &req)
	assert.EqualError(t, err, "invalid runtime_response_url")
}
