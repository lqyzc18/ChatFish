package chat

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

const (
	miniMaxBaseURL = "https://api.minimaxi.com/v1"
	miniMaxModel   = "MiniMax-M2.7"
)

type Service struct {
	chatModel model.BaseChatModel
}

func NewService(ctx context.Context, apiKey string) (*Service, error) {
	cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  apiKey,
		Model:   miniMaxModel,
		BaseURL: miniMaxBaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("new chat model: %w", err)
	}

	return &Service{chatModel: cm}, nil
}

func (s *Service) Chat(ctx context.Context, question string) error {
	msgs := []*schema.Message{
		{Role: schema.User, Content: question},
	}

	stream, err := s.chatModel.Stream(ctx, msgs)
	if err != nil {
		return fmt.Errorf("stream: %w", err)
	}

	fmt.Print("AI: ")
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			stream.Close()
			return fmt.Errorf("recv: %w", err)
		}
		if len(chunk.Content) > 0 {
			fmt.Print(chunk.Content)
		}
	}
	stream.Close()
	fmt.Println()
	return nil
}

func (s *Service) REPL(ctx context.Context) error {
	reader := bufio.NewReader(os.Stdin)
	history := make([]*schema.Message, 0)

	fmt.Println("=== ChatFish AI 对话 (输入 /quit 或 /exit 退出, Ctrl+C 安全退出) ===")
	fmt.Println()

	for {
		fmt.Print("You: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("read input: %w", err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if input == "/quit" || input == "/exit" {
			fmt.Println("再见!")
			return nil
		}

		history = append(history, &schema.Message{
			Role:    schema.User,
			Content: input,
		})

		stream, err := s.chatModel.Stream(ctx, history)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Print("AI: ")
		var responseContent strings.Builder
		for {
			chunk, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("\nError: %v\n", err)
				break
			}
			if len(chunk.Content) > 0 {
				fmt.Print(chunk.Content)
				responseContent.WriteString(chunk.Content)
			}
		}
		stream.Close()

		history = append(history, &schema.Message{
			Role:    schema.Assistant,
			Content: responseContent.String(),
		})

		fmt.Println()
	}
}
