FROM golang:1.26.1 AS builder
WORKDIR /usr/src/app
COPY backend/realTimeChat/go.mod ./
COPY backend/realTimeChat/go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download
COPY backend/realTimeChat .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /usr/src/app/main .
EXPOSE 8001
CMD [ "./main" ]