package main

import (
	"fmt"

	"agent/config"
	"agent/supervisor"

	"github.com/rs/zerolog/log"
)

func main() {
	// ========== 1. 加载配置 ==========
	// 从 config/config.yaml 读取 base_url, token, model_name 等
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("加载配置失败")
	}
	fmt.Println("✓ 配置加载成功")

	// ========== 2. 创建模型 ==========
	// ChatModel 负责与 LLM API 通信
	model := config.NewChatModel(cfg)
	fmt.Println("✓ 模型创建成功")

	// ========== 3. 创建 Agent ==========
	// 每个 Agent 专门处理一类任务
	hotelAssistant := supervisor.NewHotelAssistant(model)   // 酒店助手
	flightAssistant := supervisor.NewFlightAssistant(model) // 机票助手
	fmt.Println("✓ Agent 创建成功")

	// ========== 4. 创建 Supervisor 工作流 ==========
	// Supervisor 负责协调多个 Agent
	workflow := supervisor.CreateSupervisor(
		[]*supervisor.Agent{hotelAssistant, flightAssistant},
		model,
		"您是团队主管，负责管理酒店预订助手和机票预订助手。"+
			"如需预定酒店，请交由 hotel_assistant 处理。"+
			"如需预定机票，请交由 flight_assistant 处理。"+
			"**注意**，你每次最多只能调用一个助理！",
	)
	fmt.Println("✓ Supervisor 工作流创建成功")

	// ========== 5. 编译工作流 ==========
	sv := workflow.Compile()
	fmt.Println("✓ 工作流编译成功")

	// ========== 6. 执行 ==========
	userQuery := "请帮我预定一个北京到上海的机票，然后预定一个当地的酒店。"
	fmt.Printf("\n用户: %s\n\n", userQuery)
	fmt.Println("========== 开始执行 ==========")

	result, err := sv.Invoke(userQuery)
	if err != nil {
		log.Error().Err(err).Msg("执行失败")
		return
	}

	fmt.Println("========== 执行完成 ==========")
	fmt.Printf("\n结果: %s\n", result)
}
