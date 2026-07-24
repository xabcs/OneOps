# 应用权限管理架构设计方案

## 问题分析

不同应用的权限模型差异很大：
- **Jenkins**: 用户 + 全局角色 + 项目角色 + 角色分配
- **JumpServer**: 用户 + 用户组 + 授权规则（无需角色）
- **Nacos**: 角色 + 用户
- **GitLab**: 用户 + 组 + 项目 + 成员角色

现有问题：
- ❌ 表结构膨胀：每个应用都要加专门表
- ❌ 代码复杂：大量重复的CRUD逻辑
- ❌ 维护困难：添加新应用需要理解所有表结构

## 推荐架构：通用化 + 适配器模式

### 核心思想

1. **抽象通用实体**：用户、角色、用户组 → 统一为"主体"（Subject）
2. **JSON存储扩展**：特定应用的结构化数据存储在JSON字段
3. **适配器隔离差异**：每个应用适配器处理自己的数据模型

### 数据库设计

#### 1. 应用表（已存在，保持不变）
```sql
CREATE TABLE applications (
    id BIGINT PRIMARY KEY,
    name VARCHAR(100),
    type VARCHAR(50),     -- jenkins, jumpserver, nacos, gitlab...
    base_url VARCHAR(255),
    auth_config JSON,     -- 认证配置
    endpoints JSON,       -- 端点配置
    ...
);
```

#### 2. 应用实体表（通用，替代多个专门表）
```sql
CREATE TABLE application_entities (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    app_id BIGINT NOT NULL,
    entity_type VARCHAR(20) NOT NULL,  -- user, role, group, project...
    entity_id VARCHAR(100) NOT NULL,    -- 外部系统的ID
    entity_name VARCHAR(100) NOT NULL,  -- 显示名称
    entity_data JSON NULL,              -- 扩展数据（特定应用的额外字段）
    parent_type VARCHAR(20) NULL,       -- 父实体类型（用于层级关系）
    parent_id VARCHAR(100) NULL,        -- 父实体ID
    sync_time DATETIME,
    created_at DATETIME,
    updated_at DATETIME,
    UNIQUE KEY uk_app_entity (app_id, entity_type, entity_id),
    KEY idx_app_type (app_id, entity_type)
);
```

**设计说明：**
- `entity_type` 替代多个专门表：`user`、`role`、`group`、`project`等
- `entity_data` 存储应用特定的扩展数据
  - Jenkins用户: `{"fullName": "...", "email": "..."}`
  - JumpServer用户: `{"name": "...", "email": "...", "is_active": true}`
  - Jenkins项目角色: `{"project": "...", "permissions": [...]}`
- `parent_type` + `parent_id` 处理层级关系（如用户属于组）

#### 3. 应用权限映射表（通用）
```sql
CREATE TABLE application_bindings (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    app_id BIGINT NOT NULL,
    subject_type VARCHAR(20) NOT NULL,  -- user, group
    subject_id VARCHAR(100) NOT NULL,   -- 用户/用户组ID
    object_type VARCHAR(20) NOT NULL,   -- role, project, resource
    object_id VARCHAR(100) NOT NULL,    -- 角色/项目/资源ID
    binding_data JSON NULL,             -- 绑定扩展数据
    created_at DATETIME,
    updated_at DATETIME,
    KEY idx_app_subject (app_id, subject_type, subject_id),
    KEY idx_app_object (app_id, object_type, object_id)
);
```

**设计说明：**
- `subject_type` + `subject_id`：谁（用户/用户组）
- `object_type` + `object_id`：什么（角色/项目/资源）
- `binding_data`：绑定相关的扩展数据
  - Jenkins: `{"permissions": ["hudson.model.Hudson.Read"]}`
  - JumpServer授权规则: `{"assets": [...], "actions": [...]}`

### 应用层设计

#### 1. 统一的数据模型

