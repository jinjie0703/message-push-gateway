# web_ws（WebSocket 告警通知中心）

基于 **Go + Gin + Gorilla WebSocket** 的轻量级“告警通知中心”。

- 提供 **HTTP Webhook** 接口接收外部系统推送的告警/事件消息
- 通过 **WebSocket** 将消息实时下发给已认证的客户端
- 使用 **JWT** 进行 WebSocket 握手鉴权

> 运行日志会打印 WebSocket 入口与推送 Webhook 地址，便于快速联调。


## 功能特性

- **Webhook 推送入口**：`POST /api/push` 接收标准化 JSON 消息
- **WebSocket 实时下发**：`GET /ws?token=...` 或 `GET /api/ws/notifications?token=...`
- **JWT 鉴权**：WebSocket 连接必须携带 `token`（URL Query）
- **健康检查**：`GET /api/health`
- **系统统计**：`GET /api/stats`
- **联调接口**：`GET /api/test/push` 一键生成测试消息并推送
- **Docker / Compose**：提供容器化构建与启动


## 技术栈

- **Go**：见 `go.mod`（项目使用 Go 1.24.x）
- **HTTP 框架**：`github.com/gin-gonic/gin`
- **WebSocket**：`github.com/gorilla/websocket`
- **JWT**：`github.com/golang-jwt/jwt/v5`
- **配置**：YAML（`gopkg.in/yaml.v3`）+ 环境变量覆盖


## 快速开始

### 1) 本地运行（Go）

1. 准备配置文件（优先级见下文“配置说明”）：

- 开发时通常使用：`config/config.yaml`

2. 启动：

```bash
go run ./cmd/server
```

服务默认监听 `:8080`（或由配置/环境变量决定）。


### 2) 编译运行

```bash
go build -o bin/web_ws ./cmd/server
./bin/web_ws
```


### 3) 使用 Makefile

```bash
make build
```


### 4) Docker 运行

```bash
docker build -t web_ws:latest .
docker run --rm -p 8080:8080 -e PORT=8080 -e JWT_SECRET=CHANGE_ME_IN_PROD web_ws:latest
```


### 5) Docker Compose

```bash
docker compose up -d --build
```

对应配置见 `docker-compose.yml`。


## 接口说明

### WebSocket

- **路径（两种都支持）**
  - `GET /ws?token=YOUR_JWT_TOKEN`
  - `GET /api/ws/notifications?token=YOUR_JWT_TOKEN`

- **鉴权方式**
  - 必须通过 URL Query 传入 `token`
  - 服务端会用配置中的 `JWT_SECRET` 校验签名

> 具体 claims 字段以服务端 `internal/infrastructure/jwt` 的解析逻辑为准（至少包含 `UserID`、`Username` 用于标识连接）。


### 推送 Webhook

- **地址**：`POST /api/push`
- **Content-Type**：`application/json`

请求体示例：

```json
{
  "project_id": "project_001",
  "type": "alarm",
  "source_id": "device_01",
  "data": {
    "level": "high",
    "message": "temperature too high",
    "temperature": 87.2
  }
}
```

字段说明：

- `project_id`：项目/租户/业务域标识（用于路由与筛选）
- `type`：消息类型（如 alarm / event / test_message 等）
- `source_id`：来源标识（设备、系统、规则等）
- `data`：业务数据（任意 JSON 对象）


### 联调测试推送

- **地址**：`GET /api/test/push`
- **参数（可选）**
  - `project_id`：默认 `test_project`
  - `type`：默认 `test_message`
  - `source_id`：默认 `test_device_01`

示例：

```bash
curl "http://localhost:8080/api/test/push?project_id=p1&type=alarm&source_id=d1"
```


### 健康检查

- `GET /api/health`


### 系统统计

- `GET /api/stats`


## 配置说明

### 配置文件查找优先级

服务启动时按以下顺序查找 `config.yaml`：

1. **可执行文件同目录**：`./config.yaml`（发布/双击运行最稳定）
2. **项目根目录**：`config/config.yaml`（本地开发常用）
3. **当前工作目录**：`./config.yaml`

参考文件：`config/config.go`。


### 环境变量（最高优先级）

- `PORT`：服务端口
- `JWT_SECRET`：JWT 密钥
- `PUBLIC_BASE_URL`：对外可访问 Base URL（用于日志打印 WS/Webhook 地址）
- `WS_PATH`：WS 路径（默认 `/ws`）
- `ALARM_PUSH_PATH`：推送路径（默认 `/api/push`）


### 示例配置

见：`config/config.yaml`。


## 目录结构

```text
.
├── cmd/server/main.go              # 程序入口
├── config/
│   ├── config.go                   # 配置加载（文件 + env 覆盖 + 默认值）
│   └── config.yaml                 # 示例/生产配置
├── internal/
│   ├── api/                        # Gin 路由与 HTTP/WS Handler
│   ├── application/                # Hub/Client/MessageService 等应用层逻辑
│   ├── domain/                     # 领域模型与接口
│   ├── infrastructure/             # JWT 解析、WebSocket upgrader 等基础设施
│   └── pkg/middleware/             # 中间件（CORS）
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── test_ws.html                    # 简单 WebSocket 联调页面
```


## 联调方式

- **浏览器**：打开 `test_ws.html`，将连接地址改为你的服务地址与 token
- **curl**：调用 `POST /api/push` 或 `GET /api/test/push` 触发消息


## 安全注意事项

- **不要在生产环境使用默认 `JWT_SECRET`**（示例配置仅用于本地/演示）
- 建议通过环境变量注入密钥，并配合网关/反向代理做访问控制


## 常见问题（FAQ）

### 1. 连接 WebSocket 提示 `missing token` / `invalid token`

- 确认 URL 中带了 `?token=...`
- 确认 token 的签名密钥与服务端 `JWT_SECRET` 一致

### 2. 端口不对 / 日志打印的 URL 不对

- `PORT` 会覆盖配置文件端口
- 如需打印公网域名，设置 `PUBLIC_BASE_URL`


## License

MIT License. 详见根目录 `LICENSE`。


