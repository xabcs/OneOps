.PHONY: dev clean air run build test lint

# 开发模式 - 清理缓存并启动Air
dev: clean-air
	@echo "🚀 启动开发服务器..."
	air

# 清理Air缓存
clean-air:
	@echo "🧹 清理Air缓存..."
	@rm -rf backend/tmp
	@go clean -cache 2>/dev/null || true

# 仅启动后端（使用Air）
air:
	air

# 直接运行（不使用Air）
run:
	cd backend && go run main.go

# 构建
build:
	cd backend && go build -o bin/oneops main.go

# 测试
test:
	cd backend && go test ./... -v

# 代码检查
lint:
	cd backend && golangci-lint run

# 完全清理
clean: clean-air
	@echo "🧹 完全清理..."
	@rm -rf backend/bin
	@rm -rf backend/logs
	@go clean -modcache 2>/dev/null || true

# 强制重新编译
rebuild: clean-air
	@echo "🔄 强制重新编译..."
	@cd backend && go build -o tmp/main .
	@echo "✅ 编译完成，启动Air..."
	air
