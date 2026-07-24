# 权限映射设计优化方案

## 一、当前架构分析

### 现有模型
```
授权中心侧（OneOps内部）:
├── AuthUser         - 授权中心用户
├── AuthGroup        - 授权中心用户组
└── AuthUserGroup    - 用户-组关联

外部应用侧（从外部同步）:
├── Application              - 应用注册信息
├── ApplicationUser         - 同步的外部用户
├── ApplicationGroup        - 同步的外部用户组
├── ApplicationRole         - 同步的外部角色
└── ApplicationAuthorizationRule - 同步的授权规则（JumpServer）
```

### 核心问题

#### 1. **单向数据流缺陷**
```
❌ 当前：外部应用 → 授权中心（只读同步）
✅ 需要：授权中心 ←→ 外部应用（双向交互）
```

#### 2. **缺少权限映射层**
- 授权中心用户组无法直接分配外部应用权限
- 没有映射表关联：`AuthGroup` → `Application` 的权限关系
- 无法实现"授权中心的一个用户组 → 多个外部应用权限"

#### 3. **不同系统权限模型差异**
```
JumpServer: 用户组 + 授权规则(用户/组 → 资产/节点 → 操作)
Jenkins:    用户 + 角色/权限
GitLab:     用户 + 组 + 访问级别(Project/Group)
```

#### 4. **数据冗余问题**
- 同步大量用户/角色数据，但实际用于权限映射的价值有限
- 占用存储空间，同步维护成本高

---

## 二、优化方案设计

### 方案A：权限映射表架构（推荐）

#### 核心设计理念
```
授权中心用户组 ← 映射层 → 外部应用权限对象

职责分离：
- 授权中心管理：用户、用户组
- 映射层管理：权限映射关系
- 外部应用管理：实际权限执行
```

#### 数据库表设计

##### 1. 权限映射表（新增）
```go
// AuthGroupPermissionMapping 授权中心用户组权限映射
type AuthGroupPermissionMapping struct {
    ID          uint      `json:"id" gorm:"primaryKey"`

    // 授权中心侧
    AuthGroupID uint      `json:"authGroupId" gorm:"not null;index:idx_auth_app"`
    AuthGroupIDField AuthGroup `json:"authGroup,omitempty" gorm:"foreignKey:AuthGroupID"`

    // 外部应用侧
    AppID       uint      `json:"appId" gorm:"not null;index:idx_auth_app;index:idx_app"`
    AppIDField  Application `json:"-" gorm:"foreignKey:AppID"`

    // 映射对象（支持多种类型）
    MappingType string    `json:"mappingType" gorm:"size:20;not null"` // role, rule, asset, node, project
    ExternalID  string    `json:"externalId" gorm:"size:100"`            // 外部系统中的ID
    ExternalName string   `json:"externalName" gorm:"size:200"`         // 外部系统中的名称（用于显示）

    // 权限详情（JSON格式，灵活支持不同系统）
    PermissionDetail string `json:"permissionDetail" gorm:"type:json"` // {"actions":["connect","upload"], "assets":["server-1"]}

    // 状态和管理
    IsEnabled   bool      `json:"isEnabled" gorm:"default:true"`
    Priority    int       `json:"priority" gorm:"default:0"`
    GrantedBy   string    `json:"grantedBy" gorm:"size:50"`
    GrantedAt   time.Time `json:"grantedAt"`
    ExpireTime  *time.Time `json:"expireTime"`

    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}
```

