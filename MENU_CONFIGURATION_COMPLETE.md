# 菜单配置完成说明

## 完成时间
2026-07-24

## 执行的操作

### 1. 数据库菜单添加

已成功在数据库 `menus` 表中添加两个新菜单：

```sql
-- 用户身份映射菜单 (ID: 106)
INSERT INTO menus (name, icon, path, parent_id, sort, status, created_at, updated_at)
VALUES ('用户身份映射', 'mdi:account-switch', '/auth/user-identities', 7, 7, 1, NOW(), NOW());

-- 用户有效权限菜单 (ID: 107)
INSERT INTO menus (name, icon, path, parent_id, sort, status, created_at, updated_at)
VALUES ('用户有效权限', 'mdi:shield-check', '/auth/user-permissions', 7, 8, 1, NOW(), NOW());
```

### 2. 菜单列表（授权中心）

现在授权中心的完整菜单结构：

| ID  | 菜单名称     | 图标                  | 路径                     | 排序 |
|-----|-------------|-----------------------|--------------------------|------|
| 100 | 用户         | mdi:account           | /auth/users              | 1    |
| 101 | 用户组       | mdi:shield-account    | /auth/roles              | 2    |
| 102 | 应用         | mdi:application       | /auth/applications       | 3    |
| 103 | 权限映射     | mdi:link              | /auth/rolebindings       | 4    |
| 104 | 用户授权     | mdi:account-key       | /auth/userauthorization  | 5    |
| 105 | 操作日志     | mdi:file-document     | /auth/operationlogs      | 6    |
| 106 | 用户身份映射 | mdi:account-switch    | /auth/user-identities    | 7    | ⭐ NEW
| 107 | 用户有效权限 | mdi:shield-check      | /auth/user-permissions   | 8    | ⭐ NEW

### 3. 角色权限更新

已为超级管理员角色添加新菜单权限：

- 更新前：50个菜单
- 更新后：52个菜单
- 新增菜单ID：106, 107

```sql
UPDATE roles
SET menu_ids = JSON_ARRAY_APPEND(menu_ids, '$', 106, '$', 107)
WHERE id = 1;
```

### 4. 前端路由配置

前端路由已自动识别新页面：

**imports.ts 中的配置：**
```typescript
"auth_user-identities": () => import("@/views/auth/user-identities/index.vue"),
"auth_user-permissions": () => import("@/views/auth/user-permissions/index.vue"),
```

### 5. 国际化配置

已完成中英文翻译：

**中文：**
- 用户身份映射
- 用户有效权限

**英文：**
- User Identity Mappings
- User Effective Permissions

## 验证步骤

### 查看菜单配置
```bash
mysql -h 60.191.116.75 -P 38089 -u root -p123456 nexops \
  -e "SELECT id, name, icon, path, parent_id, sort FROM menus WHERE parent_id = 7 ORDER BY sort;"
```

### 查看角色权限
```bash
mysql -h 60.191.116.75 -P 38089 -u root -p123456 nexops \
  -e "SELECT JSON_CONTAINS(menu_ids, '106') as has_menu_106, JSON_CONTAINS(menu_ids, '107') as has_menu_107 FROM roles WHERE id = 1;"
```

## 如何访问新菜单

1. 登录系统（使用超级管理员账号）
2. 在左侧菜单找到"授权中心"
3. 展开后可以看到新的菜单项：
   - 用户身份映射
   - 用户有效权限

## 可能需要的操作

如果菜单仍然没有显示，请尝试：

1. **清除浏览器缓存**：
   - 按 `Ctrl + Shift + Delete` 清除缓存
   - 或使用无痕模式重新登录

2. **重新登录**：
   - 退出当前登录
   - 重新登录刷新菜单权限

3. **检查后端服务**：
   ```bash
   # 确保后端服务正在运行
   curl http://localhost:8082/api/v1/system/user-identity-mappings
   ```

## SQL脚本位置

完整的菜单添加脚本已保存在：
```
/Users/mpm/Desktop/doc/go/OneOpsV2/OneOps/backend/scripts/add_user_identity_menus.sql
```

## 注意事项

1. **菜单命名规范**：
   - 数据库中的 `name` 字段使用中文显示名称
   - 路由中的 `name` 使用英文标识（如 `auth_user-identities`）

2. **权限控制**：
   - 只有超级管理员默认拥有新菜单权限
   - 其他角色需要单独配置菜单权限

3. **菜单排序**：
   - `sort` 字段控制菜单显示顺序
   - 数值越小，显示越靠前

## 下一步

如果需要为其他角色添加菜单权限：

```sql
-- 查看角色
SELECT id, name FROM roles;

-- 为特定角色添加菜单权限（假设角色ID为2）
UPDATE roles
SET menu_ids = JSON_ARRAY_APPEND(menu_ids, '$', 106, '$', 107)
WHERE id = 2;
```

## 故障排查

如果菜单不显示，检查以下几点：

1. **数据库菜单是否存在**：
   ```sql
   SELECT * FROM menus WHERE path IN ('/auth/user-identities', '/auth/user-permissions');
   ```

2. **用户角色是否有权限**：
   ```sql
   SELECT r.name, JSON_CONTAINS(r.menu_ids, '106') as has_menu_106
   FROM roles r
   JOIN user_roles ur ON r.id = ur.role_id
   WHERE ur.user_id = [当前用户ID];
   ```

3. **前端路由文件是否正确**：
   - 检查 `frontend/src/router/elegant/imports.ts`
   - 确认包含 `"auth_user-identities"` 和 `"auth_user-permissions"`

## 总结

✅ 菜单已添加到数据库
✅ 超级管理员权限已更新
✅ 前端路由已配置
✅ 国际化已完成

现在刷新页面或重新登录即可看到新菜单！
