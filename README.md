# RealWorld Vue Golang Social App

> Vue 3 + Go 全栈社交应用：REST API、gRPC、WebSocket 实时通信，Docker Compose 一键启动。

<p align="center">
	<img src="https://img.shields.io/badge/Vue-3-42b883?logo=vue.js&logoColor=white" alt="Vue 3" />
	<img src="https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white" alt="Go" />
	<img src="https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white" alt="Docker Compose" />
	<img src="https://img.shields.io/badge/MongoDB-7.0-47A248?logo=mongodb&logoColor=white" alt="MongoDB" />
</p>

## 简介

前端为 Vue 3 单页应用，后端拆分为 API、实时聊天、实时通知三个 Go 服务，同时提供 gRPC 与 WebSocket 能力，覆盖前后端分离、服务拆分、实时交互与容器化部署的完整实践。

主启动方式为 Docker Compose：根目录 [.env](.env) 注入变量，[docker-compose.yml](docker-compose.yml) 启动前端、三个后端服务与 MongoDB，并通过 healthcheck 保证依赖顺序就绪。

## 主要特性

- 🔐 注册 / 登录、JWT 鉴权与受保护接口
- 📝 发帖、帖子列表与分页
- 💬 WebSocket 一对一实时聊天
- 🔔 消息与状态变化实时通知
- 🧩 API / Chat / Notification 服务拆分
- 🐳 Compose 一键启动，适合本地开发、联调与演示

## 技术栈

| 层级 | 技术 |
| --- | --- |
| 前端 | Vue 3、Vue Router、Vuex、Quasar、Axios、Cypress、pnpm、Nginx |
| 后端 | Go、Fiber、gRPC、WebSocket、MongoDB、Swagger |
| 运维 | Docker Compose、healthcheck、.env、Nginx |

## 界面预览

![首页](docs/screenshots/home-desktop.png)
![帖子详情](docs/screenshots/post-detail-desktop.png)
![聊天](docs/screenshots/chat-desktop.png)

## 快速开始

**环境要求**：Docker Desktop 或 Docker Engine + Compose v2；本地开发另需 Go 1.26+、Node.js 22.13+ 与 pnpm 11+。

```bash
# 1. 在项目根目录创建 .env（格式见「环境变量配置」）
# 2. 启动整套服务
docker compose up --build -d
# 3. 查看状态与日志
docker compose ps
docker compose logs -f GolangApiServer
```

| 入口 | 地址 |
| --- | --- |
| 前端 | `http://localhost`（`8080` 端口同样指向前端） |
| API | `http://localhost:5000` |
| Swagger | `http://localhost:5000/swagger/index.html` |

局域网内其他设备访问时把 `localhost` 换成本机 IP 即可：页面内的 API 与 WebSocket 请求都走同源相对路径（`/api/`、`/ws-notify/`、`/ws-chat/`），由 nginx 反代到后端，无需额外配置。

停止服务：`docker compose down`。加 `-v` 会同时删除 `mongo-data` 数据卷（数据将丢失）。

## 本地开发

单服务调试时可用「容器 + 本地进程」组合。本地 MongoDB 建议监听 `27018`，避开容器版的 `27017`。

```bash
# 基础依赖（本机已有 MongoDB 可跳过）
docker compose up -d mongodb

# 后端 API（读取 backend/api/.env）
cd backend/api && go run main.go

# 实时聊天服务
cd backend/realTimeChat && go run main.go

# 前端
cd frontend && pnpm install && pnpm run serve
```

前端 dev server 会把相对路径转发到后端：`/api/*` → `:5000`，`/ws-notify/*` → `:8088`，`/ws-chat/*` → `:8001`。`pnpm run serve` 默认监听 80 端口，被占用时改 `package.json` 里的 `--port`。

## 环境变量配置

根目录 `.env` 由 Compose 读取，已被 `.gitignore` 忽略，需手动创建：

```dotenv
MONGO_URI=mongodb://admin:CHANGE_ME@mongodb:27017
JWT_SECRET=CHANGE_ME
POSTS_PAGE_SIZE=10
POSTS_MAX_PAGE_SIZE=50
MONGO_INITDB_ROOT_USERNAME=admin
MONGO_INITDB_ROOT_PASSWORD=CHANGE_ME
GOLANG_NOTIFY_SERVICE_ADDR=GolangNotifyService:8090
GOLANG_API_SERVER_ADDR=GolangApiServer:5001
```

`backend/api/.env`（本地开发，连接本机 27018）：

```dotenv
MONGO_URI=mongodb://localhost:27018
JWT_SECRET=CHANGE_ME
POSTS_PAGE_SIZE=10
POSTS_MAX_PAGE_SIZE=50
GOLANG_NOTIFY_SERVICE_ADDR=localhost:8090
```

`backend/realTimeChat/.env`：

```dotenv
GOLANG_API_SERVER_ADDR=localhost:5001
JWT_SECRET=CHANGE_ME
```

> [!IMPORTANT]
> `JWT_SECRET` 必须配置：API 启动时校验，缺失即退出；聊天与通知服务在 WS 握手时自行校验，缺失或与 API 不一致会拒绝所有连接。Compose 部署由根目录 `.env` 统一注入。

| 变量 | 作用 |
| --- | --- |
| `JWT_SECRET` | 签发 / 校验 JWT，必填 |
| `GRPC_REFLECTION` | 设为 `1` 才开启 gRPC 反射，默认关闭 |
| `CORS_ALLOWED_ORIGINS` | 跨域白名单，逗号分隔；同源部署无需配置 |

> [!NOTE]
> `MONGO_INITDB_ROOT_USERNAME` / `MONGO_INITDB_ROOT_PASSWORD` 只在数据卷**首次初始化**时生效。之后修改密码需同步修改卷内用户（或 `docker compose down -v` 重置），否则 API 会因 SCRAM 认证失败而无法读写数据库。

`frontend/.env` 与 `backend/realTimeNotification` 均无需额外配置。

## 项目结构

```text
.
├── docker-compose.yml
├── .env                       # Compose 变量（不入库）
├── backend
│   ├── api/                   # REST API + gRPC server
│   ├── realTimeChat/          # 聊天 WebSocket 服务
│   └── realTimeNotification/  # 通知 WebSocket 服务
└── frontend/                  # Vue 3 SPA、nginx 配置与构建产物
```

## API 与实时端点

| 功能 | 端点 |
| --- | --- |
| 注册 / 登录 | `POST /user/signup`、`POST /user/signin` |
| 帖子 | `GET /posts`、`POST /posts`、`PATCH /posts/:id` |
| 聊天（前端） | `ws://<host>/ws-chat/:id` |
| 通知（前端） | `ws://<host>/ws-notify/:userId` |
| 聊天 / 通知（后端原始） | `:8001/ws/:id`、`:8088/ws/:userId` |

Swagger UI：`http://localhost:5000/swagger/index.html`；接口定义位于 `backend/api/docs/`。

## 生产部署

`docker compose up -d --build` 即可。建议前置 Nginx 或云负载均衡做统一入口与 TLS 终止，持久化 `mongo-data` 卷，并采集容器日志、配置健康检查告警。

## 贡献

欢迎提交 Issue 与 Pull Request：Fork 仓库并建分支 → 本地 `docker compose up --build` 验证 → 提交 PR 并说明变更内容与验证结果。

新增环境变量时请同步更新根目录 `.env` 与本文件；新增容器服务时请补充 healthcheck 和启动说明。
