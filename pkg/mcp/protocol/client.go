package protocol

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MCPClient interface {
	Connect(transport Transport) error

	ListTools(ctx context.Context) ([]Tool, error)

	ListResources(ctx context.Context) ([]Resource, error)

	CallTool(ctx context.Context, name string, params map[string]interface{}) (interface{}, error)

	GetServerCapabilities() *ServerCapabilities

	HealthCheck(ctx context.Context) error

	Disconnect() error

	IsConnected() bool
}

type Client struct {
	transport       Transport
	clientInfo      ClientInfo
	capabilities    *ServerCapabilities
	mutex           sync.RWMutex
	protocolVersion string
}

func NewClient(clientInfo ClientInfo) *Client {
	return &Client{
		clientInfo:      clientInfo,
		protocolVersion: "1.0",
	}
}

func (c *Client) Connect(transport Transport) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.transport != nil && c.transport.IsConnected() {
		return errors.New("client already connected")
	}

	if err := transport.Start(); err != nil {
		return fmt.Errorf("failed to start transport: %w", err)
	}

	c.transport = transport

	if err := c.performHandshake(); err != nil {
		c.transport.Close()
		c.transport = nil
		return err
	}

	c.capabilities = &ServerCapabilities{
		Tools:     &ToolsCapability{ListChanged: true},
		Resources: &ResourcesCapability{ListChanged: true},
	}

	if err := c.discoverCapabilities(); err != nil {
		c.transport.Close()
		c.transport = nil
		return err
	}

	return nil
}

func (c *Client) performHandshake() error {
	handshakeParams := map[string]interface{}{
		"version": c.protocolVersion,
		"client": map[string]interface{}{
			"name":    c.clientInfo.Name,
			"version": c.clientInfo.Version,
		},
	}

	requestID := uuid.New().String()
	request := NewRequest(requestID, "mcp.handshake", handshakeParams)

	if err := c.transport.Send(request); err != nil {
		return fmt.Errorf("handshake request failed: %w", err)
	}

	response, err := c.transport.Receive()
	if err != nil {
		return fmt.Errorf("handshake response failed: %w", err)
	}

	if response.Error != nil {
		return fmt.Errorf("handshake error: %s (code: %d)",
			response.Error.Message, response.Error.Code)
	}

	result, ok := response.Result.(map[string]interface{})
	if !ok {
		return errors.New("invalid handshake response format")
	}

	version, ok := result["version"].(string)
	if !ok {
		return errors.New("missing protocol version in handshake response")
	}

	if version != c.protocolVersion {
		return fmt.Errorf("incompatible protocol version: got %s, expected %s",
			version, c.protocolVersion)
	}

	return nil
}

func (c *Client) discoverCapabilities() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	_, err := c.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover tools: %w", err)
	}

	_, err = c.ListResources(ctx)
	if err != nil {
		fmt.Printf("Warning: failed to discover resources: %v\n", err)
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.capabilities.Tools = &ToolsCapability{ListChanged: true}
	c.capabilities.Resources = &ResourcesCapability{ListChanged: true}

	return nil
}

func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
	c.mutex.RLock()
	transport := c.transport
	c.mutex.RUnlock()

	if transport == nil || !transport.IsConnected() {
		return nil, errors.New("client not connected")
	}

	requestID := uuid.New().String()
	request := NewRequest(requestID, "mcp.list_tools", map[string]interface{}{})

	if err := transport.SendWithContext(ctx, request); err != nil {
		return nil, fmt.Errorf("list_tools request failed: %w", err)
	}

	response, err := transport.Receive()
	if err != nil {
		return nil, fmt.Errorf("list_tools response failed: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("list_tools error: %s (code: %d)",
			response.Error.Message, response.Error.Code)
	}

	result, ok := response.Result.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid list_tools response format")
	}

	toolsData, ok := result["tools"].([]interface{})
	if !ok {
		return nil, errors.New("invalid or missing tools array in response")
	}

	tools := make([]Tool, 0, len(toolsData))
	for _, item := range toolsData {
		toolMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := toolMap["name"].(string)
		description, _ := toolMap["description"].(string)
		inputSchema := toolMap["input_schema"]

		tools = append(tools, Tool{
			Name:        name,
			Description: description,
			InputSchema: inputSchema.(map[string]interface{}),
		})
	}

	return tools, nil
}

