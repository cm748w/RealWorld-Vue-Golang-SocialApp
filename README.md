# RealWorld-Vue-Golang-SocialApp

基于 Go + Fiber + MongoDB 的社交应用后端示例，提供认证、用户、帖子、聊天、通知等 API，并内置 Swagger 文档。

## 功能概览

- 用户注册与登录（JWT）
- 用户资料与关系相关接口
- 帖子发布、查询与分页
- 聊天与消息功能
- 通知功能
- Swagger 接口文档

## 快速开始

1. 进入后端目录：

```bash
cd backend/api
```

2. 安装依赖：

```bash
go mod tidy
```

3. 创建环境变量文件 `backend/api/.env`：

```env
# MongoDB 连接地址（可选，默认 mongodb://localhost:27017）
MONGO_URI=mongodb://localhost:27017

# JWT 密钥（必填，生产环境请使用高强度随机字符串）
JWT_SECRET=your-secret-key

# 帖子列表默认分页大小（可选，默认 2）
POSTS_PAGE_SIZE=2

# 帖子列表最大分页大小（可选，默认 100）
POSTS_MAX_PAGE_SIZE=100
```

4. 启动服务：

```bash
go run main.go
```

默认监听地址为 `http://localhost:5000`。

## Swagger 文档

- 访问地址：`http://localhost:5000/swagger/index.html`
- 重新生成文档（需先安装 swag）：

```bash
make swag-init
```

## 环境变量说明

- `MONGO_URI`：MongoDB 连接字符串；未设置时使用 `mongodb://localhost:27017`。
- `JWT_SECRET`：JWT 签名密钥。
- `POSTS_PAGE_SIZE`：帖子列表默认 `limit`。
- `POSTS_MAX_PAGE_SIZE`：帖子列表允许的最大 `limit`。

## 说明

- 数据库名固定为 `social`。
- 项目启动时会加载 `.env`，若缺失会导致服务启动失败。