```go
// 应用实体（通用）
type ApplicationEntity struct {
    ID          uint            `json:"id"`
    AppID       uint            `json:"appId"`
    EntityType  string          `json:"entityType"`   // user, role, group, project...
    EntityID    string          `json:"entityId"`
    EntityName  string          `json:"entityName"`
    EntityData  json.RawMessage `json:"entityData"`    // 应用特定数据
    ParentType  string          `json:"parentType,omitempty"`
    ParentID    string          `json:"parentId,omitempty"`
    SyncTime    time.Time       `json:"syncTime"`
    CreatedAt   time.Time       `json:"createdAt"`
    UpdatedAt   time.Time       `json:"updatedAt"`
}

// 应用权限绑定（通用）
type ApplicationBinding struct {
    ID          uint            `json:"id"`
    AppID       uint            `json:"appId"`
    SubjectType string          `json:"subjectType"`   // user, group
    SubjectID   string          `json:"subjectId"`
    ObjectType  string          `json:"objectType"`    // role, project, resource
    ObjectID    string          `json:"objectId"`
    BindingData json.RawMessage `json:"bindingData"`
    CreatedAt   time.Time       `json:"createdAt"`
    UpdatedAt   time.Time       `json:"updatedAt"`
}
```

#### 2. 适配器接口（增强）

```go
type ApplicationAdapter interface {
    // === 基本信息 ===
    ValidateConfig(config map[string]interface{}) error
    GetConfigTemplate() map[string]interface{}
    GetDisplayName() string

    // === 实体管理（通用）===
    FetchEntities(baseURL string, authConfig map[string]interface{}, entityType string) ([]ApplicationEntity, error)
    CreateEntity(baseURL string, authConfig map[string]interface{}, entityType string, entity *ApplicationEntity) error

    // === 权限绑定（通用）===
    CreateBinding(baseURL string, authConfig map[string]interface{}, binding *ApplicationBinding) error
    DeleteBinding(baseURL string, authConfig map[string]interface{}, binding *ApplicationBinding) error

    // === 支持的实体类型（元数据）===
    GetSupportedEntityTypes() []string  // 如: ["user", "role", "group", "project"]
    GetEntityMetadata(entityType string) *EntityMetadata  // 获取实体类型的元数据
}

// 实体类型元数据
type EntityMetadata struct {
    TypeName         string              // "user", "role", "group"
    DisplayName       string              // "用户", "角色", "用户组"
    SyncSupported    bool                // 是否支持同步
    CreateSupported  bool                // 是否支持创建
    BindSupported    bool                // 是否支持绑定
    Fields           []FieldDefinition  // 该实体类型的字段定义
}

// 字段定义
type FieldDefinition struct {
    FieldName    string      // 字段名
    DisplayName  string      // 显示名称
    FieldType    string      // string, number, boolean, array
    Required     bool        // 是否必需
    Searchable   bool        // 是否可搜索
    DisplayInList bool       // 是否在列表中显示
}
```

#### 3. 适配器实现示例

```go
// JumpServer 适配器
type JumpserverAdapter struct{}

func (j *JumpserverAdapter) GetSupportedEntityTypes() []string {
    return []string{"user", "group"}  // JumpServer 不支持角色
}

func (j *JumpserverAdapter) GetEntityMetadata(entityType string) *EntityMetadata {
    switch entityType {
    case "user":
        return &EntityMetadata{
            TypeName:   "user",
            DisplayName: "用户",
            Fields: []FieldDefinition{
                {FieldName: "username", DisplayName: "用户名", FieldType: "string", Required: true, Searchable: true, DisplayInList: true},
                {FieldName: "name", DisplayName: "姓名", FieldType: "string", Required: false, Searchable: false, DisplayInList: true},
                {FieldName: "email", DisplayName: "邮箱", FieldType: "string", Required: false, Searchable: true, DisplayInList: true},
                {FieldName: "is_active", DisplayName: "是否激活", FieldType: "boolean", Required: false, Searchable: true, DisplayInList: true},
            },
        }
    case "group":
        return &EntityMetadata{
            TypeName:   "group",
            DisplayName: "用户组",
            Fields: []FieldDefinition{
                {FieldName: "name", DisplayName: "组名", FieldType: "string", Required: true, Searchable: true, DisplayInList: true},
                {FieldName: "comment", DisplayName: "备注", FieldType: "string", Required: false, Searchable: false, DisplayInList: false},
            },
        }
    }
    return nil
}

func (j *JumpserverAdapter) FetchEntities(baseURL string, authConfig map[string]interface{}, entityType string) ([]ApplicationEntity, error) {
    switch entityType {
    case "user":
        return j.fetchUsers(baseURL, authConfig)
    case "group":
        return j.fetchGroups(baseURL, authConfig)
    default:
        return nil, fmt.Errorf("JumpServer 不支持实体类型: %s", entityType)
    }
}

// Jenkins 适配器
type JenkinsAdapter struct{}

func (j *JenkinsAdapter) GetSupportedEntityTypes() []string {
    return []string{"user", "role", "project"}  // Jenkins 支持项目角色
}

func (j *JenkinsAdapter) GetEntityMetadata(entityType string) *EntityMetadata {
    switch entityType {
    case "user":
        return &EntityMetadata{
            TypeName:   "user",
            DisplayName: "用户",
            Fields: []FieldDefinition{
                {FieldName: "username", DisplayName: "用户名", FieldType: "string", Required: true, Searchable: true, DisplayInList: true},
                {FieldName: "fullName", DisplayName: "全名", FieldType: "string", Required: false, Searchable: false, DisplayInList: true},
            },
        }
    case "project":
        return &EntityMetadata{
            TypeName:   "project",
            DisplayName: "项目",
            Fields: []FieldDefinition{
                {FieldName: "name", DisplayName: "项目名", FieldType: "string", Required: true, Searchable: true, DisplayInList: true},
                {FieldName: "url", DisplayName: "项目URL", FieldType: "string", Required: false, Searchable: false, DisplayInList: true},
            },
        }
    }
    // ... 其他实体类型
    return nil
}
```

