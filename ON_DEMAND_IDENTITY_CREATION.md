# 用户身份管理的按需创建方案

## 一、实际使用场景分析

### 典型场景
```
场景1：新员工入职
1. 创建用户"张三"
2. 暂时只需要Jenkins权限（参与开发项目）
3. 3个月后可能需要JumpServer权限（参与运维工作）
4. 不需要在入职时就创建所有应用账号

场景2：角色转换
1. 用户"李四"原来是开发人员，只需要Jenkins权限
2. 转岗为运维人员，需要增加JumpServer权限
3. 此时再创建JumpServer账号

场景3：项目需求
1. 用户"王五"参与项目A，需要GitLab项目权限
2. 项目A结束后，不再需要GitLab权限
3. 参与项目B时，重新激活GitLab权限
```

### 核心需求
```
✅ 用户创建时不需要立即同步到所有应用
✅ 根据实际需求，按需创建外部应用账号
✅ 支持随时添加或删除外部身份映射
✅ 权限分配时可以触发账号创建（这是合理的时机）
```

---

## 二、重新设计的方案

### 方案：按需创建 + 智能提示

#### 核心思想
```
1. 用户创建时不强制要求同步到外部应用
2. 权限分配时自动创建缺失的身份映射（这是合理的时机）
3. 提供灵活的身份管理工具
4. 清晰的权限分配状态反馈
```

---

## 三、具体的业务流程

### 流程1：用户创建（简化版）
```bash
POST /api/users
{
  "username": "zhangsan",
  "nickname": "张三",
  "email": "zhangsan@example.com"
  # 不强制要求 syncToApps
}

# 响应
{
  "success": true,
  "message": "用户创建成功",
  "user": {
    "id": 5,
    "username": "zhangsan",
    "externalIdentities": []  # 空的，还没有外部身份
  }
}
```

### 流程2：权限分配（智能处理）
```bash
POST /api/auth-groups/1/permissions
{
  "appId": 1,  # JumpServer
  "mappingType": "asset",
  "permissionDetail": {...}
}

# 系统处理逻辑：
# 1. 检查"运维组"成员的身份映射
# 2. 发现"张三"缺少JumpServer身份映射
# 3. 自动为"张三"创建JumpServer账号（合理时机）
# 4. 继续权限分配流程

# 响应
{
  "success": true,
  "message": "权限分配成功",
  "result": {
    "totalMembers": 5,
    "processedMembers": 5,
    "createdIdentities": [
      {
        "username": "zhangsan",
        "app": "JumpServer",
        "action": "created",
        "message": "自动创建JumpServer账号"
      }
    ],
    "appliedPermissions": 5
  }
}
```

### 流程3：手动添加外部身份
```bash
# 为用户手动添加外部应用账号
POST /api/users/5/identities
{
  "appId": 2,  # Jenkins
  "externalUsername": "zhangsan",
  "linkMethod": "create"  # create/link/import
}

# 响应
{
  "success": true,
  "message": "已为用户创建Jenkins账号",
  "identity": {
    "appId": 2,
    "externalUsername": "zhangsan",
    "mappingType": "manual",
    "mappingStatus": "active"
  }
}
```

### 流程4：查看用户的外部身份
```bash
GET /api/users/5/identities

# 响应
{
  "userId": 5,
  "username": "zhangsan",
  "identities": [
    {
      "appId": 1,
      "appName": "JumpServer",
      "externalUsername": "zhangsan",
      "mappingStatus": "active",
      "createdAt": "2026-01-15 10:30:00",
      "source": "auto_created_by_permission"  # 权限分配时自动创建
    },
    {
      "appId": 2,
      "appName": "Jenkins",
      "externalUsername": "zhangsan",
      "mappingStatus": "active",
      "createdAt": "2026-01-20 14:20:00",
      "source": "manual_created"  # 手动创建
    }
  ]
}
```

---

## 四、实现代码

