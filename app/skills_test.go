package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSkill(t *testing.T) {
	tmpDir := t.TempDir()

	filePath := filepath.Join(tmpDir, "test-skill.md")
	content := `---
	name: apple
	description: Deploys the apple service to production.
	---

	Run the deploy script and report the version that was deployed.
`
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	skill, err := parseSkill(filePath)
	if err != nil {
		t.Fatalf("failed to parse skill: %v", err)
	}

	expectedName := "apple"
	if skill.Name != expectedName {
		t.Errorf("skill name: got=%v, want=%v", skill.Name, expectedName)
	}

	expectedDescription := "Deploys the apple service to production."
	if skill.Description != expectedDescription {
		t.Errorf("skill description: got=%v, want=%v", skill.Description, expectedDescription)
	}
}
