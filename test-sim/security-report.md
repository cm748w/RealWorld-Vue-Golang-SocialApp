# 安全评估报告 — 恶意用户 + 黑客攻击实录

- 目标：运行中的系统 `http://localhost:8080`（前端 dev server → 后端 API `:5000`、通知 WS `:8088`、聊天 WS `:8001`）
- 方式：授权测试（系统属主本人要求），**未改动任何源码**
- 工具：`test-sim/attack.mjs`（全量攻击）、`test-sim/verify-critical.mjs`、`test-sim/verify-final.mjs`（裁决验证）
- 时间：2026-08-27

---

## 一、确认的攻击成功（按严重度排序）

### 🔴 严重 1 — 任意用户资料 + bcrypt 密码哈希公开泄露
- **路径**：`GET /api/user/getUser/:id`（路由未挂 `AuthMiddleware`）
- **证据**：匿名（无任何 token）请求 `user/getUser/{victimId}` → `200`，响应体：
  ```
  {"_id":"...","name":"VV A","email":"...","password":"$2a$10$bhtQsBh01wrha...","bio":"","followers":[...]}
  ```
  `password` 字段携带 bcrypt 哈希。
- **影响**：攻击者可遍历/枚举任意用户 ID，批量抓取**密码哈希**做离线爆破（弱密码必破）、撞库或哈希重用；同时姓名、邮箱、关注关系全部公开。

### 🔴 严重 2 — 通知接口完全公开（可读可改）
- **路径**：`GET /api/notification/:userid`、`PATCH /api/notification/mark-notification-as-readed/:userid`（均未挂鉴权中间件）
- **证据**：B 关注 A 产生一条通知后，**匿名**请求 `GET notification/{A.id}` → `200`，返回 A 的 1 条通知；**匿名** `PATCH mark-notification-as-readed/{A.id}` → `200`。
- **影响**：任意人可读取任意用户收到的通知（关注、点赞、评论动态），并可把他人通知标记为已读，破坏提醒完整性。

### 🔴 严重 3 — 通知 WebSocket 实时窃听（零鉴权）
- **路径**：`ws://localhost:8088/ws/{userId}`（通知服务未做身份校验）
- **证据**：无任何 token 直连 `ws://localhost:8088/ws/{victimId}` → 连接建立；触发一次关注后，**实时收到受害者事件**：
  ```json
  {"_id":"...","details":"FA B Start Following You!","mainUserId":"...victim...","targetId":"...victim..."}
  ```
- **影响**：攻击者可同时订阅任意多个受害者的通知流，**实时监控**对方的一举一动（谁关注/互动了谁），无需任何凭据。

### 🟡 中危 4 — 登录爆破无防护
- **路径**：`POST /api/user/signin`
- **证据**：同一账号连续 10/12 次错误密码，全部**立即**返回 401，无任何限速、锁定或验证码。
- **影响**：可对弱密码账号离线式暴力破解（配合第 1 条泄露的哈希，风险叠加）。

### 🟡 中危 5 — 存储型 XSS payload 原样存储/返回
- **路径**：`POST /api/posts` 等写接口
- **证据**：发帖内容 `<img src=x onerror=alert(1)>` 存库后原样读回（API 不做任何 HTML 消毒）。
- **影响**：当前 Vue 前端全部使用文本插值（已核查**无任何 v-html/innerHTML**），故现有客户端不直接执行；但这是**纵深防御缺口**——任何未来引入 `v-html`、或第三方以 HTML 渲染该数据的消费方都会中招。属"修复成本极低、必须修"类。

### 🟡 中危 6 — 邮箱枚举
- **路径**：`POST /api/user/signup`
- **证据**：重复注册已存在邮箱 → `409`；新邮箱 → `201`。状态码可区分账号是否存在。
- **影响**：可批量验证哪些邮箱已注册，配合钓鱼/定向攻击。

### 🟡 中危 7 — 账号/内容灌水无限制
- **证据**：6 次快速注册全部成功（`201`）；15 次快速发帖全部成功（`201`）；发帖接口无任何频率限制。
- **影响**：垃圾账号、垃圾帖刷屏，污染 feed 与数据库。

