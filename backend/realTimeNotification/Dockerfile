FROM golang:1.26.1 AS builder
WORKDIR /usr/src/app
COPY backend/realTimeNotification/go.mod ./
COPY backend/realTimeNotification/go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download
COPY backend/realTimeNotification .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /usr/src/app/main .
COPY --from=builder /usr/src/app/.env .
EXPOSE 8090
EXPOSE 8088
CMD [ "./main" ]