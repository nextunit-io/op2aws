package awsvault

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	outputReturnValue *string

	outputCallCount  int
	commandCallCount int

	commandInput []string

	testCases []testCaseStruct = []testCaseStruct{
		{
			call:               "GetAccessKeyId",
			expectedCliCommand: []string{"op", "read", "op://test-vault/test-item/aws_access_key_id"},
		},
		{
			call:               "GetSecretAccessKey",
			expectedCliCommand: []string{"op", "read", "op://test-vault/test-item/aws_secret_access_key"},
		},
		{
			call:               "GetOTP",
			expectedCliCommand: []string{"op", "item", "get", "test-item", "--vault", "test-vault", "--otp"},
		},
	}
)

type commandLineClientTest struct {
	CommandInterface
}

type cmdClientTest struct {
	CmdInterface
}

type testCaseStruct struct {
	call               string
	expectedCliCommand []string
}

func (cmdClientTest) Output() ([]byte, error) {
	outputCallCount++
	if outputReturnValue == nil {
		return nil, fmt.Errorf("Test error")
	}

	return []byte(*outputReturnValue), nil
}

func (commandLineClientTest) Command(name string, arg ...string) CmdInterface {
	commandCallCount++
	commandInput = append([]string{name}, arg...)

	return &cmdClientTest{}
}

func setupTestCases() {
	// Set defaults
	outputReturnValueString := "test-value"

	outputReturnValue = &outputReturnValueString

	outputCallCount = 0
	commandCallCount = 0

	commandInput = []string{}
}

func TestGetAccessKeyId(t *testing.T) {
	for _, v := range testCases {
		t.Run(v.call, func(t *testing.T) {
			setupTestCases()
			t.Helper()

			vault := NewOnePasswordVault(&commandLineClientTest{}, "test-vault", "test-item")

			method := reflect.ValueOf(vault).MethodByName(v.call)
			returnValue := method.Call(nil)

			value := returnValue[0].Interface().(string)
			err := returnValue[1].Interface()

			assert.Nilf(t, err, "%s should run without problems", v.call)
			assert.Equal(t, "test-value", value)

			assert.Equal(t, 1, commandCallCount)
			assert.Equal(t, 1, outputCallCount)

			assert.Equal(t, v.expectedCliCommand, commandInput)
		})
	}
}

func TestGetVaults(t *testing.T) {
	setupTestCases()
	t.Helper()

	expectedOutput := []OpVault{}
	i := 1
	for i < 10 {
		i++
		expectedOutput = append(expectedOutput, OpVault{
			Id:      fmt.Sprintf("test-id-%d", i),
			Name:    fmt.Sprintf("test-name-%d", i),
			Version: i,
		})
	}

	outputByte, err := json.Marshal(expectedOutput)
	outputString := string(outputByte)
	outputReturnValue = &outputString

	vaults, err := GetVaults(&commandLineClientTest{})
	assert.Nil(t, err, "No errors expected")
	assert.Equal(t, expectedOutput, vaults)
	assert.Equal(t, []string{"op", "vault", "list", "--format", "json"}, commandInput)
}

func TestGetItems(t *testing.T) {
	setupTestCases()
	t.Helper()

	expectedOutput := []OpItem{}
	i := 1
	for i < 10 {
		i++
		expectedOutput = append(expectedOutput, OpItem{
			Id:      fmt.Sprintf("test-id-%d", i),
			Version: i,
		})
	}

	outputByte, err := json.Marshal(expectedOutput)
	outputString := string(outputByte)
	outputReturnValue = &outputString

	items, err := GetItems(&commandLineClientTest{}, "vault-test")
	assert.Nil(t, err, "No errors expected")
	assert.Equal(t, expectedOutput, items)
	assert.Equal(t, []string{"op", "item", "list", "--vault", "vault-test", "--format", "json"}, commandInput)
}

func TestGetEntries(t *testing.T) {
	setupTestCases()
	t.Helper()

	expectedOutput := []OpEntry{}
	i := 1
	for i < 10 {
		i++
		expectedOutput = append(expectedOutput, OpEntry{
			Id: fmt.Sprintf("test-id-%d", i),
		})
	}

	outputByte, err := json.Marshal(expectedOutput)
	outputString := fmt.Sprintf("{\"fields\":%s}", string(outputByte))
	outputReturnValue = &outputString

	items, err := GetEntries(&commandLineClientTest{}, "vault-test", "item-test")
	assert.Nil(t, err, "No errors expected")
	assert.Equal(t, expectedOutput, items)
	assert.Equal(t, []string{"op", "item", "get", "item-test", "--vault", "vault-test", "--format", "json"}, commandInput)
}
