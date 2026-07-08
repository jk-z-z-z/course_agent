package retrieval

import (
	"errors"
	"os"
	"strings"

	agenttypes "course_agent_backend/internal/agent/types"
)

func ReadMaterialContent(material agenttypes.MaterialDocument) (string, error) {
	if !IsTextLike(material) {
		return "", nil
	}
	contentBytes, err := os.ReadFile(material.StoragePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	return normalizeContent(string(contentBytes)), nil
}

func ReadMaterialExcerpt(material agenttypes.MaterialDocument, limit int) (string, error) {
	content, err := ReadMaterialContent(material)
	if err != nil || content == "" {
		return content, err
	}
	return truncateText(content, limit), nil
}

func IsTextLike(material agenttypes.MaterialDocument) bool {
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
