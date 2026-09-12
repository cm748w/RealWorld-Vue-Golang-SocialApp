# 构建阶段用 bookworm（Debian 12），而不是 bullseye（Debian 11）。
#
# 为什么：bullseye 已进入 EOL，其 security 仓库元数据不再刷新，Release 文件的
# Valid-Until 永久过期 —— 实测在**官方源**上 `apt-get update` 同样以 exit 100 失败
# （"Release file ... is expired"），换任何镜像站都一样。之前阿里云那次 .deb 404
# 只是同一个病的二阶症状，锅不在阿里云。Debian 12 的安全仓库仍在正常更新：
# 同一台机器、同一网络下实测 apt-get update 21s 通过。
FROM node:24-bookworm-slim AS build
WORKDIR /usr/src/app
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
# Cypress 二进制从官方 CDN 下载很慢，改用 npmmirror 加速（这是 npm 生态的镜像，
# 与 apt 源无关，不会把"构建能否成功"绑在第三方镜像站的同步状态上）
ENV CYPRESS_INSTALL_BINARY=https://registry.npmmirror.com/-/binary/cypress/15.14.1/linux-x64/cypress.zip
#
# 这里**不换 apt 源、也不装 apt 工具链**。原实现是「无条件换阿里云源 + 装
# python3/build-essential/git」，两个问题：
#   1) 硬编码单一镜像源，构建成败被第三方镜像站的同步健康度绑架（阿里云那次同步不完整，
#      .deb 直接 404，apt 以 exit 100 失败，前端镜像在本地和 CI 上都构建不出来）；
#   2) 本项目没有任何依赖需要从源码编译（lockfile 无 node-gyp / requiresBuild，
#      sass 是纯 JS 的 dart-sass，Cypress 只下预编译二进制），工具链纯属无用负担。
#
# 结论：直接用基镜像自带的官方源即可 —— 实测本机与 CI 都不需要任何代理
# （deb.debian.org 国内直连 8 秒内可拉完全部 Packages 列表）。
# 将来若确实要装编译工具链，请用官方源 + 由构建环境注入代理，不要写死镜像源：
# Docker 有预定义代理参数，**不需要在 Dockerfile 里声明 ARG**，--build-arg 即可生效：
#   docker build --build-arg HTTP_PROXY=http://host.docker.internal:7897 \
#                --build-arg HTTPS_PROXY=http://host.docker.internal:7897 .
# 或者一次性写进 ~/.docker/config.json 的 proxies.default（pull 与 build 都生效）。
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
