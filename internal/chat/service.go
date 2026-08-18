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
