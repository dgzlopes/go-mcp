package prompts_test

import (
	"encoding/json"
	"testing"

	"go-mcp/pkg/mcp/prompts"
	"go-mcp/pkg/mcp/protocol"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptMessage(t *testing.T) {
	t.Run("serializes basic text message", func(t *testing.T) {
		msg := prompts.PromptMessage{
			Role: protocol.RoleUser,
			Content: &protocol.TextContent{
				Type: string(protocol.ContentTypeText),
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
		msg := prompts.PromptMessage{
			Role: protocol.RoleAssistant,
			Content: protocol.TextContent{
				Type: string(protocol.ContentTypeText),
				Text: "Response",
				Annotations: &protocol.Annotation{
					Audience: []protocol.Role{protocol.RoleUser},
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

		var msg prompts.PromptMessage
		err := json.Unmarshal([]byte(input), &msg)
		require.NoError(t, err)

		assert.Equal(t, protocol.RoleUser, msg.Role)
		textContent, ok := msg.Content.(protocol.TextContent)
		require.True(t, ok)
		assert.Equal(t, "Hello, world!", textContent.Text)
	})
}

func TestPrompt(t *testing.T) {
	t.Run("validates required arguments", func(t *testing.T) {
		prompt := prompts.Prompt{
			Name: "greet",
			Arguments: []prompts.PromptArgument{
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
		prompt := prompts.Prompt{
			Name: "greet",
			Arguments: []prompts.PromptArgument{
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

func TestMessageLifecycle(t *testing.T) {
	t.Run("full prompt execution flow", func(t *testing.T) {

		prompt := prompts.Prompt{
			Name: "calculate",
			Arguments: []prompts.PromptArgument{
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

		msg := prompts.PromptMessage{
			Role: protocol.RoleUser,
			Content: protocol.TextContent{
				Type: string(protocol.ContentTypeText),
				Text: result,
			},
		}

		data, err := json.Marshal(msg)
		require.NoError(t, err)

		var decoded prompts.PromptMessage
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, msg.Role, decoded.Role)
		assert.Equal(t,
			msg.Content.(protocol.TextContent).Text,
			decoded.Content.(protocol.TextContent).Text)
	})
}
