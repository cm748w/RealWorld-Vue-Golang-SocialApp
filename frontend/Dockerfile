FROM node:24-bullseye-slim AS build
WORKDIR /usr/src/app
COPY frontend/package.json frontend/package-lock.json ./
RUN sed -i 's|deb.debian.org|mirrors.aliyun.com|g' /etc/apt/sources.list || true && \
	sed -i 's|security.debian.org|mirrors.aliyun.com|g' /etc/apt/sources.list || true && \
	apt-get update && \
	apt-get install -y --no-install-recommends python3 build-essential git ca-certificates && \
	rm -rf /var/lib/apt/lists/* && \
	npm config set registry https://registry.npmmirror.com/ && \
	npm ci

COPY frontend .
ARG VUE_APP_API_URL=http://localhost:5000/
ARG VUE_APP_RealTimeNotificationUrl=ws://localhost:8088/ws/
ARG VUE_APP_RealTimeChatUrl=ws://localhost:8001/ws/
ENV VUE_APP_API_URL=$VUE_APP_API_URL
ENV VUE_APP_RealTimeNotificationUrl=$VUE_APP_RealTimeNotificationUrl
ENV VUE_APP_RealTimeChatUrl=$VUE_APP_RealTimeChatUrl
RUN npm run build
FROM nginx:1.27.4-alpine
COPY --from=build /usr/src/app/dist /usr/share/nginx/html
COPY frontend/.nginx/nginx.conf /etc/nginx/nginx.conf