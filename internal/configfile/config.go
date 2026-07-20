package configfile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type GeneralConfig struct {
	IconTheme string `json:"icon_theme,omitempty"`
	Theme     string `json:"theme,omitempty"`
	FontTheme string `json:"font_theme,omitempty"`
	FontSize  string `json:"font_size,omitempty"`
	Vibrance  int    `json:"vibrance,omitempty"`
}

var (
	mu       sync.RWMutex
	filePath string
)

func init() {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "ada-love-ide")
	os.MkdirAll(dir, 0o755)
	filePath = filepath.Join(dir, "config.json")
}

func path() string {
	mu.RLock()
	defer mu.RUnlock()
	return filePath
}

func Load() GeneralConfig {
	var cfg GeneralConfig
	data, err := os.ReadFile(path())
	if err != nil {
		return cfg
	}
	json.Unmarshal(data, &cfg)
	return cfg
}

func Save(cfg GeneralConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path(), data, 0o644)
}

func Update(fn func(cfg *GeneralConfig)) {
	cfg := Load()
	fn(&cfg)
	Save(cfg)
}
