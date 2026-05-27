package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/shared"
)

func main() {
	var prompt string
	flag.StringVar(&prompt, "p", "", "Prompt to send to LLM")
	flag.Parse()

	if prompt == "" {
		panic("Prompt must not be empty")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	baseUrl := os.Getenv("OPENROUTER_BASE_URL")
	if baseUrl == "" {
		baseUrl = "https://openrouter.ai/api/v1"
	}

	if apiKey == "" {
		panic("Env variable OPENROUTER_API_KEY not found")
	}

	readFunction := openai.ChatCompletionToolUnionParam{
		OfFunction: &openai.ChatCompletionFunctionToolParam{
			Function: shared.FunctionDefinitionParam{
				Name:        "Read",
				Description: param.NewOpt("Read and return the contents of a file"),
				Parameters: shared.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"file_path": map[string]any{
							"type":        "string",
							"description": "The path to the file to read",
						},
					},
				},
			},
		},
	}

	client := openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseUrl))

	messages := []openai.ChatCompletionMessageParamUnion{
		{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: openai.String(prompt),
				},
			},
		},
	}

	is_finished := false
	for !is_finished {
		resp, err := client.Chat.Completions.New(context.Background(),
			openai.ChatCompletionNewParams{
				Model:    "anthropic/claude-haiku-4.5",
				Messages: messages,
				Tools:    []openai.ChatCompletionToolUnionParam{readFunction},
			},
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		if len(resp.Choices) == 0 {
			panic("No choices in response")
		}

		resp_message := resp.Choices[0].Message

		is_no_tool_calls := len(resp_message.ToolCalls) == 0
		if is_no_tool_calls {
			content := resp_message.Content
			fmt.Println(string(content))
			is_finished = true
		} else {
			toolCall := resp.Choices[0].Message.ToolCalls[0]
			args := toolCall.Function.Arguments

			var argsMap map[string]any
			err := json.Unmarshal([]byte(args), &argsMap)
			if err != nil {
				panic(err)
			}

			filePath := argsMap["file_path"].(string)
			contents, err := os.ReadFile(filePath)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(contents))

			toolMessage := openai.ChatCompletionMessageParamUnion{
				OfTool: &openai.ChatCompletionToolMessageParam{
					ToolCallID: toolCall.ID,
					Content: openai.ChatCompletionToolMessageParamContentUnion{
						OfString: openai.String(string(contents)),
					},
				},
			}
			messages = append(messages, toolMessage)
			fmt.Fprintln(os.Stderr, "Messages: %q", messages)
		}
	}

	fmt.Fprintln(os.Stderr, "Finished")
}
