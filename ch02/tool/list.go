package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
)

type ListTool struct{}

func NewListTool() *ListTool {
	return &ListTool{}
}

type ListToolParam struct {
	Path          string `json:"path"`
	IncludeHidden bool   `json:"include_hidden"`
	SortBy        string `json:"sort_by"`
	SortOrder     string `json:"sort_order"`
	MaxEntries    int    `json:"max_entries"`
}

type listEntry struct {
	Name    string
	IsDir   bool
	Size    int64
	ModTime int64
}

func (t *ListTool) ToolName() AgentTool {
	return AgentToolList
}

func (t *ListTool) Info() openai.ChatCompletionToolUnionParam {
	return openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
		Name:        string(AgentToolList),
		Description: openai.String("list files and directories in a given directory"),
		Parameters: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "directory path to inspect; use . for the current directory",
				},
				"include_hidden": map[string]any{
					"type":        "boolean",
					"description": "whether to include hidden files such as .env",
				},
				"sort_by": map[string]any{
					"type":        "string",
					"enum":        []string{"name", "size", "modified_time"},
					"description": "how to sort the entries; defaults to name",
				},
				"sort_order": map[string]any{
					"type":        "string",
					"enum":        []string{"asc", "desc"},
					"description": "whether to sort ascending or descending; defaults to asc except modified_time which defaults to desc",
				},
				"max_entries": map[string]any{
					"type":        "integer",
					"description": "maximum number of entries to return; defaults to 200",
				},
			},
		},
	})
}

func (t *ListTool) Execute(ctx context.Context, argumentsInJSON string) (string, error) {
	p := ListToolParam{Path: ".", SortBy: "name", MaxEntries: 200}
	if strings.TrimSpace(argumentsInJSON) != "" {
		if err := json.Unmarshal([]byte(argumentsInJSON), &p); err != nil {
			return "", err
		}
	}
	if strings.TrimSpace(p.Path) == "" {
		p.Path = "."
	}
	if strings.TrimSpace(p.SortBy) == "" {
		p.SortBy = "name"
	}
	if strings.TrimSpace(p.SortOrder) == "" {
		if p.SortBy == "modified_time" {
			p.SortOrder = "desc"
		} else {
			p.SortOrder = "asc"
		}
	}
	switch p.SortBy {
	case "name", "size", "modified_time":
	default:
		return "", fmt.Errorf("unsupported sort_by: %s", p.SortBy)
	}
	switch p.SortOrder {
	case "asc", "desc":
	default:
		return "", fmt.Errorf("unsupported sort_order: %s", p.SortOrder)
	}
	if p.MaxEntries <= 0 {
		p.MaxEntries = 200
	}

	entries, err := os.ReadDir(p.Path)
	if err != nil {
		return "", err
	}

	items := make([]listEntry, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if !p.IncludeHidden && strings.HasPrefix(name, ".") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return "", err
		}
		items = append(items, listEntry{
			Name:    name,
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Unix(),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		less := false
		switch p.SortBy {
		case "size":
			if items[i].Size == items[j].Size {
				less = items[i].Name < items[j].Name
				break
			}
			less = items[i].Size < items[j].Size
		case "modified_time":
			if items[i].ModTime == items[j].ModTime {
				less = items[i].Name < items[j].Name
				break
			}
			less = items[i].ModTime < items[j].ModTime
		default:
			less = items[i].Name < items[j].Name
		}
		if p.SortOrder == "desc" {
			return !less
		}
		return less
	})

	if len(items) > p.MaxEntries {
		items = items[:p.MaxEntries]
	}

	lines := make([]string, 0, len(items)+2)
	lines = append(lines, fmt.Sprintf("directory: %s", filepath.Clean(p.Path)))
	lines = append(lines, fmt.Sprintf("sort_by: %s, sort_order: %s, max_entries: %d", p.SortBy, p.SortOrder, p.MaxEntries))
	for _, entry := range items {
		kind := "file"
		if entry.IsDir {
			kind = "dir"
		}
		lines = append(lines, fmt.Sprintf("%s\t%s\tsize=%d\tmodified_time=%s", kind, entry.Name, entry.Size, time.Unix(entry.ModTime, 0).UTC().Format(time.RFC3339)))
	}

	return strings.Join(lines, "\n"), nil
}
