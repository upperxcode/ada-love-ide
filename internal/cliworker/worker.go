package cliworker

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"ada-love-ide/internal/config/worker"
)

type Runtime struct {
	Config        worker.WorkerConfig
	WorkspaceDir  string
}

func New(cfg worker.WorkerConfig) *Runtime {
	return &Runtime{Config: cfg}
}

func (r *Runtime) WithWorkspace(dir string) *Runtime {
	r.WorkspaceDir = dir
	return r
}

type Command struct {
	Path    string
	Args    []string
	Env     []string
	Dir     string
	Timeout time.Duration
}

func (r *Runtime) BuildCommand(message, model string) Command {
	return r.buildArgs(message, model, false)
}

func (r *Runtime) BuildListModelsCommand() Command {
	return r.buildArgs("", "", true)
}

func (r *Runtime) buildArgs(message, model string, listModels bool) Command {
	path := r.Config.Command
	if path == "" {
		path = "opencode"
	}

	var args []string
	if listModels {
		modelsCmd := r.Config.ModelsCommand
		if modelsCmd == "" {
			modelsCmd = "models"
		}
		args = append(args, splitArgs(modelsCmd)...)
	} else {
		if r.Config.Arguments != "" {
			args = append(args, splitArgs(r.Config.Arguments)...)
		}
		if model != "" {
			args = append(args, "--model", model)
		}
		args = append(args, message)
	}

	env := buildEnv(r.Config.Environment)

	return Command{
		Path:    path,
		Args:    args,
		Env:     env,
		Dir:     r.WorkspaceDir,
		Timeout: 120 * time.Second,
	}
}

func splitArgs(raw string) []string {
	var args []string
	for _, part := range strings.Fields(raw) {
		part = strings.TrimSpace(part)
		if part != "" {
			args = append(args, part)
		}
	}
	return args
}

func buildEnv(envJSON string) []string {
	base := os.Environ()

	if envJSON == "" || envJSON == "{}" {
		return base
	}

	var extra map[string]string
	if err := json.Unmarshal([]byte(envJSON), &extra); err != nil {
		return base
	}

	merge := make(map[string]string, len(base)+len(extra))
	for _, e := range base {
		if k, v, ok := strings.Cut(e, "="); ok {
			merge[k] = v
		}
	}
	for k, v := range extra {
		merge[k] = v
	}

	result := make([]string, 0, len(merge))
	for k, v := range merge {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}
