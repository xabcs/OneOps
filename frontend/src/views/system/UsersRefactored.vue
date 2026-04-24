<template>
  <div class="users-container">
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
            placeholder="搜索用户名/昵称"
            :prefix-icon="Search"
            clearable
            style="width: 200px"
            @keyup.enter="fetchUsers"
          />
          <el-select
            v-model="queryParams.status"
            placeholder="状态"
            clearable
            style="width: 100px"
            @change="fetchUsers"
          >
            <el-option label="启用" value="active" />
            <el-option label="禁用" value="disabled" />
          </el-select>
          <el-button type="primary" :icon="Plus" @click="handleAdd">
            新增用户
          </el-button>
        </el-space>
      </template>

      <el-table :data="displayUsers" border style="width: 100%">
        <el-table-column label="用户名" width="140">
          <template #default="{ row }">
            <div class="user-cell">
              <el-avatar :size="32" :src="row.avatar">
                {{ row.nickname?.charAt(0).toUpperCase() }}
              </el-avatar>
              <span class="data-value user-name">{{ row.username }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="nickname" label="昵称" width="150" />

        <el-table-column prop="roleNames" label="分配角色" min-width="200">
          <template #default="{ row }">
            <div class="role-tags">
              <el-tag
                v-for="roleName in row.roleNames"
                :key="roleName"
                size="small"
                type="success"
                class="role-tag"
              >
                {{ roleName }}
              </el-tag>
              <span v-if="!row.roleNames?.length" class="text-tertiary">
                未分配
              </span>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="email" label="邮箱" min-width="200" />

        <el-table-column prop="homePath" label="家目录" width="150" />

        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <SwitchStatus
              v-model="row.status"
              active-value="active"
              inactive-value="disabled"
              :readonly="row.id === store.state.user?.id"
              @change="(val) => handleStatusChange(row, val)"
            />
          </template>
        </el-table-column>

        <el-table-column prop="createdAt" label="创建时间" width="180" />

        <el-table-column label="操作" width="200" align="center">
          <template #default="{ row }">
            <TableActions
              :row="row"
              :buttons="[
                { text: '编辑', click: handleEdit },
                { text: '重置密码', click: handleResetPwd, class: 'text-accent' },
                {
                  text: '删除',
                  type: 'danger',
                  click: handleDelete,
                  disabled: (row) => row.id === store.state.user?.id
                }
              ]"
            />
          </template>
        </el-table-column>
      </el-table>
    </TablePage>

    <!-- User Form Dialog -->
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

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="昵称" prop="nickname">
            <el-input v-model="form.nickname" placeholder="请输入昵称" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item label="邮箱" prop="email">
        <el-input v-model="form.email" placeholder="请输入邮箱" />
      </el-form-item>

      <el-form-item label="分配角色" prop="roleIds">
        <el-select
          v-model="form.roleIds"
          placeholder="请先选择角色，系统将根据角色权限显示可用的家目录"
          style="width: 100%"
          multiple
          collapse-tags
          collapse-tags-tooltip
          @change="handleRoleIdsChange"
        >
          <el-option
            v-for="role in roleOptions"
            :key="role.id"
            :label="role.name"
            :value="role.id"
          />
        </el-select>
      </el-form-item>

      <div v-if="form.roleIds.length > 0 && menuOptions.length > 0" class="role-permission-hint">
        <el-icon><InfoFilled /></el-icon>
        <span>
          已为所选角色配置了 <strong>{{ menuOptions.length }}</strong> 个可访问的家目录，请从下方选择
        </span>
      </div>

      <el-form-item label="家目录" prop="homePath">
        <template #label>
          <span>家目录</span>
          <el-tooltip
            v-if="form.roleIds.length > 0"
            content="系统已根据您选择的角色权限，仅显示可访问的路径"
            placement="top"
          >
            <el-icon class="label-icon"><QuestionFilled /></el-icon>
          </el-tooltip>
        </template>
        <el-select
          v-model="form.homePath"
          :placeholder="homePathPlaceholder"
          :disabled="form.roleIds.length === 0"
          clearable
          style="width: 100%"
          @change="handleHomePathChange"
        >
          <el-option
            v-for="item in menuOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          >
            <span style="float: left">{{ item.label }}</span>
            <span style="float: right; color: var(--el-color-success); font-size: 12px">
              <el-icon><Select /></el-icon> 可访问
            </span>
          </el-option>
        </el-select>
      </el-form-item>

      <el-alert
        v-if="hasPermissionWarning"
        type="warning"
        :closable="false"
        style="margin-bottom: 15px"
      >
        <template #title>
          <el-icon><Warning /></el-icon> 权限已自动调整
        </template>
        原设置的家目录不在用户权限范围内，系统已自动调整为有权限的路径。
      </el-alert>

      <el-form-item label="状态" prop="status">
        <el-radio-group v-model="form.status">
          <el-radio label="active">启用</el-radio>
          <el-radio label="disabled">禁用</el-radio>
        </el-radio-group>
      </el-form-item>
    </FormDialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { useStore } from 'vuex'
