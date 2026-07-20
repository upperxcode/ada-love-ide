package tool

type ToolProfile struct {
	ID    int      `json:"id"`
	Name  string   `json:"name"`
	Color string   `json:"color"`
	Icon  string   `json:"icon"`
	Tools []string `json:"tools"`
}

type ToolUIInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Enabled     bool   `json:"enabled"`
}

func NewProfile(name, color, icon string) ToolProfile {
	return ToolProfile{Name: name, Color: color, Icon: icon, Tools: []string{}}
}
