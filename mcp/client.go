package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"agent/tools"

	"github.com/rs/zerolog/log"
)

// MCPClient MCP 协议客户端
// MCP (Model Context Protocol) 是一种标准化的 AI 工具调用协议
// 通过 SSE (Server-Sent Events) 方式获取可用工具列表
type MCPClient struct {
	Name      string            // 客户端名称，如 "map_search"
	URL       string            // SSE 服务地址
	Headers   map[string]string // 请求头（如 Authorization）
	Transport string            // 传输方式，目前只支持 "sse"
	client    *http.Client
}

// MCPServerConfig MCP 服务配置
type MCPServerConfig struct {
	URL       string            // SSE 地址
	Headers   map[string]string // 额外的请求头
	Transport string            // 传输方式: "sse"
}

// MultiServerMCPClient 多服务 MCP 客户端
// 可以同时连接多个 MCP 服务
type MultiServerMCPClient struct {
	Servers map[string]*MCPClient // 服务名 -> 客户端
}

// NewMultiServerMCPClient 创建多服务 MCP 客户端
// 对应 Python: MultiServerMCPClient({"map_search": {...}, "other_search": {...}})
func NewMultiServerMCPClient(configs map[string]MCPServerConfig) *MultiServerMCPClient {
	client := &MultiServerMCPClient{
		Servers: make(map[string]*MCPClient),
	}

	for name, cfg := range configs {
		client.Servers[name] = &MCPClient{
			Name:      name,
			URL:       cfg.URL,
			Headers:   cfg.Headers,
			Transport: cfg.Transport,
			client:    &http.Client{},
		}
	}

	return client
}

// GetTools 获取所有服务的工具列表
// 对应 Python: await map_client.get_tools()
func (m *MultiServerMCPClient) GetTools() ([]*tools.Tool, error) {
	var allTools []*tools.Tool

	for name, server := range m.Servers {
		serverTools, err := server.GetTools()
		if err != nil {
			log.Error().Err(err).Str("server", name).Msg("获取工具列表失败")
			continue
		}
		allTools = append(allTools, serverTools...)
	}

	return allTools, nil
}

// GetTools 从单个 MCP 服务获取工具列表
func (c *MCPClient) GetTools() ([]*tools.Tool, error) {
	// 构建请求
	req, err := http.NewRequest("GET", c.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Accept", "text/event-stream")
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("服务返回错误: %d", resp.StatusCode)
	}

	// 解析 SSE 响应，获取工具列表
	return c.parseSSETools(resp)
}

// parseSSETools 解析 SSE 响应中的工具定义
func (c *MCPClient) parseSSETools(resp *http.Response) ([]*tools.Tool, error) {
	var mcpTools []*tools.Tool

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// SSE 格式: data: {...}
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// 尝试解析工具列表
			var toolsResponse struct {
				Tools []struct {
					Name        string                 `json:"name"`
					Description string                 `json:"description"`
					InputSchema map[string]interface{} `json:"inputSchema"`
				} `json:"tools"`
			}

			if err := json.Unmarshal([]byte(data), &toolsResponse); err != nil {
				continue // 不是工具列表消息，跳过
			}

			// 转换为我们的 Tool 格式
			for _, t := range toolsResponse.Tools {
				mcpTools = append(mcpTools, &tools.Tool{
					Name:        t.Name,
					Description: t.Description,
					Parameters:  t.InputSchema,
					Func:        c.createToolExecutor(t.Name), // 创建执行器
				})
			}
		}
	}

	return mcpTools, nil
}

// createToolExecutor 为 MCP 工具创建执行器
// 当 Agent 调用工具时，会通过 MCP 协议远程执行
func (c *MCPClient) createToolExecutor(toolName string) tools.ToolFunc {
	return func(args map[string]interface{}) string {
		result, err := c.ExecuteTool(toolName, args)
		if err != nil {
			return fmt.Sprintf("工具执行失败: %v", err)
		}
		return result
	}
}

// ExecuteTool 通过 MCP 协议执行工具
func (c *MCPClient) ExecuteTool(toolName string, args map[string]interface{}) (string, error) {
	// 构建执行请求
	reqBody := map[string]interface{}{
		"method": "tools/call",
		"params": map[string]interface{}{
			"name":      toolName,
			"arguments": args,
		},
	}

	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", c.URL, strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 解析响应
	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Content) > 0 {
		return result.Content[0].Text, nil
	}

	return "", nil
}