### 权限分配服务（改进版）
```go
package services

import (
    "fmt"
    "oneops/backend/logger"
    "oneops/backend/models"
    "go.uber.org/zap"
)

type PermissionMappingService struct {
    db              *gorm.DB
    adapterFactory *AdapterFactory
    identityService *UserIdentityService
}

// AssignPermissionToAuthGroup 为用户组分配权限（智能处理身份映射）
func (s *PermissionMappingService) AssignPermissionToAuthGroup(
    authGroupID, appID uint,
    mappingType, externalID, externalName string,
    permissionDetail map[string]interface{},
    grantedBy string,
    autoCreateIdentities bool,  // 新增参数：是否自动创建身份映射
) (*PermissionAssignmentDetailedResult, error) {

    result := &PermissionAssignmentDetailedResult{
        AuthGroupID: authGroupID,
        AppID:       appID,
        MappingType: mappingType,
        ExternalID:  externalID,
    }

    // 1. 获取用户组成员
    members, err := s.getAuthGroupMembers(authGroupID)
    if err != nil {
        return nil, err
    }
    result.TotalMembers = len(members)

    // 2. 检查成员的身份映射状态
    var membersWithIdentity []models.AuthUser
    var membersWithoutIdentity []models.AuthUser

    for _, member := range members {
        mapping := s.identityService.GetUserIdentityMapping(member.ID, appID)
        if mapping != nil && mapping.MappingStatus == "active" {
            membersWithIdentity = append(membersWithIdentity, member)
        } else {
            membersWithoutIdentity = append(membersWithoutIdentity, member)
        }
    }

    result.MembersWithIdentity = len(membersWithIdentity)
    result.MembersWithoutIdentity = len(membersWithoutIdentity)

    // 3. 处理缺少身份映射的成员
    if len(membersWithoutIdentity) > 0 {
        if autoCreateIdentities {
            // 自动创建身份映射
            logger.Info("自动为成员创建外部身份",
                zap.Int("count", len(membersWithoutIdentity)),
                zap.Uint("appId", appID))

            for _, member := range membersWithoutIdentity {
                if err := s.identityService.CreateExternalIdentityForMember(member.ID, appID, grantedBy); err != nil {
                    logger.Warn("创建外部身份失败",
                        zap.Uint("userId", member.ID),
                        zap.String("username", member.Username),
                        zap.Uint("appId", appID),
                        zap.Error(err))

                    result.FailedIdentities = append(result.FailedIdentities, IdentityCreationResult{
                        UserID:   member.ID,
                        Username: member.Username,
                        Error:    err.Error(),
                    })
                } else {
                    logger.Info("成功创建外部身份",
                        zap.Uint("userId", member.ID),
                        zap.String("username", member.Username))

                    result.CreatedIdentities = append(result.CreatedIdentities, IdentityCreationResult{
                        UserID:   member.ID,
                        Username: member.Username,
                        Status:   "created",
                    })

                    // 添加到有身份的成员列表
                    membersWithIdentity = append(membersWithIdentity, member)
                }
            }
        } else {
            // 不自动创建，记录待处理成员
            for _, member := range membersWithoutIdentity {
                result.PendingMembers = append(result.PendingMembers, PendingMemberInfo{
                    UserID:   member.ID,
                    Username: member.Username,
                    Reason:   "缺少外部应用身份映射",
                })
            }
        }
    }

    // 4. 为有身份的成员应用权限
    for _, member := range membersWithIdentity {
        if err := s.applyPermissionToMember(member.ID, appID, permissionDetail); err != nil {
            logger.Warn("应用权限失败",
                zap.Uint("userId", member.ID),
                zap.String("username", member.Username),
                zap.Error(err))
        } else {
            result.AppliedMembers++
        }
    }

    // 5. 创建权限映射记录
    permissionMapping := &models.AuthGroupPermissionMapping{
        AuthGroupID:      authGroupID,
        AppID:            appID,
        MappingType:      mappingType,
        ExternalID:       externalID,
        ExternalName:     externalName,
        PermissionDetail: permissionDetailJSON,
        IsEnabled:        true,
        GrantedBy:        grantedBy,
        GrantedAt:        time.Now(),
    }

    if err := s.db.Create(permissionMapping).Error; err != nil {
        return nil, err
    }

    result.MappingID = permissionMapping.ID

    // 6. 生成处理结果摘要
    if len(result.PendingMembers) > 0 {
        result.Message = fmt.Sprintf(
            "权限映射已创建，但有 %d 个成员缺少身份映射。已为 %d 个成员应用权限。",
            len(result.PendingMembers),
            result.AppliedMembers,
        )
        result.Status = "partial_success"
    } else if len(result.FailedIdentities) > 0 {
        result.Message = fmt.Sprintf(
            "权限分配部分成功。已创建 %d 个身份映射，为 %d 个成员应用权限，%d 个失败。",
            len(result.CreatedIdentities),
            result.AppliedMembers,
            len(result.FailedIdentities),
        )
        result.Status = "partial_success"
    } else {
        result.Message = fmt.Sprintf(
            "权限分配成功。已创建 %d 个身份映射，为 %d 个成员应用权限。",
            len(result.CreatedIdentities),
            result.AppliedMembers,
        )
        result.Status = "success"
    }

    return result, nil
}

// 权限分配详细结果
type PermissionAssignmentDetailedResult struct {
    MappingID              uint                   `json:"mappingId"`
    AuthGroupID            uint                   `json:"authGroupId"`
    AppID                  uint                   `json:"appId"`
    MappingType            string                 `json:"mappingType"`
    ExternalID             string                 `json:"externalId"`

    // 统计信息
    TotalMembers           int                    `json:"totalMembers"`
    MembersWithIdentity    int                    `json:"membersWithIdentity"`
    MembersWithoutIdentity int                    `json:"membersWithoutIdentity"`
    AppliedMembers         int                    `json:"appliedMembers"`

    // 详细结果
    CreatedIdentities      []IdentityCreationResult `json:"createdIdentities"`
    FailedIdentities       []IdentityCreationResult `json:"failedIdentities"`
    PendingMembers         []PendingMemberInfo      `json:"pendingMembers"`

    // 处理状态
    Status                 string                 `json:"status"`  // success, partial_success, failed
    Message                string                 `json:"message"`
}

type IdentityCreationResult struct {
    UserID   uint   `json:"userId"`
    Username string `json:"username"`
    Status   string `json:"status"`  // created, failed
    Error    string `json:"error,omitempty"`
}

type PendingMemberInfo struct {
    UserID   uint   `json:"userId"`
    Username string `json:"username"`
    Reason   string `json:"reason"`
}
```

