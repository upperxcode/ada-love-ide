package workspace

import "time"

type WorkspaceTemplate struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Personality string    `json:"personality"`
	CreatedAt   time.Time `json:"created_at"`
}
