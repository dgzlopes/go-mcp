# go-mcp

A powerful, flexible Go implementation of the Model Context Protocol (MCP) - enabling seamless integration between LLMs and the tools they need.

## Why MCP Matters

Large Language Models (LLMs) need to interact with external tools to be truly useful. But each AI provider implements tool calling differently, creating a fragmented ecosystem.

The Model Context Protocol (MCP) solves this by providing:

- **Standardized Tool Interface**: Run the same tools across different LLM providers
- **Two-way Communication**: Let AI models access real-world capabilities through a uniform API
- **Reduced Integration Complexity**: Write tools once, use them with any MCP-compatible AI

This Go implementation makes it easy to connect your Go applications with AI tools, regardless of which LLM provider you're using.

## Architecture

The MCP architecture creates a standardized bridge between LLMs and tool implementations:

```mermaid
graph LR
    A[LLM Providers<br>OpenAI, Anthropic, etc] <--> B[go-mcp Client<br>This Library]
    B <--> C[MCP Servers<br>Tool Implementations]
    
    classDef primary fill:#4285f4,stroke:#333,stroke-width:2px,color:white
    classDef secondary fill:#34a853,stroke:#333,stroke-width:2px,color:white
    classDef tertiary fill:#fbbc05,stroke:#333,stroke-width:2px,color:white
    
    class A primary
    class B secondary
    class C tertiary
```

## Key Features

- **Simplified Tool Integration**: Connect to MCP tool servers with minimal code
- **Provider Agnostic**: Works with multiple LLM providers 
- **Server Management**: Launch, monitor, and shut down tool servers
- **Transport Flexibility**: Connect via stdio, HTTP, WebSockets, and more
- **Tool Discovery**: Automatically find and use available tools
- **Schema Validation**: Ensure data consistency through JSON Schema validation

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/user/go-mcp/pkg/mcp/protocol"
)

func main() {
	// Create a client and connect to an MCP server
	client := protocol.NewClient(protocol.ClientInfo{Name: "go-mcp-example", Version: "1.0.0"})
	transport := protocol.NewStdioTransport("python path/to/mcp_server.py")
	
	if err := client.Connect(transport); err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer client.Disconnect()
	
	// Discover available tools
	ctx := context.Background()
	tools, err := client.ListTools(ctx)
	if err != nil {
		log.Fatalf("Tool discovery failed: %v", err)
	}
	
	for _, tool := range tools {
		fmt.Printf("Found tool: %s - %s\n", tool.Name, tool.Description)
	}
	
	// Call a tool
	result, err := client.CallTool(ctx, "search", map[string]interface{}{
		"query": "climate change solutions",
		"limit": 5,
	})
	
	if err != nil {
		log.Fatalf("Tool execution failed: %v", err)
	}
	
	fmt.Printf("Results: %v\n", result)
}
```

## MCP Protocol Flow

The protocol follows a standard sequence for tool discovery and execution:

```mermaid
sequenceDiagram
    participant U as User
    participant H as Host Application (LLM)
    participant C as MCP Client
    participant S1 as MCP Server 1
    participant S2 as MCP Server 2
    
    %% Connection Initialization
    C->>S1: Protocol handshake
    S1-->>C: Version & capabilities
    C->>S2: Protocol handshake
    S2-->>C: Version & capabilities
    
    %% Tool Discovery
    C->>S1: mcp.list_tools
    S1-->>C: Available tools
    C->>S2: mcp.list_tools
    S2-->>C: Available tools
    C-->>H: Register available tools
    
    %% User Interaction and Tool Calling
    U->>H: Send request
    H->>H: Process request
    H->>C: Generate tool call
    
    %% Tool Execution Loop
    C->>S1: Call tool with parameters
    S1->>S1: Execute tool
    S1-->>C: Return result
    C-->>H: Provide result
    
    %% Agent can decide to use another tool
    H->>H: Process tool result
    H->>C: Generate another tool call
    C->>S2: Call tool with parameters
    S2->>S2: Execute tool
    S2-->>C: Return result
    C-->>H: Provide result
    
    %% Final Response
    H->>H: Process all tool results
    H-->>U: Generate final response
    
    %% Connection Termination
    C->>S1: Close connection
    C->>S2: Close connection
```

The protocol enables multi-step, agentic workflows where:
1. The LLM processes user inputs and determines which tools to use
2. Tools are executed via MCP servers and results returned to the LLM
3. The LLM can use multiple tools in sequence to complete complex tasks
4. Each tool interaction follows the same standardized messaging pattern

## Installation

```bash
go get github.com/dhruvshrma/go-mcp
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.