### 用户身份服务（按需创建）
```go
package services

import (
    "oneops/backend/models"
    "gorm.io/gorm"
)

type UserIdentityService struct {
    db              *gorm.DB
    adapterFactory *AdapterFactory
}

// CreateExternalIdentityForMember 为成员创建外部身份（按需）
func (s *UserIdentityService) CreateExternalIdentityForMember(authUserID, appID uint, createdBy string) error {
    // 1. 检查是否已存在映射
    var existingMapping models.UserIdentityMapping
    err := s.db.Where("auth_user_id = ? AND app_id = ?", authUserID, appID).First(&existingMapping).Error
    if err == nil {
        // 映射已存在，更新为active状态
        existingMapping.MappingStatus = "active"
        existingMapping.DeletedAt = nil
        s.db.Save(&existingMapping)
        return nil
    }

    // 2. 获取授权中心用户信息
    var authUser models.AuthUser
    if err := s.db.First(&authUser, authUserID).Error; err != nil {
        return fmt.Errorf("用户不存在: %w", err)
    }

    // 3. 获取应用信息
    var app models.Application
    if err := s.db.First(&app, appID).Error; err != nil {
        return fmt.Errorf("应用不存在: %w", err)
    }

    // 4. 调用适配器创建外部用户
    adapter, err := s.adapterFactory.GetAdapter(app.Type)
    if err != nil {
        return fmt.Errorf("获取应用适配器失败: %w", err)
    }

    // 解析认证配置
    var authConfig map[string]interface{}
    if err := json.Unmarshal([]byte(app.AuthConfig), &authConfig); err != nil {
        return fmt.Errorf("解析认证配置失败: %w", err)
    }

    // 解析端点配置
    if app.Endpoints != "" && app.Endpoints != "null" {
        var endpoints map[string]interface{}
        if err := json.Unmarshal([]byte(app.Endpoints), &endpoints); err == nil {
            authConfig["endpoints"] = endpoints
        }
    }

    // 检查适配器是否支持创建用户
    type UserCreator interface {
        CreateUser(baseURL string, authConfig map[string]interface{}, user *UserCreateRequest) error
    }

    userCreator, ok := adapter.(UserCreator)
    if !ok {
        return fmt.Errorf("应用类型 %s 不支持自动创建用户", app.Type)
    }

    // 生成随机密码
    randomPassword := generateRandomPassword()

    // 构建创建请求
    createReq := &UserCreateRequest{
        Username:  authUser.Username,
        FullName:  authUser.Nickname,
        Email:     authUser.Email,
        Password:  randomPassword,
    }

    // 创建外部用户
    if err := userCreator.CreateUser(app.BaseURL, authConfig, createReq); err != nil {
        return fmt.Errorf("创建外部用户失败: %w", err)
    }

    // 5. 创建身份映射记录
    mapping := &models.UserIdentityMapping{
        AuthUserID:       authUserID,
        AppID:            appID,
        ExternalUsername: authUser.Username,
        ExternalUserID:   authUser.Username,  // 假设外部系统返回的用户ID就是username
        MappingType:      "auto",
        MappingStatus:    "active",
        SyncMethod:       "api",
        CreatedBy:        createdBy,
    }

    if err := s.db.Create(mapping).Error; err != nil {
        return fmt.Errorf("创建身份映射记录失败: %w", err)
    }

    logger.Info("成功为用户创建外部身份",
        zap.Uint("authUserId", authUserID),
        zap.String("username", authUser.Username),
        zap.Uint("appId", appID),
        zap.String("appName", app.Name))

    return nil
}

// GetUserExternalIdentities 获取用户的所有外部身份
func (s *UserIdentityService) GetUserExternalIdentities(authUserID uint) ([]models.UserIdentityMapping, error) {
    var identities []models.UserIdentityMapping

    err := s.db.Where("auth_user_id = ? AND deleted_at IS NULL", authUserID).
        Preload("AppIDField").
        Find(&identities).Error

    return identities, err
}

// GetUserIdentityMapping 获取用户在特定应用的身份映射
func (s *UserIdentityService) GetUserIdentityMapping(authUserID, appID uint) *models.UserIdentityMapping {
    var mapping models.UserIdentityMapping
    err := s.db.Where("auth_user_id = ? AND app_id = ? AND mapping_status = ? AND deleted_at IS NULL",
        authUserID, appID, "active").First(&mapping).Error

    if err != nil {
        return nil
    }
    return &mapping
}

// DeleteUserIdentity 删除用户的外部身份
func (s *UserIdentityService) DeleteUserIdentity(authUserID, appID uint, cleanupStrategy string) error {
    var mapping models.UserIdentityMapping
    if err := s.db.Where("auth_user_id = ? AND app_id = ?", authUserID, appID).First(&mapping).Error; err != nil {
        return err
    }

    switch cleanupStrategy {
    case "delete":
        // 删除外部账号
        s.deleteExternalAccount(&mapping)

    case "disable":
        // 禁用外部账号
        s.disableExternalAccount(&mapping)

    case "unmap":
        // 只删除映射，保留外部账号
        mapping.DeletedAt = time.Now()
        return s.db.Delete(&mapping).Error
    }

    // 删除映射记录
    return s.db.Delete(&mapping).Error
}
```

