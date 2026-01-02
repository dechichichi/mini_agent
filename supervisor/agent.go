package supervisor

import (
	"agent/config"
	"agent/tools"
)

// Agent 代表一个专门处理特定任务的助手
// 每个 Agent 有自己的名称、提示词、模型和工具集
type Agent struct {
	Name   string            // Agent 名称，如 "hotel_assistant"
	Prompt string            // 系统提示词，定义 Agent 的职责
	Model  *config.ChatModel // LLM 模型，用于生成回复
	Tools  []*tools.Tool     // Agent 可用的工具列表
}

// Run 执行 Agent 任务
// 接收用户任务描述，返回执行结果
//
// 执行流程：
// 1. 构建消息（system + user）
// 2. 调用 LLM
// 3. 如果 LLM 返回 tool_call → 执行工具 → 继续循环
// 4. 如果没有 tool_call → 返回结果
func (a *Agent) Run(task string) (string, error) {
	// 构建初始消息
	messages := []config.Message{
		{Role: "system", Content: a.Prompt}, // 告诉 LLM 它的角色
		{Role: "user", Content: task},       // 用户的任务
	}

	// 把 Tool 转成 OpenAI API 格式
	openaiTools := tools.ToolsToOpenAI(a.Tools)

	// ReAct 循环：推理 → 行动 → 观察 → 推理...
	for {
		// 调用 LLM
		content, toolCalls, err := a.Model.Invoke(messages, openaiTools)
		if err != nil {
			return "", err
		}

		// 没有工具调用 = 任务完成，返回 LLM 的回复
		if len(toolCalls) == 0 {
			return content, nil
		}

		// 有工具调用，先把 assistant 的消息加入历史
		messages = append(messages, config.Message{Role: "assistant", Content: content})

		// 执行每个工具调用
		for _, tc := range toolCalls {
			// 根据名称找到对应的工具
			tool := tools.FindTool(a.Tools, tc.Function.Name)
			if tool == nil {
				continue
			}

			// 执行工具，获取结果
			result, _ := tool.Execute(tc.Function.Arguments)

			// 把工具结果加入消息历史，让 LLM 知道执行结果
			messages = append(messages, config.Message{Role: "tool", Content: result})
		}
		// 继续循环，让 LLM 根据工具结果决定下一步
	}
}

// ==================== Agent 工厂函数 ====================

// NewHotelAssistant 创建酒店预定助手
func NewHotelAssistant(model *config.ChatModel) *Agent {
	return &Agent{
		Name:   "hotel_assistant",
		Prompt: "你能帮助用户预定酒店。注意，你只需要回答用户有关酒店预定的问题，不需要对其他问题做任何回应或追问。",
		Model:  model,
		Tools:  []*tools.Tool{tools.BookHotelTool},
	}
}

// NewFlightAssistant 创建机票预定助手
func NewFlightAssistant(model *config.ChatModel) *Agent {
	return &Agent{
		Name:   "flight_assistant",
		Prompt: "你能帮助用户预定机票。注意，你只需要回答用户有关机票预定的问题，不需要对其他问题做任何回应或追问。",
		Model:  model,
		Tools:  []*tools.Tool{tools.BookFlightTool},
	}
}
