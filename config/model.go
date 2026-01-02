package config

import (
	"encoding/json"

	"agent/utils"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"net/http"
)

type ChatModel struct {
	model       string
	baseURL     string
	apiKey      string
	temperature float64
	client      *http.Client
}

func NewChatModel(cfg *Config) *ChatModel {
	return &ChatModel{
		model:       cfg.ModelName,
		baseURL:     cfg.BaseURL,
		apiKey:      cfg.Token,
		temperature: cfg.Temperature,
		client:      &http.Client{},
	}
}

func (m *ChatModel) Invoke(messages []Message, tools []map[string]interface{}) (string, []ToolCall, error) {
	reqBody := map[string]interface{}{
		"model":       m.model,       // 模型名称
		"messages":    messages,      // 对话消息（注意：是 "messages" 不是 "data"）
		"temperature": m.temperature, // 温度参数
	}
	if len(tools) > 0 {
		reqBody["tools"] = tools
	}

	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+m.apiKey).
		SetBody(reqBody).
		Post(m.baseURL + utils.ChatCompletionsPath)

	if err != nil {
		log.Error().Err(err).Msg("调用 ChatModel 失败")
		return "", nil, errors.New("Error calling ChatModel: " + err.Error())
	}

	if resp.StatusCode() != http.StatusOK {
		log.Error().Int("status", resp.StatusCode()).Str("body", string(resp.Body())).Msg("API 返回错误")
		return "", nil, errors.New("API error: " + string(resp.Body()))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content   string     `json:"content"`
				ToolCalls []ToolCall `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		log.Error().Err(err).Msg("解析响应失败")
		return "", nil, err
	}

	if len(result.Choices) == 0 {
		return "", nil, errors.New("no response from model")
	}
	// choices[0]: 默认只请求一个回复(n=1)，取第一个即可
	msg := result.Choices[0].Message
	return msg.Content, msg.ToolCalls, nil

}