---

## 五、API接口设计

### 权限分配API（智能版）
```bash
# 自动创建身份映射
POST /api/auth-groups/1/permissions
{
  "appId": 1,
  "mappingType": "asset",
  "externalId": "all-assets",
  "externalName": "所有资产",
  "permissionDetail": {...},
  "autoCreateIdentities": true  # 自动创建缺失的身份映射
}

# 不自动创建，只记录
POST /api/auth-groups/1/permissions
{
  "appId": 1,
  "mappingType": "asset",
  "externalId": "all-assets",
  "externalName": "所有资产",
  "permissionDetail": {...},
  "autoCreateIdentities": false  # 不自动创建
}
```

### 用户身份管理API
```bash
# 获取用户的外部身份列表
GET /api/users/5/identities

# 手动添加外部身份
POST /api/users/5/identities
{
  "appId": 2,
  "externalUsername": "zhangsan",
  "linkMethod": "create"  # create/link/import
}

# 删除外部身份
DELETE /api/users/5/identities/2
{
  "cleanupStrategy": "disable"  # delete/disable/unmap
}

# 批量处理待处理的身份映射
POST /api/permission-mappings/{mappingId}/process-pending
```

---

## 六、前端交互优化

### 权限分配界面（智能版）
```vue
<template>
  <div class="permission-assignment">
    <el-form>
      <el-form-item label="权限详情">
        <!-- 权限选择 -->
      </el-form-item>

      <!-- 新增：自动创建身份映射选项 -->
      <el-form-item>
        <el-checkbox v-model="autoCreateIdentities">
          自动为缺少身份映射的成员创建外部账号
        </el-checkbox>
        <el-tooltip>
          启用后，系统会自动为没有外部账号的成员创建账号。
          如果禁用，权限只分配给已有账号的成员。
        </el-tooltip>
      </el-form-item>

      <el-button @click="assignPermission">分配权限</el-button>
    </el-form>

    <!-- 分配结果展示 -->
    <div v-if="assignmentResult" class="assignment-result">
      <el-alert
        :type="getResultType()"
        :title="assignmentResult.message"
      >
        <template #default>
          <div v-if="assignmentResult.createdIdentities.length > 0">
            <h4>已创建的外部账号：</h4>
            <ul>
              <li v-for="item in assignmentResult.createdIdentities">
                {{ item.username }} - {{ item.status }}
              </li>
            </ul>
          </div>

          <div v-if="assignmentResult.pendingMembers.length > 0">
            <h4>待处理成员：</h4>
            <ul>
              <li v-for="item in assignmentResult.pendingMembers">
                {{ item.username }} - {{ item.reason }}
              </li>
            </ul>
            <el-button size="small" @click="createPendingIdentities">
              为待处理成员创建账号
            </el-button>
          </div>
        </template>
      </el-alert>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      autoCreateIdentities: true,  // 默认启用
      assignmentResult: null
    }
  },

  methods: {
    async assignPermission() {
      const result = await this.$api.assignPermissionToAuthGroup({
        autoCreateIdentities: this.autoCreateIdentities,
        // ... 其他参数
      })

      this.assignmentResult = result

      this.$message({
        type: this.getResultType(),
        message: result.message
      })
    },

    getResultType() {
      if (!this.assignmentResult) return 'info'
      switch (this.assignmentResult.status) {
        case 'success': return 'success'
        case 'partial_success': return 'warning'
        case 'failed': return 'error'
        default: return 'info'
      }
    }
  }
}
</script>
```