### 🟡 低危 8 — 帖子内容完全公开
- **路径**：`GET /api/posts`、`GET /api/posts/search`、`GET /api/posts/:id`（无鉴权）
- **证据**：匿名即可拉取全部帖子列表与详情、执行搜索。
- **影响**：若产品预期"登录可见"，则为越权信息泄露；若预期公开社区，则属设计使然（但建议至少隐藏邮箱等敏感字段）。

---

## 二、已攻击但被防御住的（误报排除记录）

| 攻击 | 结果 | 说明 |
|---|---|---|
| JWT `alg:none` 伪造 | 🟢 拒绝 | 受保护接口（如 `getSug`、`createPost`）返回 401；之前误判为通过，实因 `getUser` 本身不鉴权 |
| JWT 空密钥 / 常见弱密钥伪造 | 🟢 拒绝 | 密钥非空，签名校验正常；伪造 `iss=受害者` 的 token 调 `createPost` → 401 |
| 过期 token 重放 | 🟢 拒绝 | `exp` 校验生效 |
| IDOR 资料写入（body 伪造 `_id`） | 🟢 拒绝 | `PATCH user/update` 使用 token 身份，忽略 body `_id`（测试中把自己改名了） |
| 越权修改/删除他人帖子 | 🟢 拒绝 | `PATCH/DELETE /posts/:id` 有 creator 校验（403） |
| 聊天跨用户读取会话 | 🟢 拒绝 | `getmsgsbynums` 校验请求方身份，非会话双方返回空 |
| 聊天伪造发送者 | 🟢 拒绝 | 消息 sender 以后端取 token 身份为准 |
| NoSQL 注入（signin `$ne`/`$gt`、query 操作符嵌套） | 🟢 拒绝 | 参数被安全处理 |
| 限流绕过（`X-Forwarded-For` 伪造 IP） | 🟢 拒绝 | `getUser` 限流按真实 IP 计数 |
| 发帖伪造 creator 归属 | 🟢 拒绝 | creator 取 token 身份 |
| 聊天 WebSocket 无鉴权窃听 | 🟢 拒绝 | 无 token 连 `ws-chat/{id}` 收不到事件（与通知 WS 形成对比） |

---

## 三、建议修复优先级（本次未改任何代码）

1. **P0**：给 `GET /user/getUser/:id` 挂 `AuthMiddleware`，且**响应剔除 `password` 字段**（所有 user 序列化处都应 `omitempty`/过滤密码哈希）。
2. **P0**：给通知两个接口挂鉴权，并校验 `userid` 与 token 身份一致。
3. **P0**：通知 WS 接入鉴权（握手校验 token，校验 `userId` 与身份一致）。
4. **P1**：`signin` 加登录限速/失败锁定；`signup`、发帖加频率限制。
5. **P1**：写接口对富文本做服务端消毒（或前端统一转义 + 后端白名单校验）。
6. **P2**：`signup` 对"邮箱已存在"返回统一提示，消除枚举。
7. **P2**：评估 feed/search/详情是否应要求登录。

---

## 五、修复状态（已按 P0→P1→P2 全部修复并验证）

| # | 发现 | 修复 | 验证 |
|---|---|---|---|
| 1 | 密码哈希公开泄露 | getUser 挂鉴权；所有 user 响应剔除 password（含搜索接口）+ `json:"password,omitempty"` | 🟢 `password=undefined` |
| 2 | 通知接口公开 | 挂鉴权 + 校验 userid 与 token 一致 + 精确匹配过滤 | 🟢 匿名/跨用户均 401 |
| 3 | 通知 WS 窃听 | 两个 WS 服务校验 `?token=` JWT（iss 与路径 id 一致） | 🟢 无 token 拒绝、有 token 连通 |
| 4 | 登录爆破 | 每 IP 20 次/分 + 15 分钟内 5 次失败锁定 1 分钟 | 🟢 第 6 次 429 |
| 5 | 存储型 XSS | 服务端消毒（script 块、事件属性、javascript:/data: URI），覆盖帖子/评论/资料/聊天 | 🟢 载荷入库后为空 |
| 6 | 邮箱枚举 | 重复注册返回统一 400 通用错误（不回显邮箱）+ 注册限流 10 次/分/IP | 🟢 通用消息 |
| 7 | 灌水 | signup 10 次/分/IP；发帖 30 次/分/用户 | 🟢 触发 429 |
| 8 | 帖子公开 | feed/search/详情挂鉴权 | 🟢 匿名 401 |

