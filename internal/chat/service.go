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

// GUIStreamCallbacks 定义流式对话过程中的 GUI 回调接口。
type GUIStreamCallbacks struct {
	OnStart  func()
	OnChunk  func(text string)
	OnFinish func()
	OnError  func(err error)
}

// Service 封装 AI 对话服务，管理对话历史和流式响应。
type Service struct {
	chatModel model.BaseChatModel
	guiOutput GUIStreamCallbacks
	history   []*schema.Message
	mu        sync.RWMutex
	cancel    context.CancelFunc
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
func (s *Service) streamResponse(ctx context.Context, msgs []*schema.Message) (string, error) {
	stream, err := s.chatModel.Stream(ctx, msgs)
	if err != nil {
		return "", fmt.Errorf("stream: %w", err)
	}
	defer stream.Close()

	if s.guiOutput.OnStart != nil {
		s.guiOutput.OnStart()
	}

	var sb strings.Builder
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			// 流中断：通知 GUI 并返回已累积的内容和错误
			if s.guiOutput.OnFinish != nil {
				s.guiOutput.OnFinish()
			}
			return sb.String(), fmt.Errorf("recv: %w", err)
		}
		if len(chunk.Content) > 0 {
			if s.guiOutput.OnChunk != nil {
				s.guiOutput.OnChunk(chunk.Content)
			}
			sb.WriteString(chunk.Content)
		}
	}

	if s.guiOutput.OnFinish != nil {
		s.guiOutput.OnFinish()
	}
	return sb.String(), nil
}

// Chat 发送用户消息并以流式方式接收 AI 回复，回复会追加到对话历史中。
func (s *Service) Chat(question string) error {
	ctx, cancel := context.WithTimeout(context.Background(), streamTimeout)
	defer cancel()
	defer func() {
		s.mu.Lock()
		s.cancel = nil
		s.mu.Unlock()
	}()

	// 在同一把锁内完成 cancel 赋值和 history 拷贝，消除竞态窗口
	s.mu.Lock()
	s.cancel = cancel
	msgs := make([]*schema.Message, len(s.history), len(s.history)+1)
	copy(msgs, s.history)
	s.mu.Unlock()

	msgs = append(msgs, &schema.Message{
		Role:    schema.User,
		Content: question,
	})

	response, err := s.streamResponse(ctx, msgs)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		// 通知 GUI 流式错误
		if s.guiOutput.OnError != nil {
			s.guiOutput.OnError(err)
		}
		return err
	}

	s.mu.Lock()
	s.history = append(s.history,
		&schema.Message{Role: schema.User, Content: question},
		&schema.Message{Role: schema.Assistant, Content: response},
	)

	if len(s.history) > maxHistoryMessages {
		s.history = s.history[len(s.history)-maxHistoryMessages:]
	}
	s.mu.Unlock()

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
	history := make([]*schema.Message, len(s.history))
	copy(history, s.history)
	return history
}

// Cancel 取消正在进行的流式请求。
func (s *Service) Cancel() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
	}
}
