.PHONY: swag-init swag-fmt build run tidy vet clean help

# Go 相关环境变量：
# - GOTOOLCHAIN=auto：允许按 go.mod 要求自动下载对应版本工具链
# - GOSUMDB=sum.golang.org：启用校验（toolchain 下载需要）
# - GOPROXY：使用国内代理加速
export GOTOOLCHAIN := auto
export GOSUMDB     := sum.golang.org
export GOPROXY     := https://goproxy.cn,direct

# swag CLI 路径（优先用 GOBIN/go/bin 下的）
SWAG := $(shell command -v swag 2> /dev/null)
MAIN := cmd/api/main.go

# 生成 Swagger 文档
swag-init:
ifndef SWAG
	$(error "未找到 swag CLI，请先运行: go install github.com/swaggo/swag/cmd/swag@latest")
endif
	swag init -g $(MAIN) -o docs --parseDependency --parseInternal --parseDepth 5

# 格式化 swag 注释
swag-fmt:
ifndef SWAG
	$(error "未找到 swag CLI，请先运行: go install github.com/swaggo/swag/cmd/swag@latest")
endif
	swag fmt -g $(MAIN) -d .

# 生成文档 + 编译
build: swag-init
	go build -o main ./cmd/api

# 生成文档 + 运行
run: swag-init
	go run ./cmd/api

# 整理依赖
tidy:
	go mod tidy

# 静态检查
vet:
	go vet ./...

clean:
	rm -f main

help:
	@echo "可用目标:"
	@echo "  swag-init  生成 Swagger 文档到 docs/"
	@echo "  swag-fmt   格式化 swag 注释"
	@echo "  build      生成文档并编译"
	@echo "  run        生成文档并运行"
	@echo "  tidy       整理 go.mod 依赖"
	@echo "  vet        静态检查"
