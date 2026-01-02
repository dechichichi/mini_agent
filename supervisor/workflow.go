package supervisor

import (
	"encoding/json"
	"fmt"

	"agent/config"
)

// SupervisorWorkflow 未编译的工作流
// 对应 Python 的 create_supervisor() 返回值
type SupervisorWorkflow struct {
	agents []*Agent          // 管理的 Agent 列表
	model  *config.ChatModel // LLM 模型
	prompt string            // Supervisor 的系统提示词
}

// Supervisor 编译后可执行的工作流
// 对应 Python 的 workflow.compile() 返回值
type Supervisor struct {
	Agents []*Agent
	Model  *config.ChatModel
	Prompt string
}

// ==================== 创建与编译 ====================

// CreateSupervisor 创建 Supervisor 工作流
// 对应 Python: create_supervisor([agents], model, prompt)
func CreateSupervisor(agents []*Agent, model *config.ChatModel, prompt string) *SupervisorWorkflow {
	return &SupervisorWorkflow{
		agents: agents,
		model:  model,
		prompt: prompt,
	}
}

// Compile 编译工作流，返回可执行的 Supervisor
// 对应 Python: workflow.compile()
func (w *SupervisorWorkflow) Compile() *Supervisor {
	return &Supervisor{
		Agents: w.agents,
		Model:  w.model,
		Prompt: w.prompt,
	}
}

// ==================== Supervisor 内部方法 ====================

// agentToTool 将 Agent 包装成 Tool 格式
// 这样 Supervisor 就可以通过 function calling "调用" Agent
//
// 原理：对 LLM 来说，Agent 就是一个 tool
// 当 LLM 决定委派任务时，会返回 tool_call
func (s *Supervisor) agentToTool(agent *Agent) map[string]interface{} {
	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        agent.Name, // Agent 名称作为 tool 名称
			"description": fmt.Sprintf("委派任务给 %s: %s", agent.Name, agent.Prompt),
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "要委派给该助手的任务描述",
					},
				},
				"required": []string{"task"},
			},
		},
	}
}

// buildTools 把所有 Agent 转成 Tools 格式
func (s *Supervisor) buildTools() []map[string]interface{} {
	result := make([]map[string]interface{}, len(s.Agents))
	for i, agent := range s.Agents {
		result[i] = s.agentToTool(agent)
	}
	return result
}

// findAgent 根据名称查找 Agent
func (s *Supervisor) findAgent(name string) *Agent {
	for _, a := range s.Agents {
		if a.Name == name {
			return a
		}
	}
	return nil
}

// ==================== Supervisor 执行方法 ====================

// Invoke 执行 Supervisor 工作流
// 对应 Python: supervisor.invoke({"messages": [...]})
//
// 执行流程：
// 1. Supervisor 收到用户请求
// 2. 调用 LLM，LLM 决定委派给哪个 Agent
// 3. 执行对应 Agent
// 4. 把结果返回给 LLM，LLM 决定是否继续委派
// 5. 直到 LLM 不再返回 tool_call，任务完成
func (s *Supervisor) Invoke(userMessage string) (string, error) {
	// 初始化消息
	messages := []config.Message{
		{Role: "system", Content: s.Prompt},  // Supervisor 的角色定义
		{Role: "user", Content: userMessage}, // 用户请求
	}

	// 把 Agent 列表转成 Tools 格式
	agentTools := s.buildTools()

	// 主循环
	for {
		// 调用 LLM，让它决定下一步
		content, toolCalls, err := s.Model.Invoke(messages, agentTools)
		if err != nil {
			return "", err
		}

		// 没有 tool_call = 任务全部完成
		if len(toolCalls) == 0 {
			return content, nil
		}

		// 添加 assistant 消息到历史
		messages = append(messages, config.Message{Role: "assistant", Content: content})

		// 处理每个委派（通常只有一个，因为 prompt 里说了每次只调用一个）
		for _, tc := range toolCalls {
			// 找到被委派的 Agent
			agent := s.findAgent(tc.Function.Name)
			if agent == nil {
				continue
			}

			// 解析任务参数
			var args struct {
				Task string `json:"task"`
			}
			json.Unmarshal([]byte(tc.Function.Arguments), &args)

			// 执行 Agent，获取结果
			result, err := agent.Run(args.Task)
			if err != nil {
				result = "执行失败: " + err.Error()
			}

			// 把结果加入消息历史
			messages = append(messages, config.Message{Role: "tool", Content: result})
		}
		// 继续循环，让 Supervisor 决定下一步
	}
}
