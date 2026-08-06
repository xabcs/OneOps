# 数据库初始化架构优化方案

## 当前问题

### 1. 代码组织混乱
- **文件过大**：`init.go` 文件达到 1268 行
- **职责不清**：初始化、同步、迁移逻辑混杂
- **数据分散**：菜单、权限、属性等数据散落在各处

### 2. 初始化函数过多
- 18 个初始化相关函数
- 函数命名不规范（init*、sync*、create* 混用）
- 依赖关系不明确

### 3. 缺乏统一管理
- 硬编码的数据定义
- 没有版本控制
- 迁移脚本分散

## 优化方案

### 架构分层

```
backend/services/
├── init/
│   ├── initializer.go          # 初始化协调器
│   ├── migrator.go            # 数据库迁移管理
│   ├── data_loader.go         # 数据加载器
│   └── modules/               # 模块初始化器
│       ├── system.go          # 系统模块（用户、角色、菜单）
│       ├── permission.go      # 权限初始化
│       ├── cmdb.go            # CMDB模块初始化
│       ├── monitor.go         # 监控模块初始化
│       ├── k8s.go             # K8s模块初始化
│       └── diagnostic.go      # 诊断模块初始化
├── data/                       # 数据定义文件（JSON/YAML）
│   ├── menus.json             # 菜单配置
│   ├── permissions.json       # 权限配置
│   ├── attributes.json        # 属性定义
│   └── roles.json             # 角色配置
└── migrations/                 # SQL迁移脚本
    ├── 001_initial_schema.sql
    ├── 002_add_disk_partitions.sql
    └── 003_add_agent_metrics.sql
```

### 初始化阶段设计

#### 阶段 1：数据库模式（Schema）
- 创建表结构
- 添加索引
- 设置外键约束

#### 阶段 2：基础数据（Seed Data）
- 内置用户
- 内置角色
- 菜单数据

#### 阶段 3：权限数据（Permission Data）
- 权限定义
- Casbin策略同步
- 角色权限分配

#### 阶段 4：模块数据（Module Data）
- CMDB：属性定义、Agent版本
- Monitor：告警规则、通知渠道
- K8s：集群配置
- Diagnostic：诊断配置

### 数据定义标准化

#### JSON Schema 示例（permissions.json）

```json
{
  "version": "1.0.0",
  "modules": [
    {
      "code": "system",
      "name": "系统管理",
      "level": 1,
      "resources": [
        {
          "code": "user",
          "name": "用户管理",
          "level": 2,
          "actions": [
            {"code": "list", "name": "用户列表", "level": 3},
            {"code": "view", "name": "查看用户", "level": 3},
            {"code": "create", "name": "创建用户", "level": 3}
          ]
        }
      ]
    }
  ]
}
```

## 重构步骤

### 步骤 1：创建初始化协调器
```go
// init/initializer.go
type Initializer struct {
    migrator     *Migrator
    dataLoader   *DataLoader
    modules      []ModuleInitializer
}

func (i *Initializer) Initialize() error {
    // 阶段1：模式迁移
    if err := i.migrator.MigrateSchema(); err != nil {
        return err
    }

    // 阶段2：基础数据
    if err := i.dataLoader.LoadSeedData(); err != nil {
        return err
    }

    // 阶段3：模块初始化
    for _, module := range i.modules {
        if err := module.Initialize(); err != nil {
            return err
        }
    }

    return nil
}
```

### 步骤 2：提取数据定义
- 将 `permissions_data.go` 改为 `data/permissions.json`
- 创建 `data/menus.json`
- 创建 `data/attributes.json`

### 步骤 3：模块化初始化
```go
// init/modules/permission.go
type PermissionModule struct {
    db *gorm.DB
}

func (m *PermissionModule) Initialize() error {
    // 1. 加载权限数据
    permissions, err := m.loadPermissions()
    if err != nil {
        return err
    }

    // 2. 同步到数据库
    if err := m.syncPermissions(permissions); err != nil {
        return err
    }

    // 3. 同步到Casbin
    if err := m.syncToCasbin(); err != nil {
        return err
    }

    return nil
}
```

## 优势

### 1. 可维护性
- ✅ 清晰的分层架构
- ✅ 单一职责原则
- ✅ 易于定位问题

### 2. 可扩展性
- ✅ 新增模块只需添加初始化器
- ✅ 数据定义与代码分离
- ✅ 支持版本化迁移

### 3. 可测试性
- ✅ 模块独立测试
- ✅ Mock数据加载
- ✅ 集成测试简化

### 4. 可读性
- ✅ 代码组织清晰
- ✅ 配置文件直观
- ✅ 文档自动生成

## 实施建议

### 优先级 P0（立即实施）
1. 创建初始化协调器
2. 提取权限数据为JSON
3. 分离菜单初始化逻辑

### 优先级 P1（近期实施）
1. 创建数据加载器
2. 模块化初始化器
3. 添加版本控制

### 优先级 P2（长期优化）
1. SQL迁移脚本管理
2. 数据验证机制
3. 回滚支持

## 兼容性保障

### 保持现有API
- `InitService.InitDatabase()` 保持不变
- 内部调用重构后的模块

### 渐进式迁移
```go
// services/init.go (保留)
func (s *InitService) InitDatabase() error {
    initializer := init.NewInitializer(db)
    return initializer.Initialize()
}
```

## 预期效果

| 指标 | 优化前 | 优化后 |
|------|--------|--------|
| 代码行数 | 1268行 | ~300行（协调器） |
| 文件数量 | 1个 | 10+个模块化文件 |
| 数据定义 | 硬编码 | JSON配置文件 |
| 可维护性 | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| 可测试性 | ⭐⭐ | ⭐⭐⭐⭐ |
