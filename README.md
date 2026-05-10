# RealWorld Vue Golang Social App

> 一套支持 REST、gRPC、WebSocket 的 Vue 3 + Go 社交应用，适合演示现代全栈架构、实时通信和 Docker Compose 一键交付。

<p align="center">
	<img src="https://img.shields.io/badge/Vue-3.0-42b883?logo=vue.js&logoColor=white" alt="Vue 3" />
	<img src="https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white" alt="Go" />
	<img src="https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white" alt="Docker Compose" />
	<img src="https://img.shields.io/badge/MongoDB-7.0-47A248?logo=mongodb&logoColor=white" alt="MongoDB" />
	<img src="https://img.shields.io/badge/gRPC-Enabled-6D67E4" alt="gRPC" />
	<img src="https://img.shields.io/badge/WebSocket-Real--Time-F59E0B" alt="WebSocket" />
	<img src="https://img.shields.io/badge/version-dev-6B7280" alt="Version" />
	<img src="https://img.shields.io/badge/license-TBD-lightgrey" alt="License" />
</p>

## 项目简介

这是一个面向社交场景的全栈项目，前端使用 Vue 3 构建单页应用，后端由多个 Go 服务组成，分别负责 REST API、实时聊天和实时通知。系统同时提供 gRPC 与 WebSocket 能力，适合展示前后端分离、服务拆分、实时交互和容器化部署等完整工程实践。

项目当前以 Docker Compose 作为主启动方式：
- 通过根目录 [.env](.env) 统一注入 Compose 和前端构建所需变量
- 通过 [docker-compose.yml](docker-compose.yml) 一键启动前端、API、聊天服务、通知服务和 MongoDB
- 通过健康检查保证服务按依赖顺序就绪后再对外提供访问

> [!NOTE]
> 当前仓库没有单独的线上 Demo 地址，建议先用 Docker Compose 在本地启动整套系统。

## 在线预览 / 演示地址

- 在线 Demo：暂无
- 本地演示：`http://localhost`
- Swagger：`http://localhost:5000/swagger/index.html`

> [!TIP]
> 如果你后续部署到公网服务器，可以在这里补充正式演示地址、API 域名和 WebSocket 地址。

## 主要特性

- 🔐 账户体系：支持注册、登录、JWT 鉴权与受保护接口访问。
- 📝 内容流：支持发帖、浏览帖子和分页加载。
- 💬 实时聊天：基于 WebSocket 的一对一消息通信。
- 🔔 实时通知：用户消息和状态变化可即时推送。
- 🧩 服务拆分：API、Chat、Notification 三个后端服务独立运行，职责清晰。
- 🐳 容器友好：Compose 一键启动，适合本地开发、联调和演示环境。

## 技术栈

| 层级 | 技术栈 | 说明 |
| --- | --- | --- |
| Frontend | Vue 3、Vue Router、Vuex、Quasar、Axios、Cypress、Nginx | 单页应用、状态管理、路由、端到端测试与静态资源交付 |
| Backend | Go、Fiber、gRPC、WebSocket、MongoDB、Swagger、godotenv | REST API、实时服务、数据存储、接口文档与环境变量管理 |
| DevOps | Docker、Docker Compose、.env、healthcheck、Nginx | 容器化交付、服务编排、健康检查、前端静态托管 |

## 项目截图

请将截图放到例如 `docs/screenshots/` 目录下，并按下面的占位图替换：

### PC 端

![首页 PC 端截图](docs/screenshots/home-desktop.png)
![帖子详情 PC 端截图](docs/screenshots/post-detail-desktop.png)
![聊天 PC 端截图](docs/screenshots/chat-desktop.png)

### 手机端

![首页 手机端截图](docs/screenshots/home-mobile.png)
![聊天 手机端截图](docs/screenshots/chat-mobile.png)

> [!NOTE]
> 截图建议覆盖登录、发帖、聊天和通知弹窗等核心路径，方便新用户快速理解产品。

## 快速开始

### 环境要求

- Docker Desktop 或 Docker Engine + Docker Compose v2
- Node.js 18+（仅在本地前端开发时需要）
- Go 1.26+（仅在本地后端开发时需要）
- MongoDB（由 Compose 自动启动，无需手动安装）

