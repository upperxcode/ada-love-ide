package adapters

import (
	"context"

	core "ada-love-core"
	wiki "github.com/upperxcode/ada-llm-wiki"
)

type WikiAdapter struct {
	mgr *wiki.WikiManager
}

func NewWikiAdapter(mgr *wiki.WikiManager) *WikiAdapter {
	return &WikiAdapter{mgr: mgr}
}

func (a *WikiAdapter) Search(ctx context.Context, query string) []core.WikiArticle {
	articles := a.mgr.Search(ctx, query)
	result := make([]core.WikiArticle, len(articles))
	for i, art := range articles {
		result[i] = core.WikiArticle{
			Title:   art.Title,
			Content: art.Content,
			Tags:    art.Tags,
		}
	}
	return result
}
