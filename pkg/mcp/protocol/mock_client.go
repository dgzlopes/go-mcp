package protocol

import (
	"context"
	"fmt"
	"sync"
)

type MockClient struct {
	connected      bool
	capabilities   *ServerCapabilities
	tools          []Tool
	resources      []Resource
	callToolResult interface{}
	callToolError  error
	mutex          sync.RWMutex
}

func NewMockClient() *MockClient {
	return &MockClient{
		connected:    false,
		capabilities: nil,
		tools:        make([]Tool, 0),
		resources:    make([]Resource, 0),
	}
}

func (c *MockClient) Connect(transport Transport) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.connected = true
	return nil
}

func (c *MockClient) Disconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.connected = false
	return nil
}

func (c *MockClient) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.connected
}

func (c *MockClient) GetServerCapabilities() *ServerCapabilities {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.capabilities
}

func (c *MockClient) ListTools(ctx context.Context) ([]Tool, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.tools, nil
}

func (c *MockClient) SetTools(tools []Tool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.tools = tools
}

func (c *MockClient) CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.callToolResult, c.callToolError
}

func (c *MockClient) SetCallToolResult(result interface{}, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.callToolResult = result
	c.callToolError = err
}

func (c *MockClient) ListResources(ctx context.Context) ([]Resource, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client is not connected")
	}

	return c.resources, nil
}

func (c *MockClient) SetResources(resources []Resource) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.resources = resources
}

func (c *MockClient) HealthCheck(ctx context.Context) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if !c.connected {
		return fmt.Errorf("client is not connected")
	}
	return nil
}
