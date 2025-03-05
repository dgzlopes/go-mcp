package server

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go-mcp/pkg/mcp/protocol"
)

var transportFactory = func(cmdStr string) protocol.Transport {
	return protocol.NewStdioTransport(cmdStr)
}

var (
	ErrServerNotFound = errors.New("server not found")
	ErrServerExists   = errors.New("server already exists")
	ErrNoServers      = errors.New("no servers available")
)

type ServerConfig struct {
	Name string

	Command string

	Args []string

	Env map[string]string

	WorkDir string
}

type Server struct {
	Name string

	Client protocol.MCPClient

	Tools []protocol.Tool

	Capabilities *protocol.ServerCapabilities

	Transport protocol.Transport

	Config ServerConfig
}

func (s *Server) IsRunning() bool {
	return s.Client != nil && s.Client.IsConnected()
}

type Manager struct {
	servers map[string]*Server
	mutex   sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		servers: make(map[string]*Server),
	}
}

func (m *Manager) LaunchServer(ctx context.Context, config ServerConfig) (*Server, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.servers[config.Name]; exists {
		return nil, fmt.Errorf("%w: %s", ErrServerExists, config.Name)
	}

	cmdStr := config.Command
	for _, arg := range config.Args {
		cmdStr += " " + arg
	}

	transport := transportFactory(cmdStr)

	if len(config.Env) > 0 {
		if t, ok := transport.(*protocol.StdioTransport); ok {
			t.SetEnv(config.Env)
		}
	}

	// Set working directory if provided
	// Note: If StdioTransport doesn't have SetWorkDir, we can handle this
	// through environment variables or by modifying the transport interface later
	if config.WorkDir != "" {
		// Working directory will be handled by the transport
		// implementation if supported
	}

	// Start the transport
	if err := transport.Start(); err != nil {
		return nil, fmt.Errorf("failed to start transport: %w", err)
	}

	// Create client
	client := protocol.NewClient(protocol.ClientInfo{
		Name:    "go-mcp",
		Version: "0.1.0",
	})

	// Connect to the server with the transport
	if err := client.Connect(transport); err != nil {
		// Clean up on connect failure
		transport.Close()
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	// Create server instance
	server := &Server{
		Name:         config.Name,
		Client:       client,
		Tools:        nil, // Will be populated below
		Capabilities: client.GetServerCapabilities(),
		Transport:    transport,
		Config:       config,
	}

	// Get tools
	tools, err := client.ListTools(ctx)
	if err != nil {
		// Non-fatal error, for now we'll just set an empty tools list
		server.Tools = []protocol.Tool{}
	} else {
		server.Tools = tools
	}

	// Add to server map
	m.servers[config.Name] = server

	return server, nil
}

func (m *Manager) GetServer(name string) (*Server, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	server, exists := m.servers[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrServerNotFound, name)
	}

	return server, nil
}

func (m *Manager) ShutdownServer(ctx context.Context, name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	server, exists := m.servers[name]
	if !exists {
		return fmt.Errorf("%w: %s", ErrServerNotFound, name)
	}

	if server.Client != nil {
		if err := server.Client.Disconnect(); err != nil {
			return fmt.Errorf("failed to disconnect from server: %w", err)
		}
	}

	delete(m.servers, name)

	return nil
}

func (m *Manager) ShutdownAll(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var lastErr error
	for name, server := range m.servers {
		if server.Client != nil {
			if err := server.Client.Disconnect(); err != nil {
				lastErr = fmt.Errorf("failed to disconnect from server %s: %w", name, err)
			}
		}
	}

	// Clear the map
	m.servers = make(map[string]*Server)

	return lastErr
}

func (m *Manager) ListServers() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	names := make([]string, 0, len(m.servers))
	for name := range m.servers {
		names = append(names, name)
	}

	return names
}

func (m *Manager) DiscoverTools(ctx context.Context) (map[string][]protocol.Tool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(m.servers) == 0 {
		return nil, ErrNoServers
	}

	tools := make(map[string][]protocol.Tool)

	for name, server := range m.servers {
		if !server.IsRunning() {
			continue
		}

		// Try to get tools from server
		serverTools, err := server.Client.ListTools(ctx)
		if err != nil {
			continue // Skip servers that fail to list tools
		}

		server.Tools = serverTools

		// Add tools to map
		tools[name] = serverTools
	}

	return tools, nil
}

func (m *Manager) MonitorHealth(ctx context.Context) map[string]error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	results := make(map[string]error)

	for name, server := range m.servers {
		if !server.IsRunning() {
			results[name] = errors.New("server not running")
			continue
		}

		// Check server health
		if err := server.Client.HealthCheck(ctx); err != nil {
			results[name] = err
		} else {
			results[name] = nil
		}
	}

	return results
}
