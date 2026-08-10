.PHONY: help run build test lint fmt clean docker

# 默认目标
help:
	@echo "可用命令:"
	@echo "  make run      - 运行应用"
	@echo "  make build    - 构建应用"
	@echo "  make test     - 运行测试"
	@echo "  make lint     - 代码检查"
	@echo "  make fmt      - 格式化代码"
	@echo "  make clean    - 清理构建"
	@echo "  make docker   - 构建Docker镜像"

# 运行应用
run:
	go run main.go

# 构建应用
build:
	go build -o bin/oneops main.go

# 运行测试
test:
	go test -v ./...

# 代码检查
lint:
	@echo "检查golangci-lint是否安装..."
	@which golangci-lint > /dev/null || (echo "请安装golangci-lint: brew install golangci-lint" && exit 1)
	golangci-lint run

# 格式化代码
fmt:
	@echo "格式化Go代码..."
	go fmt ./...
	@echo "运行goimports..."
	@which goimports > /dev/null || (echo "请安装goimports: go install golang.org/x/tools/cmd/goimports@latest" && exit 1)
	goimports -w .

# 清理构建
clean:
	rm -rf bin/ tmp/ logs/

# 构建Docker镜像
docker:
	docker build -t oneops:latest .

# 安装开发工具
install-tools:
	@echo "安装开发工具..."
	go install github.com/air-verse/air@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@which golangci-lint > /dev/null || brew install golangci-lint
	@echo "开发工具安装完成"
