package fastgpt

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"testing"
)

func TestFastGPTService_FastGPTChat(t *testing.T) {
	// 初始化配置
	viper.SetConfigName("config.dev")
	viper.AddConfigPath("../../../")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("读取配置文件失败: %v", err)
	}

	// 从配置文件获取真实的 API Key 和 Base URL
	apiKey := viper.GetString("fastgpt.api_key")
	baseURL := viper.GetString("fastgpt.base_url")
	// 打印调试信息
	t.Logf("使用的 FastGPT API Key: %s", apiKey)
	t.Logf("使用的 FastGPT Base URL: %s", baseURL)

	if apiKey == "" {
		t.Skip("跳过测试：fastgpt.api_key 未配置")
	}
	if baseURL == "" {
		t.Skip("跳过测试：fastgpt.base_url 未配置")
	}

	// Create service with real config
	service := NewFastGPTService(apiKey, baseURL)
	u7, _ := uuid.NewV7()
	// Test request
	req := Request{
		ChatID:    "mundotest_" + u7.String(),
		Stream:    false,
		Detail:    false,
		Variables: map[string]interface{}{},
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
		},
	}

	// Print conversation content
	t.Logf("=== 普通聊天测试 ===")
	t.Logf("用户: %s", req.Messages[0].Content)

	// Execute test
	responseBytes, err := service.FastGPTChat(req)
	if err != nil {
		t.Fatalf("FastGPTChat failed: %v", err)
	}

	// Parse response
	var response Response
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Print assistant response
	if len(response.Choices) > 0 {
		t.Logf("助手: %s", response.Choices[0].Message.Content)
	}

	// Verify response structure
	if response.Id == "" {
		t.Errorf("Response ID should not be empty")
	}
	if len(response.Choices) == 0 {
		t.Errorf("Expected at least 1 choice, got %d", len(response.Choices))
	}
	if len(response.Choices) > 0 && response.Choices[0].Message.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got %s", response.Choices[0].Message.Role)
	}
	if len(response.Choices) > 0 && response.Choices[0].Message.Content == "" {
		t.Errorf("Expected non-empty content from assistant")
	}
}

func TestFastGPTService_FastGPTStreamChat(t *testing.T) {
	// 初始化配置
	viper.SetConfigName("config.dev")
	viper.AddConfigPath("../../../")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("读取配置文件失败: %v", err)
	}

	// 从配置文件获取真实的 API Key 和 Base URL
	apiKey := viper.GetString("fastgpt.api_key")
	baseURL := viper.GetString("fastgpt.base_url")

	if apiKey == "" {
		t.Skip("跳过测试：fastgpt.api_key 未配置")
	}
	if baseURL == "" {
		t.Skip("跳过测试：fastgpt.base_url 未配置")
	}

	// Create service with real config
	service := NewFastGPTService(apiKey, baseURL)
	u7, _ := uuid.NewV7()
	// Test request
	req := Request{
		ChatID: "mundotest_" + u7.String(),
		Stream: false, // Will be set to true by the method
		Detail: false,
		Variables: map[string]interface{}{
			"query": "Tell me a story",
		},
		Messages: []Message{
			{
				Role:    "user",
				Content: "Tell me a short story",
			},
		},
	}

	// Print conversation content
	t.Logf("=== 流式聊天测试 ===")
	t.Logf("用户: %s", req.Messages[0].Content)
	t.Logf("助手: ")

	// Collect streaming chunks
	var chunks []string
	var fullResponse string
	callback := func(data []byte) {
		chunks = append(chunks, string(data))

		// Parse streaming chunk to extract content
		var streamChunk map[string]interface{}
		if err := json.Unmarshal(data, &streamChunk); err == nil {
			if choices, ok := streamChunk["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, ok := delta["content"].(string); ok {
							fullResponse += content
							t.Logf("%s", content)
						}
					}
				}
			}
		}
	}

	// Execute test
	err := service.FastGPTStreamChat(req, callback)
	if err != nil {
		t.Fatalf("FastGPTStreamChat failed: %v", err)
	}

	t.Logf("\n完整回复: %s", fullResponse)

	// Verify we received some chunks and content
	if len(chunks) == 0 {
		t.Errorf("Expected to receive streaming chunks, got 0")
	}
	if fullResponse == "" {
		t.Errorf("Expected to receive content from streaming response, got empty string")
	}

	t.Logf("收到 %d 个流式数据块", len(chunks))
}
