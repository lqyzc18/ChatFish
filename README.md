# ChatFish

基于 Go + [cloudwego/eino](https://github.com/cloudwego/eino) 框架的智能 AI
对话桌面应用，使用 [fyne](https://github.com/fyne-io/fyne) 构建现代化 GUI 界面，支持流式输出和多轮上下文对话。

## 功能特性

- **流式输出** — 实时逐 token 显示 AI 响应，无需等待完整生成
- **多轮对话** — 维护完整对话历史，支持上下文理解
- **Markdown 渲染** — AI 响应支持 Markdown 格式渲染，代码块、列表、链接等清晰展示
- **现代化 GUI** — 使用 fyne 框架构建的跨平台桌面应用
- **消息气泡** — 用户消息和 AI 响应以气泡样式展示
- **配置灵活** — 支持 GUI 界面配置 API Key、Base URL 和模型
- **简洁设计** — 浅色主题，清爽的界面风格

## 技术栈

| 层次     | 技术选型                                                                     |
|--------|--------------------------------------------------------------------------|
| GUI 框架 | [fyne](https://github.com/fyne-io/fyne)                                  |
| AI 框架  | [cloudwego/eino](https://github.com/cloudwego/eino)                      |
| 模型接入   | [cloudwego/eino-ext](https://github.com/cloudwego/eino-ext)（OpenAI 兼容接口） |
| 模型     | MiniMax M3                                                               |
| 配置解析   | gopkg.in/yaml.v3                                                         |

## 项目结构

```
ChatFish/
├── main.go                    # GUI 应用入口
├── internal/
│   ├── chat/
│   │   └── service.go         # AI 对话服务（流式 + 多轮历史）
│   ├── config/
│   │   └── config.go          # YAML 配置加载和保存
│   └── gui/
│       ├── app.go             # 主应用窗口
│       ├── chat_view.go       # 聊天界面组件
│       ├── markdown.go        # Markdown 渲染器（goldmark → Fyne RichText）
│       ├── settings_view.go   # 设置界面组件
│       ├── theme.go           # 主题管理
│       └── custom_theme.go    # 自定义浅色主题
└── go.mod
```

## 快速开始

### 1. 配置密钥

推荐直接在应用 GUI 的设置界面中配置 API Key。

配置文件会自动保存在系统用户配置目录下：
- **Windows**: `%APPDATA%\ChatFish\config.yaml`
- **macOS/Linux**: `~/.config/ChatFish/config.yaml`

如果需要手动创建，可以按照上述路径创建文件并填入：

```yaml
api_key: "your-minimax-api-key"
```

### 2. 编译运行

```bash
go build -o chatfish.exe .
./chatfish.exe
```

## 使用说明

### 主界面

- **发送消息**: 在输入框中输入消息，按回车或点击发送按钮
- **清除对话**: 点击工具栏的清除按钮（垃圾桶图标）
- **设置**: 点击工具栏的设置按钮（齿轮图标）配置 API Key
- **帮助**: 点击工具栏的帮助按钮（问号图标）查看帮助

### 消息样式

- 用户消息显示为蓝色气泡，靠右对齐
- AI 响应显示为灰色气泡，靠左对齐
- 流式输出时实时更新 AI 响应内容

### 快捷键

- **Enter**: 换行
- **Ctrl+Enter**: 发送消息

## 开发说明

```bash
# 安装依赖
go mod tidy

# 编译
go build -o chatfish.exe .

# 直接运行
go run .
```

## 注意事项

- 配置文件统一存放在系统用户目录中，安全且不会被意外提交到版本库
- 首次编译 fyne 应用可能需要较长时间（约 10 分钟），后续编译会很快
- 应用支持跨平台运行（Windows、macOS、Linux）
- 当前仅支持浅色主题