package protocol

import (
	"encoding/json"
	"fmt"
)

const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

type ProgressToken interface{}

type Cursor string

type RequestID interface{}

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type RequestMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type ResponseMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ErrorMessage   `json:"error,omitempty"`
}

type NotificationMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type ErrorMessage struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type RequestMeta struct {
	ProgressToken ProgressToken `json:"progressToken,omitempty"`
}

type RequestParams struct {
	Meta *RequestMeta           `json:"_meta,omitempty"`
	Data map[string]interface{} `json:"-"`
}

type NotificationParams struct {
	Meta map[string]interface{} `json:"_meta,omitempty"`
	Data map[string]interface{} `json:"-"`
}

type ResultMeta struct {
	Meta map[string]interface{} `json:"_meta,omitempty"`
	Data map[string]interface{} `json:"-"` // Additional fields
}

type Implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ClientCapabilities struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
	Roots        *RootsCapability       `json:"roots,omitempty"`
	Sampling     *struct{}              `json:"sampling,omitempty"`
}

type ServerCapabilities struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
	Logging      *struct{}              `json:"logging,omitempty"`
	Prompts      *PromptsCapability     `json:"prompts,omitempty"`
	Resources    *ResourcesCapability   `json:"resources,omitempty"`
	Tools        *ToolsCapability       `json:"tools,omitempty"`
}

type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type InitializeParams struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      Implementation     `json:"clientInfo"`
}

type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      Implementation     `json:"serverInfo"`
	Instructions    string             `json:"instructions,omitempty"`
}

type Resource struct {
	URI         string                 `json:"uri"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	MimeType    string                 `json:"mimeType,omitempty"`
	Annotations *Annotation            `json:"annotations,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type ResourceTemplate struct {
	URITemplate string      `json:"uriTemplate"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	MimeType    string      `json:"mimeType,omitempty"`
	Annotations *Annotation `json:"annotations,omitempty"`
}

type ResourceContents struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
}

type TextResourceContents struct {
	ResourceContents
	Text string `json:"text"`
}

type BlobResourceContents struct {
	ResourceContents
	Blob string `json:"blob"`
}

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

func (t *Tool) ValidateAndExecute(args map[string]interface{}) (*CallToolResult, error) {
	if err := t.ValidateArguments(args); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	//TODO: Mock Result
	return &CallToolResult{
		Content: []Content{
			TextContent{
				Type: string(ContentTypeText),
				Text: "Tool execution result",
			},
		},
	}, nil
}

func (t *Tool) ValidateArguments(args map[string]interface{}) error {
	schema := t.InputSchema

	if required, ok := schema["required"].([]string); ok {
		for _, field := range required {
			if _, exists := args[field]; !exists {
				return fmt.Errorf("missing required field: %s", field)
			}
		}
	}

	if props, ok := schema["properties"].(map[string]interface{}); ok {
		for name, value := range args {
			if propSchema, exists := props[name]; exists {
				if err := ValidateType(propSchema.(map[string]interface{}), value); err != nil {
					return fmt.Errorf("invalid argument %s: %w", name, err)
				}
			}
		}
	}

	return nil
}

func ValidateType(schema map[string]interface{}, value interface{}) error {
	expectedType, ok := schema["type"].(string)
	if !ok {
		return fmt.Errorf("schema missing type")
	}

	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "number":
		switch v := value.(type) {
		case float64, float32, int, int32, int64:
		default:
			return fmt.Errorf("expected number, got %T", v)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	default:
		return fmt.Errorf("unsupported type: %s", expectedType)
	}

	return nil
}

type CallToolResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

type ContentType string

const (
	ContentTypeText     ContentType = "text"
	ContentTypeImage    ContentType = "image"
	ContentTypeResource ContentType = "resource"
)

type Content interface {
	GetType() ContentType
}

type Annotation struct {
	Audience []Role  `json:"audience,omitempty"`
	Priority float64 `json:"priority"`
}

type TextContent struct {
	Type        string      `json:"type"`
	Text        string      `json:"text,omitempty"`
	Annotations *Annotation `json:"annotations,omitempty"`
}

func (tc TextContent) GetType() ContentType {
	return ContentTypeText
}

type ImageContent struct {
	Type        ContentType `json:"type"`
	Data        string      `json:"data"`
	MimeType    string      `json:"mimeType"`
	Annotations *Annotation `json:"annotations,omitempty"`
}

func (ic ImageContent) GetType() ContentType {
	return ContentTypeImage
}

type EmbeddedResource struct {
	Type        ContentType      `json:"type"`
	Resource    ResourceContents `json:"resource"`
	Annotations *Annotation      `json:"annotations,omitempty"`
}

func (er EmbeddedResource) GetType() ContentType { return ContentTypeResource }

type LoggingLevel string

const (
	LoggingLevelDebug     LoggingLevel = "debug"
	LoggingLevelInfo      LoggingLevel = "info"
	LoggingLevelNotice    LoggingLevel = "notice"
	LoggingLevelWarning   LoggingLevel = "warning"
	LoggingLevelError     LoggingLevel = "error"
	LoggingLevelCritical  LoggingLevel = "critical"
	LoggingLevelAlert     LoggingLevel = "alert"
	LoggingLevelEmergency LoggingLevel = "emergency"
)

type ModelPreferences struct {
	Hints                []ModelHint `json:"hints,omitempty"`
	CostPriority         float64     `json:"costPriority,omitempty"`
	SpeedPriority        float64     `json:"speedPriority,omitempty"`
	IntelligencePriority float64     `json:"intelligencePriority,omitempty"`
}

type ModelHint struct {
	Name string `json:"name,omitempty"`
}

type Root struct {
	URI  string `json:"uri"`
	Name string `json:"name,omitempty"`
}

type ListToolsResponse struct {
	Tools []Tool `json:"tools"`
}

type ListResourcesResponse struct {
	Resources []Resource `json:"resources"`
}

type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type CallToolRequest struct {
	Method string   `json:"method"`
	Params ToolCall `json:"params"`
}

type HandshakeRequest struct {
	Version string `json:"version"`
	Client  struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"client"`
}

type HandshakeResponse struct {
	Version string `json:"version"`
	Server  struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"server"`
}

type ClientInfo struct {
	Name    string
	Version string
}
