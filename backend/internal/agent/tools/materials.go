package tools

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"

	"course_agent_backend/internal/agent/retrieval"
	agenttypes "course_agent_backend/internal/agent/types"
)

type materialSourceCollector struct {
	mu      sync.Mutex
	sources []agenttypes.Source
	seen    map[uint64]struct{}
}

func newMaterialSourceCollector() *materialSourceCollector {
	return &materialSourceCollector{seen: make(map[uint64]struct{})}
}

func (c *materialSourceCollector) Add(source agenttypes.Source) {
	if source.MaterialNodeID == 0 || strings.TrimSpace(source.Snippet) == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.seen[source.MaterialNodeID]; ok {
		return
	}
	c.seen[source.MaterialNodeID] = struct{}{}
	c.sources = append(c.sources, source)
}

func (c *materialSourceCollector) List() []agenttypes.Source {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]agenttypes.Source, len(c.sources))
	copy(result, c.sources)
	return result
}

type materialSummary struct {
	MaterialNodeID uint64 `json:"materialNodeId"`
	FileName       string `json:"fileName"`
	MimeType       string `json:"mimeType"`
}

type listMaterialsInput struct {
	Limit int `json:"limit"`
}

type searchMaterialsInput struct {
	Query      string `json:"query"`
	MaxResults int    `json:"maxResults"`
}

type searchMaterialsOutput struct {
	MaterialNodeID uint64 `json:"materialNodeId"`
	FileName       string `json:"fileName"`
	Snippet        string `json:"snippet"`
}

type readMaterialInput struct {
	MaterialNodeID uint64 `json:"materialNodeId"`
	MaxChars       int    `json:"maxChars"`
}

type readMaterialOutput struct {
	MaterialNodeID uint64 `json:"materialNodeId"`
	FileName       string `json:"fileName"`
	Content        string `json:"content"`
}

func BuildCourseTools(materials []agenttypes.MaterialDocument) ([]tool.BaseTool, func() []agenttypes.Source, error) {
	collector := newMaterialSourceCollector()
	indexed := make(map[uint64]agenttypes.MaterialDocument, len(materials))
	for _, material := range materials {
		indexed[material.MaterialNodeID] = material
	}

	listTool, err := toolutils.InferTool("list_course_materials", "列出当前课程中可供检索的资料列表。适合在回答前先了解有哪些文件。", func(ctx context.Context, input listMaterialsInput) ([]materialSummary, error) {
		_ = ctx
		limit := input.Limit
		if limit <= 0 || limit > 50 {
			limit = 20
		}
		result := make([]materialSummary, 0, min(limit, len(materials)))
		for i, material := range materials {
			if i >= limit {
				break
			}
			result = append(result, materialSummary{MaterialNodeID: material.MaterialNodeID, FileName: material.FileName, MimeType: material.MimeType})
		}
		return result, nil
	})
	if err != nil {
		return nil, nil, err
	}

	searchTool, err := toolutils.InferTool("search_course_materials", "根据问题或关键词检索当前课程资料，返回最相关的文件片段。用于先定位相关资料。", func(ctx context.Context, input searchMaterialsInput) ([]searchMaterialsOutput, error) {
		_ = ctx
		query := strings.TrimSpace(input.Query)
		if query == "" {
			return []searchMaterialsOutput{}, nil
		}
		sources := retrieval.RetrieveSources(query, materials)
		limit := input.MaxResults
		if limit <= 0 || limit > 6 {
			limit = 4
		}
		if len(sources) > limit {
			sources = sources[:limit]
		}
		result := make([]searchMaterialsOutput, 0, len(sources))
		for _, source := range sources {
			collector.Add(source)
			result = append(result, searchMaterialsOutput{MaterialNodeID: source.MaterialNodeID, FileName: source.FileName, Snippet: source.Snippet})
		}
		return result, nil
	})
	if err != nil {
		return nil, nil, err
	}

	readTool, err := toolutils.InferTool("read_course_material", "读取某个课程资料文件的正文片段。适合在已经知道文件 ID 后继续查看细节。", func(ctx context.Context, input readMaterialInput) (*readMaterialOutput, error) {
		_ = ctx
		material, ok := indexed[input.MaterialNodeID]
		if !ok {
			return nil, fmt.Errorf("material %d not found", input.MaterialNodeID)
		}
		limit := input.MaxChars
		if limit <= 0 || limit > 4000 {
			limit = 1600
		}
		excerpt, err := retrieval.ReadMaterialExcerpt(material, limit)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(excerpt) == "" {
			return &readMaterialOutput{MaterialNodeID: material.MaterialNodeID, FileName: material.FileName, Content: "该资料当前无法提取可读文本，可能不是文本文件。"}, nil
		}
		collector.Add(agenttypes.Source{MaterialNodeID: material.MaterialNodeID, FileName: material.FileName, Snippet: excerpt})
		return &readMaterialOutput{MaterialNodeID: material.MaterialNodeID, FileName: material.FileName, Content: excerpt}, nil
	})
	if err != nil {
		return nil, nil, err
	}

	return []tool.BaseTool{listTool, searchTool, readTool}, collector.List, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
