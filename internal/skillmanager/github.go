package skillmanager

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ada-love-ide/internal/config/skill"
)

// GitHubRegistry implementa SkillRegistry para o ecossistema GitHub.
type GitHubRegistry struct {
	baseURL   string
	apiURL    string
	authToken string
	client    *http.Client
}

// NewGitHubRegistry cria um registry GitHub.
// baseURL: URL base do GitHub (ex: "https://github.com").
// authToken: token de autenticação opcional (GitHub PAT).
func NewGitHubRegistry(baseURL, authToken string) *GitHubRegistry {
	if baseURL == "" {
		baseURL = "https://github.com"
	}
	return &GitHubRegistry{
		baseURL:   strings.TrimRight(baseURL, "/"),
		apiURL:    "https://api.github.com",
		authToken: authToken,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        5,
				IdleConnTimeout:     30 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
	}
}

func (r *GitHubRegistry) Name() string { return "github" }

// ── busca ──────────────────────────────────────────────────────────

type gitHubSearchItem struct {
	Path    string  `json:"path"`
	HTMLURL string  `json:"html_url"`
	Score   float64 `json:"score"`
	Repo    struct {
		FullName      string `json:"full_name"`
		Name          string `json:"name"`
		Description   string `json:"description"`
		DefaultBranch string `json:"default_branch"`
	} `json:"repository"`
}

type gitHubSearchResponse struct {
	Items []gitHubSearchItem `json:"items"`
}

func (r *GitHubRegistry) Search(ctx context.Context, query string, limit int) ([]skill.SearchResult, error) {
	u, _ := url.Parse(r.apiURL + "/search/code")
	q := u.Query()
	q.Set("q", fmt.Sprintf("%s filename:SKILL.md", query))
	q.Set("per_page", fmt.Sprintf("%d", limit))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("github: criar requisição: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if r.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.authToken)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github: requisição: %w", err)
	}
	defer resp.Body.Close()

	// Se não autenticado, GitHub pode retornar 403/401 com rate limit
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		// Loga o aviso mas retorna vazio sem erro — o usuário pode não ter token
		fmt.Printf("[GitHubRegistry] Aviso: busca GitHub sem autenticação pode ser limitada. Status %d: %s\n",
			resp.StatusCode, strings.TrimSpace(string(body)))
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("github: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // 2MB max
	if err != nil {
		return nil, fmt.Errorf("github: ler resposta: %w", err)
	}

	var searchResp gitHubSearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("github: decodificar JSON: %w", err)
	}

	// Deduplica por slug (full_name + diretório), mantendo maior score
	seen := make(map[string]float64)
	deduped := make(map[string]gitHubSearchItem)

	for _, item := range searchResp.Items {
		// Slug: owner/repo/diretório
		dir := filepath.Dir(item.Path)
		slug := item.Repo.FullName
		if dir != "" && dir != "." {
			slug = item.Repo.FullName + "/" + dir
		}

		if existingScore, ok := seen[slug]; ok && item.Score <= existingScore {
			continue
		}
		seen[slug] = item.Score
		deduped[slug] = item
	}

	out := make([]skill.SearchResult, 0, len(deduped))
	for slug, item := range deduped {
		// Nome de exibição: diretório ou nome do repo
		displayName := filepath.Base(slug)
		summary := item.Repo.Description
		version := item.Repo.DefaultBranch

		// Se o item tem um subdiretório, usa o nome do diretório
		dir := filepath.Dir(item.Path)
		if dir != "" && dir != "." {
			displayName = filepath.Base(dir)
		}

		out = append(out, skill.SearchResult{
			Name:         slug,
			DisplayName:  displayName,
			RegistryName: "github",
			Summary:      summary,
			Description:  summary,
			Slug:         slug,
			Version:      version,
			Score:        item.Score,
		})
	}

	return out, nil
}

// ── download e instalação ──────────────────────────────────────────

func (r *GitHubRegistry) DownloadAndInstall(ctx context.Context, slug, version, targetDir string) error {
	// slug: "owner/repo" ou "owner/repo/subdir"
	parts := strings.SplitN(slug, "/", 3)
	if len(parts) < 2 {
		return fmt.Errorf("github: slug inválido: %q", slug)
	}
	owner := parts[0]
	repo := parts[1]
	subdir := ""
	if len(parts) > 2 {
		subdir = parts[2]
	}

	if version == "" {
		version = "HEAD"
	}

	// URL do archive ZIP: https://api.github.com/repos/{owner}/{repo}/zipball/{ref}
	archiveURL := fmt.Sprintf("%s/repos/%s/%s/zipball/%s", r.apiURL, owner, repo, version)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, archiveURL, nil)
	if err != nil {
		return fmt.Errorf("github: criar req archive: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if r.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.authToken)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("github: download archive: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("github: archive status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	// Limite de 50MB para o ZIP
	zipData, err := io.ReadAll(io.LimitReader(resp.Body, 50<<20+1))
	if err != nil {
		return fmt.Errorf("github: ler zip: %w", err)
	}
	if len(zipData) > 50<<20 {
		return fmt.Errorf("github: zip muito grande (>50MB)")
	}

	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("github: abrir zip: %w", err)
	}

	// O ZIP do GitHub tem um diretório raiz type "repo-branch/" que precisamos pular
	// Ex: "ada-love-ai/my-skill-123abc/SKILL.md"
	// Precisamos extrair os arquivos do subdiretório correto
	skipPrefix := ""
	if subdir != "" {
		// Encontra o prefixo do diretório raiz no ZIP
		for _, f := range zipReader.File {
			if strings.HasSuffix(f.Name, "/") {
				continue
			}
			// Ex: "owner-repo-abc123/subdir/SKILL.md"
			// Precisamos do prefixo: "owner-repo-abc123/subdir/"
			parts := strings.SplitN(f.Name, "/", 3)
			if len(parts) >= 3 && parts[1] == subdir {
				skipPrefix = parts[0] + "/" + subdir + "/"
				break
			}
			if len(parts) >= 2 {
				skipPrefix = parts[0] + "/"
				break
			}
		}
	} else {
		// Encontra o prefixo do diretório raiz
		for _, f := range zipReader.File {
			if strings.HasSuffix(f.Name, "/") {
				continue
			}
			parts := strings.SplitN(f.Name, "/", 2)
			if len(parts) >= 2 {
				skipPrefix = parts[0] + "/"
				break
			}
		}
	}

	// Extrai os arquivos, pulando o prefixo
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("github: criar diretório: %w", err)
	}

	for _, f := range zipReader.File {
		name := f.Name
		// Pula o diretório raiz do ZIP
		if skipPrefix != "" && strings.HasPrefix(name, skipPrefix) {
			name = strings.TrimPrefix(name, skipPrefix)
		}

		// Sanitize path
		fpath := filepath.Join(targetDir, name)
		if !strings.HasPrefix(filepath.Clean(fpath), filepath.Clean(targetDir)+string(os.PathSeparator)) {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0o755)
			continue
		}

		os.MkdirAll(filepath.Dir(fpath), 0o755)

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("github: abrir arquivo no zip: %w", err)
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return fmt.Errorf("github: criar arquivo: %w", err)
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return fmt.Errorf("github: extrair arquivo: %w", err)
		}
	}

	return nil
}
