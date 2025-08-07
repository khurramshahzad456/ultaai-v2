package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"ultahost-ai-assistant/internal/config"
	"ultahost-ai-assistant/pkg/models"

	"github.com/sashabaranov/go-openai"
)

var singleCommandFunction = openai.FunctionDefinition{
	Name:        "run_approved_command",
	Description: "Generate a safe Linux command with explanation",
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "Shell command to execute",
			},
			"explanation": map[string]any{
				"type":        "string",
				"description": "What this command does",
			},
		},
		"required": []string{"command", "explanation"},
	},
}

var multiStepFunction = openai.FunctionDefinition{
	Name:        "execute_steps",
	Description: "Run multiple safe server commands step-by-step",
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"description": map[string]any{
				"type":        "string",
				"description": "Task overview",
			},
			"steps": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"command":     map[string]any{"type": "string"},
						"explanation": map[string]any{"type": "string"},
					},
					"required": []string{"command"},
				},
			},
		},
		"required": []string{"description", "steps"},
	},
}

func ExtractCommandFromPrompt(prompt string) (*models.AICommand, error) {
	client := openai.NewClient(config.Get().OpenAIKey)

	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "You're an assistant that returns only safe, pre-approved Linux commands.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Functions:    []openai.FunctionDefinition{singleCommandFunction},
		FunctionCall: openai.FunctionCall{Name: "run_approved_command"},
	})
	if err != nil {
		return nil, fmt.Errorf("OpenAI error: %w", err)
	}

	var out models.AICommand
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.FunctionCall.Arguments), &out); err != nil {
		return nil, fmt.Errorf("function decode error: %w", err)
	}
	return &out, nil
}

func ExtractStepsFromPrompt(prompt string) (*models.StructuredAIResponse, error) {
	client := openai.NewClient(config.Get().OpenAIKey)

	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: "You're an AI assistant that outputs only safe commands in a step-by-step format."},
			{Role: "user", Content: prompt},
		},
		Functions:    []openai.FunctionDefinition{multiStepFunction},
		FunctionCall: openai.FunctionCall{Name: "execute_steps"},
	})
	if err != nil {
		return nil, fmt.Errorf("OpenAI error: %w", err)
	}

	var out models.StructuredAIResponse
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.FunctionCall.Arguments), &out); err != nil {
		return nil, fmt.Errorf("function decode error: %w", err)
	}
	return &out, nil
}
