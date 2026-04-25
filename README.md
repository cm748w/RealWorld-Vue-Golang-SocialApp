# 社交项目
## 1. 环境
- MongoDB（端口 27017）
- Go 1.16+
- Node.js 14+
- npm
## 2. 后端
```bash
cd backend/api
go mod tidy
```
启动服务：
```bash
go run main.go
```
后端地址：`http://localhost:5000`
## 3. 前端
```bash
cd frontend
```
安装依赖：
```bash
npm install
```
启动前端：
```bash
npm run serve
```
## 4. 使用
- 前端访问：`http://localhost:80`
- 后端接口：`http://localhost:5000`
- swagger：`http://localhost:5000/swagger/index.html`