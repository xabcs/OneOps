# OneOps 组件库

这是一套为 OneOps 平台定制的高质量 Vue 3 组件库，专注于管理系统和审计日志场景。

## 📦 安装

组件库已经集成在项目中，无需额外安装：

```javascript
import { PageHeader, TablePage, FormDialog } from '@/components'
```

---

## 🎨 组件列表

### 页面布局组件

#### PageHeader
统一页面头部组件

**Props**:
- `title`: 页面标题
- `subtitle`: 页面描述（可选）

**示例**:
```vue
<PageHeader
  title="用户管理"
  subtitle="管理系统登录账号及所属角色"
/>
```

---

#### TablePage
完整的表格页面组件，包含头部、操作区、表格和分页

**Props**:
- `title`: 页面标题
- `subtitle`: 页面描述
- `loading`: 加载状态
- `total`: 数据总数
- `currentPage`: 当前页码（支持v-model）
- `pageSize`: 每页大小（支持v-model）
- `pageSizes`: 每页大小选项，默认 `[10, 20, 50, 100]`
- `showPagination`: 是否显示分页，默认 `true`
- `paginationLayout`: 分页布局，默认 `'total, sizes, prev, pager, next, jumper'`

**Events**:
- `update:currentPage`: 页码变化
- `update:pageSize`: 每页大小变化
- `size-change`: 每页大小变化事件
- `current-change`: 页码变化事件

**Slots**:
- `actions`: 操作按钮区域
- `default`: 表格内容

**示例**:
```vue
<TablePage
  title="用户管理"
  subtitle="管理系统登录账号及所属角色"
  :loading="loading"
  :total="total"
  v-model:current-page="currentPage"
  v-model:page-size="pageSize"
  @size-change="fetchUsers"
  @current-change="fetchUsers"
>
  <template #actions>
    <el-space>
      <el-input
        v-model="queryParams.username"
        placeholder="搜索用户名"
        :prefix-icon="Search"
        clearable
        style="width: 200px"
        @keyup.enter="fetchUsers"
      />
      <el-button type="primary" :icon="Plus" @click="handleAdd">
        新增用户
      </el-button>
    </el-space>
  </template>

  <el-table :data="displayUsers" border style="width: 100%">
    <el-table-column prop="username" label="用户名" />
    <el-table-column prop="email" label="邮箱" />
  </el-table>
</TablePage>
```

---

### 表单组件

#### FormDialog
统一的表单对话框组件

**Props**:
- `modelValue`: 对话框显示状态（支持v-model）
- `title`: 对话框标题
- `formModel`: 表单数据对象
- `rules`: 表单验证规则
- `width`: 对话框宽度，默认 `'500px'`
- `labelWidth`: 表单标签宽度，默认 `'80px'`
- `labelPosition`: 表单标签位置，默认 `'top'`
- `submitting`: 提交中状态
- `cancelText`: 取消按钮文本，默认 `'取消'`
- `confirmText`: 确定按钮文本，默认 `'确定'`
- `showFooter`: 是否显示底部，默认 `true`
- `destroyOnClose`: 关闭时销毁内容，默认 `true`

**Events**:
- `submit`: 表单提交事件（已通过验证）
- `close`: 对话框关闭事件

**Exposed Methods**:
- `clearValidate()`: 清除验证
- `resetFields()`: 重置表单
- `formRef`: 表单引用

**示例**:
```vue
<FormDialog
  v-model="dialogVisible"
  :title="form.id ? '编辑用户' : '新增用户'"
  :form-model="form"
  :rules="rules"
  :submitting="submitting"
  @submit="submitForm"
  width="600px"
>
  <el-row :gutter="20">
    <el-col :span="12">
      <el-form-item label="用户名" prop="username">
        <el-input
          v-model="form.username"
          placeholder="请输入用户名"
          :disabled="!!form.id"
        />
      </el-form-item>
    </el-col>
    <el-col :span="12">
      <el-form-item label="密码" prop="password" v-if="!form.id">
        <el-input
          v-model="form.password"
          type="password"
          placeholder="请输入密码"
          show-password
        />
      </el-form-item>
    </el-col>
  </el-row>

  <el-form-item label="邮箱" prop="email">
    <el-input v-model="form.email" placeholder="请输入邮箱" />
  </el-form-item>

  <el-form-item label="状态" prop="status">
    <el-radio-group v-model="form.status">
      <el-radio label="active">启用</el-radio>
      <el-radio label="disabled">禁用</el-radio>
    </el-radio-group>
  </el-form-item>
</FormDialog>

<script setup>
const dialogVisible = ref(false)
const submitting = ref(false)

const form = ref({
  id: null,
  username: '',
  password: '',
  email: '',
  status: 'active'
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  email: [{ type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }]
}

const submitForm = async (formData) => {
  // FormDialog 已经处理了验证，这里直接使用数据
  submitting.value = true
  try {
    await api.submit(formData)
    ElMessage.success('保存成功')
    dialogVisible.value = false
  } finally {
    submitting.value = false
  }
}
</script>
```