- 回归：`full-flow.mjs` 45/45、`module-test.mjs` 20/20、`verify-fixed.mjs` 24/24 全部通过
- 三个后端服务已用修复后的代码重启（API :5000、通知 :8088/8090、聊天 :8001）

---

## 六、第二轮攻击（修复后重攻，2026-08-27）—— 新发现 9 项，未动源码

| # | 严重度 | 发现 | 证据 |
|---|---|---|---|
| 2.1 | 中 | **注册竞态 TOCTOU — 同邮箱可创建重复账号**（users 集合无 email 唯一索引） | 6 个并发注册同一邮箱 → `201×2`（**两个不同 _id**）+ 429×4；该邮箱可正常登录 |
| 2.2 | 中 | 点赞无频率限制 | 25 次快速点赞全部 200 |
| 2.3 | 中 | 评论无频率限制 | 25 次快速评论全部 201（配合 2.2 可刷爆帖互动） |
| 2.4 | 中 | 关注/取关无频率限制 | 20 次快速切换全部 200（可对目标轰炸通知） |
| 2.5 | 中 | 私信无频率限制 | 12 条快速消息全部 201（消息洪泛） |
| 2.6 | 低 | Swagger 文档公开 | `GET /api/swagger/index.html` → 200（为攻击者提供完整路由图） |
| 2.7 | 低 | 错误信息泄露 | 畸形 JSON → 500 + 原始 Go 解析错误文本（暴露框架内部） |
| 2.8 | 低 | CORS 全开 | `Access-Control-Allow-Origin: http://evil.example.com`（任意来源）+ `Allow-Credentials: true` |
| 2.9 | 低 | imageUrl/selectedFile 未消毒 | `javascript:alert(1)`、`data:image/svg+xml` 原样入库（浏览器对 `<img :src>` 阻断执行，影响有限） |

**验证过的安全项（本轮未突破）**：登录锁定不影响正确密码登录（只拦截失败尝试）；聊天已读标记无法跨用户；getUser/通知/WS/限流等 P0-P2 修复全部保持生效。

### 修复状态（已全部修复并验证，2026-08-27）

| # | 修复 | 验证 |
|---|---|---|
| 2.1 | users.email 唯一索引 + DuplicateKey 统一错误（先清理了既有重复账号） | 🟢 5 并发同邮箱注册 → `201×1 + 400×4` |
| 2.2 | likePost 限流 30 次/分/用户 | 🟢 第 31 次 429 |
| 2.3 | commentPost 限流 10 次/分/用户 | 🟢 超限 429 |
| 2.4 | follow 限流 10 次/分/用户 | 🟢 超限 429 |
| 2.5 | sendChatMessage 限流 20 条/分/用户 | 🟢 第 21 条 429 |
| 2.6 | Swagger 文档公开 | ~~挂鉴权~~（**应属主要求已回退为公开**：改用 Apifox 测试，且所有业务接口本身都有 AuthMiddleware 保护，文档仅暴露路由结构） | 🟡 已知可接受 |
| 2.7 | 自定义 Fiber 错误处理，500 返回通用信息 | 🟢 不再泄露解析错误 |
| 2.8 | CORS 收紧为本地/内网来源（API + 两个 WS） | 🟢 任意来源不再反射 |
| 2.9 | imageUrl/selectedFile 协议白名单（http/https/data:image） | 🟢 javascript:/SVG 载荷被清空 |

回归：`attack2.mjs` 0 发现；`verify-fixed.mjs` 24/24；`full-flow.mjs` 45/45；`module-test.mjs` 20/20。

---

## 复现命令

```bash
node test-sim/attack.mjs          # 全量攻击（含误报与命中标注）
node test-sim/verify-final.mjs    # 裁决验证（只留实证）
node test-sim/verify-fixed.mjs    # 修复后验证（24 项全部关闭）
node test-sim/attack2.mjs         # 第二轮攻击（9 项新发现）
```
