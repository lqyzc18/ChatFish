# ChatFish

基于 Go + [cloudwego/eino](https://github.com/cloudwego/eino) 框架的智能 AI 对话 CLI 工具，支持流式输出（token-by-token）和多轮上下文对话。

## 功能特性

- **流式输出** — 实时逐 token 打印 AI 响应，无需等待完整生成
- **多轮对话** — REPL 模式下维护完整对话历史，支持上下文理解
- **优雅退出** — 支持 `/quit` / `/exit` 命令和 Ctrl+C 信号拦截，安全关闭连接
- **配置灵活** — 支持命令行参数、环境变量、配置文件三种密钥配置方式
- **依赖注入** — 基于接口编程，低耦合，易于扩展

## 技术栈

| 层次 | 技术选型 |
|------|---------|
| CLI 框架 | [spf13/cobra](https://github.com/spf13/cobra) |
| AI 框架 | [cloudwego/eino](https://github.com/cloudwego/eino) |
| 模型接入 | [cloudwego/eino-ext](https://github.com/cloudwego/eino-ext)（OpenAI 兼容接口） |
| 模型 | MiniMax M2.7 |
| 配置解析 | gopkg.in/yaml.v3 |

## 项目结构

```
ChatFish/
├── main.go                    # 程序入口
├── config/
│   └── config.yaml            # API 密钥配置文件（不提交到 Git）
├── internal/
│   ├── chat/
│   │   └── service.go         # AI 对话服务（流式 + 多轮历史）
│   └── config/
│       └── config.go          # YAML 配置加载（支持 exe 同级目录）
├── cmd/
│   ├── root.go                # Cobra 根命令（--version）
│   └── chat.go                # chat 子命令（密钥优先级链）
└── go.mod
```

## 快速开始

### 1. 配置密钥

在可执行文件同级目录创建 `config/config.yaml`：

```yaml
api_key: "your-minimax-api-key"
```

### 2. 编译运行

```bash
go build -o chatfish.exe .
./chatfish.exe chat "你好，用一句话介绍自己"
```

### 3. 交互式对话

```bash
./chatfish.exe chat -i
# 或直接
./chatfish.exe chat
```

退出命令：`/quit` `/exit` 或 `Ctrl+C`。

## 密钥配置优先级

```
命令行 -k > 环境变量 MINIMAX_API_KEY > config/config.yaml
```

优先级从高到低，优先使用高优先级的配置源。

## 常用命令

```bash
# 查看版本
chatfish.exe --version

# 单次提问
chatfish.exe chat "你好"

# 交互式对话
chatfish.exe chat -i

# 手动指定密钥
chatfish.exe chat -k "your-api-key" "你好"
```

## 开发说明

```bash
# 安装依赖
go mod tidy

# 编译
go build -o chatfish.exe .

# 直接运行（需确保 config/config.yaml 存在）
go run .
```

## 注意事项

- `config/config.yaml` 包含敏感信息，已加入 `.gitignore`，不要提交到版本库
- 流式输出依赖终端 UTF-8 编码，Windows PowerShell 默认可能显示乱码，建议使用 Windows Terminal 或 VS Code 终端
