package urlworker

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type ServerManager struct {
	mu     sync.Mutex
	procs  map[string]*managedServer
}

type managedServer struct {
	WorkerName string
	Cmd        *exec.Cmd
	Port       int
	StartedAt  time.Time
	BaseURL    string
}

var DefaultServerManager = &ServerManager{
	procs: make(map[string]*managedServer),
}

// Start starts the URL worker's server process in the background.
func (sm *ServerManager) Start(cfg Runtime) error {
	sm.mu.Lock()
	if existing, ok := sm.procs[cfg.Config.Name]; ok {
		sm.mu.Unlock()
		if existing.Cmd != nil && existing.Cmd.Process != nil && existing.Cmd.ProcessState == nil {
			return fmt.Errorf("server for worker %q is already running on port %d", cfg.Config.Name, existing.Port)
		}
		delete(sm.procs, cfg.Config.Name)
	}
	sm.mu.Unlock()

	cmdStr := strings.TrimSpace(cfg.URLPaths.StartCommand)
	if cmdStr == "" {
		return fmt.Errorf("no start command configured for worker %q", cfg.Config.Name)
	}

	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return fmt.Errorf("invalid start command for worker %q", cfg.Config.Name)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server for %q: %w", cfg.Config.Name, err)
	}

	sm.mu.Lock()
	ms := &managedServer{
		WorkerName: cfg.Config.Name,
		Cmd:        cmd,
		StartedAt:  time.Now(),
		BaseURL:    cfg.BaseURL,
	}
	sm.procs[cfg.Config.Name] = ms
	sm.mu.Unlock()

	// Try to extract port from base URL
	if h := strings.Split(cfg.BaseURL, ":"); len(h) >= 2 {
		portStr := strings.Split(h[len(h)-1], "/")[0]
		fmt.Sscanf(portStr, "%d", &ms.Port)
	}

	// Wait briefly for the server to be ready
	for i := 0; i < 15; i++ {
		if sm.IsRunning(cfg.Config.Name) {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("server for %q started but not responding after 7.5s", cfg.Config.Name)
}

// Stop kills the URL worker's server process.
func (sm *ServerManager) Stop(workerName string) error {
	sm.mu.Lock()
	ms, ok := sm.procs[workerName]
	if !ok {
		sm.mu.Unlock()
		return nil
	}
	delete(sm.procs, workerName)
	sm.mu.Unlock()

	if ms.Cmd != nil && ms.Cmd.Process != nil {
		return ms.Cmd.Process.Kill()
	}
	return nil
}

// StopAll kills all managed server processes.
func (sm *ServerManager) StopAll() {
	sm.mu.Lock()
	names := make([]string, 0, len(sm.procs))
	for name := range sm.procs {
		names = append(names, name)
	}
	sm.mu.Unlock()

	for _, name := range names {
		_ = sm.Stop(name)
	}
}

// IsRunning checks if the server is alive via TCP + HTTP health check.
func (sm *ServerManager) IsRunning(workerName string) bool {
	sm.mu.Lock()
	ms, ok := sm.procs[workerName]
	sm.mu.Unlock()
	if !ok || ms == nil {
		return false
	}

	// Check if process is still alive
	if ms.Cmd != nil && ms.Cmd.Process != nil && ms.Cmd.ProcessState != nil {
		return false
	}

	// Try TCP dial first
	host := "localhost"
	port := ms.Port
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err == nil {
		conn.Close()
		return true
	}

	// Try HTTP GET as fallback
	if ms.BaseURL != "" {
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get(ms.BaseURL)
		if err == nil {
			resp.Body.Close()
			return true
		}
	}

	return false
}

// Status returns the current server status.
func (sm *ServerManager) Status(workerName string) (running bool, port int, baseURL string, uptime string) {
	sm.mu.Lock()
	ms, ok := sm.procs[workerName]
	sm.mu.Unlock()
	if !ok || ms == nil {
		return false, 0, "", ""
	}
	uptime = time.Since(ms.StartedAt).Truncate(time.Second).String()
	return sm.IsRunning(workerName), ms.Port, ms.BaseURL, uptime
}
