// Package chat 提供对话服务封装。
//
// 主要职责：
// 1) 维护多轮对话历史（有限长度，避免上下文无限增长）
// 2) 以流式方式接收模型输出，并通过 GUI 回调增量渲染
// 3) 支持取消：Cancel 或发起新请求后，旧请求的迟到结果不会写回历史
package chat

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

const (
	defaultBaseURL     = "https://api.minimaxi.com/v1"
	defaultModel       = "MiniMax-M3"
	maxHistoryMessages = 20
	streamTimeout      = 60 * time.Second
)

var (
	ErrEmptyQuestion     = errors.New("question cannot be empty")
	ErrRequestInProgress = errors.New("a chat request is already in progress")
)

// GUIStreamCallbacks 定义流式对话过程中的 GUI 回调接口。
type GUIStreamCallbacks struct {
	// requestID 用于区分并发/取消场景下的“过期回调”，GUI 侧会根据 requestID 决定是否渲染。
	OnStart  func(requestID uint64)
	OnChunk  func(requestID uint64, text string)
	OnFinish func(requestID uint64)
	OnError  func(requestID uint64, err error)
}

// Service 封装 AI 对话服务，管理对话历史和流式响应。
type Service struct {
	chatModel model.BaseChatModel
	guiOutput GUIStreamCallbacks
	history   []*schema.Message
	mu        sync.RWMutex
	cancel    context.CancelFunc
	// requestID 自增，用于识别“当前正在进行的请求”。
	// 当请求被 Cancel 或新的 Chat 开始后，旧请求的迟到结果会被丢弃。
	requestID uint64
}

// Option 是 Service 的功能选项函数。
type Option func(*Service)

// WithGUIOutput 设置流式对话的 GUI 回调。
func WithGUIOutput(callbacks GUIStreamCallbacks) Option {
	return func(s *Service) { s.guiOutput = callbacks }
}

// NewService 创建并返回一个新的对话服务实例。
func NewService(apiKey, baseURL, modelName string, opts ...Option) (*Service, error) {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if modelName == "" {
		modelName = defaultModel
	}

	ctx := context.Background()
	cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  apiKey,
		Model:   modelName,
		BaseURL: baseURL,
		Timeout: streamTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("new chat model: %w", err)
	}

	s := &Service{
		chatModel: cm,
		history:   make([]*schema.Message, 0),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s, nil
}

// streamResponse 执行流式请求并逐 chunk 回调 GUI。
func (s *Service) streamResponse(ctx context.Context, msgs []*schema.Message, requestID uint64) (string, error) {
	// 该方法负责“读流 + 回调增量文本 + 汇总成最终回复”。
	// 是否把最终回复写回 history，由 Chat() 决定（它会再次校验 requestID）。
	stream, err := s.chatModel.Stream(ctx, msgs)
	if err != nil {
		return "", fmt.Errorf("stream: %w", err)
	}
	defer stream.Close()

	if s.guiOutput.OnStart != nil {
		s.guiOutput.OnStart(requestID)
	}
	defer func() {
		if s.guiOutput.OnFinish != nil {
			s.guiOutput.OnFinish(requestID)
		}
	}()

	var sb strings.Builder
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return sb.String(), fmt.Errorf("recv: %w", err)
		}
		if len(chunk.Content) > 0 {
			if s.guiOutput.OnChunk != nil {
				s.guiOutput.OnChunk(requestID, chunk.Content)
			}
			sb.WriteString(chunk.Content)
		}
	}

	return sb.String(), nil
}

// Chat 发送用户消息并以流式方式接收 AI 回复，回复会追加到对话历史中。
func (s *Service) Chat(question string) error {
	// 生命周期：
	// - 通过 requestID 锁定“当前请求”
	// - 流式读取模型回复，同时把 chunk 交给 GUI（GUI 再用 requestID 做过期过滤）
	// - 当流结束后，只有 requestID 仍然匹配且 ctx 未取消时才把结果写入历史
	question = strings.TrimSpace(question)
	if question == "" {
		return ErrEmptyQuestion
	}

	ctx, cancel := context.WithTimeout(context.Background(), streamTimeout)

	s.mu.Lock()
	if s.cancel != nil {
		s.mu.Unlock()
		cancel()
		return ErrRequestInProgress
	}

	// 该请求成为“当前请求”，后续 Cancel/新请求会让 requestID 发生变化，从而跳过迟到历史写入。
	s.requestID++
	requestID := s.requestID
	s.cancel = cancel
	msgs := cloneMessages(s.history)
	s.mu.Unlock()
	defer func() {
		cancel()
		s.mu.Lock()
		if s.requestID == requestID {
			s.cancel = nil
		}
		s.mu.Unlock()
	}()

	msgs = append(msgs, &schema.Message{
		Role:    schema.User,
		Content: question,
	})

	response, err := s.streamResponse(ctx, msgs, requestID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		// 通知 GUI 流式错误
		if s.guiOutput.OnError != nil {
			s.guiOutput.OnError(requestID, err)
		}
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if ctx.Err() != nil || s.requestID != requestID {
		return nil
	}
	s.history = append(s.history,
		&schema.Message{Role: schema.User, Content: question},
		&schema.Message{Role: schema.Assistant, Content: response},
	)

	if len(s.history) > maxHistoryMessages {
		s.history = s.history[len(s.history)-maxHistoryMessages:]
	}

	return nil
}

// ClearHistory 清空对话历史。
func (s *Service) ClearHistory() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.history = make([]*schema.Message, 0)
}

// GetHistory 返回当前对话历史的副本。
func (s *Service) GetHistory() []*schema.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneMessages(s.history)
}

// Cancel 取消正在进行的流式请求。
func (s *Service) Cancel() uint64 {
	s.mu.RLock()
	cancel := s.cancel
	requestID := s.requestID
	s.mu.RUnlock()
	if cancel != nil {
		cancel()
	}
	return requestID
}

func cloneMessages(messages []*schema.Message) []*schema.Message {
	// 复制每条消息结构，避免外部并发读到内部切片被追加/截断后的引用别名问题。
	// 注意：*schema.Message 的字段里包含字符串等值类型，本实现按结构体浅拷贝即可满足本项目的并发需求。
	cloned := make([]*schema.Message, len(messages))
	for i, message := range messages {
		if message == nil {
			continue
		}
		copy := *message
		cloned[i] = &copy
	}
	return cloned
}
