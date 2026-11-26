package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openai/openai-go/v3"
)

type Skill struct {
	Name        string
	Description string
	Content     string
}

func ParseSkills() ([]Skill, error) {
	matches, err := filepath.Glob(".claude/skills/*/SKILL.md")
	if err != nil {
		return nil, fmt.Errorf("failed to glob skills: %w", err)
	}

	skills := make([]Skill, 0)
	for _, path := range matches {
		skill, err := parseSkill(path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse skill: %w", err)
		}
		skills = append(skills, *skill)
	}

	return skills, nil
}

func parseSkill(path string) (*Skill, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("file is empty: %v", path)
	}

	frontMatter, lastFrontMatter, err := readFrontMatter(lines)
	if err != nil {
		return nil, fmt.Errorf("failed to read frontmatter for file %v: %w", path, err)
	}
	skillContent := strings.Join(lines[lastFrontMatter+1:], "\n")

	name := frontMatter["name"]
	description := frontMatter["description"]
	return &Skill{Name: name, Description: description, Content: strings.TrimSpace(skillContent)}, nil
}

func readFrontMatter(lines []string) (map[string]string, int, error) {
	if strings.TrimSpace(lines[0]) != "---" {
		return nil, 0, fmt.Errorf("frontmatter is missing or misplaced")
	}
	kvPairs := make(map[string]string)
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return kvPairs, i, nil
		}

		parts := strings.SplitN(lines[i], ":", 2)
		if len(parts) < 2 {
			return nil, 0, fmt.Errorf("invalid frontmatter line: %q", lines[i])
		}
		key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		kvPairs[key] = value
	}

	return nil, 0, fmt.Errorf("frontmatter is unfinished")
}

func BuildSkillPrompt(skills []Skill) openai.ChatCompletionMessageParamUnion {
	var b bytes.Buffer

	b.WriteString("You have access to the following skills:\n\n")
	for _, skill := range skills {
		fmt.Fprintf(&b, "- %s: %s\n", skill.Name, skill.Description)
	}
	return openai.ChatCompletionMessageParamUnion{
		OfSystem: &openai.ChatCompletionSystemMessageParam{
			Content: openai.ChatCompletionSystemMessageParamContentUnion{
				OfString: openai.String(b.String()),
			},
		},
	}
}

func TransformPrompt(skills []Skill, prompt string) string {
	for _, skill := range skills {
		slashName := "/" + skill.Name
		if strings.HasPrefix(prompt, slashName) {
			argsJoined := strings.TrimSpace(strings.TrimPrefix(prompt, slashName))
			return substituteArgs(skill.Content, argsJoined)
		}
	}
	return prompt
}

func substituteArgs(content string, argsJoined string) string {
	args := strings.Fields(argsJoined)
	result := strings.Replace(content, "$ARGUMENTS", argsJoined, -1)
	for i, arg := range args {
		result = strings.Replace(result, fmt.Sprintf("$%d", i), arg, -1)
	}

	for i := len(args); i < 10; i++ {
		result = strings.Replace(result, fmt.Sprintf("$%d", i), "", -1)
	}
	return result
}
