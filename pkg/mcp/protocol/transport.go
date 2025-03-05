package protocol

import (
	"context"
	"io"
)

type Transport interface {
	Send(request *JSONRPCRequest) error

	SendWithContext(ctx context.Context, request *JSONRPCRequest) error

	Receive() (*JSONRPCResponse, error)

	Start() error

	Close() error

	IsConnected() bool
}

type ReadWriteCloser interface {
	io.Reader
	io.Writer
	io.Closer
}
