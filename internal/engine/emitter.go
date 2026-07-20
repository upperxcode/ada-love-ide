package engine

import (
	"context"

	"ada-love-ide/internal/chat"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Emitter é a interface usada pelo pacote chat, re-exposta aqui
// para App poder referenciá-la sem acoplar diretamente a `chat`.
type Emitter = chat.Emitter

// wailsEmitter adapta o context do Wails para a interface chat.Emitter.
type wailsEmitter struct{ ctx context.Context }

// NewEmitter devolve um chat.Emitter que envia via runtime.EventsEmit.
func NewEmitter(ctx context.Context) Emitter { return wailsEmitter{ctx: ctx} }

func (w wailsEmitter) Emit(event string, data ...any) {
	if w.ctx == nil {
		return
	}
	runtime.EventsEmit(w.ctx, event, data...)
}
