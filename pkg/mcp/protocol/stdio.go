package protocol

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type StdioTransport struct {
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	scanner    *bufio.Scanner
	connected  bool
	mutex      sync.Mutex
	lineBuffer []string // For debug and error reporting
	env        map[string]string
	cmdStr     string
}

func NewStdioTransport(cmdStr string) *StdioTransport {
	return &StdioTransport{
		cmdStr:     cmdStr,
		connected:  false,
		lineBuffer: make([]string, 0, 10),
		env:        make(map[string]string),
	}
}

func (t *StdioTransport) SetEnv(env map[string]string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Copy the environment variables
	for k, v := range env {
		t.env[k] = v
	}
}

func (t *StdioTransport) Start() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.connected {
		return errors.New("transport already started")
	}

	args := strings.Fields(t.cmdStr)
	if len(args) == 0 {
		return errors.New("empty command string")
	}

	cmdName := args[0]
	var cmdArgs []string
	if len(args) > 1 {
		cmdArgs = args[1:]
	}
	t.cmd = exec.Command(cmdName, cmdArgs...)

	if len(t.env) > 0 {
		t.cmd.Env = os.Environ()

		for k, v := range t.env {
			t.cmd.Env = append(t.cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Set up pipes for stdin and stdout
	var err error
	t.stdin, err = t.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	t.stdout, err = t.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	t.scanner = bufio.NewScanner(t.stdout)

	if err := t.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	t.connected = true
	return nil
}

func (t *StdioTransport) Send(request *JSONRPCRequest) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.connected {
		return fmt.Errorf("transport not connected")
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	requestJSON = append(requestJSON, '\n')

	_, err = t.stdin.Write(requestJSON)
	if err != nil {
		t.connected = false
		return fmt.Errorf("failed to write to stdin: %w", err)
	}

	return nil
}

func (t *StdioTransport) SendWithContext(ctx context.Context, request *JSONRPCRequest) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Continue with send
	}

	done := make(chan error, 1)

	go func() {
		done <- t.Send(request)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

func (t *StdioTransport) Receive() (*JSONRPCResponse, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.connected {
		return nil, fmt.Errorf("transport not connected")
	}

	if !t.scanner.Scan() {
		t.connected = false
		if err := t.scanner.Err(); err != nil {
			return nil, fmt.Errorf("error reading from stdout: %w", err)
		}
		return nil, fmt.Errorf("EOF reached")
	}

	text := t.scanner.Text()

	t.bufferLine(text)

	var response JSONRPCResponse
	if err := json.Unmarshal([]byte(text), &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w, raw response: %s", err, text)
	}

	return &response, nil
}

func (t *StdioTransport) Close() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.connected {
		return nil
	}

	t.connected = false

	if t.stdin != nil {
		t.stdin.Close()
	}

	if t.cmd.Process != nil {
		return t.cmd.Process.Kill()
	}

	return nil
}

func (t *StdioTransport) IsConnected() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.connected
}

func (t *StdioTransport) bufferLine(line string) {
	if len(t.lineBuffer) >= 10 {
		t.lineBuffer = t.lineBuffer[1:]
	}
	t.lineBuffer = append(t.lineBuffer, line)
}

func (t *StdioTransport) GetBufferedLines() []string {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return append([]string{}, t.lineBuffer...)
}
