package executor

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// ExecutionResult is the result of running code.
type ExecutionResult struct {
	Status        string `json:"status"` // "ok", "error", "timeout"
	Stdout        string `json:"stdout"`
	Stderr        string `json:"stderr"`
	ExecutionTime int64  `json:"execution_time_ms"`
	Compiled      bool   `json:"compiled"` // true if used go run, false if yaegi
	Error         string `json:"error,omitempty"`
}

// Executor runs Go code with a timeout and security constraints.
type Executor struct {
	Timeout time.Duration
}

// NewExecutor returns an Executor with default 30s timeout.
func NewExecutor() *Executor {
	return &Executor{Timeout: 30 * time.Second}
}

// Execute runs the code and returns the result. It tries yaegi first, then falls back to go run.
func (e *Executor) Execute(ctx context.Context, code string) (*ExecutionResult, error) {
	ctx, cancel := context.WithTimeout(ctx, e.Timeout)
	defer cancel()
	start := time.Now()
	if err := ValidateCode(code); err != nil {
		return &ExecutionResult{
			Status:        "error",
			Error:         err.Error(),
			ExecutionTime: time.Since(start).Milliseconds(),
		}, nil
	}
	// Wrap in main if needed for yaegi
	wrapped := wrapMain(code)
	var stdout, stderr bytes.Buffer
	i := interp.New(interp.Options{Stdout: &stdout, Stderr: &stderr})
	i.Use(stdlib.Symbols)
	_, err := i.EvalWithContext(ctx, wrapped)
	elapsed := time.Since(start).Milliseconds()
	if ctx.Err() == context.DeadlineExceeded {
		return &ExecutionResult{Status: "timeout", ExecutionTime: elapsed, Compiled: false}, nil
	}
	if err != nil {
		// Fallback to go run for code that needs full Go
		result, _ := RunCompiled(ctx, code, e.Timeout)
		if result != nil && result.Status == "ok" {
			return result, nil
		}
		return &ExecutionResult{
			Status:        "error",
			Stderr:        stderr.String(),
			Error:         err.Error(),
			ExecutionTime: elapsed,
			Compiled:      false,
		}, nil
	}
	return &ExecutionResult{
		Status:        "ok",
		Stdout:        stdout.String(),
		Stderr:        stderr.String(),
		ExecutionTime: elapsed,
		Compiled:      false,
	}, nil
}

func wrapMain(code string) string {
	code = strings.TrimSpace(code)
	if strings.HasPrefix(code, "package ") && strings.Contains(code, "func main()") {
		return code
	}
	return "package main\n\nfunc main() {\n" + code + "\n}\n"
}

// RunCompiled runs code via "go run" in a temp directory. Used as fallback.
func RunCompiled(ctx context.Context, code string, timeout time.Duration) (*ExecutionResult, error) {
	start := time.Now()
	if err := ValidateCode(code); err != nil {
		return &ExecutionResult{
			Status:        "error",
			Error:         err.Error(),
			ExecutionTime: time.Since(start).Milliseconds(),
			Compiled:      true,
		}, nil
	}
	dir, err := os.MkdirTemp("", "glad-run-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	fpath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(fpath, []byte(code), 0600); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "run", fpath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	elapsed := time.Since(start).Milliseconds()
	if ctx.Err() == context.DeadlineExceeded {
		return &ExecutionResult{Status: "timeout", ExecutionTime: elapsed, Compiled: true}, nil
	}
	if err != nil {
		return &ExecutionResult{
			Status:        "error",
			Stdout:        stdout.String(),
			Stderr:        stderr.String(),
			Error:         err.Error(),
			ExecutionTime: elapsed,
			Compiled:      true,
		}, nil
	}
	return &ExecutionResult{
		Status:        "ok",
		Stdout:        stdout.String(),
		Stderr:        stderr.String(),
		ExecutionTime: elapsed,
		Compiled:      true,
	}, nil
}
