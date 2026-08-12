# Swagger API 文档

本项目使用 [swaggo/swag](https://github.com/swaggo/swag) 通过代码注释自动生成 OpenAPI(Swagger 2.0)文档,并通过 [gin-swagger](https://github.com/swaggo/gin-swagger) 提供 Swagger UI。

## 快速开始

### 1. 安装 swag CLI(仅需一次)

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

安装后 `swag` 二进制位于 `$GOBIN`(通常是 `~/go/bin`)。确保该路径在 `$PATH` 中。

### 2. 生成文档

```bash
make swag-init
```

或直接运行:

```bash
swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal --parseDepth 5
```

生成的文件:
- `docs/docs.go` —— Go 包,内嵌 swagger spec
- `docs/swagger.json` —— OpenAPI JSON
- `docs/swagger.yaml` —— OpenAPI YAML

### 3. 启动服务并访问 UI

```bash
make run
```

浏览器打开:

```
http://localhost:<port>/swagger/index.html
```

> Swagger UI **仅在非生产环境开放**(由 `config.App.Environment != "production"` 控制)。生产环境不会注册 `/swagger/*` 路由,避免接口暴露。

## 环境约束

本项目 `go.mod` 声明 `go 1.26.0`。若本机工具链版本较低,需设置以下环境变量(已固化在 `Makefile` 中):

```bash
export GOTOOLCHAIN=auto          # 按需自动下载对应工具链
export GOSUMDB=sum.golang.org    # 启用校验(toolchain 下载需要)
export GOPROXY=https://goproxy.cn,direct
```

## 为接口添加文档

在每个 HTTP handler 函数上方添加 swag 注释。模板:

```go
// GetUser godoc
// @Summary      获取用户列表
// @Description  分页获取系统用户列表,支持搜索
// @Tags         系统管理-用户
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "页码"    default(1)
// @Param        pageSize  query     int     false  "每页数量" default(20)
// @Param        username  query     string  false  "用户名"
// @Success      200  {object}  utils.Response{data=dto.PageResult}
// @Failure      200  {object}  utils.Response  "业务错误"
// @Router       /system/users [get]
// @Security     BearerAuth
func (ctrl *UserController) GetUsers(c *gin.Context) {
```

### 关键约定

| 项目 | 约定 |
|------|------|
| `@Router` 路径 | 写完整路径,**不含** `/api` 前缀(BasePath 已设为 `/api`) |
| `@Tags` | 按模块分组,如 `系统管理-用户`、`CMDB-服务器`、`K8s-集群管理`、`监控-告警` |
| 响应外层 | 统一用 `utils.Response{data=...}`;分页用 `utils.Response{data=dto.PageResult}` |
| 认证 | 除 `/login`、常量路由、agent 心跳、WebSocket 外,都加 `@Security BearerAuth` |
| `@Produce` | 仅接受:`json`、`xml`、`plain`、`html`、`octet-stream` 等,**不要用 `csv`** |
| 类型引用 | 注释中引用的类型必须在**所在 controller 文件已 import**,否则 swag 无法解析;跨包未导入的类型用 `object` 代替 |

### 特殊接口

- **文件下载(CSV)**:`@Produce plain` + `@Success 200 {file} binary`,如审计日志导出。
- **WebSocket**:`@Success 101 {string} string "升级为 WebSocket 连接"`,token 作为 query 参数,不加 `@Security`。

## 常用命令

| 命令 | 说明 |
|------|------|
| `make swag-init` | 生成 Swagger 文档到 `docs/` |
| `make swag-fmt` | 格式化 swag 注释 |
| `make build` | 生成文档 + 编译 |
| `make run` | 生成文档 + 运行 |
| `make vet` | 静态检查 |

## 完整文档生成后如何更新

新增或修改接口注释后,重新运行 `make swag-init`,提交更新后的 `docs/` 目录即可。

## API 元信息

元信息定义在 `cmd/api/main.go` 文件顶部(`@title`、`@version`、`@host`、`@BasePath` 等)。修改后需重新生成文档。
