package tool

import (
	"testing"

	"go-mcp/pkg/mcp/protocol"

	"github.com/stretchr/testify/assert"
)

type MockTool struct {
	protocol.Tool
	ExecuteFn func(args map[string]interface{}) (*protocol.CallToolResult, error)
}

func (m *MockTool) ValidateAndExecute(args map[string]interface{}) (*protocol.CallToolResult, error) {
	if m.ExecuteFn != nil {
		return m.ExecuteFn(args)
	}
	return &protocol.CallToolResult{}, nil
}

func createTestTools() []*protocol.Tool {
	return []*protocol.Tool{
		{
			Name:        "echo",
			Description: "Echo back the input",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
		{
			Name:        "add",
			Description: "Add two numbers",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"a": map[string]interface{}{
						"type": "number",
					},
					"b": map[string]interface{}{
						"type": "number",
					},
				},
			},
		},
	}
}

func createTestProtocolTools() []protocol.Tool {
	return []protocol.Tool{
		{
			Name:        "list-files",
			Description: "List files in a directory",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
		{
			Name:        "read-file",
			Description: "Read a file",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
	}
}

func TestRegistry(t *testing.T) {
	t.Run("RegisterTool", func(t *testing.T) {
		registry := NewRegistry()
		tool := createTestTools()[0]

		// Register the tool
		err := registry.RegisterTool(tool, "test-source")
		assert.NoError(t, err, "RegisterTool should not return an error")

		// Check if tool exists
		retrievedTool, exists := registry.GetTool("echo")
		assert.True(t, exists, "Tool should exist")
		assert.Equal(t, tool, retrievedTool, "Retrieved tool should be the same as registered")

		// Check the source
		source, exists := registry.GetToolSource("echo")
		assert.True(t, exists, "Tool source should exist")
		assert.Equal(t, "test-source", source, "Source should match")

		// Try to register the same tool again (should fail)
		err = registry.RegisterTool(tool, "another-source")
		assert.Error(t, err, "Registering duplicate tool should fail")
	})

	t.Run("RegisterProtocolTool", func(t *testing.T) {
		registry := NewRegistry()
		protocolTool := createTestProtocolTools()[0]

		// Register the protocol tool
		err := registry.RegisterProtocolTool(protocolTool, "server1")
		assert.NoError(t, err, "RegisterProtocolTool should not return an error")

		// Check if tool exists
		tool, exists := registry.GetTool("list-files")
		assert.True(t, exists, "Tool should exist")
		assert.Equal(t, "list-files", tool.Name, "Tool name should match")
		assert.Equal(t, "List files in a directory", tool.Description, "Description should match")
	})

	t.Run("UnregisterTool", func(t *testing.T) {
		registry := NewRegistry()
		tool := createTestTools()[0]

		// Register the tool
		err := registry.RegisterTool(tool, "test-source")
		assert.NoError(t, err, "RegisterTool should not return an error")

		// Unregister the tool
		registry.UnregisterTool("echo")

		// Check if tool exists (should not)
		_, exists := registry.GetTool("echo")
		assert.False(t, exists, "Tool should not exist after unregistering")

		// Check if source exists (should not)
		_, exists = registry.GetToolSource("echo")
		assert.False(t, exists, "Tool source should not exist after unregistering")
	})

	t.Run("ListTools", func(t *testing.T) {
		registry := NewRegistry()
		tools := createTestTools()

		// Register tools
		registry.RegisterTool(tools[0], "source1")
		registry.RegisterTool(tools[1], "source2")

		// List all tools
		allTools := registry.ListTools()
		assert.Equal(t, 2, len(allTools), "Should list 2 tools")

		// List tools from source1
		source1Tools := registry.ListToolsFromSource("source1")
		assert.Equal(t, 1, len(source1Tools), "Should list 1 tool from source1")
		assert.Equal(t, "echo", source1Tools[0].Name, "Tool name should be 'echo'")

		// List tools from source2
		source2Tools := registry.ListToolsFromSource("source2")
		assert.Equal(t, 1, len(source2Tools), "Should list 1 tool from source2")
		assert.Equal(t, "add", source2Tools[0].Name, "Tool name should be 'add'")
	})

	t.Run("ExecuteTool", func(t *testing.T) {
		registry := NewRegistry()

		// Create a mock tool with a custom execution function
		mockTool := &MockTool{
			Tool: protocol.Tool{
				Name:        "mock-tool",
				Description: "A mock tool",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"input": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
			ExecuteFn: func(args map[string]interface{}) (*protocol.CallToolResult, error) {
				// Create a simple text content object
				content := protocol.TextContent{
					Type: string(protocol.ContentTypeText),
					Text: "Mock result",
				}
				return &protocol.CallToolResult{
					Content: []protocol.Content{content},
				}, nil
			},
		}

		// Register the mock tool
		registry.RegisterTool(&mockTool.Tool, "test-source")

		// Create a tool call
		toolCall := &protocol.ToolCall{
			Name: "mock-tool",
			Arguments: map[string]interface{}{
				"input": "test",
			},
		}

		// Execute the tool
		result, err := registry.ExecuteTool(toolCall)
		assert.NoError(t, err, "ExecuteTool should not return an error")
		assert.NotNil(t, result, "Result should not be nil")
		assert.Len(t, result.Content, 1, "Result should have one content item")

		assert.NotNil(t, result.Content[0], "Content item should not be nil")

		// Try with non-existent tool
		notFoundCall := &protocol.ToolCall{
			Name: "not-found",
			Arguments: map[string]interface{}{
				"foo": "bar",
			},
		}

		_, err = registry.ExecuteTool(notFoundCall)
		assert.Error(t, err, "ExecuteTool should return an error for non-existent tool")
	})
}
