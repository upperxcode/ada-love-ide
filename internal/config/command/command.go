package command

type SubCommandInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ArgsUsage   string `json:"args_usage"`
}

type CommandInfo struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Usage       string           `json:"usage"`
	Aliases     []string         `json:"aliases"`
	SubCommands []SubCommandInfo `json:"sub_commands"`
}
