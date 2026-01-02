package main

import (
	"fmt"

	"agent/config"
	"agent/mcp"
	"agent/supervisor"

	"github.com/rs/zerolog/log"
)

// MCP 服务配置
const (
	// 腾讯位置服务 Token
	// 获取地址: https://lbs.qq.com/dev/console/application/mine
	LBS_TOKEN = "your_lbs_token_here"

	// Venus Token
	// 获取地址: https://venus.woa.com/#/openapi/accountManage/personalAccount
	VENUS_TOKEN = "your_venus_token_here"

	// MCP 搜索服务 SSE 地址
	// 创建实例: https://ai.woa.com/#/mcp/mcp-market/detail/mcp_OKpYCrxz5B
	SEARCH_SSE_URL = "your_search_sse_url_here"
)

func main() {
	// ========== 1. 加载配置 ==========
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("加载配置失败")
	}

	// ========== 2. 创建模型 ==========
	model := config.NewChatModel(cfg)

	// ========== 3. 创建 MCP 客户端并获取工具 ==========

	// 3.1 腾讯地图 MCP 服务
	mapClient := mcp.NewMultiServerMCPClient(map[string]mcp.MCPServerConfig{
		"map_search": {
			URL:       fmt.Sprintf("https://mcp.map.qq.com/sse?key=%s", LBS_TOKEN),
			Transport: "sse",
		},
	})
	mapTools, err := mapClient.GetTools()
	if err != nil {
		log.Error().Err(err).Msg("获取地图工具失败")
	}
	fmt.Printf("✓ 获取到 %d 个地图工具\n", len(mapTools))

	// 3.2 综合搜索 MCP 服务
	searchClient := mcp.NewMultiServerMCPClient(map[string]mcp.MCPServerConfig{
		"other_search": {
			URL: SEARCH_SSE_URL,
			Headers: map[string]string{
				"Authorization": "Bearer " + VENUS_TOKEN,
			},
			Transport: "sse",
		},
	})
	searchTools, err := searchClient.GetTools()
	if err != nil {
		log.Error().Err(err).Msg("获取搜索工具失败")
	}
	fmt.Printf("✓ 获取到 %d 个搜索工具\n", len(searchTools))

	// ========== 4. 创建 Agent ==========

	// 地图助手 - 负责位置查询和路线规划
	mapAgent := &supervisor.Agent{
		Name:   "map_assistant",
		Prompt: "你是一个地图助手，负责位置查询和路线规划。",
		Model:  model,
		Tools:  mapTools,
	}

	// 搜索助手 - 负责各种信息搜索
	searchAgent := &supervisor.Agent{
		Name:   "search_assistant",
		Prompt: "你是一个能搜索各种信息的助手。注意，你不负责位置服务相关的信息收集，如果有类似需求路线规划需求，请不要做任何回应。",
		Model:  model,
		Tools:  searchTools,
	}

	fmt.Println("✓ Agent 创建成功")

	// ========== 5. 创建 Supervisor ==========
	workflow := supervisor.CreateSupervisor(
		[]*supervisor.Agent{mapAgent, searchAgent},
		model,
		"您是团队主管，负责管理信息搜索助手和路线规划助手。"+
			"如需搜索各种信息，请交由 search_assistant 处理。"+
			"如需规划路线，查询位置，请交由 map_assistant 处理。"+
			"**注意**，你每次只能调用一个助理agent！",
	)

	sv := workflow.Compile()
	fmt.Println("✓ Supervisor 创建成功")

	// ========== 6. 执行查询 ==========
	userQuery := "北京最出名的老北京火锅在哪里，从腾讯北京总部大楼怎么过去呢？"
	fmt.Printf("\n用户: %s\n\n", userQuery)
	fmt.Println("========== 开始执行 ==========")

	result, err := sv.Invoke(userQuery)
	if err != nil {
		log.Error().Err(err).Msg("执行失败")
		return
	}

	fmt.Println("========== 执行完成 ==========")
	fmt.Printf("\n结果:\n%s\n", result)
}
