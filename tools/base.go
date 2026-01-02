package tools

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type ToolFunc func(args map[string]interface{}) string

type Tool struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
	Func        ToolFunc
}

func (t *Tool) ToOpenAITool() map[string]interface{} {
	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        t.Name,
			"description": t.Description,
			"parameters":  t.Parameters,
		},
	}
}

func (t *Tool) Execute(argsJSON string) (string, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		log.Error().Err(err).Str("args", argsJSON).Msg("")
		return "", errors.New("invalid argument: " + err.Error())
	}
	return t.Func(args), nil
}

func ToolsToOpenAI(tools []*Tool) []map[string]interface{} {
	result := make([]map[string]interface{}, len(tools))
	for i, t := range tools {
		result[i] = t.ToOpenAITool()
	}
	return result
}

func FindTool(tools []*Tool, name string) *Tool {
	for _, t := range tools {
		if t.Name == name {
			return t
		}
	}
	return nil
}
