FROM node:24-bullseye-slim AS build
WORKDIR /usr/src/app
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
# Cypress 二进制从官方 CDN 下载很慢，改用 npmmirror 镜像加速
ENV CYPRESS_INSTALL_BINARY=https://registry.npmmirror.com/-/binary/cypress/15.14.1/linux-x64/cypress.zip
# 这里**刻意不安装 apt 工具链**（python3 / build-essential / git）。
#
# 原因：原实现为「换阿里云源 + apt-get install 工具链」，而阿里云的 bullseye-security
# 镜像同步不完整——元数据能取到，但大量 .deb 直接 404（实测报
# "Failed to fetch ... 404 Not Found / Unable to fetch some archives"），
# apt-get 以 exit 100 失败，导致整个前端镜像在本地和 CI 上都构建不出来。
# 而实测确认：本项目的依赖树里没有任何包需要从源码编译（lockfile 中无 node-gyp /
# requiresBuild 记录，sass 是纯 JS 的 dart-sass，Cypress 只下载预编译二进制），
# 因此这套工具链纯属无用负担——它没帮上忙，却把构建绑死在第三方镜像站的健康状态上。
#
# 若将来引入需要 node-gyp 的依赖，再按下面方式补回（并保留回退逻辑，不要再写死单一镜像源）：
#   RUN set -eu; cp /etc/apt/sources.list /tmp/orig.list; \
#       sed -i 's|deb.debian.org|mirrors.aliyun.com|g;s|security.debian.org|mirrors.aliyun.com|g' /etc/apt/sources.list || true; \
#       (apt-get update -o Acquire::Check-Valid-Until=false && \
#        apt-get install -y --no-install-recommends python3 build-essential git ca-certificates) || \
#       (cp /tmp/orig.list /etc/apt/sources.list && apt-get update && \
#        apt-get install -y --no-install-recommends python3 build-essential git ca-certificates); \
#       rm -rf /var/lib/apt/lists/*
RUN npm install -g pnpm@11.7.0 --registry=https://registry.npmmirror.com/ && \
	pnpm config set registry https://registry.npmmirror.com/ && \
	pnpm install --frozen-lockfile

COPY frontend .
RUN pnpm run build
FROM nginx:1.27.4-alpine

# ---------------------------------------------------------------------------
#  运行时非 root 化
#  nginx 官方 alpine 镜像自带 nginx 用户/组（uid/gid 101）。
#  以非 root 启动时，nginx 仍需要写这几个位置，因此先把属主改掉：
#    - /var/cache/nginx ：client_body/proxy 等临时文件目录（编译期默认路径）
#    - /var/log/nginx   ：access.log / error.log（镜像里软链到 /dev/stdout|stderr）
#    - /var/run/nginx.pid：master 进程 pid 文件（编译期默认 pid 路径）
#  静态资源目录无需写入权限，保持 root 属主即可。
# ---------------------------------------------------------------------------
RUN chown -R nginx:nginx /var/cache/nginx /var/log/nginx \
	&& touch /var/run/nginx.pid \
	&& chown nginx:nginx /var/run/nginx.pid

COPY --from=build /usr/src/app/dist /usr/share/nginx/html/app
COPY frontend/.nginx/nginx.conf /etc/nginx/nginx.conf

# 以非 root 用户运行：即使应用/静态文件被投毒，攻击者也无法直接写系统目录。
# 说明：nginx.conf 顶部的 `user nginx;` 在非 root 启动时会被 nginx 忽略
# （启动日志会打印一条 "the \"user\" directive makes sense only if the master
#  process runs with super-user privileges, ignored" 的 warn），这是 nginx 的既定行为：
# master 本身就不是 root，没有权限也不需要在 worker 前再 setuid，worker 直接
# 继承启动用户（nginx）。因此该指令保留即可，不必删除。
# 另外 compose 把 80 端口映射到宿主机，容器内仍监听 80：Docker 20.10+ 默认在容器内
# 设置 net.ipv4.ip_unprivileged_port_start=0，非 root 进程可以绑定 80。
USER nginx

# 健康检查：命中 nginx.conf 里新增的 /healthz（200 + "ok"）。
# nginx:alpine 自带 busybox wget，因此不需要额外安装工具。
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
	CMD wget -qO- http://127.0.0.1/healthz >/dev/null 2>&1 || exit 1
