package mcp

import (
	"context"
	"go-mcp/pkg/mcp/protocol"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	ctx := context.Background()

	t.Run("Initialize", func(t *testing.T) {
		client := NewClient()

		assert.False(t, client.initialized)

		err := client.Initialize(ctx)
		require.NoError(t, err)
		assert.True(t, client.initialized)

		// Attempting to initialize again should error
		err = client.Initialize(ctx)
		assert.Equal(t, ErrAlreadyInitialized, err)
	})

	t.Run("Shutdown", func(t *testing.T) {
		client := setupClient(t)

		err := client.Shutdown(ctx)
		require.NoError(t, err)
		assert.False(t, client.initialized)
	})

	t.Run("ListTools", func(t *testing.T) {
		client := setupClient(t)

		client.tools["tool1"] = &protocol.Tool{Name: "tool1"}
		client.tools["tool2"] = &protocol.Tool{Name: "tool2"}
		client.toolSources["tool1"] = "server1"
		client.toolSources["tool2"] = "server2"

		tools := client.ListTools()
		assert.Len(t, tools, 2)

		toolNames := make(map[string]bool)
		for _, tool := range tools {
			toolNames[tool.Name] = true
		}
		assert.True(t, toolNames["tool1"])
		assert.True(t, toolNames["tool2"])
	})

	t.Run("GetTool", func(t *testing.T) {
		client := setupClient(t)

		client.tools["tool1"] = &protocol.Tool{Name: "tool1"}
		client.toolSources["tool1"] = "server1"

		tool, err := client.GetTool("tool1")
		require.NoError(t, err)
		assert.Equal(t, "tool1", tool.Name)

		_, err = client.GetTool("non-existent")
		assert.Equal(t, ErrToolNotFound, err)
	})
}

func setupClient(t *testing.T) *Client {
	client := NewClient()
	err := client.Initialize(context.Background())
	require.NoError(t, err)
	return client
}
