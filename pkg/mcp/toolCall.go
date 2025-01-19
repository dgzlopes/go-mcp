package mcp

import (
	"fmt"
)

type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type CallToolRequest struct {
	Method string   `json:"method"`
	Params ToolCall `json:"params"`
}

func (t *Tool) ValidateAndExecute(args map[string]interface{}) (*CallToolResult, error) {
	if err := t.validateArguments(args); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	//TODO: Mock Result
	return &CallToolResult{
		Content: []Content{
			TextContent{
				Type: string(ContentTypeText),
				Text: "Tool execution result",
			},
		},
	}, nil
}

func (t *Tool) validateArguments(args map[string]interface{}) error {
	schema := t.InputSchema

	if required, ok := schema["required"].([]string); ok {
		for _, field := range required {
			if _, exists := args[field]; !exists {
				return fmt.Errorf("missing required field: %s", field)
			}
		}
	}

	if props, ok := schema["properties"].(map[string]interface{}); ok {
		for name, value := range args {
			if propSchema, exists := props[name]; exists {
				if err := validateType(propSchema.(map[string]interface{}), value); err != nil {
					return fmt.Errorf("invalid argument %s: %w", name, err)
				}
			}
		}
	}

	return nil
}

func validateType(schema map[string]interface{}, value interface{}) error {
	expectedType, ok := schema["type"].(string)
	if !ok {
		return fmt.Errorf("schema missing type")
	}

	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "number":
		switch v := value.(type) {
		case float64, float32, int, int32, int64:
			// These are acceptable number types
		default:
			return fmt.Errorf("expected number, got %T", v)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	default:
		return fmt.Errorf("unsupported type: %s", expectedType)
	}

	return nil
}
