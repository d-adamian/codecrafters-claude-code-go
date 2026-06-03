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
)

func logJson(value any) {
	jsonData, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(os.Stderr, string(jsonData))
}

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

	tools := make([]openai.ChatCompletionToolUnionParam, 0)
	for _, tool := range AllFunctions {
		tools = append(tools, asToolParam(tool))
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

	fmt.Fprintln(os.Stderr, "Initial message")
	logJson(messages)

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
		fmt.Fprintln(os.Stderr, "LLM response")
		logJson(resp_message)
		toolCalls := make([]openai.ChatCompletionMessageToolCallUnionParam, 0)
		for _, toolCall := range resp_message.ToolCalls {
			toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallUnionParam{
				OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
					ID: toolCall.ID,
					Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
						Name:      toolCall.Function.Name,
						Arguments: toolCall.Function.Arguments,
					},
				},
			})
		}
		messages = append(messages, openai.ChatCompletionMessageParamUnion{
			OfAssistant: &openai.ChatCompletionAssistantMessageParam{
				Content:   openai.ChatCompletionAssistantMessageParamContentUnion{OfString: param.NewOpt(resp_message.Content)},
				ToolCalls: toolCalls,
			},
		})
		fmt.Fprintln(os.Stderr, "All messages")
		logJson(messages)

		is_no_tool_calls := len(resp_message.ToolCalls) == 0
		if is_no_tool_calls {
			content := resp_message.Content
			fmt.Println(string(content))
			is_finished = true
		} else {
			toolCall := resp.Choices[0].Message.ToolCalls[0]
			function := mustFindFunctionByName(toolCall.Function.Name)
			args := toolCall.Function.Arguments

			var params map[string]any
			err := json.Unmarshal([]byte(args), &params)
			if err != nil {
				panic(err)
			}

			result := function.Function(params)

			toolMessage := openai.ChatCompletionMessageParamUnion{
				OfTool: &openai.ChatCompletionToolMessageParam{
					ToolCallID: toolCall.ID,
					Content: openai.ChatCompletionToolMessageParamContentUnion{
						OfString: openai.String(result),
					},
				},
			}
			messages = append(messages, toolMessage)
			fmt.Fprintln(os.Stderr, "All messages")
			logJson(messages)
		}
	}

	fmt.Fprintln(os.Stderr, "Finished")
}