func (c *Client) ListResources(ctx context.Context) ([]Resource, error) {
	c.mutex.RLock()
	transport := c.transport
	c.mutex.RUnlock()

	if transport == nil || !transport.IsConnected() {
		return nil, errors.New("client not connected")
	}

	requestID := uuid.New().String()
	request := NewRequest(requestID, "mcp.list_resources", map[string]interface{}{})

	if err := transport.SendWithContext(ctx, request); err != nil {
		return nil, fmt.Errorf("list_resources request failed: %w", err)
	}

	response, err := transport.Receive()
	if err != nil {
		return nil, fmt.Errorf("list_resources response failed: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("list_resources error: %s (code: %d)",
			response.Error.Message, response.Error.Code)
	}

	result, ok := response.Result.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid list_resources response format")
	}

	resourcesData, ok := result["resources"].([]interface{})
	if !ok {
		return nil, errors.New("invalid or missing resources array in response")
	}

	resources := make([]Resource, 0, len(resourcesData))
	for _, item := range resourcesData {
		resourceMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := resourceMap["name"].(string)
		description, _ := resourceMap["description"].(string)
		resourceType, _ := resourceMap["type"].(string)
		metadata, _ := resourceMap["metadata"].(map[string]interface{})

		resources = append(resources, Resource{
			Name:        name,
			Description: description,
			Type:        resourceType,
			Metadata:    metadata,
		})
	}

	return resources, nil
}

func (c *Client) CallTool(ctx context.Context, name string, params map[string]interface{}) (interface{}, error) {
	c.mutex.RLock()
	transport := c.transport
	c.mutex.RUnlock()

	if transport == nil || !transport.IsConnected() {
		return nil, errors.New("client not connected")
	}

	requestID := uuid.New().String()
	request := NewRequest(requestID, name, params)

	if err := transport.SendWithContext(ctx, request); err != nil {
		return nil, fmt.Errorf("tool call request failed: %w", err)
	}

	response, err := transport.Receive()
	if err != nil {
		return nil, fmt.Errorf("tool call response failed: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("tool call error: %s (code: %d)",
			response.Error.Message, response.Error.Code)
	}

	return response.Result, nil
}

func (c *Client) GetServerCapabilities() *ServerCapabilities {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.capabilities == nil {
		return nil
	}

	return &ServerCapabilities{
		Tools:     c.capabilities.Tools,
		Resources: c.capabilities.Resources,
	}
}

func (c *Client) HealthCheck(ctx context.Context) error {
	c.mutex.RLock()
	transport := c.transport
	c.mutex.RUnlock()

	if transport == nil || !transport.IsConnected() {
		return errors.New("client not connected")
	}

	requestID := uuid.New().String()
	request := NewRequest(requestID, "mcp.ping", map[string]interface{}{})

	if err := transport.SendWithContext(ctx, request); err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}

	response, err := transport.Receive()
	if err != nil {
		return fmt.Errorf("health check response failed: %w", err)
	}

	if response.Error != nil {
		return fmt.Errorf("health check error: %s (code: %d)",
			response.Error.Message, response.Error.Code)
	}

	return nil
}

func (c *Client) Disconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.transport == nil || !c.transport.IsConnected() {
		return nil
	}

	err := c.transport.Close()
	c.transport = nil
	return err
}

func (c *Client) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.transport != nil && c.transport.IsConnected()
}

const defaultTimeout = 10 * time.Second
