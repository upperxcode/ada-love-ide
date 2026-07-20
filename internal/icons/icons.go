package icons

import (
	"strconv"
	"sync"
)

type IconData struct {
	Elements   string
	StrokeW    float64
	StrokeLine string
}

type IconSet map[string]IconData

var (
	currentSet  IconSet
	currentName string
	mu          sync.RWMutex
)

func init() {
	currentSet = Lucide()
	currentName = "lucide"
}

func Register(set IconSet) {
	mu.Lock()
	defer mu.Unlock()
	currentSet = set
}

func SetTheme(name string) {
	mu.Lock()
	defer mu.Unlock()
	switch name {
	case "material":
		currentSet = Material()
	default:
		currentSet = Lucide()
	}
	currentName = name
}

func CurrentTheme() string {
	mu.RLock()
	defer mu.RUnlock()
	return currentName
}

func AvailableThemes() []string {
	return []string{"lucide", "material"}
}

func Get(name string) IconData {
	mu.RLock()
	defer mu.RUnlock()
	if currentSet == nil {
		return IconData{}
	}
	if icon, ok := currentSet[name]; ok {
		return icon
	}
	return IconData{}
}

func SVG(name string, class string, color ...string) string {
	icon := Get(name)
	if icon.Elements == "" {
		return ""
	}
	sw := "2"
	if icon.StrokeW > 0 {
		sw = strconv.FormatFloat(icon.StrokeW, 'f', -1, 64)
	}
	sl := "round"
	if icon.StrokeLine != "" {
		sl = icon.StrokeLine
	}
	style := ""
	if len(color) > 0 && color[0] != "" {
		style = ` style="color:` + color[0] + `;"`
	}
	classAttr := ""
	if class != "" {
		classAttr = ` class="` + class + `"`
	}
	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="` + sw + `" stroke-linecap="` + sl + `" stroke-linejoin="` + sl + `"` + classAttr + style + `>` + icon.Elements + `</svg>`
}

const (
	Sparkle     = "sparkle"
	Plus        = "plus"
	Edit        = "edit"
	Delete      = "delete"
	ChevronDown = "chevron-down"
	Settings    = "settings"
	User        = "user"
	PenTool     = "pen-tool"
	Sliders     = "sliders"
	Shield      = "shield"
	Layers      = "layers"
	Home        = "home"
	Sun         = "sun"
	PenEdit     = "pen-edit"
	MonitorCode = "monitor-code"
	Compass     = "compass"
	AlertCircle = "alert-circle"
	Book        = "book"
	Logout      = "logout"
	File        = "file"
	Sidebar     = "sidebar"
	Paperclip   = "paperclip"
	Crosshair   = "crosshair"
	ContextRing = "context-ring"
	Send        = "send"
	Expand      = "expand"
	Collapse    = "collapse"
	Cog         = "cog"
)
