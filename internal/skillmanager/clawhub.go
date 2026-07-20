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

// ClawHubRegistry implementa SkillRegistry para o ecossistema ClawHub.
type ClawHubRegistry struct {
	baseURL   string
	authToken string
	client    *http.Client
}

// NewClawHubRegistry cria um registry ClawHub.
// baseURL: URL base da API (ex: "https://clawhub.ai").
// authToken: token de autenticação opcional.
func NewClawHubRegistry(baseURL, authToken string) *ClawHubRegistry {
	if baseURL == "" {
		baseURL = "https://clawhub.ai"
	}
	return &ClawHubRegistry{
		baseURL:   strings.TrimRight(baseURL, "/"),
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

func (r *ClawHubRegistry) Name() string { return "clawhub" }

// ── busca ──────────────────────────────────────────────────────────

type clawhubSearchResult struct {
	Score       float64 `json:"score"`
	Slug        string  `json:"slug"`
	DisplayName string  `json:"display_name"`
	Summary     string  `json:"summary"`
	Version     string  `json:"version"`
}

type clawhubSearchResponse struct {
	Results []clawhubSearchResult `json:"results"`
}

func (r *ClawHubRegistry) Search(ctx context.Context, query string, limit int) ([]skill.SearchResult, error) {
	u, _ := url.Parse(r.baseURL + "/api/v1/search")
	q := u.Query()
	q.Set("q", query)
	q.Set("limit", fmt.Sprintf("%d", limit))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("clawhub: criar requisição: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if r.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.authToken)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("clawhub: requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("clawhub: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // 2MB max
	if err != nil {
		return nil, fmt.Errorf("clawhub: ler resposta: %w", err)
	}

	var searchResp clawhubSearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("clawhub: decodificar JSON: %w", err)
	}

	out := make([]skill.SearchResult, 0, len(searchResp.Results))
	for _, sr := range searchResp.Results {
		if sr.Slug == "" {
			continue
		}
		out = append(out, skill.SearchResult{
			Name:         sr.Slug,
			DisplayName:  sr.DisplayName,
			RegistryName: "clawhub",
			Summary:      sr.Summary,
			Version:      sr.Version,
			Slug:         sr.Slug,
			Score:        sr.Score,
		})
	}
	return out, nil
}

// ── download e instalação ──────────────────────────────────────────

func (r *ClawHubRegistry) DownloadAndInstall(ctx context.Context, slug, version, targetDir string) error {
	u, _ := url.Parse(r.baseURL + "/api/v1/download")
	q := u.Query()
	q.Set("slug", slug)
	if version != "" {
		q.Set("version", version)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("clawhub: criar req download: %w", err)
	}
	req.Header.Set("Accept", "application/zip")
	if r.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.authToken)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("clawhub: download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("clawhub: download status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	// Limite de 50MB para o ZIP
	zipData, err := io.ReadAll(io.LimitReader(resp.Body, 50<<20+1))
	if err != nil {
		return fmt.Errorf("clawhub: ler zip: %w", err)
	}
	if len(zipData) > 50<<20 {
		return fmt.Errorf("clawhub: zip muito grande (>50MB)")
	}

	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("clawhub: abrir zip: %w", err)
	}

	// Extrai o ZIP no diretório alvo
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("clawhub: criar diretório: %w", err)
	}

	for _, f := range zipReader.File {
		// Sanitize path
		fpath := filepath.Join(targetDir, f.Name)
		if !strings.HasPrefix(filepath.Clean(fpath), filepath.Clean(targetDir)+string(os.PathSeparator)) {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0o755)
			continue
		}

		// Cria diretório pai
		os.MkdirAll(filepath.Dir(fpath), 0o755)

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("clawhub: abrir arquivo no zip: %w", err)
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return fmt.Errorf("clawhub: criar arquivo: %w", err)
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return fmt.Errorf("clawhub: extrair arquivo: %w", err)
		}
	}

	return nil
}
