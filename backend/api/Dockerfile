FROM golang:1.26.1 AS builder
WORKDIR /usr/src/app
COPY backend/api/go.mod ./
COPY backend/api/go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download
COPY backend/api .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl
WORKDIR /root/
COPY --from=builder /usr/src/app/main .
EXPOSE 5000
CMD [ "./main" ]