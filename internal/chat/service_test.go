package chat

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type fakeChatModel struct {
	chunks []string
}

func (m fakeChatModel) Generate(context.Context, []*schema.Message, ...model.Option) (*schema.Message, error) {
	return nil, errors.New("not implemented")
}

func (m fakeChatModel) Stream(context.Context, []*schema.Message, ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	chunks := make([]*schema.Message, len(m.chunks))
	for i, chunk := range m.chunks {
		chunks[i] = &schema.Message{Role: schema.Assistant, Content: chunk}
	}
	return schema.StreamReaderFromArray(chunks), nil
}

type blockingChatModel struct {
	started chan struct{}
}

func (m *blockingChatModel) Generate(context.Context, []*schema.Message, ...model.Option) (*schema.Message, error) {
	return nil, errors.New("not implemented")
}

func (m *blockingChatModel) Stream(ctx context.Context, _ []*schema.Message, _ ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	close(m.started)
	<-ctx.Done()
	return nil, ctx.Err()
}

func TestChatAppendsStreamedResponse(t *testing.T) {
	var started, finished int
	var chunks []string
	service := &Service{
		chatModel: fakeChatModel{chunks: []string{"hello", " world"}},
		guiOutput: GUIStreamCallbacks{
			OnStart:  func(uint64) { started++ },
			OnChunk:  func(_ uint64, text string) { chunks = append(chunks, text) },
			OnFinish: func(uint64) { finished++ },
		},
	}

	if err := service.Chat(" question "); err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	history := service.GetHistory()
	if len(history) != 2 {
		t.Fatalf("history length = %d, want 2", len(history))
	}
	if got, want := history[0].Content, "question"; got != want {
		t.Fatalf("user message = %q, want %q", got, want)
	}
	if got, want := history[1].Content, "hello world"; got != want {
		t.Fatalf("assistant message = %q, want %q", got, want)
	}
	if started != 1 || finished != 1 {
		t.Fatalf("callbacks started=%d finished=%d, want 1 each", started, finished)
	}
	if got, want := len(chunks), 2; got != want {
		t.Fatalf("chunk count = %d, want %d", got, want)
	}
}

func TestChatRejectsConcurrentRequestAndCancellationSkipsHistory(t *testing.T) {
	model := &blockingChatModel{started: make(chan struct{})}
	service := &Service{chatModel: model}
	done := make(chan error, 1)
	go func() { done <- service.Chat("first") }()

	select {
	case <-model.started:
	case <-time.After(time.Second):
		t.Fatal("first request did not start")
	}

	if err := service.Chat("second"); !errors.Is(err, ErrRequestInProgress) {
		t.Fatalf("concurrent Chat() error = %v, want %v", err, ErrRequestInProgress)
	}

	service.Cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("cancelled Chat() error = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancelled request did not finish")
	}
	if got := service.GetHistory(); len(got) != 0 {
		t.Fatalf("history after cancellation = %#v, want empty", got)
	}
}

func TestChatHistoryIsBoundedAndCopied(t *testing.T) {
	service := &Service{chatModel: fakeChatModel{chunks: []string{"answer"}}}

	for i := 0; i < 11; i++ {
		if err := service.Chat("question"); err != nil {
			t.Fatalf("Chat() iteration %d error = %v", i, err)
		}
	}
	history := service.GetHistory()
	if len(history) != maxHistoryMessages {
		t.Fatalf("history length = %d, want %d", len(history), maxHistoryMessages)
	}

	history[0].Content = "mutated"
	if got := service.GetHistory()[0].Content; got == "mutated" {
		t.Fatal("GetHistory() returned a mutable history entry")
	}
}

func TestChatRejectsBlankQuestion(t *testing.T) {
	service := &Service{chatModel: fakeChatModel{}}

	if err := service.Chat(" \n\t "); !errors.Is(err, ErrEmptyQuestion) {
		t.Fatalf("Chat() error = %v, want %v", err, ErrEmptyQuestion)
	}
}
