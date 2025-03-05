package protocol_test

import (
	"encoding/json"
	"go-mcp/pkg/mcp/protocol"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolCall(t *testing.T) {
	t.Run("deserializes basic tool call", func(t *testing.T) {
		input := `{
			"name": "get_weather",
			"arguments": {
				"location": "London",
				"units": "celsius"
			}
		}`

		var call protocol.ToolCall
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

		var call protocol.ToolCall
		err := json.Unmarshal([]byte(input), &call)
		require.NoError(t, err)

		assert.Equal(t, "list_files", call.Name)
		assert.Nil(t, call.Arguments)
	})
}

func TestTool(t *testing.T) {
	t.Run("validates argument types", func(t *testing.T) {
		tool := protocol.Tool{
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
		tool := protocol.Tool{
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
