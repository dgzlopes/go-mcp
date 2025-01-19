package mcp_test

import (
	"encoding/json"
	"go-mcp/pkg/mcp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptMessage(t *testing.T) {
	t.Run("serializes basic text message", func(t *testing.T) {
		msg := mcp.PromptMessage{
			Role: mcp.RoleUser,
			Content: mcp.TextContent{
				Type: string(mcp.ContentTypeText),
				Text: "Hello, world!",
			},
		}

		data, err := json.Marshal(msg)
		require.NoError(t, err)

		expected := `{
			"role": "user",
			"content": {
				"type": "text",
				"text": "Hello, world!"
			}
		}`
		assert.JSONEq(t, expected, string(data))
	})

	t.Run("serializes message with annotations", func(t *testing.T) {
		msg := mcp.PromptMessage{
			Role: mcp.RoleAssistant,
			Content: mcp.TextContent{
				Type: string(mcp.ContentTypeText),
				Text: "Response",
				Annotations: &mcp.Annotation{
					Audience: []mcp.Role{mcp.RoleUser},
					Priority: 0.8,
				},
			},
		}

		data, err := json.Marshal(msg)
		require.NoError(t, err)

		expected := `{
			"role": "assistant",
			"content": {
				"type": "text",
				"text": "Response",
				"annotations": {
					"audience": ["user"],
					"priority": 0.8
				}
			}
		}`
		assert.JSONEq(t, expected, string(data))
	})

	t.Run("deserializes from JSON correctly", func(t *testing.T) {
		input := `{
			"role": "user",
			"content": {
				"type": "text",
				"text": "Hello, world!"
			}
		}`

		var msg mcp.PromptMessage
		err := json.Unmarshal([]byte(input), &msg)
		require.NoError(t, err)

		assert.Equal(t, mcp.RoleUser, msg.Role)
		textContent, ok := msg.Content.(mcp.TextContent)
		require.True(t, ok)
		assert.Equal(t, "Hello, world!", textContent.Text)
	})
}

func TestToolCall(t *testing.T) {
	t.Run("deserializes basic tool call", func(t *testing.T) {
		input := `{
			"name": "get_weather",
			"arguments": {
				"location": "London",
				"units": "celsius"
			}
		}`

		var call mcp.ToolCall
		err := json.Unmarshal([]byte(input), &call)
		require.NoError(t, err)

		assert.Equal(t, "get_weather", call.Name)
		assert.Equal(t, "London", call.Arguments["location"])
		assert.Equal(t, "celsius", call.Arguments["units"])
	})

	t.Run("handles missing arguments", func(t *testing.T) {
		input := `{
			"name": "list_files"
		}`

		var call mcp.ToolCall
		err := json.Unmarshal([]byte(input), &call)
		require.NoError(t, err)

		assert.Equal(t, "list_files", call.Name)
		assert.Nil(t, call.Arguments)
	})
}

func TestPrompt(t *testing.T) {
	t.Run("validates required arguments", func(t *testing.T) {
		prompt := mcp.Prompt{
			Name: "greet",
			Arguments: []mcp.PromptArgument{
				{
					Name:        "name",
					Description: "Person to greet",
					Required:    true,
				},
			},
		}

		err := prompt.ValidateArguments(map[string]string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required argument: name")
	})

	t.Run("executes template with valid arguments", func(t *testing.T) {
		prompt := mcp.Prompt{
			Name: "greet",
			Arguments: []mcp.PromptArgument{
				{
					Name:        "name",
					Description: "Person to greet",
					Required:    true,
				},
			},
			Template: "Hello, {name}!",
		}

		result, err := prompt.Execute(map[string]string{
			"name": "Alice",
		})
		require.NoError(t, err)
		assert.Equal(t, "Hello, Alice!", result)
	})
}

func TestTool(t *testing.T) {
	t.Run("validates argument types", func(t *testing.T) {
		tool := mcp.Tool{
			Name:        "add_numbers",
			Description: "Add two numbers together",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"a": map[string]interface{}{"type": "number"},
					"b": map[string]interface{}{"type": "number"},
				},
				"required": []string{"a", "b"},
			},
		}

		_, err := tool.ValidateAndExecute(map[string]interface{}{
			"a": "not a number",
			"b": 3,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid argument a")
	})

	t.Run("executes with valid arguments", func(t *testing.T) {
		tool := mcp.Tool{
			Name:        "add_numbers",
			Description: "Add two numbers together",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"a": map[string]interface{}{"type": "number"},
					"b": map[string]interface{}{"type": "number"},
				},
				"required": []string{"a", "b"},
			},
		}

		result, err := tool.ValidateAndExecute(map[string]interface{}{
			"a": 5,
			"b": 3,
		})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Content, 1)
	})
}

func TestMessageLifecycle(t *testing.T) {
	t.Run("full prompt execution flow", func(t *testing.T) {

		prompt := mcp.Prompt{
			Name: "calculate",
			Arguments: []mcp.PromptArgument{
				{
					Name:     "operation",
					Required: true,
				},
				{
					Name:     "values",
					Required: true,
				},
			},
			Template: "Please {operation} these values: {values}",
		}

		result, err := prompt.Execute(map[string]string{
			"operation": "add",
			"values":    "1, 2, 3",
		})
		require.NoError(t, err)

		msg := mcp.PromptMessage{
			Role: mcp.RoleUser,
			Content: mcp.TextContent{
				Type: string(mcp.ContentTypeText),
				Text: result,
			},
		}

		data, err := json.Marshal(msg)
		require.NoError(t, err)

		var decoded mcp.PromptMessage
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, msg.Role, decoded.Role)
		assert.Equal(t,
			msg.Content.(mcp.TextContent).Text,
			decoded.Content.(mcp.TextContent).Text)
	})
}
