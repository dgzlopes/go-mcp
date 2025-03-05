package prompts

import (
	"encoding/json"
	"fmt"
	"strings"

	"go-mcp/pkg/mcp/protocol"
)

type PromptMessage struct {
	Role    protocol.Role    `json:"role"`
	Content protocol.Content `json:"content"`
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

func (pm *PromptMessage) UnmarshalJSON(data []byte) error {
	type Alias PromptMessage
	aux := struct {
		*Alias
		Content json.RawMessage `json:"content"`
	}{
		Alias: (*Alias)(pm),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var contentMap map[string]interface{}
	if err := json.Unmarshal(aux.Content, &contentMap); err != nil {
		return err
	}

	contentType, ok := contentMap["type"].(string)
	if !ok {
		return fmt.Errorf("content type not found or invalid")
	}

	switch protocol.ContentType(contentType) {
	case protocol.ContentTypeText:
		var textContent protocol.TextContent
		if err := json.Unmarshal(aux.Content, &textContent); err != nil {
			return err
		}
		pm.Content = textContent
	case protocol.ContentTypeImage:
		var imageContent protocol.ImageContent
		if err := json.Unmarshal(aux.Content, &imageContent); err != nil {
			return err
		}
		pm.Content = imageContent
	case protocol.ContentTypeResource:
		var resourceContent protocol.EmbeddedResource
		if err := json.Unmarshal(aux.Content, &resourceContent); err != nil {
			return err
		}
		pm.Content = resourceContent
	default:
		return fmt.Errorf("unknown content type: %s", contentType)
	}

	return nil
}
