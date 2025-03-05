package server

import (
	"context"
	"sync"

	"go-mcp/pkg/mcp/protocol"
)

type MockManager struct {
	servers     map[string]*Server
	mutex       sync.RWMutex
	callResults map[string]interface{}
	callErrors  map[string]error
}

func NewMockManager() *MockManager {
	return &MockManager{
		servers:     make(map[string]*Server),
		callResults: make(map[string]interface{}),
		callErrors:  make(map[string]error),
	}
}

func (m *MockManager) LaunchServer(ctx context.Context, config ServerConfig) (*Server, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	mockClient := protocol.NewMockClient()

	server := &Server{
		Name:         config.Name,
		Client:       mockClient,
		Tools:        []protocol.Tool{},
		Capabilities: &protocol.ServerCapabilities{},
		Config:       config,
	}

	m.servers[config.Name] = server

	return server, nil
}

func (m *MockManager) GetServer(name string) (*Server, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	server, exists := m.servers[name]
	if !exists {
		return nil, ErrServerNotFound
	}

	return server, nil
}

func (m *MockManager) ShutdownServer(ctx context.Context, name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.servers[name]; !exists {
		return ErrServerNotFound
	}

	delete(m.servers, name)
	return nil
}

func (m *MockManager) ShutdownAll(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.servers = make(map[string]*Server)
	return nil
}

func (m *MockManager) ListServers() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	names := make([]string, 0, len(m.servers))
	for name := range m.servers {
		names = append(names, name)
	}

	return names
}

func (m *MockManager) DiscoverTools(ctx context.Context) (map[string][]protocol.Tool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(m.servers) == 0 {
		return nil, ErrNoServers
	}

	tools := make(map[string][]protocol.Tool)

	for name, server := range m.servers {
		tools[name] = server.Tools
	}

	return tools, nil
}

func (m *MockManager) MonitorHealth(ctx context.Context) map[string]error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	health := make(map[string]error)
	for name := range m.servers {
		health[name] = nil
	}

	return health
}

func (m *MockManager) SetCallToolResult(serverName string, result interface{}, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.callResults[serverName] = result
	m.callErrors[serverName] = err

	if server, exists := m.servers[serverName]; exists {
		if mockClient, ok := server.Client.(*protocol.MockClient); ok {
			mockClient.SetCallToolResult(result, err)
		}
	}
}
