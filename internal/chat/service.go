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
	defaultModel       = "MiniMax-M2.7"
	maxHistoryMessages = 20
	streamTimeout      = 60 * time.Second
)

type GUIStreamCallbacks struct {
	OnStart  func()
	OnChunk  func(text string)
	OnFinish func()
}

type Service struct {
	chatModel model.BaseChatModel
	guiOutput GUIStreamCallbacks
	history   []*schema.Message
	mu        sync.RWMutex
}

type Option func(*Service)

func WithGUIOutput(callbacks GUIStreamCallbacks) Option {
	return func(s *Service) { s.guiOutput = callbacks }
}

func NewService(apiKey, baseURL, model string, opts ...Option) (*Service, error) {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if model == "" {
		model = defaultModel
	}

	ctx := context.Background()
	cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  apiKey,
		Model:   model,
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
		select {
		case <-ctx.Done():
			if s.guiOutput.OnFinish != nil {
				s.guiOutput.OnFinish()
			}
			return sb.String(), ctx.Err()
		default:
		}

		chunk, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
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

func (s *Service) Chat(question string) error {
	ctx, cancel := context.WithTimeout(context.Background(), streamTimeout)
	defer cancel()

	s.mu.RLock()
	msgs := make([]*schema.Message, len(s.history), len(s.history)+1)
	copy(msgs, s.history)
	s.mu.RUnlock()

	msgs = append(msgs, &schema.Message{
		Role:    schema.User,
		Content: question,
	})

	response, err := s.streamResponse(ctx, msgs)
	if err != nil {
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

func (s *Service) ClearHistory() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.history = make([]*schema.Message, 0)
}

func (s *Service) GetHistory() []*schema.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	history := make([]*schema.Message, len(s.history))
	copy(history, s.history)
	return history
}
