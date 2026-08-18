# devin-2api

> [English](README.md) | **中文**

`devin-2api` 是一个把 [Devin](https://app.devin.ai/) / Windsurf 后端封装成 OpenAI 兼容协议的多协议网关。它对外同时提供 OpenAI Responses API、OpenAI Chat Completions API 和 Anthropic Messages API 三类接口，并内置一个 Web 管理面板。

## 特性

- **多协议兼容**：`POST /v1/responses`、`POST /v1/chat/completions`、`POST /v1/messages`，以及对应的 `/v1/models` 模型发现接口。
- **流式与一次性响应**：均支持 JSON 和 SSE 流式输出。
- **OpenAI Responses WebSocket**：支持 `GET /v1/responses` 升级为 WebSocket，供 Codex 等客户端使用。
- **工具调用（function calling）**：Responses / Chat / Anthropic 均支持函数工具。
- **推理/thinking 内容**：自动把上游思考过程映射到各协议对应字段。
- **图片输入**：支持 base64 data URL 形式的图片；Responses 兼容 `input_image` 字符串与 `image_url` 对象、`image_url` part 等写法。
- **管理面板**：访问 `/panel` 查看模型列表、渠道/供应商、账户用量与价格筛选。
- **流控与连接模型**：`server.max_concurrency` 限制并发；`devin.force_http1` 默认强制 HTTP/1.1，避免 HTTP/2 单连接多 stream 被上游串行处理。
- **代理**：支持 `http://`、`https://`、`socks5://`、`socks5h://` 代理，或留空走系统 `HTTP_PROXY` / `HTTPS_PROXY`。
- **请求级调试日志**：开启 `debug.enabled` 后在 `logs/` 目录下输出每请求的完整链路。

## 快速开始

### 1. 获取 Devin token

`devin-2api` 需要 Devin / Windsurf 的 session token，格式为 `devin-session-token$...`。

macOS 示例：

```bash
sqlite3 ~/Library/"Application Support"/Devin/User/globalStorage/state.vscdb \
  "SELECT json_extract(value, '$.apiKey') FROM ItemTable WHERE key='windsurfAuthStatus';"
```
Windows 示例（PowerShell）：

```powershell
$state = "$env:APPDATA\devin\User\globalStorage\state.vscdb"
# 如果使用 Devin - Next 客户端，请尝试：
# $state = "$env:APPDATA\Devin - Next\User\globalStorage\state.vscdb"
sqlite3 $state "SELECT json_extract(value, '\$.apiKey') FROM ItemTable WHERE key='windsurfAuthStatus';"
```

Token 保存在 Devin / Devin - Next 客户端本地状态 `%APPDATA%` 下。如果上述路径不存在，可在 `%APPDATA%` 下搜索所有 `state.vscdb`：

```powershell
Get-ChildItem -Path $env:APPDATA -Recurse -Filter state.vscdb -ErrorAction SilentlyContinue | Select-Object -ExpandProperty FullName
```
Windsurf / Devin 扩展同样把 token 存在 `%APPDATA%\Code\User\globalStorage\state.vscdb`。如果该路径不存在，可在 `%APPDATA%` 或 Windsurf 安装目录下搜索 `state.vscdb`。

Linux 用户可以在对应应用目录的 `state.vscdb` 中查找同样的键，或直接从 Devin / Windsurf 客户端的本地存储中提取。

### 2. 配置

```bash
cp config.example.yaml config.yaml
```

编辑 `config.yaml`，填入真实 token，其余字段示例值通常无需改动：

```yaml
server:
  listen: ":8080"           # HTTP 监听地址
  max_concurrency: 1024     # /v1/* 并发上限，默认 1024

devin:
  base_url: "https://server.codeium.com"
  token: "devin-session-token$..."   # 必填
  model: "glm-5-2"                   # 默认模型 UID，可按需修改
  # proxy: "socks5://127.0.0.1:1080" # 可选代理
  force_http1: true                  # 默认强制 HTTP/1.1

dashboard:
  password: ""              # 管理面板密码；留空直接访问 /panel

auth:
  api_key: ""               # /v1/* 访问密钥；留空不校验
```

### 3. 启动

本地：

```bash
go run ./cmd/devin-2api -config config.yaml
```

Docker：

```bash
docker run --rm -p 8080:8080 \
  -v "$PWD/config.yaml:/app/config.yaml" \
  leokun123/devin-2api --config /app/config.yaml
```

### 4. 验证

```bash
curl http://localhost:8080/healthz
# {"status":"ok"}
```

## 支持的 API 端点

| 端点 | 方法 | 说明 |
| --- | --- | --- |
| `/healthz` | GET | 健康检查 |
| `/v1/models` | GET | OpenAI 兼容模型列表 |
| `/v1/models/{model}` | GET | 单个模型信息 |
| `/v1/responses` | POST | OpenAI Responses API（JSON / typed SSE） |
| `/v1/responses` | GET | 升级 WebSocket（子协议 `responses_websockets=2026-02-06`） |
| `/v1/chat/completions` | POST | OpenAI Chat Completions API（JSON / SSE） |
| `/v1/messages` | POST | Anthropic Messages API（JSON / SSE） |
| `/panel` | GET | 管理面板 |
| `/panel/login` | POST | 面板登录 |
| `/panel/api/*` | GET | 面板数据 API |

`/v1/*` 端点默认无需鉴权。设置 `auth.api_key` 后，客户端需通过 `Authorization: Bearer <key>` 或 `X-Api-Key: <key>` 传递。

## 用法示例

### OpenAI Responses

非流式：

```bash
curl http://localhost:8080/v1/responses \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-api-key" \
  -d '{
    "model": "glm-5-2",
    "input": "你好"
  }'
```

流式：

```bash
curl -N http://localhost:8080/v1/responses \
  -H "Content-Type: application/json" \
  -d '{
    "model": "glm-5-2",
    "input": "你好",
    "stream": true
  }'
```

多模态（base64 data URL）：

```bash
curl http://localhost:8080/v1/responses \
  -H "Content-Type: application/json" \
  -d '{
    "model": "glm-5-2",
    "input": [
      {"type": "message", "role": "user", "content": [
        {"type": "input_text", "text": "描述这张图"},
        {"type": "input_image", "image_url": {"url": "data:image/png;base64,..."}}
      ]}
    ]
  }'
```

### OpenAI Chat Completions

非流式：

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "glm-5-2",
    "messages": [{"role": "user", "content": "你好"}]
  }'
