package utils

const (
	// API 路径
	// 注意：阿里云 DashScope 的 base_url 已包含 /v1
	// 所以这里只需要 /chat/completions
	ChatCompletionsPath = "/chat/completions"

	// 如果使用腾讯 Venus 或 OpenAI，base_url 不含 /v1，则用：
	// ChatCompletionsPath = "/v1/chat/completions"

	// 角色常量
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)
