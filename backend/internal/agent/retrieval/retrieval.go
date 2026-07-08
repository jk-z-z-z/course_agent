package retrieval

import (
	"errors"
	"os"
	"regexp"
	"sort"
	"strings"

	agenttypes "course_agent_backend/internal/agent/types"
)

var tokenPattern = regexp.MustCompile(`[\p{Han}]{1,12}|[a-zA-Z0-9_]{2,}`)

type scoredSource struct {
	agenttypes.Source
	score int
}

func RetrieveSources(question string, materials []agenttypes.MaterialDocument) []agenttypes.Source {
	keywords := extractKeywords(question)
	scored := make([]scoredSource, 0)
	for _, material := range materials {
		snippet, score, err := extractSnippet(material, question, keywords)
		if err != nil || score <= 0 || strings.TrimSpace(snippet) == "" {
			continue
		}
		scored = append(scored, scoredSource{
			Source: agenttypes.Source{MaterialNodeID: material.MaterialNodeID, FileName: material.FileName, Snippet: snippet},
			score:  score,
		})
	}
	sort.SliceStable(scored, func(i, j int) bool { return scored[i].score > scored[j].score })
	limit := 4
	if len(scored) < limit {
		limit = len(scored)
	}
	result := make([]agenttypes.Source, 0, limit)
	for i := 0; i < limit; i++ {
		result = append(result, scored[i].Source)
	}
	return result
}

func extractKeywords(question string) []string {
	matches := tokenPattern.FindAllString(strings.ToLower(question), -1)
	keywords := make([]string, 0, len(matches))
	seen := make(map[string]struct{})
	for _, match := range matches {
		trimmed := strings.TrimSpace(match)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		keywords = append(keywords, trimmed)
	}
	return keywords
}

func extractSnippet(material agenttypes.MaterialDocument, question string, keywords []string) (string, int, error) {
	if !isTextLike(material) {
		return "", 0, nil
	}
	contentBytes, err := os.ReadFile(material.StoragePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", 0, nil
		}
		return "", 0, err
	}
	content := normalizeContent(string(contentBytes))
	if content == "" {
		return "", 0, nil
	}
	lines := strings.Split(content, "\n")
	bestLine := ""
	bestScore := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		score := scoreLine(strings.ToLower(line), strings.ToLower(question), keywords)
		if score > bestScore {
			bestScore = score
			bestLine = line
		}
	}
	if bestScore == 0 {
		bestLine = truncateText(content, 220)
		bestScore = 1
	}
	return truncateText(bestLine, 240), bestScore, nil
}

func scoreLine(line, question string, keywords []string) int {
	score := 0
	if question != "" && strings.Contains(line, question) {
		score += 8
	}
	for _, keyword := range keywords {
		if strings.Contains(line, keyword) {
			score += 3
		}
	}
	return score
}

func isTextLike(material agenttypes.MaterialDocument) bool {
	mimeType := strings.ToLower(material.MimeType)
	if strings.HasPrefix(mimeType, "text/") || strings.Contains(mimeType, "json") || strings.Contains(mimeType, "xml") {
		return true
	}
	name := strings.ToLower(material.FileName)
	switch {
	case strings.HasSuffix(name, ".txt"), strings.HasSuffix(name, ".md"), strings.HasSuffix(name, ".markdown"), strings.HasSuffix(name, ".csv"), strings.HasSuffix(name, ".json"), strings.HasSuffix(name, ".yaml"), strings.HasSuffix(name, ".yml"), strings.HasSuffix(name, ".log"), strings.HasSuffix(name, ".html"), strings.HasSuffix(name, ".htm"):
		return true
	default:
		return false
	}
}

func normalizeContent(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	return strings.TrimSpace(content)
}

func truncateText(text string, limit int) string {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= limit {
		return string(runes)
	}
	return string(runes[:limit]) + "..."
}