##### 2. 外部系统权限模板表（新增）
```go
// ApplicationPermissionTemplate 外部应用权限定义模板
type ApplicationPermissionTemplate struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    AppID       uint      `json:"appId" gorm:"not null;index"`
    AppIDField  Application `json:"-" gorm:"foreignKey:AppID"`

    // 权限定义
    PermissionCode   string `json:"permissionCode" gorm:"size:100;not null"`   // admin, developer, reader
    PermissionName   string `json:"permissionName" gorm:"size:100;not null"`
    PermissionType   string `json:"permissionType" gorm:"size:50;not null"`     // role, rule, preset
    Description      string `json:"description" gorm:"size:200"`

    // 权限内容模板（JSON，不同系统有不同结构）
    Template        string `json:"template" gorm:"type:json"` // JumpServer: {"actions":["connect"],"assets":["all"]}

    IsSystemPreset  bool   `json:"isSystemPreset" gorm:"default:false"` // 是否系统预设

    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}
```

#### 业务流程设计

##### 流程1：权限分配流程
```
1. 管理员在授权中心创建用户组（如"运维组"）
2. 选择要分配权限的外部应用（如 JumpServer）
3. 从应用的权限模板中选择或自定义权限
4. 系统自动调用外部应用API执行实际授权
5. 保存映射关系到权限映射表
```

##### 流程2：权限同步与检查
```
定时任务：
1. 读取所有启用的权限映射
2. 调用外部API检查权限状态
3. 检测到差异时记录告警或自动修复
4. 审计日志记录所有变更
```

##### 流程3：用户权限继承
```
当用户被添加到用户组时：
1. 查询该用户组的所有权限映射
2. 实时计算用户的有效权限列表
3. 返回给前端展示（该用户在哪些应用有哪些权限）
```

---

### 方案B：轻量级映射架构

如果不想大规模改造，可以采用简化版本：

#### 简化映射表
```go
// AuthGroupAppPermission 授权中心用户组-应用权限映射（简化版）
type AuthGroupAppPermission struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    AuthGroupID uint      `json:"authGroupId" gorm:"not null;index:idx_group_app"`
    AppID       uint      `json:"appId" gorm:"not null;index:idx_group_app"`

    // 直接存储外部系统的权限对象ID
    ExternalObjectType string `json:"externalObjectType"` // role, rule
    ExternalObjectID   string `json:"externalObjectId"`

    // 备注
    Description string `json:"description"`

    CreatedAt   time.Time `json:"createdAt"`
}
```

---

## 三、具体优化建议

### 推荐实施路径

#### 阶段1：基础映射能力（2-3天）
1. 创建 `auth_group_permission_mappings` 表
2. 实现基础API：
   - `POST /api/auth-groups/{groupId}/permissions` - 分配权限
   - `GET /api/auth-groups/{groupId}/permissions` - 查看权限
   - `DELETE /api/auth-groups/{groupId}/permissions/{id}` - 撤销权限
3. 实现权限继承查询API

#### 阶段2：外部系统集成（3-5天）
1. 扩展适配器接口，添加权限操作方法：
   ```go
   type ApplicationAdapter interface {
       // 现有方法...
       FetchUsers() ...
       FetchRoles() ...

       // 新增权限操作方法
       AssignPermissionToGroup(groupID, permission) error
       RevokePermissionFromGroup(groupID, permission) error
       CheckGroupPermission(groupID, permission) (bool, error)
   }
   ```
2. 为每个适配器实现具体的权限操作逻辑
3. 在分配权限时自动调用外部API

#### 阶段3：权限模板与可视化（3-4天）
1. 创建应用权限模板管理界面
2. 提供可视化权限选择器（替代手动输入ID）
3. 权限审计与监控面板

---

## 四、数据对比

### 优化前 vs 优化后

| 维度 | 优化前 | 优化后 |
|------|--------|--------|
| **数据流向** | 单向同步（外部→内部） | 双向交互（内部→外部） |
| **权限分配** | 无法操作 | 可直接分配 |
| **权限可见性** | 需要登录外部系统查看 | 统一视图查看所有权限 |
| **维护成本** | 同步数据维护成本高 | 映射关系维护成本低 |
| **存储空间** | 大量冗余数据 | 只存必要映射 |
| **权限一致性** | 难以保证 | 自动检查与修复 |
| **扩展性** | 新增系统需同步大量数据 | 新增系统只需定义权限模板 |

