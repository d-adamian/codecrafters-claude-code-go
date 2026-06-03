package main

import (
	"os"
	"os/exec"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/shared"
)

type FuncParam struct {
	Name        string
	Type        string
	Description string
	Required    bool
}

type Function struct {
	Name        string
	Description string
	Params      []FuncParam
	Function    func(params map[string]any) string
}

const FilePathParam = "file_path"
const ContentsParam = "content"

var ReadFunction = Function{
	Name:        "Read",
	Description: "Read and return the contents of a file",
	Params: []FuncParam{
		{Name: "file_path", Type: "string", Description: "The path to the file to read", Required: true},
	},
	Function: doRead,
}

var WriteFunction = Function{
	Name:        "Write",
	Description: "Write the contents to a file",
	Params: []FuncParam{
		{Name: "file_path", Type: "string", Description: "The path of the file to write to", Required: true},
		{Name: "content", Type: "string", Description: "The content to write to the file", Required: true},
	},
	Function: doWrite,
}

var BashFunction = Function{
	Name:        "Bash",
	Description: "Execute a shell command",
	Params: []FuncParam{
		{Name: "command", Type: "string", Description: "The command to execute", Required: true},
	},
	Function: doBash,
}

var AllFunctions = []Function{ReadFunction, WriteFunction, BashFunction}

func doRead(params map[string]any) string {
	filePath := params[FilePathParam].(string)
	contents, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	return string(contents)
}

func doWrite(params map[string]any) string {
	filePath := params[FilePathParam].(string)
	contents := params[ContentsParam].(string)
	err := os.WriteFile(filePath, []byte(contents), 0644)
	if err != nil {
		panic(err)
	}
	return "File written successfully"
}

func doBash(params map[string]any) string {
	command := params["command"].(string)
	output, err := exec.Command("bash", "-c", command).CombinedOutput()
	if err != nil {
		panic(err)
	}
	return string(output)
}

func asToolParam(f Function) openai.ChatCompletionToolUnionParam {
	requiredParams := make([]string, 0)
	properties := make(map[string]any)
	for _, p := range f.Params {
		if p.Required {
			requiredParams = append(requiredParams, p.Name)
		}
		properties[p.Name] = map[string]any{
			"type":        p.Type,
			"description": p.Description,
		}
	}
	return openai.ChatCompletionToolUnionParam{
		OfFunction: &openai.ChatCompletionFunctionToolParam{
			Function: shared.FunctionDefinitionParam{
				Name:        f.Name,
				Description: param.NewOpt(f.Description),
				Parameters: shared.FunctionParameters{
					"type":       "object",
					"required":   requiredParams,
					"properties": properties,
				},
			},
		},
	}
}

func mustFindFunctionByName(name string) Function {
	for _, f := range AllFunctions {
		if f.Name == name {
			return f
		}
	}
	panic("function not found")
}