> [!TIP]
> 如果你只想体验完整功能，优先使用 Docker Compose，不需要单独安装前后端依赖。

### 一键启动整套系统（推荐）

1. 克隆仓库并进入项目根目录。
2. 确认根目录 [.env](.env) 已填写正确的本地值。
3. 启动整套服务：

```bash
docker compose up --build -d
```

4. 查看运行状态：

```bash
docker compose ps
```

5. 打开应用：
- 前端：`http://localhost`
- API：`http://localhost:5000`
- Swagger：`http://localhost:5000/swagger/index.html`

6. 查看日志：

```bash
docker compose logs -f GolangApiServer
```

7. 停止并清理：

```bash
docker compose down
```

> [!NOTE]
> Compose 会自动读取根目录 [.env](.env) 中的变量，其中前端构建地址应使用浏览器可访问的 `localhost`，而后端服务间通信则使用容器网络里的服务名。

### 本地开发部署

如果你需要单独调试某个服务，可以采用“容器 + 本地进程”的组合方式：

#### 1. 启动基础依赖

只启动 MongoDB 和相关服务，或直接启动整套 Compose 环境：

```bash
docker compose up -d mongodb
```

如果你希望一起调试 API、Chat 和 Notification，也可以直接使用完整 Compose。

#### 2. 启动后端 API

```bash
cd backend/api
go run main.go
```

API 本地启动时会优先读取当前目录下的 `.env`，用于加载 MongoDB、JWT 和通知服务地址等配置。

#### 3. 启动实时聊天服务

```bash
cd backend/realTimeChat
go run main.go
```

聊天服务本地启动时会读取当前目录下的 `.env`，用于加载 API gRPC 地址。

#### 4. 启动前端

```bash
cd frontend
npm install
npm run serve
```

前端默认读取 `frontend/.env`，本地开发时会连接：
- `http://localhost:5000/`
- `ws://localhost:8088/ws/`
- `ws://localhost:8001/ws/`

#### 5. 数据库迁移说明

当前项目使用 MongoDB，未引入独立的迁移工具或 SQL migration 文件。

> [!NOTE]
> 也就是说，项目启动时不需要额外执行“数据库迁移”命令，核心数据集合会在服务运行和首次写入时创建。

## 生产环境部署

### Docker Compose 部署

这是当前最推荐的部署方式，适合自建服务器、测试环境和演示环境。

```bash
docker compose up -d --build
```

建议在服务器上配套以下资源：
- 反向代理层：Nginx 或云负载均衡
- 持久化存储：MongoDB 数据卷
- 监控与日志：容器日志采集、健康检查和告警

### Docker + Nginx

如果你希望拆分前后端，也可以保留：
- 前端静态站点交给 Nginx 托管
- API、Chat、Notification 仍然以容器方式部署
- 通过 Nginx 或网关做统一入口和 TLS 终止

> [!TIP]
> 当前仓库已经把前端打包进 Nginx 镜像，适合直接使用 Compose 对外提供服务。

### 不太建议的方案

- Vercel：更适合纯前端静态站点，不适合当前这种 gRPC + WebSocket + 多服务后端结构。
- Railway / 纯函数平台：对长连接和多服务间内部发现的支持通常不如容器编排直观。

## 环境变量配置

### 根目录 `.env`

该文件由 Docker Compose 读取，同时也用于前端构建参数注入。

```dotenv
MONGO_URI=mongodb://admin:123456@mongodb:27017
JWT_SECRET=your-jwt-secret
POSTS_PAGE_SIZE=2
POSTS_MAX_PAGE_SIZE=50
MONGO_INITDB_ROOT_USERNAME=admin
MONGO_INITDB_ROOT_PASSWORD=123456
GOLANG_NOTIFY_SERVICE_ADDR=GolangNotifyService:8090
GOLANG_API_SERVER_ADDR=GolangApiServer:5001
VUE_APP_API_URL=http://localhost:5000/
VUE_APP_RealTimeNotificationUrl=ws://localhost:8088/ws/
VUE_APP_RealTimeChatUrl=ws://localhost:8001/ws/
```

### `backend/api/.env`

用于 API 服务本地开发：

