# MCP (Model Context Protocol)

A Go implementation of the Model Context Protocol, a protocol for structured communication with AI language models.

## Overview

MCP is a JSON-RPC based protocol that provides a structured way to:

- Format and manage prompts
- Handle tool calls and executions
- Manage resources and content types
- Control message flow between users and AI assistants

## Features

### Message Types

#### PromptMessage

A structured message for communication between users and AI assistants.

- Supports different roles (user/assistant)
- Handles various content types:
  - Text
  - Images
  - Embedded Resources
- Supports annotations for message metadata

### Content Management

The protocol supports multiple content types:

```go
const (
	ContentTypeText ContentType = "text"
	ContentTypeImage ContentType = "image"
	ContentTypeResource ContentType = "resource"
)
```
### Tools and Execution

- **Tool** - A function that can be called by the AI model.
- **ToolCall** - A message that instructs the AI model to call a tool.
- **ToolResult** - A message that contains the result of a tool call.

Example tool definition:
```go
tool := Tool{
    "Name": "add_numbers",
    "Description": "Add two numbers together",
    "InputSchema": map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "a": map[string]interface{}{"type": "number"},
            "b": map[string]interface{}{"type": "number"},
        },
        "required": []string{"a", "b"},
    },
}
```

## Protocol Structure 

### Request/Response Format 

The protocol uses JSON-RPC 2.0 for request/response communication.

### Capabilities 

Supports negotiation of features between clients and servers.
```go 
type ServerCapabilities struct {
    Experimental map[string]interface{} `json:"experimental,omitempty"`
    Logging *struct{} `json:"logging,omitempty"`
    Prompts *PromptsCapability `json:"prompts,omitempty"`
    Resources *ResourcesCapability `json:"resources,omitempty"`
    Tools *ToolsCapability `json:"tools,omitempty"`
}
```

## Usage 

### Basic Message Creation 
```go
msg := PromptMessage{
    Role: RoleUser,
    Content: TextContent{
        Type: string(ContentTypeText),
        Text: "Hello, world!",
    },
}
```

### Executing Prompts

### IN DEVELOPMENT