```

流式并在最后返回 usage：

```bash
curl -N http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "glm-5-2",
    "messages": [{"role": "user", "content": "你好"}],
    "stream": true,
    "stream_options": {"include_usage": true}
  }'
```

多模态：

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "glm-5-2",
    "messages": [{"role": "user", "content": [
      {"type": "text", "text": "描述这张图"},
      {"type": "image_url", "image_url": {"url": "data:image/png;base64,..."}}
    ]}]
  }'
```

### Anthropic Messages

```bash
curl -N http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-api-key" \
  -d '{
    "model": "glm-5-2",
    "max_tokens": 1024,
    "messages": [{"role": "user", "content": [
      {"type": "image", "source": {"type": "base64", "media_type": "image/png", "data": "..."}},
      {"type": "text", "text": "描述这张图"}
    ]}]
  }'
```

### 模型列表

```bash
curl http://localhost:8080/v1/models
```

## 配置说明

所有配置集中在 `config.yaml`，启动时加载一次，未知字段会被拒绝。

| 字段 | 说明 | 必填 |
| --- | --- | --- |
| `server.listen` | HTTP 监听地址 | 是 |
| `server.max_concurrency` | 同时处理的 `/v1/*` 请求数上限；0 使用默认值 1024 | 否 |
| `devin.base_url` | Devin Connect 上游地址 | 配置了 `devin.token` 后必填 |
| `devin.token` | Devin session token（`devin-session-token$...`） | 否；未设置时 `/v1/*` 返回 503 |
| `devin.model` | 默认 chat model UID（如 `glm-5-2`） | 配置了 `devin.token` 后必填 |
| `devin.proxy` | 代理地址；支持 `http/https/socks5/socks5h` | 否 |
| `devin.force_http1` | 强制 HTTP/1.1，避免 HTTP/2 单连接多 stream 并发瓶颈 | 否；默认 `true` |
| `debug.enabled` | 在配置文件同目录 `logs/` 下写请求级调试日志 | 否 |
| `dashboard.password` | 管理面板密码；留空无需登录 | 否 |
| `auth.api_key` | `/v1/*` 接口访问密钥；留空不鉴权 | 否 |

完整示例：

```yaml
server:
  listen: ":8080"
  max_concurrency: 1024

devin:
  base_url: "https://server.codeium.com"
  token: "devin-session-token$..."
  model: "glm-5-2"
  force_http1: true

debug:
  enabled: false

dashboard:
  password: ""

auth:
  api_key: ""
```

## 管理面板

启动后访问 `http://localhost:8080/panel`：

- `dashboard.password` 为空时直接进入；
- 设置密码后，首次访问需登录，会话 cookie 24 小时有效；
- 展示当前可用模型、所属渠道/供应商、图片能力；
- 展示账户用量、价格区间与筛选。

## 注意事项

- `devin.token` 为空时，`/v1/*` 返回 `503 provider_configuration`；
- 鉴权头同时支持 `Authorization: Bearer <key>` 和 `X-Api-Key: <key>`；
- token 等敏感字段在调试日志中会被脱敏为 `<redacted>`；
- 图片只支持 base64 data URL，不支持外部 URL、`file_id` 或本地文件路径；
- 每轮请求仍需携带完整消息历史；`previous_response_id` 字段会被接收，但当前不会启用服务端状态保存；
- `config.yaml` 已加入 `.gitignore`；即使示例文件被 git 追踪，真实 token 也不要提交；
- 请求体上限为 32 MiB，满足多图 base64 场景。

## 文档

- **架构、API 字段子集、proto 提取等技术细节**：[Contributing guide](CONTRIBUTING.md)
- **开源协议**：[MIT](LICENSE)