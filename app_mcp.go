package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"ada-love-ide/internal/config/mcp"
)

// MCPServerRegistry represents a server entry from the MCP registry.
type MCPServerRegistry struct {
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Language    string   `json:"language"`
	Stars       int      `json:"stars"`
	Topics      []string `json:"topics"`
	Tags        string   `json:"tags"`
}

// gitHubRepo represents a single item from the GitHub search API.
type gitHubRepo struct {
	FullName        string   `json:"full_name"`
	Description     string   `json:"description"`
	HTMLURL         string   `json:"html_url"`
	Language        string   `json:"language"`
	StargazersCount int      `json:"stargazers_count"`
	Topics          []string `json:"topics"`
}

type gitHubSearchResponse struct {
	TotalCount int          `json:"total_count"`
	Items      []gitHubRepo `json:"items"`
}

// SearchMCPServers busca servidores MCP no GitHub com topic:mcp-server.
func (a *App) SearchMCPServers(query string) []MCPServerRegistry {
	entries := fetchMCPServersFromGitHub()
	if query == "" {
		return entries
	}
	q := strings.ToLower(query)
	var out []MCPServerRegistry
	for _, e := range entries {
		if strings.Contains(strings.ToLower(e.Name), q) ||
			strings.Contains(strings.ToLower(e.Description), q) {
			out = append(out, e)
		}
	}
	return out
}

func fetchMCPServersFromGitHub() []MCPServerRegistry {
	client := &http.Client{Timeout: 15 * time.Second}

	// Busca repositórios com topic "mcp-server" ordenados por stars
	url := "https://api.github.com/search/repositories?q=topic:mcp-server+in:name&sort=stars&order=desc&per_page=50"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil
	}

	var searchResp gitHubSearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil
	}

	entries := make([]MCPServerRegistry, 0, len(searchResp.Items))
	for _, item := range searchResp.Items {
		desc := item.Description
		if len(desc) > 200 {
			desc = desc[:200] + "..."
		}
		tags := strings.Join(item.Topics, ", ")
		if len(tags) > 60 {
			tags = tags[:60] + "..."
		}
		entries = append(entries, MCPServerRegistry{
			Name:        item.FullName,
			URL:         item.HTMLURL,
			Description: desc,
			Language:    item.Language,
			Stars:       item.StargazersCount,
			Topics:      item.Topics,
			Tags:        tags,
		})
	}
	return entries
}

// InstallMCPFromRepo clona um repositório GitHub, instala dependências e configura o MCP server.
func (a *App) InstallMCPFromRepo(repoUrl string) error {
	// Extrai owner/repo da URL
	repoPath := extractRepoPath(repoUrl)
	if repoPath == "" {
		return fmt.Errorf("URL de repositório inválida: %s", repoUrl)
	}

	home, _ := os.UserHomeDir()
	mcpDir := filepath.Join(home, ".ada-love-ide", "mcp-servers")
	os.MkdirAll(mcpDir, 0o755)

	// Nome do diretório: owner-repo
	dirName := strings.ReplaceAll(repoPath, "/", "-")
	targetDir := filepath.Join(mcpDir, dirName)

	// Se já existe, faz git pull; senão, clona
	if _, err := os.Stat(targetDir); err == nil {
		runCmd(targetDir, "git", "pull", "origin", "main")
		runCmd(targetDir, "git", "pull", "origin", "master")
	} else {
		if err := runCmd("", "git", "clone", fmt.Sprintf("https://github.com/%s.git", repoPath), targetDir); err != nil {
			return fmt.Errorf("falha ao clonar repositório: %w", err)
		}
	}

	// Detecta package manager e instala
	pm, runCmdStr, args := detectPackageManager(targetDir)
	if pm != "" {
		fmt.Printf("[MCP] Instalando com %s em %s\n", pm, targetDir)
		if err := runCmd(targetDir, pm, "install"); err != nil {
			return fmt.Errorf("falha ao instalar dependências com %s: %w", pm, err)
		}
	}

	// Detecta o comando para executar o MCP server
	command, detectedArgs := detectRunCommand(targetDir, pm, runCmdStr, args)

	// Nome amigável
	name := filepath.Base(repoPath)
	if name == "" {
		name = dirName
	}

	// Salva no banco
	server := mcp.MCPServerUI{
		Command: command,
		Args:    detectedArgs,
		Env:     map[string]string{},
		URL:     "",
		Enabled: true,
		Icon:    "🔌",
		Color:   "#6b7280",
	}

	a.eng.DB.SaveMCPServer(name, server)
	fmt.Printf("[MCP] Servidor %s instalado e configurado. Comando: %s %v\n", name, command, detectedArgs)
	return nil
}

func extractRepoPath(repoURL string) string {
	// Aceita: https://github.com/owner/repo, https://github.com/owner/repo.git, owner/repo
	repoURL = strings.TrimSuffix(repoURL, ".git")
	if strings.Contains(repoURL, "github.com/") {
		parts := strings.SplitN(repoURL, "github.com/", 2)
		if len(parts) == 2 {
			return strings.Trim(strings.SplitN(parts[1], "/", 3)[0]+"/"+strings.SplitN(parts[1], "/", 3)[1], "/")
		}
	}
	if strings.Count(repoURL, "/") == 1 && !strings.Contains(repoURL, "://") {
		return repoURL
	}
	return ""
}

