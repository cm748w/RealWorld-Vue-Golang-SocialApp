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



##  演示地址

- 本地演示：`http://localhost`
- Swagger：`http://localhost:5000/swagger/index.html`



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
| Frontend | Vue 3、Vue Router、Vuex、Quasar、Axios、Cypress、pnpm、Nginx | 单页应用、状态管理、路由、端到端测试、依赖管理与静态资源交付 |
| Backend | Go、Fiber、gRPC、WebSocket、MongoDB、Swagger、godotenv | REST API、实时服务、数据存储、接口文档与环境变量管理 |
| DevOps | Docker、Docker Compose、.env、healthcheck、Nginx | 容器化交付、服务编排、健康检查、前端静态托管 |

## 项目截图


### PC 端

![首页 PC 端截图](docs/screenshots/home-desktop.png)
![帖子详情 PC 端截图](docs/screenshots/post-detail-desktop.png)
![聊天 PC 端截图](docs/screenshots/chat-desktop.png)

<!-- ### 手机端

![首页 手机端截图](docs/screenshots/home-mobile.png)
![聊天 手机端截图](docs/screenshots/chat-mobile.png) -->


## 快速开始

### 环境要求

- Docker Desktop 或 Docker Engine + Docker Compose v2
- Node.js 18+ 与 pnpm 9+（仅在本地前端开发时需要）
- Go 1.26+（仅在本地后端开发时需要）
- MongoDB（Docker 模式由 Compose 自动启动 `mongo:7.0`；纯本地开发需自备 MongoDB，并建议使用端口 `27018`，详见下文「MongoDB 端口规划」）


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
> Compose 会自动读取根目录 [.env](.env) 中的变量。前端已不依赖构建地址参数：浏览器访问前端时，API 与 WebSocket 请求走同源相对路径（`/api/`、`/ws-notify/`、`/ws-chat/`），由 nginx 反向代理到各后端服务，因此通过任意 IP/域名（含局域网 IP）访问均可正常连接后端。

### 本地开发部署

如果你需要单独调试某个服务，可以采用“容器 + 本地进程”的组合方式：

#### 1. 启动基础依赖

- **推荐：直接使用本地 MongoDB**。纯本地开发时，API 默认连接本机 `localhost:27018`（见 `backend/api/.env`），无需启动任何容器。
- 如果本机没有安装 MongoDB，也可以先启动 Compose 中的容器版 MongoDB（`mongo:7.0`）：

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
pnpm install
pnpm run serve
```

> [!NOTE]
> 前端使用 **pnpm** 作为包管理器：锁文件为 `pnpm-lock.yaml`，构建脚本白名单配置在 `pnpm-workspace.yaml`（`allowBuilds`，如 Cypress 二进制下载）。安装新依赖使用 `pnpm add <pkg>`，CI 与 Docker 构建均基于 pnpm。

前端默认读取 `frontend/.env`，本地开发时会连接：
- `http://localhost:5000/`
- `ws://localhost:8088/ws/`
- `ws://localhost:8001/ws/`

#### 5. 数据库迁移说明

当前项目使用 MongoDB，未引入独立的迁移工具或 SQL migration 文件。

> [!NOTE]
> 也就是说，项目启动时不需要额外执行“数据库迁移”命令，核心数据集合会在服务运行和首次写入时创建。

### MongoDB 端口规划（本地 27018 与容器 27017）

项目允许本地 MongoDB 与容器版 MongoDB 并存、互不冲突：

| 环境 | 监听地址 | 说明 |
| --- | --- | --- |
| 容器版（Compose `mongodb` 服务） | 宿主 `27017`；容器网络内为 `mongodb:27017` | 仅用于 Docker 全家桶，数据存放在数据卷 `mongo-data` |
| 本地版（纯本地开发） | `127.0.0.1:27018` | 避开容器版 `27017`，数据存放在本机 `dbPath` |

> [!NOTE]
> - 容器内的 API 通过 Compose 内部网络以服务名 `mongodb:27017` 连接容器版 MongoDB，宿主端口映射仅对外部访问可见，因此两套 MongoDB 可以各自独立工作。
> - 如需调整本地 MongoDB 端口，请修改本机 `mongod.cfg`（Windows 默认位于 `C:\Program Files\MongoDB\Server\<版本>\bin\mongod.cfg`）中的 `net.port`，并同步更新 `backend/api/.env` 的 `MONGO_URI`。

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

该文件由 Docker Compose 读取。它已被根目录 `.gitignore` 忽略，不会进入版本库；如果本地缺失，请按下方示例在项目根目录手动创建。

```dotenv
MONGO_URI=mongodb://admin:123456@mongodb:27017
JWT_SECRET=your-jwt-secret
POSTS_PAGE_SIZE=10
POSTS_MAX_PAGE_SIZE=50
MONGO_INITDB_ROOT_USERNAME=admin
MONGO_INITDB_ROOT_PASSWORD=123456
GOLANG_NOTIFY_SERVICE_ADDR=GolangNotifyService:8090
GOLANG_API_SERVER_ADDR=GolangApiServer:5001
```

### `backend/api/.env`

用于 API 服务本地开发：

```dotenv
# 本地 MongoDB 使用端口 27018，避开 Docker 版 MongoDB 的 27017
MONGO_URI=mongodb://localhost:27018
JWT_SECRET=your-jwt-secret
POSTS_PAGE_SIZE=10
POSTS_MAX_PAGE_SIZE=50
GOLANG_NOTIFY_SERVICE_ADDR=localhost:8090
```

### `backend/realTimeChat/.env`

用于实时聊天服务本地开发：

```dotenv
GOLANG_API_SERVER_ADDR=localhost:5001
```

### `frontend/.env`

前端已不再需要构建地址变量：API 与 WebSocket 均使用运行时相对路径（`/api/`、`/ws-notify/`、`/ws-chat/`），由 nginx 反向代理到后端，因此无需在此配置任何地址。

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
		├── package.json
		├── pnpm-lock.yaml
		└── pnpm-workspace.yaml
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



## 作者与联系方式

- 作者：cm748w
- GitHub：https://github.com/cm748w/RealWorld-Vue-Golang-SocialApp
- 邮箱：shanacongyun@163.com
