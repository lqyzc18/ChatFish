package cmd

import (
	"github.com/spf13/cobra"
)

const version = "0.2.0"

var rootCmd = &cobra.Command{
	Use:     "chatfish",
	Short:   "ChatFish - 基于 eino 框架的智能 AI 对话 CLI 工具",
	Version: version,
	Long: `ChatFish 是一个使用 Go 语言和 cloudwego/eino 框架构建的智能 AI 对话工具。
支持流式输出、交互式对话模式，开箱即用。`,
}

func Execute() error {
	return rootCmd.Execute()
}