func detectPackageManager(dir string) (pm, runCmd, args string) {
	if fileExists(filepath.Join(dir, "package.json")) {
		// Tenta usar npx, mas instala com npm
		if hasTool("bun") {
			return "bun", "bun", "run"
		}
		return "npm", "npx", ""
	}
	if fileExists(filepath.Join(dir, "pyproject.toml")) || fileExists(filepath.Join(dir, "setup.py")) || fileExists(filepath.Join(dir, "requirements.txt")) {
		if hasTool("uvx") {
			return "uv", "uvx", ""
		}
		if hasTool("pip3") {
			return "pip3", "python3", "-m"
		}
		return "pip", "python", "-m"
	}
	if fileExists(filepath.Join(dir, "Cargo.toml")) {
		if hasTool("cargo") {
			return "cargo", "cargo", "run"
		}
	}
	if fileExists(filepath.Join(dir, "go.mod")) {
		if hasTool("go") {
			return "go", "go", "run"
		}
	}
	return "", "", ""
}

func detectRunCommand(dir, pm, runCmd, args string) (string, []string) {
	// Tenta ler package.json para encontrar bin ou main
	if fileExists(filepath.Join(dir, "package.json")) {
		data, err := os.ReadFile(filepath.Join(dir, "package.json"))
		if err == nil {
			var pkg struct {
				Bin  map[string]string `json:"bin"`
				Main string            `json:"main"`
				Name string            `json:"name"`
			}
			if json.Unmarshal(data, &pkg) == nil {
				// Se tem bin, usa npx + nome do bin
				for name := range pkg.Bin {
					if runCmd == "npx" {
						return runCmd, []string{"-y", name}
					}
					return runCmd, []string{"run", name}
				}
				// Se tem main, usa node
				if pkg.Main != "" {
					return "node", []string{pkg.Main}
				}
				// Tenta npx -y com o nome do pacote
				if runCmd == "npx" && pkg.Name != "" {
					return runCmd, []string{"-y", pkg.Name}
				}
			}
		}
		// Fallback: npm package
		if runCmd == "npx" {
			return runCmd, []string{"-y", filepath.Base(dir)}
		}
	}

	// Tenta detectar entry point Python
	if pm == "pip" || pm == "pip3" || pm == "uv" {
		// Procura por entry_points no pyproject.toml
		if fileExists(filepath.Join(dir, "pyproject.toml")) {
			data, _ := os.ReadFile(filepath.Join(dir, "pyproject.toml"))
			re := regexp.MustCompile(`"([^=]+)=[^:]+:([^\"]+)"`)
			if m := re.FindStringSubmatch(string(data)); len(m) >= 3 {
				script := strings.TrimSpace(m[1])
				return runCmd, []string{script}
			}
		}
		// Procura por arquivos .py na raiz com CLI
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".py") {
				data, _ := os.ReadFile(filepath.Join(dir, e.Name()))
				if strings.Contains(string(data), "if __name__") || strings.Contains(string(data), "def main") {
					return runCmd, []string{strings.TrimSuffix(e.Name(), ".py")}
				}
			}
		}
	}

	// Fallback para Rust
	if pm == "cargo" {
		return "cargo", []string{"run", "--release"}
	}
	// Fallback para Go
	if pm == "go" {
		return "go", []string{"run", "."}
	}

	return runCmd, nil
}

func hasTool(name string) bool {
	return exec.Command("which", name).Run() == nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func runCmd(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// TestMCPConnection testa a conectividade de um servidor MCP.
func (a *App) TestMCPConnection(name, command, urlStr string, args []string) mcp.ConnectionTestResult {
	start := time.Now()
	if urlStr != "" {
		return testMCPURL(urlStr, start)
	}
	if command != "" {
		return testMCPCLI(command, args, start)
	}
	return mcp.ConnectionTestResult{Success: false, Message: "Nenhum comando ou URL configurado"}
}

func testMCPURL(rawURL string, start time.Time) mcp.ConnectionTestResult {
	u, err := url.Parse(rawURL)
	if err != nil {
		return mcp.ConnectionTestResult{Success: false, Message: fmt.Sprintf("URL inválida: %v", err)}
	}
	host := u.Host
	if host == "" {
		host = u.Path
	}
	_, port, _ := net.SplitHostPort(host)
	if port == "" {
		if u.Scheme == "wss" || u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
		host = net.JoinHostPort(strings.Split(host, ":")[0], port)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(rawURL)
	if err == nil {
		resp.Body.Close()
		elapsed := time.Since(start).Milliseconds()
		return mcp.ConnectionTestResult{Success: true, Message: fmt.Sprintf("Conexão HTTP OK (status %d)", resp.StatusCode), LatencyMS: int(elapsed)}
	}
	conn, err := net.DialTimeout("tcp", host, 5*time.Second)
	if err != nil {
		elapsed := time.Since(start).Milliseconds()
		return mcp.ConnectionTestResult{Success: false, Message: fmt.Sprintf("Falha na conexão: %v", err), LatencyMS: int(elapsed)}
	}
	conn.Close()
	elapsed := time.Since(start).Milliseconds()
	return mcp.ConnectionTestResult{Success: true, Message: fmt.Sprintf("Conexão TCP OK (%s)", host), LatencyMS: int(elapsed)}
}

func testMCPCLI(command string, args []string, start time.Time) mcp.ConnectionTestResult {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, command, args...)
	output, err := cmd.CombinedOutput()
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		return mcp.ConnectionTestResult{Success: false, Message: fmt.Sprintf("Falha: %v", err), LatencyMS: int(elapsed)}
	}
	msg := strings.TrimSpace(string(output))
	if len(msg) > 200 {
		msg = msg[:200] + "..."
	}
	if msg == "" {
		msg = "Comando executado com sucesso (sem output)"
	}
	return mcp.ConnectionTestResult{Success: true, Message: msg, LatencyMS: int(elapsed)}
}
