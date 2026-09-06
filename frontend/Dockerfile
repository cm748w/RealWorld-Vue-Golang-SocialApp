FROM node:24-bullseye-slim AS build
WORKDIR /usr/src/app
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
# Cypress 二进制从官方 CDN 下载很慢，改用 npmmirror 镜像加速
ENV CYPRESS_INSTALL_BINARY=https://registry.npmmirror.com/-/binary/cypress/15.14.1/linux-x64/cypress.zip
RUN sed -i 's|deb.debian.org|mirrors.aliyun.com|g' /etc/apt/sources.list || true && \
	sed -i 's|security.debian.org|mirrors.aliyun.com|g' /etc/apt/sources.list || true && \
	apt-get update && \
	apt-get install -y --no-install-recommends python3 build-essential git ca-certificates && \
	rm -rf /var/lib/apt/lists/* && \
	npm install -g pnpm@11.7.0 --registry=https://registry.npmmirror.com/ && \
	pnpm config set registry https://registry.npmmirror.com/ && \
	pnpm install --frozen-lockfile

COPY frontend .
RUN pnpm run build
FROM nginx:1.27.4-alpine
COPY --from=build /usr/src/app/dist /usr/share/nginx/html/app
COPY frontend/.nginx/nginx.conf /etc/nginx/nginx.conf