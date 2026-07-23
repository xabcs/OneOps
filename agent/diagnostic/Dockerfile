# 多阶段构建 - 构建阶段
FROM docker.1ms.run/golang:1.21-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

# 设置Go代理（使用国内镜像加速）
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

# 设置工作目录
WORKDIR /build

# 复制go mod文件（先只复制go.mod）
COPY go.mod ./

# 下载依赖并生成go.sum（使用国内代理）
RUN go mod download && go mod tidy

# 复制源代码
COPY . .

# 构建二进制文件
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o diagnostic-agent \
    main.go

# 最终运行阶段
FROM docker.xuanyuan.run/library/alpine:3.23.5

# 安装运行时依赖
RUN apk add --no-cache \
    ca-certificates \
    curl \
    openjdk8-jre-base \
    bash \
    procps \
    && rm -rf /var/cache/apk/*

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/diagnostic-agent .

# 创建Arthas缓存目录
RUN mkdir -p /tmp/arthas

# 暴露端口
EXPOSE 8888

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q -O- http://localhost:8888/health || exit 1

# 启动agent
ENTRYPOINT ["./diagnostic-agent"]