### 服务层简化

```go
// 统一的实体服务
type ApplicationEntityService struct {
    adapterFactory *AdapterFactory
    db             *gorm.DB
}

// 同步实体（通用方法）
func (s *ApplicationEntityService) SyncEntities(appID uint, entityType string, operator string) error {
    app, err := s.GetApplicationByID(appID)
    if err != nil {
        return err
    }

    adapter, err := s.adapterFactory.GetAdapter(app.Type)
    if err != nil {
        return err
    }

    // 检查应用是否支持该实体类型
    supportedTypes := adapter.GetSupportedEntityTypes()
    if !contains(supportedTypes, entityType) {
        return fmt.Errorf("应用 %s 不支持实体类型 %s", app.Type, entityType)
    }

    // 获取实体数据
    entities, err := adapter.FetchEntities(app.BaseURL, authConfig, entityType)
    if err != nil {
        return err
    }

    // 删除旧数据
    s.db.Where("app_id = ? AND entity_type = ?", appID, entityType).Delete(&ApplicationEntity{})

    // 保存新数据
    for _, entity := range entities {
        entity.AppID = appID
        entity.EntityType = entityType
        s.db.Create(&entity)
    }

    return nil
}

// 获取实体列表（通用方法）
func (s *ApplicationEntityService) GetEntities(appID uint, entityType string) ([]ApplicationEntity, error) {
    var entities []ApplicationEntity
    err := s.db.Where("app_id = ? AND entity_type = ?", appID, entityType).Find(&entities).Error
    return entities, err
}

// 创建实体（通用方法）
func (s *ApplicationEntityService) CreateEntity(appID uint, entityType string, entity *ApplicationEntity) error {
    // ... 通用创建逻辑
}
```

### 前端适配

前端根据应用的元数据动态渲染界面：

```typescript
// 获取应用支持的实体类型
const entityTypes = await getSupportedEntityTypes(appId);

// 动态生成标签页
entityTypes.forEach(type => {
  const metadata = await getEntityMetadata(appId, type);

  tabs.push({
    label: metadata.displayName,
    key: metadata.typeName,
    columns: generateColumns(metadata.fields),  // 根据字段定义生成表格列
    syncSupported: metadata.syncSupported
  });
});
```

## 迁移策略

### 阶段1：添加新表（不破坏现有功能）
1. 创建 `application_entities` 和 `application_bindings` 表
2. 保留现有表：`application_users`、`application_groups`、`application_roles`
3. 新功能使用新表，旧功能保持不变

### 阶段2：适配器改造
1. 更新适配器接口，实现新方法
2. 添加元数据支持
3. 逐步迁移现有功能

### 阶段3：前端改造
1. 根据元数据动态生成界面
2. 统一实体和权限绑定操作

### 阶段4：清理（可选）
1. 数据迁移到新表
2. 删除旧的专门表
3. 完全切换到新架构

## 优势总结

✅ **表结构简洁**：2个通用表替代多个专门表
✅ **扩展性强**：添加新应用只需实现适配器，无需改表结构
✅ **代码复用**：统一的CRUD逻辑，减少重复代码
✅ **灵活适配**：通过JSON字段支持各种应用的特殊需求
✅ **元数据驱动**：前端可以根据元数据动态生成界面
✅ **向后兼容**：渐进式迁移，不影响现有功能

这个方案既解决了表结构膨胀问题，又保持了足够的灵活性来适配不同应用的权限模型。
