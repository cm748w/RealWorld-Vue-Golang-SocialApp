# 社交应用
---
## 必要前提
- **MongoDB**：确保本地安装并运行 MongoDB（默认端口 27017）
- **Go**：后端需要 Go 1.16+ 环境
- **Node.js**：前端需要 Node.js 14+ 环境（推荐 16+）
- **Yarn 或 npm**：用于前端依赖管理
- **air（可选）**：Go 热重载工具，提升开发效率
- **搞这些还挺费劲儿的**
- **不配好运行不了**
---
## 后端
### 1. 进入后端目录
```bash
cd backend/api
```
### 2. 安装依赖
```bash
go mod tidy
```
### 3. 配置环境变量
**重要**：至少也要把 `.env.example` 文件名改为 `.env`  ！！！
**后端 .env.example 内容：**
```env
# MongoDB 连接 URI (默认: mongodb://localhost:27017)
MONGO_URI=mongodb://localhost:27017

# JWT 密钥，用于生成和验证令牌
JWT_SECRET=FbhWg2iGc3unA6rMTkylqLHZzEJmCx0s

# 帖子列表的默认每页数量（对应接口：GET /posts）
POSTS_PAGE_SIZE=2

# 帖子列表 API 允许的最大值（对应接口：GET /posts）
POSTS_MAX_PAGE_SIZE=50
```
### 4. 生成 Swagger 文档
```bash
# 先安装 swag
go install github.com/swaggo/swag/cmd/swag@latest
# 生成文档
make swag-init
```
### 5. 启动服务
#### 方法一：生产环境
```bash
go run main.go
# 启动路径：backend/api/main.go
```
#### 方法二：开发环境
```bash
# 先安装 air
go install github.com/air-verse/air@latest
# 启动热重载服务
air
# 启动路径：backend/api/main.go
```
**启动成功标志**：服务运行于 `http://localhost:5000`，Swagger 文档访问地址：`http://localhost:5000/swagger/index.html`
### 后端必看
- **数据库**：默认连接 `mongodb://localhost:27017`，数据库名固定为 `social`
- **环境变量**：`JWT_SECRET` 为必填项，生产环境请使用高强度随机字符串
  - 生成方法：
    - PowerShell：`-join ((65..90) + (97..122) + (48..57) | Get-Random -Count 32 | ForEach-Object {[char]$_})`
    - Linux/macOS：`openssl rand -base64 32`
- **端口**：后端监听 5000 端口，修改位置 => `backend/api/main.go:64`
- **热重载**：使用 `air` 时，提升开发效率（swagger文档仍需手动输入make swag-init）
---
## 前端
### 1. 进入前端目录
```bash
cd frontend
```
### 2. 安装依赖
```bash
# 使用 yarn
yarn install
# 或使用 npm
npm install
# 用一个就行了
```
### 3. 配置 API 地址
**重要**：至少也要把 `.env.example` 文件名改为 `.env`  ！！！
```bash
cp .env.example .env
# 编辑 .env 文件，设置 VUE_APP_API_BASE_URL
# 例如：VUE_APP_API_BASE_URL=http://localhost:5000
```
**前端 .env.example 内容：**
```env
# 默认是写好的，若要改动，请务必与 `backend/api/main.go:64` 同步
VUE_APP_API_URL=http://localhost:5000/
```
### 4. 启动开发服务器
```bash
# 使用 yarn
yarn serve
# 或使用 npm
npm run serve
```
**启动成功标志**：前端运行于 `http://localhost:80`
### 前端注意事项
- **端口**：前端默认使用 80 端口，如被占用会自动切换到其他端口
- **API 地址**：默认 `http://localhost:5000`，
- **浏览器**：我用的 Chrome 浏览器 [下载 Chrome](https://dl.google.com/tag/s/appguid%3D%7B8A69D345-D564-463C-AFF1-A69D9E530F96%7D%26iid%3D%7B95C015C5-143C-E71A-8407-17631385306A%7D%26lang%3Den%26browser%3D5%26usagestats%3D1%26appname%3DGoogle%2520Chrome%26needsadmin%3Dprefers%26ap%3D-arch_x64-statsdef_1%26installdataindex%3Dempty/update2/installers/ChromeSetup.exe)
---
## 常见问题
### 1. 后端启动失败
- **错误**：`Error: JWT_SECRET is required`
  **解决**：检查 `.env` 文件是否正确配置 `JWT_SECRET`
- **错误**：`Error: failed to connect to MongoDB`
  **解决**：确保 MongoDB 服务正在运行，检查连接字符串是否正确
### 2. 前端无法访问后端 API
- **错误**：`404 Not Found` 或 `CORS` 错误
  **解决**：检查后端服务是否启动，API 地址是否配置正确
### 3. 功能无法使用
- **聊天/通知**：确保 后端 5000 端口可访问
- **用户认证**：检查 JWT 配置是否正确
---
## 技术栈参考
| 层级 | 技术 |
|------|------|
| 前端框架 | Vue 3 |
| UI 组件库 | Quasar Framework |
| 状态管理 | Vuex 4 |
| 后端框架 | Fiber (Go) |
| 数据库 | MongoDB |
| 实时通信 | WebSocket |
| 接口文档 | Swagger + swag |
| 开发工具 | air（热重载） |
---
## 最后提醒
- 必开 MongoDB 服务
- 再查.env
- 前后端 API 地址保持一致
- 开发时用 `air` 
- 有问题，问豆包
