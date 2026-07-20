package db

// skillStateTable migration adds color, icon, and active fields to skills table
// This allows skills to have custom appearance and be enabled/disabled per workspace
var skillStateTable = `
-- Skills table already exists with: id, name, description, tags, content
-- We add optional fields for customization

-- Add color field for skill header background (6-char hex or named color)
ALTER TABLE skills ADD COLUMN color TEXT DEFAULT '';

-- Add icon field for skill emoji/icon display
ALTER TABLE skills ADD COLUMN icon TEXT DEFAULT '';

-- Add active field to enable/disable skills
ALTER TABLE skills ADD COLUMN active INTEGER NOT NULL DEFAULT 1;

-- Add description field if not exists (may already exist)
ALTER TABLE skills ADD COLUMN description TEXT DEFAULT '';
`

// Note: SQLite doesn't support ADD COLUMN IF NOT EXISTS before v3.35.0
// These migrations are designed to be idempotent - errors on re-run are ignored
