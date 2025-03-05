package tool

import (
	"context"
	"fmt"
	"sync"

	"go-mcp/pkg/mcp/protocol"
)

type Registry struct {
	tools map[string]*protocol.Tool

	sources map[string]string

	mutex sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		tools:   make(map[string]*protocol.Tool),
		sources: make(map[string]string),
		mutex:   sync.RWMutex{},
	}
}

func (r *Registry) RegisterTool(tool *protocol.Tool, source string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if tool.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if tool.InputSchema == nil {
		return fmt.Errorf("tool input schema cannot be nil")
	}

	if existingSource, exists := r.sources[tool.Name]; exists {
		return fmt.Errorf("tool %s already registered by source %s", tool.Name, existingSource)
	}

	r.tools[tool.Name] = tool
	r.sources[tool.Name] = source

	return nil
}

func (r *Registry) RegisterProtocolTool(protocolTool protocol.Tool, source string) error {
	mcpTool := &protocol.Tool{
		Name:        protocolTool.Name,
		Description: protocolTool.Description,
		InputSchema: protocolTool.InputSchema,
	}

	return r.RegisterTool(mcpTool, source)
}

func (r *Registry) GetTool(name string) (*protocol.Tool, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tool, exists := r.tools[name]
	return tool, exists
}

func (r *Registry) GetToolSource(name string) (string, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	source, exists := r.sources[name]
	return source, exists
}

func (r *Registry) UnregisterTool(name string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.tools, name)
	delete(r.sources, name)
}

func (r *Registry) ImportFromServer(server *protocol.Client, serverName string) error {
	ctx := context.Background()
	tools, err := server.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tools from server %s: %w", serverName, err)
	}

	// Register each tool
	for _, tool := range tools {
		err := r.RegisterProtocolTool(tool, serverName)
		if err != nil {
			return fmt.Errorf("failed to register tool %s from server %s: %w", tool.Name, serverName, err)
		}
	}

	return nil
}

func (r *Registry) ListTools() []*protocol.Tool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tools := make([]*protocol.Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

func (r *Registry) ListToolsFromSource(source string) []*protocol.Tool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var tools []*protocol.Tool
	for name, toolSource := range r.sources {
		if toolSource == source {
			tools = append(tools, r.tools[name])
		}
	}
	return tools
}

func (r *Registry) ExecuteTool(call *protocol.ToolCall) (*protocol.CallToolResult, error) {
	r.mutex.RLock()
	tool, exists := r.tools[call.Name]
	r.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool %s not found", call.Name)
	}

	return tool.ValidateAndExecute(call.Arguments)
}
