package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	storage "github.com/ada-love-ai/storage/storage"
)

// AddSessionAttachment anexa um arquivo à sessão.
// Retorna erro se o arquivo não existir em disco ou se já estiver anexado.
func (a *App) AddSessionAttachment(sessionID, filePath string) (*storage.SessionAttachment, error) {
	ctx := context.Background()

	// Verifica sessão existe
	sess, ok := a.eng.DB.GetSession(sessionID)
	if !ok {
		return nil, fmt.Errorf("sessão %s não encontrada", sessionID)
	}
	_ = sess

	// Verifica arquivo existe em disco
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("arquivo não encontrado: %s", filePath)
		}
		return nil, fmt.Errorf("erro ao acessar arquivo: %w", err)
	}

	// Verifica duplicata
	exists, err := a.eng.DB.Attachments().AttachmentExists(ctx, sessionID, filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao verificar anexo: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("arquivo já anexado a esta sessão: %s", filepath.Base(filePath))
	}

	// Detecta mime type pela extensão
	mimeType := detectMimeType(filePath)

	attachment := &storage.SessionAttachment{
		SessionID: sessionID,
		FilePath:  filePath,
		FileName:  filepath.Base(filePath),
		FileSize:  info.Size(),
		MimeType:  mimeType,
	}

	if err := a.eng.DB.Attachments().AddAttachment(ctx, attachment); err != nil {
		return nil, fmt.Errorf("erro ao salvar anexo: %w", err)
	}

	return attachment, nil
}

// RemoveSessionAttachment remove um anexo da sessão.
func (a *App) RemoveSessionAttachment(sessionID, filePath string) error {
	ctx := context.Background()
	return a.eng.DB.Attachments().DeleteAttachment(ctx, sessionID, filePath)
}

// ListSessionAttachments lista todos os anexos de uma sessão.
func (a *App) ListSessionAttachments(sessionID string) ([]storage.SessionAttachment, error) {
	ctx := context.Background()
	return a.eng.DB.Attachments().ListAttachments(ctx, sessionID)
}

// CheckAttachmentExists verifica se um arquivo já está anexado à sessão.
func (a *App) CheckAttachmentExists(sessionID, filePath string) (bool, error) {
	ctx := context.Background()
	return a.eng.DB.Attachments().AttachmentExists(ctx, sessionID, filePath)
}

func detectMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".txt", ".md", ".go", ".ts", ".tsx", ".js", ".jsx", ".py", ".rs", ".java",
		".c", ".cpp", ".h", ".hpp", ".css", ".scss", ".html", ".json", ".yaml", ".yml",
		".toml", ".xml", ".sh", ".bash", ".sql", ".rb", ".php", ".swift", ".kt", ".dart":
		return "text/plain"
	case ".pdf":
		return "application/pdf"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".svg":
		return "image/svg+xml"
	case ".csv":
		return "text/csv"
	default:
		return "application/octet-stream"
	}
}
