package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"chatfish/internal/chat"
	"chatfish/internal/config"

	"github.com/spf13/cobra"
)

var (
	apiKey      string
	interactive bool
)

var chatCmd = &cobra.Command{
	Use:   "chat [question]",
	Short: "向 AI 提问",
	Long:  `与 MiniMax AI 进行对话。使用 --interactive 或 -i 进入交互式对话模式。`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := apiKey
		if key == "" {
			key = os.Getenv("MINIMAX_API_KEY")
		}
		if key == "" {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			key = cfg.APIKey
		}

		svc, err := chat.NewService(context.Background(), key)
		if err != nil {
			return fmt.Errorf("init service: %w", err)
		}

		if interactive || len(args) == 0 {
			ctx, stop := signal.NotifyContext(context.Background(),
				syscall.SIGINT, syscall.SIGTERM)
			defer stop()
			return svc.REPL(ctx)
		}

		return svc.Chat(context.Background(), args[0])
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)

	chatCmd.Flags().StringVarP(&apiKey, "api-key", "k", "",
		"MiniMax API Key（优先级：命令行 > 环境变量 > config/config.yaml）")
	chatCmd.Flags().BoolVarP(&interactive, "interactive", "i", false,
		"进入交互式对话模式")
}
