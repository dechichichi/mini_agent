# Mini Agent

一个用 Go 实现的轻量级 Multi-Agent 框架，采用 Supervisor 模式协调多个专业 Agent 完成复杂任务。

## 特性

- **Supervisor 模式**: 主管 Agent 协调多个专业 Agent
- **ReAct 循环**: 推理 → 行动 → 观察 → 推理
- **Function Calling**: 基于 OpenAI 兼容 API 的工具调用
- **MCP 协议支持**: 可接入远程 MCP 服务获取工具
- **多模型支持**: 兼容 OpenAI、阿里云 DashScope、腾讯 Venus 等

## 项目结构

```
mini_agent/
├── main.go                 # 入口示例
├── config/
│   ├── config.go           # 配置加载 (viper)
│   ├── config.yaml         # 配置文件
│   ├── message.go          # Message, ToolCall 结构体
│   └── model.go            # ChatModel LLM 客户端
├── supervisor/
│   ├── agent.go            # Agent 定义和 ReAct 执行
│   └── workflow.go         # Supervisor 工作流
├── tools/
│   ├── base.go             # Tool 基础结构
│   └── book.go             # 示例工具 (BookHotel, BookFlight)
├── mcp/
│   └── client.go           # MCP 协议客户端
├── utils/
│   └── vars.go             # 常量定义
└── examples/
    └── mcp_example.go      # MCP 集成示例
```

## 快速开始

### 1. 配置

编辑 `config/config.yaml`:

```yaml
# 阿里云 DashScope
base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"
token: "sk-xxx"
model_name: "qwen-flash"

# 或腾讯 Venus
# base_url: "http://v2.open.venus.oa.com/llmproxy/v1"
# token: "your_token"
# model_name: "gpt-4o"

temperature: 0
```

### 2. 运行

```bash
go run main.go
```

### 3. 输出示例

```
✓ 配置加载成功
✓ 模型创建成功
✓ Agent 创建成功
✓ Supervisor 工作流创建成功
✓ 工作流编译成功

用户: 请帮我预定一个北京到上海的机票，然后预定一个当地的酒店。

========== 开始执行 ==========
[Supervisor] 委派给 flight_assistant
[flight_assistant] 调用工具 book_flight
[Supervisor] 委派给 hotel_assistant
[hotel_assistant] 调用工具 book_hotel
========== 执行完成 ==========

结果: 已为您完成预定...
```

## 核心概念

### Agent

专门处理特定任务的助手，包含：
- `Name`: 名称标识
- `Prompt`: 系统提示词
- `Model`: LLM 模型
- `Tools`: 可用工具列表

```go
agent := &supervisor.Agent{
    Name:   "hotel_assistant",
    Prompt: "你是酒店预定助手...",
    Model:  model,
    Tools:  []*tools.Tool{tools.BookHotelTool},
}
```

### Supervisor

协调多个 Agent 的主管，核心思想是**把 Agent 包装成 Tool**：

```go
workflow := supervisor.CreateSupervisor(
    []*supervisor.Agent{hotelAgent, flightAgent},
    model,
    "您是团队主管，负责管理酒店和机票助手...",
)
sv := workflow.Compile()
result, _ := sv.Invoke("帮我订机票和酒店")
```

### Tool

可被 Agent 调用的工具：

```go
var BookHotelTool = &Tool{
    Name:        "book_hotel",
    Description: "预定酒店",
    Parameters: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "city": map[string]interface{}{
                "type":        "string",
                "description": "城市名称",
            },
        },
        "required": []string{"city"},
    },
    Func: func(args map[string]interface{}) string {
        city := args["city"].(string)
        return fmt.Sprintf("已预定 %s 的酒店", city)
    },
}
```

## MCP 集成

支持通过 MCP 协议接入远程工具服务：

```go
client := mcp.NewMultiServerMCPClient(map[string]mcp.MCPServerConfig{
    "map_search": {
        URL:       "https://mcp.map.qq.com/sse?key=YOUR_KEY",
        Transport: "sse",
    },
})
tools, _ := client.GetTools()
```

完整示例见 `examples/mcp_example.go`

## 依赖

- [resty](https://github.com/go-resty/resty) - HTTP 客户端
- [viper](https://github.com/spf13/viper) - 配置管理
- [zerolog](https://github.com/rs/zerolog) - 日志

## License

MIT