```dotenv
MONGO_URI=mongodb://localhost:27017
JWT_SECRET=your-jwt-secret
POSTS_PAGE_SIZE=2
POSTS_MAX_PAGE_SIZE=50
GOLANG_NOTIFY_SERVICE_ADDR=localhost:8090
```

### `backend/realTimeChat/.env`

用于实时聊天服务本地开发：

```dotenv
GOLANG_API_SERVER_ADDR=localhost:5001
```

### `frontend/.env`

用于前端本地开发：

```dotenv
VUE_APP_API_URL=http://localhost:5000/
VUE_APP_RealTimeNotificationUrl=ws://localhost:8088/ws/
VUE_APP_RealTimeChatUrl=ws://localhost:8001/ws/
```

> [!NOTE]
> `backend/realTimeNotification` 当前不需要单独的 `.env` 文件。

## 项目结构

```text
.
├── docker-compose.yml
├── .env
├── backend
│   ├── api
│   │   ├── main.go
│   │   ├── Dockerfile
│   │   ├── controllers/
│   │   ├── database/
│   │   ├── middleware/
│   │   ├── models/
│   │   ├── routes/
│   │   ├── servergrpc/
│   │   ├── tests/
│   │   └── validation/
│   ├── realTimeChat
│   │   ├── main.go
│   │   ├── Dockerfile
│   │   ├── realtime/
│   │   └── servegrpc/
│   └── realTimeNotification
│       ├── main.go
│       ├── Dockerfile
│       ├── realtime/
│       └── servegrpc/
└── frontend
		├── src/
		├── public/
		├── Dockerfile
		├── vue.config.js
		└── package.json
```

## 核心功能使用说明

### 1. 注册与登录

访问前端首页后，先完成注册/登录，系统会签发 JWT，并在后续请求中使用该令牌。

```bash
POST /user/signup
POST /user/signin
```

### 2. 发帖与浏览

登录后可以创建帖子、浏览帖子列表和查看分页结果。

```bash
GET /posts
POST /posts
```

### 3. 实时聊天

进入聊天页面后，前端会通过 WebSocket 与聊天服务建立连接，实现消息的实时发送和接收。

```text
ws://localhost:8001/ws/:id
```

### 4. 实时通知

当用户收到消息或产生通知事件时，通知服务会通过 WebSocket 推送更新。

```text
ws://localhost:8088/ws/:userId
```

> [!TIP]
> 如果你在调试实时功能，建议同时打开前端页面、浏览器控制台和 `docker compose logs -f`，能更快定位消息链路问题。

## API 接口文档

- Swagger UI：`http://localhost:5000/swagger/index.html`
- Swagger JSON：`backend/api/docs/swagger.json`
- Swagger YAML：`backend/api/docs/swagger.yaml`

当前 API 文档由 Go Swagger 注释生成，适合配合前端联调和接口回归。

## 贡献指南

欢迎提交 Issue 和 Pull Request。

建议的贡献流程：

1. Fork 仓库并创建分支。
2. 基于当前 Compose 流程本地验证功能。
3. 提交代码前检查 `docker compose config` 与本地运行日志。
4. 通过后提交 PR，并描述变更内容与验证结果。

### 开发建议

- 新增环境变量时，同时更新根目录 `.env` 和 README 对应章节。
- 新增容器服务时，同时补充 healthcheck 和 README 启动说明。
- 修改实时服务时，优先验证 WebSocket 和 gRPC 链路。

## License

当前仓库尚未包含独立的 `LICENSE` 文件。

> [!NOTE]
> 如果你后续决定采用 MIT、Apache-2.0 或其他许可证，建议把 `LICENSE` 文件补进仓库并同步更新这一节。

## 作者与联系方式

- 作者：请在此填写你的名字或团队名称
- GitHub：请在此填写你的仓库主页
- 邮箱：请在此填写你的联系邮箱

## 致谢

感谢以下技术与项目生态提供支持：

- Vue 3
- Go 生态与 Fiber
- MongoDB
- Docker / Docker Compose
- gRPC 与 WebSocket
- Swagger / OpenAPI

<!-- > [!TIP] -->
<!-- > 如果你希望，我还可以继续帮你把 README 再细化成“开源项目风格”或“企业交付风格”两种版本。 -->