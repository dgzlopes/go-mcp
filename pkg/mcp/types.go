package mcp

import "encoding/json"

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
	Params  json.RawMessage `json:"params,omitempty`
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
	URI         string      `json:"uri"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	MimeType    string      `json:"mimeType,omitempty"`
	Annotations *Annotation `json:"annotations,omitempty"`
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
