// file: internal/services/llm_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sashabaranov/go-openai"
)

// LLMService 封装了与LLM的交互
type LLMService struct {
	client    *openai.Client
	modelName string
}

// NewLLMService 创建一个新的LLM服务实例
func NewLLMService(apiKey, baseURL, modelName string) (*LLMService, error) {
	if apiKey == "" || baseURL == "" {
		return nil, errors.New("LLM API key 或 base URL 未配置")
	}
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL

	return &LLMService{
		client:    openai.NewClientWithConfig(config),
		modelName: modelName,
	}, nil
}

// GetCompletion 调用LLM并实现指数退避重试，以提高稳定性
func (s *LLMService) GetCompletion(prompt string) (string, error) {
	var content string

	// 定义需要重试的操作
	operation := func() error {
		req := openai.ChatCompletionRequest{
			Model: s.modelName,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			MaxTokens:   4096,
			Temperature: 0.7,
		}

		resp, err := s.client.CreateChatCompletion(context.Background(), req)
		if err != nil {
			log.Printf("调用LLM API时出错，将进行重试: %v", err)
			return err // 返回错误以触发 backoff 库的重试
		}

		if len(resp.Choices) > 0 && resp.Choices[0].Message.Content != "" {
			content = resp.Choices[0].Message.Content
			return nil // 成功获取内容，停止重试
		}

		return errors.New("LLM响应为空或不包含有效内容")
	}

	// 配置指数退避重试策略
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 30 * time.Second // 设置最长重试时间，例如30秒

	err := backoff.Retry(operation, bo)

	if err != nil {
		return "", fmt.Errorf("经过多次重试后，LLM调用仍然失败: %w", err)
	}

	return content, nil
}
