package server

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"go-mcp/pkg/mcp/protocol"
)

type MockClient struct {
	connected     bool
	tools         []protocol.Tool
	capabilities  *protocol.ServerCapabilities
	mutex         sync.RWMutex
	callResults   map[string]interface{}
	healthStatus  error
	disconnectErr error
}

func NewMockClient() *MockClient {
	tools := []protocol.Tool{
		{
			Name:        "echo",
			Description: "Echo back input",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{"type": "string"},
				},
			},
		},
		{
			Name:        "add",
			Description: "Add two numbers",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"a": map[string]interface{}{"type": "number"},
					"b": map[string]interface{}{"type": "number"},
				},
			},
		},
	}

	capabilities := &protocol.ServerCapabilities{
		Tools: &protocol.ToolsCapability{ListChanged: true},
	}

	return &MockClient{
		connected:    true,
		tools:        tools,
		capabilities: capabilities,
		callResults:  make(map[string]interface{}),
	}
}

func (m *MockClient) Connect(transport protocol.Transport) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.connected = true
	return nil
}

func (m *MockClient) ListTools(ctx context.Context) ([]protocol.Tool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if !m.connected {
		return nil, fmt.Errorf("client not connected")
	}
	return m.tools, nil
}

func (m *MockClient) ListResources(ctx context.Context) ([]protocol.Resource, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if !m.connected {
		return nil, fmt.Errorf("client not connected")
	}
	return []protocol.Resource{}, nil
}

func (m *MockClient) CallTool(ctx context.Context, name string, params map[string]interface{}) (interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.connected {
		return nil, fmt.Errorf("client not connected")
	}

	if result, ok := m.callResults[name]; ok {
		return result, nil
	}

	if name == "echo" && params != nil {
		if text, ok := params["text"]; ok {
			return map[string]interface{}{"text": fmt.Sprintf("Echo: %v", text)}, nil
		}
	}

	if name == "add" && params != nil {
		a, aOk := params["a"].(float64)
		b, bOk := params["b"].(float64)
		if aOk && bOk {
			return map[string]interface{}{"sum": a + b}, nil
		}
	}

	return map[string]interface{}{"status": "ok"}, nil
}

func (m *MockClient) GetServerCapabilities() *protocol.ServerCapabilities {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.capabilities
}

func (m *MockClient) HealthCheck(ctx context.Context) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.healthStatus
}

func (m *MockClient) Disconnect() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.connected = false
	return m.disconnectErr
}

func (m *MockClient) IsConnected() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.connected
}

func (m *MockClient) SetMockToolResult(toolName string, result interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.callResults[toolName] = result
}

func (m *MockClient) SetHealthStatus(err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.healthStatus = err
}

func (m *MockClient) SetDisconnectError(err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.disconnectErr = err
}

func createMockServer(name string) *Server {
	mockClient := NewMockClient()

	return &Server{
		Name:         name,
		Client:       mockClient,
		Tools:        mockClient.tools,
		Capabilities: mockClient.capabilities,
		Transport:    nil,
		Config: ServerConfig{
			Name:    name,
			Command: "mock-command",
		},
	}
}

func TestServerManager(t *testing.T) {
	t.Run("LaunchServer", func(t *testing.T) {
		manager := NewManager()

		mockServer := createMockServer("test-server")

		manager.servers["test-server"] = mockServer

		retrievedServer, err := manager.GetServer("test-server")
		if err != nil {
			t.Fatalf("Failed to get server: %v", err)
		}

		if retrievedServer != mockServer {
			t.Fatal("Retrieved server should be the same as the mock server")
		}

		if !retrievedServer.IsRunning() {
			t.Fatal("Server should be running")
		}
	})

	t.Run("ShutdownServer", func(t *testing.T) {
		manager := NewManager()

		mockServer := createMockServer("test-server")
		manager.servers["test-server"] = mockServer

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := manager.ShutdownServer(ctx, "test-server")
		if err != nil {
			t.Fatalf("Failed to shutdown server: %v", err)
		}

		_, err = manager.GetServer("test-server")
		if err == nil {
			t.Fatal("Server should be removed after shutdown")
		}
	})

	t.Run("DiscoverTools", func(t *testing.T) {
		manager := NewManager()

		server1 := createMockServer("server1")
		server2 := createMockServer("server2")

		manager.servers["server1"] = server1
		manager.servers["server2"] = server2

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		toolMap, err := manager.DiscoverTools(ctx)
		if err != nil {
			t.Fatalf("Failed to discover tools: %v", err)
		}

		if len(toolMap) != 2 {
			t.Fatalf("Expected tools from 2 servers, got %d", len(toolMap))
		}

		if len(toolMap["server1"]) != 2 {
			t.Fatalf("Expected 2 tools from server1, got %d", len(toolMap["server1"]))
		}

		if len(toolMap["server2"]) != 2 {
			t.Fatalf("Expected 2 tools from server2, got %d", len(toolMap["server2"]))
		}
	})
}
