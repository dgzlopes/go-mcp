package mcp

import (
	"fmt"
	"strings"
)

type PromptMessage struct {
	Role    Role    `json:"role"`
	Content Content `json:"content"`
}

type Prompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
	Template    string
}

type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

type GetPromptRequest struct {
	Method string `json:"method"`
	Params struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments,omitempty"`
	} `json:"params"`
}

type GetPromptResult struct {
	Description string          `json:"description,omitempty"`
	Messages    []PromptMessage `json:"messages"`
}

func (p *Prompt) Execute(args map[string]string) (string, error) {
	for _, arg := range p.Arguments {
		if arg.Required {
			if _, ok := args[arg.Name]; !ok {
				return "", fmt.Errorf("missing required argument: %s", arg.Name)
			}
		}
	}

	result := p.Template
	for name, value := range args {
		result = strings.ReplaceAll(result, "{"+name+"}", value)
	}

	return result, nil
}

func (p *Prompt) ValidateArguments(args map[string]string) error {
	for _, arg := range p.Arguments {
		if arg.Required {
			if _, ok := args[arg.Name]; !ok {
				return fmt.Errorf("missing required argument: %s", arg.Name)
			}
		}
	}
	return nil
}
