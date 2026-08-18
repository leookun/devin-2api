# devin-2api

> **English** | [中文](README.zh-CN.md)

`devin-2api` is a multi-protocol gateway that wraps the [Devin](https://app.devin.ai/) / Windsurf backend behind an OpenAI-compatible API. It exposes OpenAI Responses API, OpenAI Chat Completions API, and Anthropic Messages API endpoints, and includes a built-in web dashboard.

## Features

- **Multi-protocol support**: `POST /v1/responses`, `POST /v1/chat/completions`, `POST /v1/messages`, plus `/v1/models` model discovery.
- **Streaming and non-streaming** responses in JSON and SSE.
- **OpenAI Responses WebSocket**: `GET /v1/responses` upgrades to WebSocket for clients such as Codex.
- **Function calling**: supported across Responses, Chat, and Anthropic protocols.
- **Reasoning / thinking content**: upstream thinking is mapped to the corresponding fields of each protocol.
- **Image input**: supports base64 data URL images; Responses also accepts `input_image` as a string or `image_url` object, as well as `image_url` parts.
- **Dashboard**: visit `/panel` for model list, providers, account usage, and price filters.
- **Concurrency and connection model**: `server.max_concurrency` limits concurrent `/v1/*` requests; `devin.force_http1` defaults to `true` to avoid HTTP/2 multi-stream serialization by the upstream.
- **Proxy support**: `http://`, `https://`, `socks5://`, `socks5h://` proxy, or leave empty to use system `HTTP_PROXY` / `HTTPS_PROXY`.
- **Request-level debug logs**: enable with `debug.enabled`; logs each request lifecycle under `logs/`.

## Quick start

### 1. Get a Devin token

`devin-2api` needs a Devin / Windsurf session token in the `devin-session-token$...` format.

On macOS:

```bash
sqlite3 ~/Library/"Application Support"/Devin/User/globalStorage/state.vscdb \
  "SELECT json_extract(value, '$.apiKey') FROM ItemTable WHERE key='windsurfAuthStatus';"
```

On Windows (PowerShell):

```powershell
$state = "$env:APPDATA\devin\User\globalStorage\state.vscdb"
# If using the Devin - Next client, try:
# $state = "$env:APPDATA\Devin - Next\User\globalStorage\state.vscdb"
sqlite3 $state "SELECT json_extract(value, '\$.apiKey') FROM ItemTable WHERE key='windsurfAuthStatus';"
```

The token is stored in the Devin / Devin - Next client local state under `%APPDATA%`. If the path above does not exist, search for `state.vscdb` under `%APPDATA%`:

```powershell
Get-ChildItem -Path $env:APPDATA -Recurse -Filter state.vscdb -ErrorAction SilentlyContinue | Select-Object -ExpandProperty FullName
```

Linux users can look for the same key in the equivalent `state.vscdb` path, or extract the token from the Devin / Windsurf client local storage.

### 2. Configure

```bash
cp config.example.yaml config.yaml
```

Edit `config.yaml` and set the real token; the other fields can usually stay at their example values:

```yaml
server:
  listen: ":8080"          # HTTP listen address
  max_concurrency: 1024    # /v1/* concurrency limit, default 1024

devin:
  base_url: "https://server.codeium.com"
  token: "devin-session-token$..."  # required
  model: "glm-5-2"                  # default model UID
  # proxy: "socks5://127.0.0.1:1080" # optional proxy
  force_http1: true                 # force HTTP/1.1 by default

dashboard:
  password: ""             # dashboard password; empty = open /panel

auth:
  api_key: ""              # /v1/* API key; empty = no auth
```

### 3. Run

Local:

```bash
go run ./cmd/devin-2api -config config.yaml
```

Docker:

```bash
docker run --rm -p 8080:8080 \
  -v "$PWD/config.yaml:/app/config.yaml" \
  leokun123/devin-2api --config /app/config.yaml
```

### 4. Verify

```bash
curl http://localhost:8080/healthz
# {"status":"ok"}
```

## Supported API endpoints

| Endpoint | Method | Description |
| --- | --- | --- |
| `/healthz` | GET | Health check |
| `/v1/models` | GET | OpenAI-compatible model list |
| `/v1/models/{model}` | GET | Single model info |
| `/v1/responses` | POST | OpenAI Responses API (JSON / typed SSE) |
| `/v1/responses` | GET | WebSocket upgrade (subprotocol `responses_websockets=2026-02-06`) |
| `/v1/chat/completions` | POST | OpenAI Chat Completions API (JSON / SSE) |
| `/v1/messages` | POST | Anthropic Messages API (JSON / SSE) |
| `/panel` | GET | Dashboard |
| `/panel/login` | POST | Dashboard login |
| `/panel/api/*` | GET | Dashboard data APIs |

`/v1/*` endpoints are open by default. Set `auth.api_key` and send it via `Authorization: Bearer <key>` or `X-Api-Key: <key>`.

## Usage examples

### OpenAI Responses

Non-streaming:

```bash
curl http://localhost:8080/v1/responses \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-api-key" \
  -d '{
    "model": "glm-5-2",
    "input": "Hello"
  }'
```

Streaming:

```bash
curl -N http://localhost:8080/v1/responses \
  -H "Content-Type: application/json" \
  -d '{
    "model": "glm-5-2",
    "input": "Hello",
    "stream": true
  }'
```

Multimodal (base64 data URL):

```bash
curl http://localhost:8080/v1/responses \
  -H "Content-Type: application/json" \
  -d '{
    "model": "glm-5-2",
    "input": [
      {"type": "message", "role": "user", "content": [
        {"type": "input_text", "text": "Describe this image"},
        {"type": "input_image", "image_url": {"url": "data:image/png;base64,..."}}
      ]}
    ]
  }'
```

### OpenAI Chat Completions

Non-streaming:

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "glm-5-2",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

Streaming with usage at the end:

```bash
curl -N http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "glm-5-2",
    "messages": [{"role": "user", "content": "Hello"}],
    "stream": true,
    "stream_options": {"include_usage": true}
  }'
```

Multimodal:

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "glm-5-2",
    "messages": [{"role": "user", "content": [
      {"type": "text", "text": "Describe this image"},
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
      {"type": "text", "text": "Describe this image"}
    ]}]
  }'
