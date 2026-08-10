
# OneOps Backend3 重构方案

## 一、目标

将 `backend/`（技术层扁平结构）重构为 `backend3/`（技术层 + 功能模块子目录），实现：

- 架构统一：model/repository/service/controller 四层均按功能模块分子目录
- 无循环依赖
- 功能完整不丢失
- 消除全局 db 变量，统一构造函数注入

## 二、最终目录结构

```
backend3/
├── cmd/api/main.go                          # 入口 + 依赖组装
├── model/                                   # 纯struct
│   ├── system/          # package modelsystem
│   ├── cmdb/            # package modelcmdb
│   ├── k8s/             # package modelk8s
│   ├── audit/           # package modelaudit
│   └── authorization/   # package modelauth
├── repository/                              # 数据库操作层
│   ├── system/
│   ├── cmdb/
│   ├── k8s/
│   ├── audit/
│   └── authorization/
├── service/                                 # 业务逻辑层
│   ├── system/
│   ├── cmdb/
│   ├── k8s/
│   ├── audit/
│   ├── monitoring/
│   │   └── notification/
│   └── authorization/
│       └── adapter/
├── controller/                              # HTTP控制器
│   ├── system/
│   ├── cmdb/
│   ├── k8s/
│   ├── audit/
│   ├── monitoring/
│   └── authorization/
├── routes/                                  # 扁平文件
│   ├── router.go
│   ├── system.go / cmdb.go / k8s.go / audit.go / monitoring.go / authorization.go
├── config/
└── pkg/
    ├── database/  logger/  middleware/  dto/  errors/  utils/
```

## 三、依赖方向（无环）

```
model:
  authorization, audit → (无外部依赖)
  system → (无外部依赖，基础层)
  cmdb → system (ServerAttribute.Definition → modelsystem.AttributeDefinition)
  k8s → system (ClusterRoleBinding.User/Role → modelsystem.User/Role)

repository → model (单向)
service → repository (单向)
controller → service (单向)
routes → controller (单向)
pkg/middleware → service (单向，service 不反向依赖)
```

## 四、代码改造清单

| # | 改造项                         | 详情                                                                        |
| - | ------------------------------ | --------------------------------------------------------------------------- |
| 1 | AttributeDefinition 关联标签   | `ServerAttributes` 从 `gorm:"foreignKey:AttributeID"` 改为 `gorm:"-"` |
| 2 | controller/route.go 下沉       | 路由生成算法移到 service/system/route_gen.go                                |
| 3 | controller/diagnostic.go 下沉  | Java分析+HTTP调用移到 service/k8s/diagnostic.go                             |
| 4 | controller/cmdb.go 拆分        | Agent管理拆到 controller/cmdb/agent.go                                      |
| 5 | controller/{menu,role,user}.go | 直接DB操作下沉到 repository/system/                                         |
| 6 | middleware/permission_db.go    | 改为调用 service 方法                                                       |
| 7 | 全局 var db 消除               | 统一改为 repository 构造函数注入                                            |
| 8 | container 删除                 | main.go 中手动组装依赖链                                                    |
| 9 | handler 并入 controller        | WebSocket 处理器移到 controller/{cmdb,k8s}/                                 |

## 五、实施步骤

### 步骤1：初始化 backend3

- 复制 backend → backend3
- 修改 go.mod: `module oneops/backend3`
- 创建所有分层子目录

### 步骤2：迁移 model 层

- models/ → model/{system,cmdb,k8s,audit,authorization}/
- 按模块拆分文件，修改 package 名
- 改造 AttributeDefinition 标签
- 跨模块引用加 import

### 步骤3：迁移 pkg 共享层

- errors/ → pkg/errors/
- utils/ → pkg/utils/
- config/ 保留
- logger/ → pkg/logger/
- dto/ → pkg/dto/
- middleware/ → pkg/middleware/
- services/db.go → pkg/database/
- services/redis.go → pkg/database/

### 步骤4：迁移 repository 层

- 从 services/ 中提取纯 CRUD 操作到 repository/{模块}/
- 统一 `NewXxxRepository(db *gorm.DB)` 构造函数

### 步骤5：迁移 service 层

- services/ → service/{system,cmdb,k8s,audit,monitoring,authorization}/
- 统一 `NewXxxService(repo)` 构造函数
- 下沉胖 controller 逻辑（route_gen, diagnostic）

### 步骤6：迁移 controller 层

- controller/ → controller/{模块}/
- handler/ 并入 controller/{cmdb,k8s}/
- 统一 `NewXxxController(service)` 构造函数

### 步骤7：迁移 routes + 入口

- routes/ 扁平文件
- main.go → cmd/api/main.go
- 手动组装依赖链

### 步骤8：编译修复

- 全局替换 import 路径
- go build ./... 迭代修复
