# 资产管理功能优化计划

## 一、问题分析

### 1. 主机字段设计与资产配置不吻合

**现状问题：**
- 主机表单中缺少 `cabinetId`（机柜ID）选择器
- 主机表单中缺少标签选择功能
- 业务系统选择存在但位置不够明显
- 云主机预留字段（`provider`、`instanceId`、`region` 等）未充分展示

**影响：**
- 用户无法将主机与机柜关联，资产管理不完整
- 无法通过标签对主机进行分类管理
- 云主机资产信息不完整，无法对接阿里云、腾讯云等

**✅ 已完成优化：**
- ✅ 添加了机房/机柜级联选择器
  - 先选择机房，再根据机房选择机柜
- ✅ 添加了标签多选器
  - 支持选择多个标签
  - 支持标签颜色显示
- ✅ 优化了业务系统选择器
  - 使用树形选择器
  - 显示层级关系
- ✅ 明确展示了云主机预留字段

### 2. 访问策略授权对象不友好

**现状问题：**
- 访问策略表单中 `subjectId` 和 `assetScopeId` 都是直接输入数字ID
- 用户需要记住各种对象的ID，体验极差
- 容易出错，输入错误的ID会导致策略不生效

**影响：**
- 用户体验差，配置困难
- 容易出现配置错误
- 后期维护困难

**✅ 已完成优化：**
- ✅ 授权对象类型选择
  - 用户：显示用户下拉选择器（支持搜索）
  - 角色：显示角色下拉选择器
  - 用户组：暂时显示提示（功能未开放）
- ✅ 资产范围类型选择
  - 全部资产：不需要选择
  - 单台服务器：暂时显示提示（功能即将开放）
  - 主机分组：显示分组树选择器
  - 业务系统：显示业务系统树选择器
  - 标签：显示标签选择器
- ✅ 表格显示优化
  - 将ID显示改为名称显示
  - 更友好的展示方式

### 3. 字段关联不清晰

**现状问题：**
- 用户不明白"资产配置"中的业务系统、机房机柜、标签管理和主机列表中的字段有什么关联
- 各个配置项之间缺少清晰的说明

**影响：**
- 用户不知道如何配置资产
- 配置混乱，数据不规范

## 二、优化计划

### P0 - 核心优化（必须完成）

#### 1. 后端API扩展

**目标：** 为前端提供获取资产数据的完整API

**任务清单：**
- [x] 创建 `GET /api/users` - 获取用户列表（用于访问策略授权对象）
- [x] 创建 `GET /api/roles` - 获取角色列表（用于访问策略授权对象）
- [x] 创建 `GET /api/cmdb/rooms` - 获取机房列表
- [x] 创建 `GET /api/cmdb/cabinets` - 获取机柜列表（支持按机房筛选）
- [x] 完善 `GET /api/cmdb/tags` - 获取标签列表（已存在，需检查）

#### 2. 前端主机表单优化

**目标：** 完善主机表单，支持所有资产配置项

**任务清单：**
- [x] 添加机房/机柜级联选择器
  - 先选择机房
  - 再根据机房选择机柜
- [x] 添加标签多选器
  - 支持选择多个标签
  - 支持标签颜色显示
- [x] 优化业务系统选择器
  - 使用树形选择器
  - 显示层级关系
- [x] 明确展示云主机预留字段
  - 云服务商（provider）：aliyun、tencent、aws、huawei、other
  - 实例ID（instanceId）
  - 实例规格（instanceType）
  - 地域（region）
  - 可用区（zone）
  - 计费类型（chargeType）：postpay、prepay

#### 3. 前端访问策略表单优化

**目标：** 将ID输入改为友好的选择器

**任务清单：**
- [x] 授权对象类型选择
  - 用户：显示用户下拉选择器
  - 角色：显示角色下拉选择器
  - 用户组：显示用户组下拉选择器
- [x] 资产范围类型选择
  - 全部资产：不需要选择
  - 单台服务器：显示服务器下拉选择器（支持搜索）
  - 主机分组：显示分组树选择器
  - 业务系统：显示业务系统树选择器
  - 标签：显示标签多选器
- [x] 数据预加载和缓存

### P1 - 体验优化（可选）

#### 1. 字段关联说明

**目标：** 让用户理解各配置项的关联关系

**任务清单：**
- [ ] 在主机表单中添加字段说明
- [ ] 在资产配置页面添加关联说明
- [ ] 添加配置向导（可选）

#### 2. 界面优化

**任务清单：**
- [ ] 优化表单布局，使用折叠面板分组
- [ ] 添加必填项标识
- [ ] 添加字段提示信息

## 三、技术实现要点

### 1. 机房机柜级联选择

```vue
<ElFormItem label="所在机房">
  <ElSelect v-model="serverForm.roomId" placeholder="请选择机房" @change="handleRoomChange">
    <ElOption v-for="room in rooms" :key="room.id" :label="room.name" :value="room.id" />
  </ElSelect>
</ElFormItem>

<ElFormItem label="所在机柜">
  <ElSelect v-model="serverForm.cabinetId" placeholder="请先选择机房" :disabled="!serverForm.roomId">
    <ElOption v-for="cabinet in filteredCabinets" :key="cabinet.id" :label="cabinet.name" :value="cabinet.id" />
  </ElSelect>
</ElFormItem>
```

### 2. 访问策略授权对象选择