```

### Model list

```bash
curl http://localhost:8080/v1/models
```

## Configuration

All settings live in `config.yaml`, loaded once at startup; unknown fields are rejected.

| Field | Description | Required |
| --- | --- | --- |
| `server.listen` | HTTP listen address | Yes |
| `server.max_concurrency` | Max concurrent `/v1/*` requests; `0` uses default 1024 | No |
| `devin.base_url` | Devin Connect upstream URL | Required once `devin.token` is set |
| `devin.token` | Devin session token (`devin-session-token$...`) | No; `/v1/*` returns 503 if unset |
| `devin.model` | Default chat model UID (e.g. `glm-5-2`) | Required once `devin.token` is set |
| `devin.proxy` | Proxy URL; supports `http/https/socks5/socks5h` | No |
| `devin.force_http1` | Force HTTP/1.1 to avoid HTTP/2 multi-stream serialization | No; defaults to `true` |
| `debug.enabled` | Write per-request debug logs under `logs/` | No |
| `dashboard.password` | Dashboard password; empty = no login | No |
| `auth.api_key` | API key for `/v1/*`; empty disables auth | No |

Full example:

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

## Dashboard

Visit `http://localhost:8080/panel` after starting:

- If `dashboard.password` is empty, the panel is open;
- With a password set, login is required on first visit; session cookie is valid for 24 hours;
- Shows available models, their provider/channel, and image support;
- Shows account usage and price filtering.

## Notes

- If `devin.token` is empty, `/v1/*` returns `503 provider_configuration`;
- Auth headers are accepted as both `Authorization: Bearer <key>` and `X-Api-Key: <key>`;
- Sensitive fields such as tokens are redacted as `<redacted>` in debug logs;
- Images must be base64 data URLs; external URLs, `file_id`, and local file paths are not supported;
- Each request must still carry the full message history; `previous_response_id` is accepted but currently does not enable server-side state retention;
- `config.yaml` is in `.gitignore`; even though the example is tracked, never commit a real token;
- Request body size is capped at 32 MiB to accommodate multi-image base64 payloads.

## Documentation

- **Architecture, supported API fields, proto extraction, and other technical details**: [Contributing guide](CONTRIBUTING.md)
- **License**: [MIT](LICENSE)