---

### 表格功能组件

#### TableActions
表格操作列组件

**Props**:
- `row`: 当前行数据
- `buttons`: 按钮配置数组

**按钮配置**:
- `text`: 按钮文本
- `type`: 按钮类型（`primary` | `success` | `warning` | `danger` | `info`）
- `icon`: 按钮图标
- `disabled`: 禁用条件函数 `(row) => boolean`
- `visible`: 显示条件函数 `(row) => boolean`
- `click`: 点击事件处理函数 `(row) => void`
- `class`: 自定义CSS类

**示例**:
```vue
<el-table-column label="操作" width="200" align="center">
  <template #default="{ row }">
    <TableActions
      :row="row"
      :buttons="[
        {
          text: '编辑',
          type: 'primary',
          click: (row) => handleEdit(row)
        },
        {
          text: '重置密码',
          click: (row) => handleResetPwd(row),
          class: 'text-accent'
        },
        {
          text: '删除',
          type: 'danger',
          click: (row) => handleDelete(row),
          disabled: (row) => row.id === currentUserId
        },
        {
          text: '详情',
          click: (row) => handleDetail(row),
          visible: (row) => row.status === 'active'
        }
      ]"
    />
  </template>
</el-table-column>
```

---

#### SwitchStatus
状态开关组件

**Props**:
- `modelValue`: 开关状态（支持v-model）
- `activeValue`: 激活值，默认 `true`
- `inactiveValue`: 未激活值，默认 `false`
- `disabled`: 是否禁用
- `readonly`: 是否只读
- `confirmMessage`: 确认消息（字符串或函数）

**Events**:
- `change`: 状态变化事件
- `update:modelValue`: 状态更新事件

**示例**:
```vue
<!-- 基础用法 -->
<SwitchStatus
  v-model="row.status"
  active-value="active"
  inactive-value="disabled"
  @change="handleStatusChange"
/>

<!-- 带确认对话框 -->
<SwitchStatus
  v-model="row.status"
  active-value="active"
  inactive-value="disabled"
  :readonly="row.id === currentUserId"
  confirm-message="确定要更改状态吗？"
  @change="(row, val) => handleStatusChange(row, val)"
/>

<!-- 动态确认消息 -->
<SwitchStatus
  v-model="row.status"
  :confirm-message="(val) => val === 'active' ? '确定要启用吗？' : '确定要禁用吗？'"
  @change="handleStatusChange"
/>
```

---

### 审计日志专用组件

#### AuditFilter
审计日志筛选组件

**Props**:
- `form`: 表单数据对象
- `showTimeRange`: 是否显示时间范围选择器，默认 `true`

**Events**:
- `search`: 搜索事件
- `reset`: 重置事件
- `time-range-change`: 时间范围变化事件

**Slots**:
- `filters`: 自定义筛选字段

**示例**:
```vue
<AuditFilter
  :form="searchForm"
  @search="handleSearch"
  @reset="resetForm"
  @time-range-change="handleTimeRangeChange"
>
  <template #filters>
    <el-form-item label="登录账号">
      <el-input
        v-model="searchForm.username"
        placeholder="输入账号"
        clearable
        style="width: 180px"
        @clear="handleSearch"
      />
    </el-form-item>
    <el-form-item label="状态">
      <el-select
        v-model="searchForm.status"
        placeholder="全部"
        clearable
        style="width: 120px"
      >
        <el-option label="成功" value="success" />
        <el-option label="失败" value="failed" />
      </el-select>
    </el-form-item>
  </template>
</AuditFilter>
```

---

#### AuditLogTable
审计日志表格组件

**Props**:
- `data`: 表格数据
- `loading`: 加载状态
- `total`: 数据总数
- `currentPage`: 当前页码（支持v-model）
- `pageSize`: 每页大小（支持v-model）
- `pageSizes`: 每页大小选项

**Events**:
- `update:currentPage`: 页码变化
- `update:pageSize`: 每页大小变化
- `sort-change`: 排序变化

**Slots**:
- `default`: 表格列

