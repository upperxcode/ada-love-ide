package knowledge

// Entry stores a single knowledge item with its embedding vector.
type Entry struct {
	KnowledgeID int64   // index within the workspace (0-based)
	WorkspaceID int64   // workspace this entry belongs to
	Text        string  // original knowledge item text
	Vector      []float32 // embedding vector
}