import { Plus, Search, InfoFilled, QuestionFilled, Select, Warning } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { systemApi } from '@/api'
import { TablePage, FormDialog, TableActions, SwitchStatus } from '@/components'

const store = useStore()

// 状态管理
const userList = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const roleOptions = ref([])
const menuOptions = ref([])
const allMenus = ref([])
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const hasPermissionWarning = ref(false)

// 查询参数
const queryParams = reactive({
  username: '',
  status: ''
})

// 表单数据
const form = ref({
  id: null,
  username: '',
  password: '',
  nickname: '',
  email: '',
  homePath: '/',
  roleIds: [],
  status: 'active'
})

// 表单验证规则
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
  email: [{ type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }],
  roleIds: [{ required: true, message: '请至少选择一个角色', trigger: 'change', type: 'array' }]
}

// 计算属性
const displayUsers = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return userList.value.slice(start, end)
})

const homePathPlaceholder = computed(() => {
  if (form.value.roleIds.length === 0) {
    return '请先选择角色'
  } else if (menuOptions.value.length === 0) {
    return '所选角色无任何可访问的家目录'
  } else {
    return `请选择家目录（共 ${menuOptions.value.length} 个可选项）`
  }
})

// 数据获取
const fetchUsers = async () => {
  loading.value = true
  try {
    const [userRes, roleRes] = await Promise.all([
      systemApi.getUsers(),
      systemApi.getRoles()
    ])

    if (userRes.code === 200 && roleRes.code === 200) {
      const roles = roleRes.data || []
      let data = userRes.data

      // 前端过滤
      if (queryParams.username) {
        const keyword = queryParams.username.toLowerCase()
        data = data.filter(u =>
          u.username.toLowerCase().includes(keyword) ||
          u.nickname.toLowerCase().includes(keyword)
        )
      }
      if (queryParams.status) {
        data = data.filter(u => u.status === queryParams.status)
      }

      userList.value = data.map(user => ({
        ...user,
        roleNames: (user.roleIds || []).map(rid =>
          roles.find(r => r.id === rid)?.name
        ).filter(Boolean)
      }))
      total.value = userList.value.length
    }
  } catch (error) {
    console.error('Error fetching users:', error)
  } finally {
    loading.value = false
  }
}

const fetchRoles = async () => {
  try {
    const res = await systemApi.getRoles()
    if (res.code === 200) {
      roleOptions.value = res.data
    }
  } catch (error) {
    console.error('Error fetching roles:', error)
  }
}

const fetchMenus = async () => {
  try {
    const res = await systemApi.getMenus()
    if (res.code === 200) {
      allMenus.value = res.data
      menuOptions.value = res.data
        .filter(menu => !menu.parentId || menu.parentId === 0)
        .map(item => ({
          label: item.name,
          value: item.path
        }))
    }
  } catch (error) {
    console.error('Error fetching menus:', error)
  }
}

// 权限相关方法
const getUserAccessibleMenuIds = (selectedRoleIds) => {
  if (!selectedRoleIds?.length) return new Set()

  const menuIds = new Set()
  selectedRoleIds.forEach(roleId => {
    const role = roleOptions.value.find(r => r.id === roleId)
    if (role?.menuIds) {
      try {
        const ids = JSON.parse(role.menuIds)
        ids.forEach(id => menuIds.add(id))
      } catch (e) {
        console.error('解析角色菜单ID失败:', e)
      }
    }
  })
  return menuIds
}

const hasRolePathPermission = (path, selectedRoleIds) => {
  if (!selectedRoleIds?.length) return false

  const accessibleMenuIds = getUserAccessibleMenuIds(selectedRoleIds)
  if (accessibleMenuIds.size === 0) return false

  const findMenuByPath = (menus, targetPath) => {
    for (const menu of menus) {
      if (menu.path === targetPath) return menu
      if (menu.children) {
        const found = findMenuByPath(menu.children, targetPath)
        if (found) return found
      }
    }
    return null
  }

  const menu = findMenuByPath(allMenus.value, path)
  return menu && accessibleMenuIds.has(menu.id)
}

// 事件处理
const handleAdd = () => {
  form.value = {
    id: null,
    username: '',
    password: '',
    nickname: '',
    email: '',
    homePath: '',
    roleIds: [],
    status: 'active'
  }
  menuOptions.value = []
  hasPermissionWarning.value = false
  dialogVisible.value = true
}

const handleEdit = (row) => {
  form.value = {
    id: row.id,
    username: row.username,
    nickname: row.nickname,
    email: row.email,
    homePath: row.homePath || '',
    roleIds: [...(row.roleIds || [])],
    status: row.status
  }

  if (form.value.roleIds.length > 0) {
    handleRoleIdsChange(form.value.roleIds)
    const originalHomePath = row.homePath || ''
    if (originalHomePath && hasRolePathPermission(originalHomePath, form.value.roleIds)) {
      form.value.homePath = originalHomePath
    } else if (originalHomePath) {
      form.value.homePath = ''
    }
  } else {
    menuOptions.value = []
  }

  dialogVisible.value = true
}

