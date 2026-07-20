package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"

	"ada-love-ide/internal/engine"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed frontend/build
var frontendAssets embed.FS

func main() {
	eng, err := engine.New()
	if err != nil {
		fmt.Printf("Erro ao inicializar engine: %v\n", err)
		os.Exit(1)
	}
	defer eng.Close()

	app := NewApp(eng)

	// Strip o prefixo "frontend/build" do embed.FS para que o
	// assetserver do Wails encontre index.html e _app/ na raiz.
	assets, _ := fs.Sub(frontendAssets, "frontend/build")

	if err := wails.Run(&options.App{
		Title:  "Ada Love IDE",
		Width:  1400,
		Height: 900,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind:      []interface{}{app},
	}); err != nil {
		fmt.Printf("Erro ao iniciar o app: %v\n", err)
		os.Exit(1)
	}
}
