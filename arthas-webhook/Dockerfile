# MsrePilot 镜像（多阶段构建，构建上下文即本目录）
#
# 构建（在 arthas-webhook/ 目录内执行）：
#   docker build --platform linux/amd64 \
#     -t mpm-image-registry.cn-hangzhou.cr.aliyuncs.com/base/msre-pilot:v1 .
#   docker push mpm-image-registry.cn-hangzhou.cr.aliyuncs.com/base/msre-pilot:v1
FROM docker.1ms.run/golang:1.26-alpine AS builder
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# 独立工程无 DB/API 依赖，仅 gin + zap + k8s api 类型，构建产物 < 20MB
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/msre-pilot .

FROM docker.1ms.run/alpine:3.20
RUN adduser -D -u 10001 msre-pilot
COPY --from=builder /out/msre-pilot /usr/local/bin/msre-pilot
USER 10001
EXPOSE 9443
ENTRYPOINT ["msre-pilot"]