const handleRoleIdsChange = (newRoleIds) => {
  if (!newRoleIds?.length) {
    menuOptions.value = []
    form.value.homePath = ''
    hasPermissionWarning.value = false
    return
  }

  const accessibleMenuIds = getUserAccessibleMenuIds(newRoleIds)
  const accessibleMenus = allMenus.value.filter(menu => {
    if (accessibleMenuIds.has(menu.id)) {
      return !menu.parentId || menu.parentId === 0
    }
    return false
  })

  if (accessibleMenus.length === 0) {
    menuOptions.value = []
    form.value.homePath = ''
    hasPermissionWarning.value = true
    return
  }

  menuOptions.value = accessibleMenus.map(item => ({
    label: item.name,
    value: item.path
  }))

  const currentHomePath = form.value.homePath
  if (currentHomePath && currentHomePath !== '/') {
    const hasPermission = hasRolePathPermission(currentHomePath, newRoleIds)
    hasPermissionWarning.value = !hasPermission
    if (!hasPermission) {
      form.value.homePath = ''
    }
  } else {
    hasPermissionWarning.value = false
  }
}

const handleHomePathChange = (newHomePath) => {
  if (!newHomePath || newHomePath === '/') {
    hasPermissionWarning.value = form.value.roleIds?.length > 0
    return
  }

  if (form.value.roleIds?.length > 0) {
    hasPermissionWarning.value = !hasRolePathPermission(newHomePath, form.value.roleIds)
  } else {
    hasPermissionWarning.value = true
  }
}

const handleResetPwd = (row) => {
  ElMessageBox.prompt(`请输入用户 "${row.username}" 的新密码`, '重置密码', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    inputPattern: /^.{6,20}$/,
    inputErrorMessage: '密码长度需在 6-20 位之间'
  }).then(async ({ value }) => {
    try {
      await systemApi.updateUser(row.id, { password: value })
      ElMessage.success('密码重置成功')
    } catch (error) {
      ElMessage.error('重置失败')
    }
  })
}

const submitForm = async () => {
  submitting.value = true
  try {
    const res = form.value.id
      ? await systemApi.updateUser(form.value.id, form.value)
      : await systemApi.addUser(form.value)

    if (res.code === 200) {
      ElMessage.success(form.value.id ? '更新成功' : '添加成功')
      dialogVisible.value = false
      fetchUsers()
      if (form.value.id === store.state.user?.id) {
        store.dispatch('fetchUserInfo')
      }
    } else {
      ElMessage.error(res.message || '操作失败')
    }
  } catch (error) {
    ElMessage.error('操作失败')
  } finally {
    submitting.value = false
  }
}

const handleStatusChange = async (row, val) => {
  try {
    await systemApi.updateUser(row.id, { status: val })
    ElMessage.success(`${val === 'active' ? '启用' : '禁用'}成功`)
    if (row.id === store.state.user?.id) {
      store.dispatch('fetchUserInfo')
    }
  } catch (error) {
    row.status = val === 'active' ? 'disabled' : 'active'
    ElMessage.error('操作失败')
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确定要删除用户 "${row.username}" 吗？`, '提示', {
    type: 'warning'
  }).then(async () => {
    try {
      await systemApi.deleteUser(row.id)
      ElMessage.success('删除成功')
      fetchUsers()
    } catch (error) {
      ElMessage.error('删除失败')
    }
  })
}

// 初始化
onMounted(() => {
  fetchUsers()
  fetchRoles()
  fetchMenus()
})
</script>

<style scoped>
.users-container {
  display: flex;
  flex-direction: column;
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-name {
  font-weight: 500;
}

.role-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.role-tag {
  margin: 2px 0;
}

.text-tertiary {
  color: var(--text-tertiary);
  font-size: 12px;
}

.text-accent {
  color: var(--accent) !important;
}

.text-accent:hover {
  color: var(--accent-hover) !important;
  text-decoration: underline;
}

.role-permission-hint {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 15px;
  margin-bottom: 15px;
  background-color: var(--el-color-info-light-9);
  border: 1px solid var(--el-color-info-light-5);
  border-radius: 4px;
  color: var(--el-color-info);
  font-size: 13px;
}

.role-permission-hint .el-icon {
  font-size: 16px;
}

.role-permission-hint strong {
  color: var(--el-color-primary);
  font-weight: 600;
}

.label-icon {
  margin-left: 4px;
  cursor: help;
  color: var(--el-color-info);
  font-size: 14px;
  vertical-align: middle;
}

.label-icon:hover {
  color: var(--el-color-primary);
}

.data-value {
  color: var(--text-primary);
}
</style>
