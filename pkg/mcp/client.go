package mcp

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go-mcp/pkg/mcp/protocol"
	"go-mcp/pkg/mcp/server"
)

var (
	ErrNotInitialized     = errors.New("MCP client not initialized")
	ErrAlreadyInitialized = errors.New("MCP client already initialized")
	ErrToolNotFound       = errors.New("tool not found")
)

type ToolResult struct {
	ToolName string
	Contents []protocol.Content
}

type MCPClient interface {
	Initialize(ctx context.Context) error
	Shutdown(ctx context.Context) error
	AddServer(config server.ServerConfig) error
	RemoveServer(serverName string) error
	GetServer(serverName string) (*server.Server, error)
	ListServers() []*server.Server
	ListTools() []*protocol.Tool
	GetTool(name string) (*protocol.Tool, error)
	ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) (*protocol.CallToolResult, error)
}

type Client struct {
	manager     *server.Manager
	tools       map[string]*protocol.Tool
	toolSources map[string]string
	initialized bool
	mu          sync.RWMutex
}

func NewClient() *Client {
	return &Client{
		manager:     server.NewManager(),
		tools:       make(map[string]*protocol.Tool),
		toolSources: make(map[string]string),
	}
}

func (c *Client) Initialize(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return ErrAlreadyInitialized
	}

	c.initialized = true
	return nil
}

func (c *Client) Shutdown(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return ErrNotInitialized
	}

	err := c.manager.ShutdownAll(ctx)
	c.initialized = false
	return err
}

func (c *Client) AddServer(config server.ServerConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return ErrNotInitialized
	}

	srv, err := c.manager.LaunchServer(context.Background(), config)
	if err != nil {
		return err
	}

	return c.importToolsFromServer(srv)
}

func (c *Client) importToolsFromServer(srv *server.Server) error {
	for _, protocolTool := range srv.Tools {
		tool := &protocol.Tool{
			Name:        protocolTool.Name,
			Description: protocolTool.Description,
			InputSchema: protocolTool.InputSchema,
		}

		c.tools[tool.Name] = tool
		c.toolSources[tool.Name] = srv.Name
	}
	return nil
}

func (c *Client) RemoveServer(serverName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return ErrNotInitialized
	}

	c.unregisterToolsFromServer(serverName)

	return c.manager.ShutdownServer(context.Background(), serverName)
}

func (c *Client) unregisterToolsFromServer(serverName string) {
	var toolsToRemove []string

	for name, source := range c.toolSources {
		if source == serverName {
			toolsToRemove = append(toolsToRemove, name)
		}
	}

	for _, name := range toolsToRemove {
		delete(c.tools, name)
		delete(c.toolSources, name)
	}
}

func (c *Client) GetServer(serverName string) (*server.Server, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.initialized {
		return nil, ErrNotInitialized
	}

	return c.manager.GetServer(serverName)
}

func (c *Client) ListServers() []*server.Server {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.initialized {
		return nil
	}

	serverNames := c.manager.ListServers()

	servers := make([]*server.Server, 0, len(serverNames))

	for _, name := range serverNames {
		if server, err := c.manager.GetServer(name); err == nil {
			servers = append(servers, server)
		}
	}

	return servers
}

func (c *Client) ListTools() []*protocol.Tool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.initialized {
		return nil
	}

	tools := make([]*protocol.Tool, 0, len(c.tools))
	for _, tool := range c.tools {
		tools = append(tools, tool)
	}
	return tools
}

func (c *Client) GetTool(name string) (*protocol.Tool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.initialized {
		return nil, ErrNotInitialized
	}

	tool, exists := c.tools[name]
	if !exists {
		return nil, ErrToolNotFound
	}

	return tool, nil
}

func (c *Client) getToolServer(name string) (*server.Server, error) {
	serverName, exists := c.toolSources[name]
	if !exists {
		return nil, ErrToolNotFound
	}

	return c.manager.GetServer(serverName)
}

func (c *Client) ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) (*protocol.CallToolResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.initialized {
		return nil, ErrNotInitialized
	}

	_, exists := c.tools[toolName]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrToolNotFound, toolName)
	}

	serverName, exists := c.toolSources[toolName]
	if !exists {
		return nil, fmt.Errorf("no server found for tool: %s", toolName)
	}

	srv, err := c.manager.GetServer(serverName)
	if err != nil {
		return nil, err
	}

	call := &protocol.ToolCall{
		Name:      toolName,
		Arguments: args,
	}

	result, err := srv.Client.CallTool(ctx, call.Name, call.Arguments)
	if err != nil {
		return nil, err
	}

	var content []protocol.Content
	var isError bool

	if m, ok := result.(map[string]interface{}); ok {
		if val, ok := m["isError"].(bool); ok {
			isError = val
		}

		if contentArray, ok := m["content"].([]interface{}); ok {
			for _, item := range contentArray {
				if contentMap, ok := item.(map[string]interface{}); ok {
					typeVal, hasType := contentMap["type"].(string)
					if !hasType {
						continue
					}

					switch typeVal {
					case string(protocol.ContentTypeText):
						if textVal, ok := contentMap["text"].(string); ok {
							content = append(content, protocol.TextContent{
								Type: string(protocol.ContentTypeText),
								Text: textVal,
							})
						}
					default:
						if textVal, ok := contentMap["text"].(string); ok {
							content = append(content, protocol.TextContent{
								Type: string(protocol.ContentTypeText),
								Text: textVal,
							})
						}
					}
				}
			}
		}
	}

	if len(content) == 0 {
		content = []protocol.Content{
			protocol.TextContent{
				Type: string(protocol.ContentTypeText),
				Text: "Tool execution completed",
			},
		}
	}

	return &protocol.CallToolResult{
		Content: content,
		IsError: isError,
	}, nil
}
