package protocol

import (
	"encoding/json"
	"fmt"
)

const JSONRPCVersion = "2.0"

const (
	ErrParseError     = -32700 // Invalid JSON
	ErrInvalidRequest = -32600 // The JSON sent is not a valid Request object
	ErrMethodNotFound = -32601 // The method does not exist / is not available
	ErrInvalidParams  = -32602 // Invalid method parameter(s)
	ErrInternalError  = -32603 // Internal JSON-RPC error
	ErrServerError    = -32000 // Generic server-defined error
	ErrConnError      = -32001 // Connection error
	ErrProtocolError  = -32002 // Protocol error
)

type JSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      string                 `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewRequest(id string, method string, params map[string]interface{}) *JSONRPCRequest {
	return &JSONRPCRequest{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Method:  method,
		Params:  params,
	}
}

func NewResponse(id string, result interface{}) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Result:  result,
	}
}

func NewErrorResponse(id string, code int, message string, data interface{}) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

func ToJSON(obj interface{}) ([]byte, error) {
	return json.Marshal(obj)
}

func FromJSON(data []byte, obj interface{}) error {
	return json.Unmarshal(data, obj)
}

func (e *JSONRPCError) Error() string {
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}