---

## 七、优势总结

### 与之前的方案对比

| 维度 | 之前的严格模式 | 按需创建方案 |
|------|----------------|--------------|
| **用户创建** | 必须同步到所有应用 | 按需创建，灵活选择 |
| **权限分配** | 缺少身份时报错 | 自动创建或记录待处理 |
| **用户体验** | 复杂，需要提前规划 | 简单，按需添加 |
| **管理成本** | 高，提前创建可能用不到的账号 | 低，只在需要时创建 |
| **灵活性** | 低，创建时必须确定 | 高，随时可以添加 |
| **合理性** | 不符合实际使用 | 符合实际业务场景 |

### 核心优势

1. **✅ 符合实际业务场景**
   - 用户不需要所有应用权限时，不强制创建账号
   - 根据实际需求，按需创建外部应用账号

2. **✅ 权限分配是合理的触发时机**
   - 权限分配时发现缺少身份映射，自动创建是合理的
   - 这正是用户真正需要外部应用的时机

3. **✅ 提供灵活的管理选项**
   - 可以选择自动创建或手动处理
   - 支持批量处理待处理的身份映射

4. **✅ 清晰的状态反馈**
   - 详细的结果信息（创建了哪些账号，哪些成员待处理）
   - 管理员可以清楚地了解处理情况

5. **✅ 降低管理成本**
   - 不提前创建可能用不到的账号
   - 只在真正需要时创建，提高资源利用率

---

## 八、结论

**按需创建方案是最佳选择**，因为：

1. **✅ 符合实际业务场景**：用户权限需求是动态的，不是静态的
2. **✅ 权限分配是合理时机**：需要权限时创建账号，逻辑清晰
3. **✅ 提供灵活选项**：支持自动和手动两种模式
4. **✅ 降低管理成本**：不提前创建无用账号
5. **✅ 用户体验友好**：简化操作，提供清晰反馈

这样就完全解决了逻辑矛盾问题，同时提供了更灵活和实用的解决方案！
