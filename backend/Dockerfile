# ============ 编译阶段 ============
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o server ./cmd/server

# ============ 运行阶段 ============
FROM alpine:3.19

WORKDIR /app

# 安装时区数据
RUN apk add --no-cache tzdata

# 从编译阶段复制二进制
COPY --from=builder /app/server .
COPY --from=builder /app/configs ./configs

# 创建非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup
USER appuser

EXPOSE 8080

ENTRYPOINT ["./server"]