**示例**:
```vue
<AuditLogTable
  :data="logs"
  :loading="loading"
  :total="total"
  :current-page="currentPage"
  :page-size="pageSize"
  :page-sizes="[10, 20, 50, 100]"
  @update:current-page="handleCurrentChange"
  @update:page-size="handleSizeChange"
  @sort-change="handleSortChange"
>
  <el-table-column prop="time" label="登录时间" width="180" sortable="custom">
    <template #default="{ row }">
      <span class="data-value">{{ row.time }}</span>
    </template>
  </el-table-column>

  <el-table-column prop="username" label="登录账号" width="120">
    <template #default="{ row }">
      <div class="user-cell">
        <el-avatar :size="24" :src="getAvatarUrl(row.username)" />
        <span>{{ row.username }}</span>
      </div>
    </template>
  </el-table-column>

  <el-table-column prop="status" label="状态" width="100">
    <template #default="{ row }">
      <StatusTag :status="row.status" />
    </template>
  </el-table-column>
</AuditLogTable>
```

---

#### TimeRangeSelector
时间范围选择器组件

**Props**:
- `modelValue`: 选择的时间范围（支持v-model）
- `options`: 时间范围选项

**Events**:
- `change`: 时间范围变化事件

**示例**:
```vue
<TimeRangeSelector
  v-model="timeRange"
  @change="handleTimeRangeChange"
/>
```

---

### 其他功能组件

#### StatusTag
状态标签组件（自动映射）

**Props**:
- `status`: 状态值
- `typeMap`: 自定义状态映射

**示例**:
```vue
<!-- 基础用法 -->
<StatusTag :status="row.status" />

<!-- 自定义映射 -->
<StatusTag
  :status="row.status"
  :type-map="{
    active: 'success',
    inactive: 'info',
    banned: 'danger'
  }"
/>
```

---

#### SearchBar
搜索栏组件

**Props**:
- `modelValue`: 搜索关键词（支持v-model）
- `placeholder`: 占位文本
- `showClear`: 是否显示清空按钮

**Events**:
- `search`: 搜索事件
- `clear`: 清空事件

**示例**:
```vue
<SearchBar
  v-model="keyword"
  placeholder="搜索..."
  @search="handleSearch"
/>
```

---

## 💡 最佳实践

### 1. 组件组合

推荐使用组件组合来构建复杂的页面：

```vue
<template>
  <TablePage
    title="用户管理"
    :loading="loading"
    :total="total"
    v-model:current-page="page"
    v-model:page-size="size"
    @current-change="fetch"
  >
    <template #actions>
      <SearchBar v-model="keyword" @search="fetch" />
      <el-button type="primary" @click="handleAdd">新增</el-button>
    </template>

    <el-table :data="data">
      <el-table-column>
        <template #default="{ row }">
          <TableActions :row="row" :buttons="actions" />
        </template>
      </el-table-column>
    </el-table>
  </TablePage>

  <FormDialog
    v-model="dialogVisible"
    title="表单"
    :form-model="form"
    :rules="rules"
    @submit="handleSubmit"
  >
    <!-- 表单项 -->
  </FormDialog>
</template>
```

### 2. 事件处理

使用组件提供的事件，而不是手动管理：

```vue
<!-- 推荐：使用组件事件 -->
<TablePage @current-change="fetchData" />

<!-- 不推荐：手动处理 -->
<el-pagination @current-change="handlePageChange" />
```

### 3. 表单验证

利用FormDialog的内置验证：

```vue
<!-- 推荐：使用FormDialog -->
<FormDialog :rules="rules" @submit="handleSubmit">
  <!-- 自动处理验证 -->
</FormDialog>

<!-- 不推荐：手动验证 -->
<el-dialog>
  <el-form ref="formRef">
    <!-- 需要手动验证 -->
  </el-form>
</el-dialog>
```

---

## 🔧 常见问题

### Q: 如何自定义表格列？

A: 使用 `AuditLogTable` 的默认插槽：

```vue
<AuditLogTable :data="logs" :total="total">
  <el-table-column prop="field" label="字段" />
</AuditLogTable>
```

### Q: 如何处理状态切换的回滚？

A: 使用 `SwitchStatus` 组件，它会自动处理回滚：

```vue
<SwitchStatus
  v-model="row.status"
  @change="async (row, val) => {
    try {
      await api.updateStatus(row.id, val)
      ElMessage.success('更新成功')
    } catch (error) {
      // SwitchStatus 会自动回滚状态
      ElMessage.error('更新失败')
    }
  }"
/>
```

### Q: 如何禁用某个操作按钮？

A: 使用 `disabled` 函数：

```vue
<TableActions
  :row="row"
  :buttons="[
    {
      text: '删除',
      type: 'danger',
      click: handleDelete,
      disabled: (row) => row.id === currentUserId
    }
  ]"
/>
```

---

## 📚 相关文档

- [重构总结](../REFACTOR_SUMMARY.md)
- [行为日志重构对比](../BEHAVIOR_REFACTOR_COMPARISON.md)
- [系统管理重构对比](../SYSTEM_REFACTOR_COMPARISON.md)

---

**组件库版本**: 1.0.0
**最后更新**: 2026-04-24
**维护者**: OneOps Team