---

## 五、架构图

### 优化后的权限映射架构

```
┌─────────────────────────────────────────────────────────────┐
│                        OneOps 授权中心                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │  AuthUser    │  │  AuthGroup   │  │ PermissionMapping│  │
│  │  (用户管理)   │  │  (用户组)    │  │   (权限映射)      │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                            ↓ 映射关系
┌─────────────────────────────────────────────────────────────┐
│                      外部应用集成层                           │
│  ┌───────────────┐  ┌───────────────┐  ┌──────────────────┐│
│  │ JumpServer    │  │    Jenkins    │  │     GitLab       ││
│  │ Adapter       │  │    Adapter    │  │     Adapter      ││
│  │ - 用户组授权   │  │ - 角色分配     │  │ - 项目访问权限    ││
│  │ - 资产权限     │  │ - Job权限     │  │ - 仓库访问        ││
│  └───────────────┘  └───────────────┘  └──────────────────┘│
└─────────────────────────────────────────────────────────────┘
                            ↓ API调用
┌─────────────────────────────────────────────────────────────┐
│                        实际外部系统                           │
│  ┌───────────────┐  ┌───────────────┐  ┌──────────────────┐│
│  │ JumpServer    │  │    Jenkins    │  │     GitLab       ││
│  │ - 授权规则API  │  │ - 角色管理API  │  │ - 成员权限API     ││
│  │ - 资产管理API  │  │ - Job配置API  │  │ - 项目设置API     ││
│  └───────────────┘  └───────────────┘  └──────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

---

## 六、优势总结

### 核心价值

1. **统一权限管理**
   - 在一个地方管理所有外部应用的权限
   - 用户组权限可视化，一目了然

2. **降低维护成本**
   - 不需要同步大量用户/角色数据
   - 只维护必要的映射关系

3. **提高安全性**
   - 权限变更自动同步到外部系统
   - 定期审计，确保权限一致性

4. **扩展性强**
   - 新增外部应用只需实现适配器接口
   - 权限模板机制，快速适配不同系统

5. **用户体验好**
   - 管理员不需要登录多个系统
   - 用户可以查看自己在所有系统的权限

### 适用场景

✅ **强烈推荐使用此方案的场景**：
- 有多个外部应用需要统一管理权限
- 外部应用支持API操作权限
- 需要定期审计权限分配情况
- 用户组相对固定，权限分配频繁

⚠️ **不太适合的场景**：
- 外部应用不支持API操作权限
- 权限分配极其复杂，无法模板化
- 外部应用权限模型与授权中心差异过大

---

## 七、下一步行动建议

### 立即可做的优化（不依赖外部API）

1. **创建权限映射表**，手动维护映射关系
2. **实现权限查询API**，展示用户在所有应用的权限
3. **前端权限展示页面**，可视化权限矩阵

### 中期优化（需要外部API支持）

1. **扩展适配器接口**，添加权限操作方法
2. **实现自动权限分配/撤销**
3. **权限模板管理**

### 长期优化

1. **权限生命周期管理**（临时权限、自动过期）
2. **权限审批流程**
3. **权限智能推荐**（基于用户行为）

---

## 结论

**建议采用方案A（权限映射表架构）**，原因是：

1. ✅ **符合实际需求**：将授权中心用户组分配外部系统权限
2. ✅ **扩展性好**：支持多种外部应用，易于扩展
3. ✅ **维护成本低**：只维护映射关系，不冗余存储
4. ✅ **用户体验好**：统一视图管理，权限清晰可见
5. ✅ **安全性高**：自动同步，审计完善

**核心思想转变**：
- 从"同步外部数据到内部"转向"从内部控制外部权限"
- 从"数据同步"转向"权限映射"
- 从"被动查看"转向"主动管理"
