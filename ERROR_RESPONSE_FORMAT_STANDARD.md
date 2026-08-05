# 前后端错误响应格式统一规范

## ✅ 结论：前端不需要兼容 `msg` 字段

### 原因

**后端统一使用 `message` 字段**，完全没有使用 `msg` 字段。

---

## 📋 后端响应格式

### 统一响应结构（backend/utils/response.go）

```go
type Response struct {
    Code    int         `json:"code"`
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Message string      `json:"message"`    // ✅ 统一使用 message
}
```

### 响应示例

**成功响应**：
```json
{
  "code": 200,
  "success": true,
  "data": { ... },
  "message": "操作成功"
}
```

**失败响应**：
```json
{
  "code": 400,
  "success": false,
  "message": "用户名已存在"
}
```

---

## 🔍 验证结果

### 后端搜索结果

```bash
# 搜索 json:"msg"
grep -rn 'json:"msg"' backend --include="*.go"
# 结果：无

# 搜索 gin.H{"msg": ...}
grep -rn 'gin.H{.*"msg"' backend --include="*.go"
# 结果：无
```

**结论**：后端**完全没有使用** `msg` 字段。

---

## 🔧 前端修改

### 已修改的文件

#### 1. `frontend/src/service/request/index.ts`

**修改前**（错误地兼容 `msg`）：
```typescript
// ❌ 错误：兼容不存在的字段
message = error.response?.data?.msg || message;
request.state.errMsgStack.filter(msg => msg !== response.data.msg);
?.confirm(response.data.msg, $t('common.error'), {
```

**修改后**（统一使用 `message`）：
```typescript
// ✅ 正确：统一使用 message
message = error.response?.data?.message || message;
request.state.errMsgStack.filter(msg => msg !== response.data.message);
?.confirm(response.data.message, $t('common.error'), {
```

---

#### 2. `frontend/src/views/manage/user/modules/user-operate-drawer.vue`

**修改前**：
```typescript
// ❌ 错误：兼容不存在的字段
const backendMessage = error?.response?.data?.message
  || error?.response?.data?.msg    // ← 兼容 msg
  || error?.message
  || '添加失败';
```

**修改后**：
```typescript
// ✅ 正确：统一使用 message
const backendMessage = error?.response?.data?.message
  || error?.message
  || '添加失败';
```

---

## 📊 前后端约定

### 错误消息字段约定

| 场景 | 后端返回 | 前端读取 | 说明 |
|------|---------|---------|------|
| **所有响应** | `message` 字段 | `response.data.message` | ✅ 统一标准 |

**不需要兼容的字段**：
- ❌ `msg` - 后端不使用
- ❌ `error` - 后端不使用
- ❌ `errMsg` - 后端不使用

---

## 🎯 统一的错误处理规范

### 后端规范

**统一使用 `utils/response.go` 中的方法**：

```go
// ✅ 正确：使用统一方法
c.JSON(200, utils.ErrorBadRequest("用户名已存在"))
c.JSON(200, utils.ErrorForbidden("权限不足"))
c.JSON(200, utils.SuccessResponse(data, "操作成功"))

// ❌ 错误：手动构造响应
c.JSON(200, gin.H{
    "code": 400,
    "msg": "错误",    // ← 错误字段名
})
```

---

### 前端规范

**错误处理模板**：

```typescript
// ✅ 正确：统一使用 message 字段
const { error } = await someApiCall();

if (!error) {
  window.$message?.success('操作成功');
} else {
  const errorMessage = error?.response?.data?.message
    || error?.message
    || '操作失败';
  window.$message?.error(errorMessage);
}
```

**禁止的做法**：
```typescript
// ❌ 错误：兼容不存在的字段
error?.response?.data?.msg

// ❌ 错误：硬编码错误消息
window.$message?.error('操作失败');
```

---

## 📝 变更记录

### 修改时间
2026-08-04

### 修改内容
1. ✅ 移除前端对 `msg` 字段的兼容
2. ✅ 统一使用 `message` 字段
3. ✅ 确认后端响应格式一致

### 修改文件
1. `frontend/src/service/request/index.ts`
2. `frontend/src/views/manage/user/modules/user-operate-drawer.vue`

---

## ✅ 总结

**前后端完全统一**：

| 维度 | 字段名 | 说明 |
|------|--------|------|
| **后端响应** | `message` | ✅ 统一标准 |
| **前端读取** | `response.data.message` | ✅ 统一标准 |
| **兼容 `msg`** | 不需要 | ✅ 后端不使用 |

**现在前后端错误响应格式完全统一，不再有字段名不一致的问题！**
