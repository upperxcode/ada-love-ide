package cliworker

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var (
	DefaultTimeout = 120 * time.Second
	ansiRegex      = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
)

type ModelInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ProviderName string `json:"provider_name"`
}

// stripANSI removes ANSI escape sequences from a string.
func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// Emitter is a function that emits an event with data to the frontend.
type EmitterFn func(event string, data ...any)

// ExecuteStream runs a CLI worker command, streaming output via emitters.
// Returns the combined response string.
func ExecuteStream(cmd Command, emit EmitterFn) (string, error) {
	timeout := cmd.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	execCmd := exec.CommandContext(ctx, cmd.Path, cmd.Args...)
	execCmd.Env = cmd.Env
	if cmd.Dir != "" {
		execCmd.Dir = cmd.Dir
	}

	stdoutPipe, _ := execCmd.StdoutPipe()
	stderrPipe, _ := execCmd.StderrPipe()

	if err := execCmd.Start(); err != nil {
		return "", fmt.Errorf("CLI worker start failed: %w", err)
	}

	var fullOutput strings.Builder
	var fullStderr strings.Builder
	done := make(chan struct{})

	// Read stdout line by line, emit as chat:delta
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdoutPipe.Read(buf)
			if n > 0 {
				chunk := string(buf[:n])
				fullOutput.WriteString(chunk)
				clean := stripANSI(chunk)
				if emit != nil {
					emit("chat:delta", map[string]any{"content": clean})
				}
			}
			if err != nil {
				break
			}
		}
		close(done)
	}()

	// Read stderr line by line, emit as chat:thinking
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stderrPipe.Read(buf)
			if n > 0 {
				chunk := string(buf[:n])
				fullStderr.WriteString(chunk)
				clean := stripANSI(chunk)
				if emit != nil {
					emit("chat:thinking", map[string]any{"content": clean, "type": "text"})
				}
			}
			if err != nil {
				break
			}
		}
	}()

	err := execCmd.Wait()
	<-done

	out := strings.TrimSpace(fullOutput.String())
	errStr := strings.TrimSpace(fullStderr.String())

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			if emit != nil {
				emit("chat:error", map[string]any{"error": "CLI worker timed out"})
			}
			return out, fmt.Errorf("CLI worker timed out after %v", timeout)
		}
		msg := errStr
		if msg == "" {
			msg = err.Error()
		}
		fmt.Printf("[CLIWorker] command failed: %s %s\n  stdout: %s\n  stderr: %s\n  err: %v\n", cmd.Path, strings.Join(cmd.Args, " "), out, errStr, err)
		if emit != nil {
			emit("chat:error", map[string]any{"error": msg})
		}
		return out, fmt.Errorf("CLI worker failed: %s", msg)
	}

	return out, nil
}

// Execute runs a CLI worker command and returns the combined output.
func Execute(cmd Command) (string, error) {
	timeout := cmd.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	execCmd := exec.CommandContext(ctx, cmd.Path, cmd.Args...)
	execCmd.Env = cmd.Env
	if cmd.Dir != "" {
		execCmd.Dir = cmd.Dir
	}

	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	err := execCmd.Run()

	out := stripANSI(strings.TrimSpace(stdout.String()))
	errStr := stripANSI(strings.TrimSpace(stderr.String()))

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return out, fmt.Errorf("CLI worker timed out after %v", timeout)
		}
		msg := errStr
		if msg == "" {
			msg = err.Error()
		}
		fmt.Printf("[CLIWorker] command failed: %s %s\n  stdout: %s\n  stderr: %s\n  err: %v\n", cmd.Path, strings.Join(cmd.Args, " "), out, errStr, err)
		return out, fmt.Errorf("CLI worker failed: %s", msg)
	}

	return out, nil
}

// ListModels runs the CLI's model listing command and parses the output.
// Expected format: one model per line in "provider/model" format.
func ListModels(rt *Runtime) ([]ModelInfo, error) {
	cmd := rt.BuildListModelsCommand()

	timeout := cmd.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	execCmd := exec.CommandContext(ctx, cmd.Path, cmd.Args...)
	execCmd.Env = cmd.Env
	if cmd.Dir != "" {
		execCmd.Dir = cmd.Dir
	}

	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	if err := execCmd.Run(); err != nil {
		errStr := strings.TrimSpace(stderr.String())
		if errStr != "" {
			return nil, fmt.Errorf("list models failed: %s (cmd: %s %s)", errStr, cmd.Path, strings.Join(cmd.Args, " "))
		}
		return nil, fmt.Errorf("list models failed: %w (cmd: %s %s)", err, cmd.Path, strings.Join(cmd.Args, " "))
	}

	raw := strings.TrimSpace(stdout.String())
	if raw == "" {
		return nil, nil
	}

	lines := strings.Split(raw, "\n")
	seen := make(map[string]bool)
	var models []ModelInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || seen[line] {
			continue
		}
		seen[line] = true

		provider, model := parseModelLine(line)
		if model == "" {
			continue
		}
		models = append(models, ModelInfo{
			ID:           line,
			Name:         model,
			ProviderName: provider,
		})
	}

	return models, nil
}

// parseModelLine splits a "provider/model" string.
func parseModelLine(line string) (provider, model string) {
	if idx := strings.LastIndex(line, "/"); idx > 0 && idx < len(line)-1 {
		return line[:idx], line[idx+1:]
	}
	return "", line
}
