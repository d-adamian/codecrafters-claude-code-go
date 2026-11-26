package main

import (
	"os"
	"path/filepath"
	"strings"
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

	expectedContent := "Run the deploy script and report the version that was deployed."
	trimmedContent := strings.TrimSpace(skill.Content)
	if trimmedContent != expectedContent {
		t.Errorf("skill content: got=%v, want=%v", trimmedContent, expectedContent)
	}
}

func TestTransformPrompt(t *testing.T) {
	appleSkill := Skill{
		Name:        "apple",
		Description: "Deploys the apple service to production.",
		Content:     "Run the deploy script and report the version that was deployed.",
	}

	skillWithArgumentSkill := Skill{
		Name:        "args",
		Description: "Test argument substitution.",
		Content:     "Positional args: 0 - $0, 1 - $1, 2 - $2. Full args: $ARGUMENTS",
	}

	testCases := []struct {
		skills   []Skill
		prompt   string
		expected string
	}{
		{
			skills:   []Skill{appleSkill},
			prompt:   "/apple",
			expected: "Run the deploy script and report the version that was deployed.",
		},
		{
			skills:   []Skill{},
			prompt:   "/apple",
			expected: "/apple",
		},
		{
			skills:   []Skill{appleSkill},
			prompt:   "/banana",
			expected: "/banana",
		},
		{
			skills:   []Skill{appleSkill},
			prompt:   "apple",
			expected: "apple",
		},
		{
			skills:   []Skill{skillWithArgumentSkill},
			prompt:   "/args 1 2",
			expected: "Positional args: 0 - 1, 1 - 2, 2 - . Full args: 1 2",
		},
	}

	for _, tc := range testCases {
		result := TransformPrompt(tc.skills, tc.prompt)
		if result != tc.expected {
			t.Errorf("transformPrompt(%v, %v): got=%v, want=%v", tc.skills, tc.prompt, result, tc.expected)
		}
	}
}
