# 前端提示方式使用指南

本文档详细说明 Element Plus 三种提示方式的使用场景和最佳实践。

## 📋 目录

- [三种提示方式概述](#三种提示方式概述)
- [1. Message 消息提示](#1-message-消息提示)
- [2. Notification 通知](#2-notification-通知)
- [3. Dialog/MessageBox 弹出框](#3-dialogmessagebox-弹出框)
- [对比总结](#对比总结)
- [最佳实践](#最佳实践)
- [项目中的实际应用](#项目中的实际应用)

---

## 三种提示方式概述

| 方式 | 组件 | 位置 | 打断性 | 信息量 |
|-----|------|------|--------|--------|
| **Message** | `ElMessage` | 页面顶部居中 | ⭐ 低 | 少 |
| **Notification** | `ElNotification` | 页面右上角 | ⭐⭐ 中 | 多 |
| **Dialog** | `ElMessageBox` | 页面中心（模态） | ⭐⭐⭐ 高 | 最多 |

---

## 1. Message 消息提示

### 特点
```
┌─────────────────────────────┐
│     ✅ 删除成功              │
└─────────────────────────────┘
```

- **位置**：页面顶部居中
- **样式**：简单的消息条，带图标
- **持续时间**：3秒后自动消失
- **交互**：无法手动关闭，等待自动消失
- **内容**：简短文字（一般不超过20字）

### 使用方法

```typescript
// 方式1：使用全局对象（推荐）
window.$message?.success('操作成功');
window.$message?.error('操作失败');
window.$message?.warning('警告信息');
window.$message?.info('提示信息');

// 方式2：直接调用
import { ElMessage } from 'element-plus';
ElMessage.success('操作成功');
```

### ✅ 适合场景

| 场景类型 | 具体示例 | 使用原因 |
|---------|---------|---------|
| **操作成功** | 保存成功、删除成功、更新成功 | 轻量反馈，快速提示，不干扰用户 |
| **操作失败** | 保存失败、删除失败、加载失败 | 快速提示错误，不阻塞用户继续操作 |
| **状态变更** | 启用成功、禁用成功、状态更新 | 简洁提示状态变化 |
| **简单警告** | 未选择任何项、输入为空 | 提醒用户注意，但不严重 |
| **表单验证** | 必填项未填写（单条提示） | 快速提示具体错误 |

### ❌ 不适合场景

- ❌ 需要显示详细原因的错误
- ❌ 需要用户确认的操作
- ❌ 重要的系统通知
- ❌ 复杂的提示信息

### 代码示例

```typescript
// 成功提示
async function handleSave() {
  const { error } = await fetchData();
  if (!error) {
    window.$message?.success('保存成功');
  }
}

// 错误提示
async function handleDelete() {
  const { error } = await fetchDelete();
  if (error) {
    window.$message?.error('删除失败');
  }
}

// 警告提示
function handleSubmit() {
  if (!form.value.name) {
    window.$message?.warning('请输入名称');
    return false;
  }
}

// 带配置的提示
window.$message?.success({
  message: '操作成功',
  duration: 2000,  // 持续时间
  showClose: false, // 不显示关闭按钮
  grouping: true    // 相同消息合并显示
});
```

---

## 2. Notification 通知

### 特点
```
┌────────────────────────────────┐
│ ⚠️ 无法删除角色        ✕      │
│ ─────────────────────────────  │
│ 该角色已关联 3 个用户：        │
│ user1、user2、user3。请先...   │
└────────────────────────────────┘
```

- **位置**：页面右上角（可配置）
- **样式**：卡片式设计，有标题和内容
- **持续时间**：可配置（默认3-5秒）
- **交互**：可手动关闭 ✕
- **内容**：标题 + 详细内容，支持多行、HTML

### 使用方法

```typescript
import { ElNotification } from 'element-plus';

// 基础用法
ElNotification({
  title: '通知标题',
  message: '通知内容',
  type: 'success'  // success | warning | info | error
});

// 完整配置
ElNotification({
  title: '无法删除角色',
  message: '该角色已关联 3 个用户',
  type: 'warning',
  duration: 5000,      // 持续时间（毫秒）
  position: 'top-right', // 位置：top-right | top-left | bottom-right | bottom-left
  showClose: true,      // 显示关闭按钮
  onClick: () => {      // 点击回调
    console.log('通知被点击');
  }
});
```

### ✅ 适合场景

| 场景类型 | 具体示例 | 使用原因 |
|---------|---------|---------|
| **重要通知** | 系统公告、重要提醒 | 醒目但不强制打断用户当前操作 |
| **带原因的错误** | 删除失败（原因：已关联3个用户） | 提供详细信息帮助用户理解问题 |
| **异步操作结果** | 数据同步完成、任务完成 | 不打扰用户但告知操作结果 |
| **需要详细说明** | 3个角色不可删除（列出每个原因） | 可以显示多条详细信息 |
| **非阻塞提醒** | 新消息、系统更新、数据过期 | 用户可以选择性查看 |
| **操作结果汇总** | 批量操作结果（成功5个，失败1个） | 可以显示详细的操作统计 |

### ❌ 不适合场景

- ❌ 需要用户立即确认的危险操作（应该用 Dialog）
- ❌ 简单的成功/失败提示（应该用 Message）
- ❌ 需要用户输入的场景（应该用 Dialog）

### 代码示例

```typescript
// 带原因的错误提示
async function handleDeleteRole(id: number) {
  const { canDelete, reason } = checkCanDelete(id);

  if (!canDelete) {
    ElNotification({
      title: '无法删除角色',
      message: reason,
      type: 'warning',
      duration: 5000
    });
    return;
  }
}

// 多行消息
ElNotification({
  title: '批量删除结果',
  message: `成功删除 5 个角色\n失败 1 个角色：\n• 测试角色: 已关联用户`,
  type: 'success',
  duration: 5000
});

// 使用 HTML
ElNotification({
  title: '系统通知',
  dangerouslyUseHTMLString: true,
  message: '<strong>新版本</strong> 已发布，<a href="#">点击查看</a>',
  type: 'info'
});

// 点击回调
ElNotification({
  title: '数据已过期',
  message: '点击刷新数据',
  type: 'warning',
  onClick: () => {
    refreshData();
  }
});
```

---

## 3. Dialog/MessageBox 弹出框

### 特点
```
┌─────────────────────────────────┐
│     ⚠️ 确认删除                 │ ← 模态对话框
│                                 │   遮罩层覆盖整个页面
│  确定要删除该角色吗？           │   强制用户操作
│  此操作不可恢复。               │
│                                 │
│    [ 取消 ]    [ 确定 ]        │
└─────────────────────────────────┘
```

- **位置**：页面中心，覆盖整个页面
- **样式**：模态对话框，带遮罩层
- **持续时间**：永久显示，直到用户操作
- **交互**：强制用户操作（点击按钮）
- **内容**：标题 + 内容 + 按钮，可自定义

### 使用方法

```typescript
import { ElMessageBox } from 'element-plus';

// 确认框
try {
  await ElMessageBox.confirm(
    '确定要删除吗？此操作不可恢复。',
    '确认删除',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  );
  // 用户点击确定
  console.log('用户确认删除');
} catch {
  // 用户点击取消
  console.log('用户取消删除');
}

// 警告框
await ElMessageBox.alert(
  '这是一条重要提示信息',
  '提示',
  {
    confirmButtonText: '我知道了',
    type: 'warning'
  }
);

// 输入框
const { value } = await ElMessageBox.prompt(
  '请输入角色名称',
  '新增角色',
  {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    inputPattern: /^[a-zA-Z0-9_]+$/,
    inputErrorMessage: '只能输入字母、数字和下划线'
  }
);
```

### ✅ 适合场景

| 场景类型 | 具体示例 | 使用原因 |
|---------|---------|---------|
| **危险操作确认** | 删除确认、重置数据、清空数据 | 防止误操作，需要明确确认 |
| **重要决策** | 保存并退出、放弃修改、离开页面 | 需要用户做出明确选择 |
| **系统级错误** | 网络断开、服务器错误、权限不足 | 严重问题，需要用户注意并处理 |
| **表单验证** | 必填项未填写、数据格式错误 | 阻止提交，要求用户处理 |
| **重要提示** | 首次使用引导、功能说明、注意事项 | 需要用户阅读并理解 |
| **二次确认** | 批量操作、不可逆操作 | 再次确认避免误操作 |

### ❌ 不适合场景

- ❌ 简单的成功提示（太重，应该用 Message）
- ❌ 频繁的操作反馈（会打断用户，应该用 Message）
- ❌ 不重要的信息（会被用户反感）

### 代码示例

```typescript
// 删除确认
async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm(
      `确定要删除角色 "${roleName}" 吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    );

    // 用户确认，执行删除
    const { error } = await fetchDelete(id);
    if (!error) {
      window.$message?.success('删除成功');
    }
  } catch {
    // 用户取消
    console.log('用户取消删除');
  }
}

// 批量操作确认
async function handleBatchDelete(items: any[]) {
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${items.length} 个项目吗？此操作不可恢复。`,
      '确认批量删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    );
    // 执行批量删除
  } catch {
    // 用户取消
  }
}

// 系统错误提示
async function handleCriticalError() {
  try {
    await ElMessageBox.alert(
      '系统连接已断开，请检查网络设置后刷新页面。',
      '连接错误',
      {
        confirmButtonText: '刷新页面',
        type: 'error',
        showClose: false  // 禁止关闭，必须处理
      }
    );
    location.reload();
  } catch {
    // 不可能执行，因为 showClose: false
  }
}

// 首次使用引导
async function showFirstTimeGuide() {
  await ElMessageBox.alert(
    '欢迎使用系统！\n\n这是您第一次使用角色管理功能。\n\n点击"确定"开始使用。',
    '功能介绍',
    {
      confirmButtonText: '开始使用',
      type: 'info'
    });
}
```

---

## 对比总结

### 详细对比表

| 对比维度 | Message | Notification | Dialog |
|---------|---------|--------------|--------|
| **打断性** | ⭐ 低（自动消失） | ⭐⭐ 中（可关闭） | ⭐⭐⭐ 高（必须处理） |
| **信息量** | 少（简短文字） | 多（标题+内容） | 最多（可包含表单） |
| **用户操作** | 无需操作 | 可选关闭 | 必须操作 |
| **适用频率** | ⭐⭐⭐⭐⭐ 高 | ⭐⭐⭐ 中 | ⭐⭐ 低 |
| **用户打扰** | 最低 | 中等 | 最高 |
| **性能开销** | 最小 | 中等 | 最大 |
| **响应速度** | 快 | 中 | 慢（需要等待用户操作） |

### 选择决策树

```
需要提示用户
    ↓
是否需要用户确认/决策？
    ↓
┌───────────┴───────────┐
│ 是                    │ 否
│                       │
↓                       ↓
是否是危险/不可逆操作？   是否需要详细说明？
↓                       ↓
┌───────┴───────┐      ┌──────┴──────┐
│ 是           │ 否    │ 是          │ 否
│              │       │             │
↓              ↓       ↓             ↓
Dialog         Dialog  Notification  Message
(确认框)      (提示框)  (通知)       (消息)
```

---

## 最佳实践

### 1. 删除操作的三层提示策略

```typescript
// 第1层：前置检查通知（Notification）
if (isBuiltinRole) {
  ElNotification({
    title: '无法删除',
    message: '该角色是系统内置角色，不能删除',
    type: 'warning'
  });
  return;
}

// 第2层：危险操作确认（Dialog）
try {
  await ElMessageBox.confirm(
    '确定要删除吗？此操作不可恢复。',
    '确认删除',
    { type: 'warning' }
  );
} catch {
  return;  // 用户取消
}

// 第3层：操作结果反馈（Message）
const { error } = await fetchDelete();
if (!error) {
  window.$message?.success('删除成功');
}
```

### 2. 批量操作的提示策略

```typescript
// 批量删除流程

// 1. 前置检查通知（Notification）
const invalidItems = checkItems(selectedItems);
if (invalidItems.length > 0) {
  ElNotification({
    title: `无法删除 ${invalidItems.length} 个项目`,
    message: invalidItems.map(i => `• ${i.name}: ${i.reason}`).join('\n'),
    type: 'warning',
    duration: 5000
  });
}

// 2. 批量操作确认（Dialog）
try {
  await ElMessageBox.confirm(
    `确定要删除选中的 ${validItems.length} 个项目吗？`,
    '确认批量删除',
    { type: 'warning' }
  );
} catch {
  return;
}

// 3. 批量操作结果（Notification）
ElNotification({
  title: '批量删除完成',
  message: `成功: ${successCount}，失败: ${failCount}`,
  type: successCount > 0 ? 'success' : 'error'
});
```

### 3. 表单验证的提示策略

```typescript
// 单条错误 → Message
if (!form.value.name) {
  window.$message?.warning('请输入名称');
  return false;
}

// 多条错误 → Dialog（汇总显示）
const errors = validateForm(form.value);
if (errors.length > 0) {
  ElMessageBox.alert(
    errors.map(e => `• ${e.field}: ${e.message}`).join('\n'),
    '表单验证失败',
    { type: 'error' }
  );
  return false;
}
```

### 4. 异步操作的提示策略

```typescript
// 开始操作
window.$message?.info('正在处理，请稍候...');

// 异步操作
try {
  await asyncOperation();

  // 成功 → Message
  window.$message?.success('操作成功');
} catch (error) {
  // 失败 → Notification（显示详细错误）
  ElNotification({
    title: '操作失败',
    message: error.message,
    type: 'error',
    duration: 5000
  });
}
```

### 5. 系统级错误的提示策略

```typescript
// 系统级错误 → Dialog（强制处理）
try {
  await criticalOperation();
} catch (error) {
  ElMessageBox.alert(
    '系统发生严重错误，建议刷新页面或联系管理员。',
    '系统错误',
    {
      confirmButtonText: '刷新页面',
      type: 'error',
      showClose: false  // 禁止关闭
    }
  );
}
```

---

## 项目中的实际应用

### 用户管理模块

```typescript
// 1. 删除单个用户
async function handleDeleteUser(id: number) {
  // 确认框
  try {
    await ElMessageBox.confirm('确定要删除该用户吗？', '确认删除', {
      type: 'warning'
    });
  } catch {
    return;
  }

  // 执行删除
  const { error } = await fetchDeleteUser(id);

  // 成功提示
  if (!error) {
    window.$message?.success('删除成功');
  } else {
    window.$message?.error('删除失败');
  }
}

// 2. 批量删除用户
async function handleBatchDeleteUsers(ids: number[]) {
  // 前置检查通知
  const adminUser = ids.find(id => isUserAdmin(id));
  if (adminUser) {
    ElNotification({
      title: '无法删除',
      message: 'admin 用户是系统管理员，不能删除',
      type: 'warning'
    });
    return;
  }

  // 确认框
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${ids.length} 个用户吗？`,
      '确认批量删除',
      { type: 'warning' }
    );
  } catch {
    return;
  }

  // 执行删除
  let successCount = 0;
  for (const id of ids) {
    const { error } = await fetchDeleteUser(id);
    if (!error) successCount++;
  }

  // 结果通知
  ElNotification({
    title: '批量删除完成',
    message: `成功删除 ${successCount} 个用户`,
    type: 'success'
  });
}
```

### 角色管理模块

```typescript
// 1. 删除角色（带详细原因）
async function handleDeleteRole(id: number) {
  const { canDelete, reason } = checkCanDeleteRole(id);

  // 不能删除 → 通知（显示原因）
  if (!canDelete) {
    ElNotification({
      title: '无法删除角色',
      message: reason,
      type: 'warning',
      duration: 5000
    });
    return;
  }

  // 可以删除 → 确认框
  try {
    await ElMessageBox.confirm(
      '确定要删除该角色吗？此操作不可恢复。',
      '确认删除',
      { type: 'warning' }
    );
  } catch {
    return;
  }

  // 执行删除
  const { error } = await fetchDeleteRole(id);

  // 结果提示
  if (!error) {
    window.$message?.success('删除成功');
  } else {
    window.$message?.error('删除失败');
  }
}

// 2. 修改角色状态
async function handleStatusChange(role: Role, status: number) {
  const { error } = await fetchUpdateRole(role.id, { status });

  if (!error) {
    // 轻量提示
    window.$message?.success(status === 1 ? '已启用' : '已禁用');
  } else {
    window.$message?.error('状态更新失败');
  }
}
```

### 登录模块

```typescript
// 1. 登录失败（表单验证）
function handleLoginError() {
  if (!form.value.username) {
    window.$message?.warning('请输入用户名');
    return;
  }

  if (!form.value.password) {
    window.$message?.warning('请输入密码');
    return;
  }
}

// 2. 登录失败（服务器错误）
async function handleLogin() {
  const { error } = await fetchLogin(form.value);

  if (error) {
    // 详细错误通知
    ElNotification({
      title: '登录失败',
      message: error.message,
      type: 'error',
      duration: 5000
    });
  }
}

// 3. 登录成功
window.$message?.success('登录成功');
```

---

## 🎯 快速选择指南

### 一句话总结

- **Message**：「操作成功了/失败了」（轻量、快速）
- **Notification**：「操作失败了，原因是XXX」（重要、详细）
- **Dialog**：「你确定要XXX吗？这事很危险」（危险、需要确认）

### 场景速查表

| 场景 | 使用方式 | 示例代码 |
|-----|---------|---------|
| 操作成功 | Message | `window.$message?.success('保存成功')` |
| 操作失败 | Message | `window.$message?.error('保存失败')` |
| 删除确认 | Dialog | `ElMessageBox.confirm('确定删除？', '确认')` |
| 不能删除+原因 | Notification | `ElNotification({ title, message, type: 'warning' })` |
| 系统错误 | Dialog | `ElMessageBox.alert('系统错误', '错误', { type: 'error' })` |
| 批量操作结果 | Notification | `ElNotification({ title: '完成', message: '成功5个' })` |
| 简单警告 | Message | `window.$message?.warning('请先选择')` |

---

## 📚 参考资源

- [Element Plus Message](https://element-plus.org/zh-CN/component/message.html)
- [Element Plus Notification](https://element-plus.org/zh-CN/component/notification.html)
- [Element Plus MessageBox](https://element-plus.org/zh-CN/component/message-box.html)

---

**最后更新**: 2026-05-11
**维护者**: 开发团队
**版本**: v1.0