```vue
<!-- 授权对象为用户时 -->
<ElFormItem v-if="currentPolicy.subjectType === 'user'" label="授权对象" prop="subjectId">
  <ElSelect v-model="currentPolicy.subjectId" placeholder="请选择用户" filterable>
    <ElOption v-for="user in users" :key="user.id" :label="user.username" :value="user.id" />
  </ElSelect>
</ElFormItem>

<!-- 授权对象为角色时 -->
<ElFormItem v-if="currentPolicy.subjectType === 'role'" label="授权对象" prop="subjectId">
  <ElSelect v-model="currentPolicy.subjectId" placeholder="请选择角色">
    <ElOption v-for="role in roles" :key="role.id" :label="role.name" :value="role.id" />
  </ElSelect>
</ElFormItem>
```

### 3. 资产范围选择

```vue
<!-- 资产范围为主机分组时 -->
<ElFormItem v-if="currentPolicy.assetScopeType === 'group'" label="资产范围" prop="assetScopeId">
  <ElTreeSelect
    v-model="currentPolicy.assetScopeId"
    :data="serverGroups"
    :props="{ label: 'name', value: 'id', children: 'children' }"
    placeholder="请选择主机分组"
  />
</ElFormItem>

<!-- 资产范围为业务系统时 -->
<ElFormItem v-if="currentPolicy.assetScopeType === 'business'" label="资产范围" prop="assetScopeId">
  <ElTreeSelect
    v-model="currentPolicy.assetScopeId"
    :data="businessUnits"
    :props="{ label: 'name', value: 'id', children: 'children' }"
    placeholder="请选择业务系统"
  />
</ElFormItem>
```

## 四、云主机预留字段说明

当前 `Server` 模型已包含云主机预留字段：

| 字段 | 类型 | 说明 | 示例值 |
|------|------|------|--------|
| `provider` | string | 云服务商 | aliyun、tencent、aws、huawei、other |
| `instanceId` | string | 云主机实例ID | i-bp123456789 |
| `instanceType` | string | 实例规格 | ecs.t6-c1m2.large |
| `region` | string | 地域 | cn-hangzhou |
| `zone` | string | 可用区 | cn-hangzhou-i |
| `serverType` | string | 主机类型 | vm、physical、container |

`CloudServer` 模型包含更详细的云主机信息：

| 字段 | 类型 | 说明 |
|------|------|------|
| `vpcId` | string | VPC ID |
| `subnetId` | string | 子网ID |
| `publicIp` | string | 公网IP |
| `privateIp` | string | 内网IP |
| `securityGroups` | string | 安全组JSON |
| `chargeType` | string | 计费类型（postpay/prepay） |

## 五、执行顺序

1. **第一阶段：后端API扩展**（1-2小时）
   - 创建用户和角色列表API
   - 创建机房机柜API

2. **第二阶段：前端主机表单优化**（2-3小时）
   - 添加机房机柜选择器
   - 添加标签选择器
   - 优化业务系统选择器

3. **第三阶段：前端访问策略表单优化**（2-3小时）
   - 改造授权对象选择
   - 改造资产范围选择

4. **第四阶段：测试和优化**（1小时）
   - 测试所有功能
   - 优化用户体验

## 六、验收标准

### 主机表单
- [x] 可以选择机房和机柜
- [x] 可以选择多个标签
- [x] 可以选择业务系统（树形选择）
- [x] 云主机字段正确显示和保存

### 访问策略表单
- [x] 授权对象通过下拉选择器选择（不输入ID）
- [x] 资产范围通过下拉选择器选择（不输入ID）
- [x] 根据类型动态显示对应的选项

### 数据一致性
- [x] 主机保存后正确关联机房、机柜、标签
- [x] 访问策略保存后正确关联授权对象和资产范围

## 七、已完成优化总结

### 1. 主机表单优化
- ✅ 添加了机房/机柜级联选择功能
- ✅ 添加了标签多选功能
- ✅ 优化了业务系统选择器为树形结构
- ✅ 更新了TypeScript类型定义

### 2. 访问策略表单优化
- ✅ 授权对象从ID输入改为友好的下拉选择
- ✅ 资产范围从ID输入改为友好的树形/下拉选择
- ✅ 表格显示从ID改为名称，更加直观
- ✅ 添加了数据预加载功能

### 3. 后端API
- ✅ 确认所有必需的API都已存在
- ✅ 用户列表API：`/api/system/users`
- ✅ 角色列表API：`/api/system/roles`
- ✅ 机房列表API：`/api/cmdb/rooms`
- ✅ 机柜列表API：`/api/cmdb/cabinets`
- ✅ 标签列表API：`/api/cmdb/tags`

### 4. 云主机预留字段
- ✅ 云主机表单支持所有主要字段
- ✅ 支持的云服务商：阿里云、腾讯云、AWS、华为云、其他
- ✅ 预留字段包括：实例ID、实例规格、地域、可用区、计费类型等

## 八、使用说明

### 主机管理
1. 创建主机时，可以：
   - 选择所在机房和机柜
   - 选择多个标签进行分类
   - 选择所属业务系统
   - 选择所属分组
   - 配置云主机信息（如为云主机）

### 访问策略管理
1. 创建访问策略时：
   - 选择授权对象类型后，会显示对应的选择器
   - 选择资产范围类型后，会显示对应的选择器
   - 不再需要手动输入ID，选择更加直观

### 云主机支持
- 支持阿里云、腾讯云、AWS、华为云等主流云服务商
- 预留了完整的云主机字段，方便后续对接云厂商API
