package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadExperts(t *testing.T) {
	tmpDir := t.TempDir()
	yamlContent := `experts:
  - id: "go-expert"
    name: "Go Expert"
    description: "Go specialist"
    language: "go"
    start_command: "./go-expert/expert"
    triggers:
      - "go.mod"
  - id: "flutter-expert"
    name: "Flutter Expert"
    description: "Flutter specialist"
    language: "flutter"
    start_command: "./flutter-expert/expert"
    triggers:
      - "pubspec.yaml"
`
	yamlPath := filepath.Join(tmpDir, "experts.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	experts, err := LoadExperts(yamlPath)
	if err != nil {
		t.Fatalf("LoadExperts failed: %v", err)
	}

	if len(experts) != 2 {
		t.Fatalf("Expected 2 experts, got %d", len(experts))
	}

	if experts[0].ID != "go-expert" {
		t.Errorf("Expected first expert ID 'go-expert', got '%s'", experts[0].ID)
	}
	if experts[0].Language != "go" {
		t.Errorf("Expected first expert language 'go', got '%s'", experts[0].Language)
	}
	if experts[1].ID != "flutter-expert" {
		t.Errorf("Expected second expert ID 'flutter-expert', got '%s'", experts[1].ID)
	}
}

func TestFindExpertByLanguage(t *testing.T) {
	experts := []*ExpertPlugin{
		{ID: "go-expert", Language: "go"},
		{ID: "flutter-expert", Language: "flutter"},
	}

	p, ok := FindExpertByLanguage("go", experts)
	if !ok || p.ID != "go-expert" {
		t.Errorf("Expected to find go-expert for language 'go'")
	}

	p, ok = FindExpertByLanguage("python", experts)
	if ok {
		t.Errorf("Expected not to find expert for language 'python'")
	}
}
