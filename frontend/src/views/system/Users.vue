<template>
    <div class="users-container">
        <el-card shadow="never" class="table-card" v-loading="loading">
            <template #header>
                <div class="card-header-content">
                    <div class="header-left">
                        <div style="display: flex; align-items: center; gap: 12px">
                            <h2 class="page-title">用户管理</h2>
                            <span class="accent-dot"></span>
                        </div>
                        <p class="page-subtitle">管理系统登录账号及所属角色</p>
                    </div>
                    <div class="header-right">
                        <el-space>
                            <el-input v-model="queryParams.username" placeholder="搜索用户名/昵称" :prefix-icon="Search" clearable style="width: 200px" @keyup.enter="fetchUsers" />
                            <el-select v-model="queryParams.status" placeholder="状态" clearable style="width: 100px" @change="fetchUsers">
                                <el-option label="启用" value="active" />
                                <el-option label="禁用" value="disabled" />
                            </el-select>
                            <el-button type="primary" :icon="Plus" @click="handleAdd">新增用户</el-button>
                        </el-space>
                    </div>
                </div>
            </template>
            <el-table :data="displayUsers" border style="width: 100%">
                <el-table-column label="用户名" width="140">
                    <template #default="{ row }">
                        <div style="display: flex; align-items: center; gap: 8px">
                            <el-avatar :size="32" :src="row.avatar">{{ row.nickname?.charAt(0).toUpperCase() }}</el-avatar>
                            <span class="data-value user-name">{{ row.username }}</span>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column prop="nickname" label="昵称" width="150" />
                <el-table-column prop="roleNames" label="分配角色" min-width="200">
                    <template #default="{ row }">
                        <div class="role-tags">
                            <el-tag v-for="roleName in row.roleNames" :key="roleName" size="small" type="success" class="role-tag">
                                {{ roleName }}
                            </el-tag>
                            <span v-if="!row.roleNames || row.roleNames.length === 0" class="text-tertiary">未分配</span>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column prop="email" label="邮箱" min-width="200" />
                <el-table-column prop="homePath" label="家目录" width="150" />
                <el-table-column prop="status" label="状态" width="100" align="center">
                    <template #default="{ row }">
                        <el-switch v-model="row.status" active-value="active" inactive-value="disabled" :disabled="row.id === store.state.user?.id" @change="(val) => handleStatusChange(row, val)" />
                    </template>
                </el-table-column>
                <el-table-column prop="createdAt" label="创建时间" width="180" />
                <el-table-column label="操作" width="200" align="center">
                    <template #default="{ row }">
                        <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
                        <el-button link class="text-accent" @click="handleResetPwd(row)">重置密码</el-button>
                        <el-button link type="danger" :disabled="row.id === store.state.user?.id" @click="handleDelete(row)">删除</el-button>
                    </template>
                </el-table-column>
            </el-table>

            <div class="pagination-container">
                <el-pagination v-model:current-page="currentPage" v-model:page-size="pageSize" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" :total="total" @size-change="handleSizeChange" @current-change="handleCurrentChange" />
            </div>
        </el-card>

        <!-- User Form Dialog -->
        <el-dialog v-model="dialogVisible" :title="form.id ? '编辑用户' : '新增用户'" width="500px">
            <el-form :model="form" :rules="rules" ref="formRef" label-width="80px" label-position="top">
                <el-row :gutter="20">
                    <el-col :span="12">
                        <el-form-item label="用户名" prop="username">
                            <el-input v-model="form.username" placeholder="请输入用户名" :disabled="!!form.id" />
                        </el-form-item>
                    </el-col>
                    <el-col :span="12">
                        <el-form-item label="密码" prop="password" v-if="!form.id">
                            <el-input v-model="form.password" type="password" placeholder="请输入密码" show-password />
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
                    <el-select v-model="form.roleIds" placeholder="请先选择角色，系统将根据角色权限显示可用的家目录" style="width: 100%" multiple collapse-tags collapse-tags-tooltip @change="handleRoleIdsChange">
                        <el-option v-for="role in roleOptions" :key="role.id" :label="role.name" :value="role.id" />
                    </el-select>
                </el-form-item>

                <!-- 角色权限提示 -->
                <div v-if="form.roleIds.length > 0 && menuOptions.length > 0" class="role-permission-hint">
                    <el-icon>
                        <InfoFilled />
                    </el-icon>
                    <span>已为所选角色配置了 <strong>{{ menuOptions.length }}</strong> 个可访问的家目录，请从下方选择</span>
                </div>

                <el-form-item label="家目录" prop="homePath">
                    <el-select v-model="form.homePath" :placeholder="homePathPlaceholder" :disabled="form.roleIds.length === 0" clearable style="width: 100%" @change="handleHomePathChange">
                        <template #label>
                            <span>家目录</span>
                            <el-tooltip v-if="form.roleIds.length > 0" content="系统已根据您选择的角色权限，仅显示可访问的路径" placement="top">
                                <el-icon class="label-icon">
                                    <QuestionFilled />
                                </el-icon>
                            </el-tooltip>
                        </template>
                        <el-option v-for="item in menuOptions" :key="item.value" :label="item.label" :value="item.value">
                            <span style="float: left">{{ item.label }}</span>
                            <span style="float: right; color: var(--el-color-success); font-size: 12px">
                                <el-icon><Select /></el-icon> 可访问
                            </span>
                        </el-option>
                    </el-select>
                </el-form-item>

                <!-- 权限冲突警告 -->
                <el-alert v-if="hasPermissionWarning" type="warning" :closable="false" style="margin-bottom: 15px">
                    <template #title>
                        <el-icon>
                            <Warning />
                        </el-icon> 权限已自动调整
                    </template>
                    原设置的家目录不在用户权限范围内，系统已自动调整为有权限的路径。
                </el-alert>
                <el-form-item label="状态" prop="status">
                    <el-radio-group v-model="form.status">
                        <el-radio label="active">启用</el-radio>
                        <el-radio label="disabled">禁用</el-radio>
                    </el-radio-group>
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="dialogVisible = false">取消</el-button>
                <el-button type="primary" @click="submitForm" :loading="submitting">确定</el-button>
            </template>
        </el-dialog>
    </div>
</template>

<script setup>
    import { ref, reactive, computed, onMounted, nextTick } from 'vue'
    import { useStore } from 'vuex'
    import { Plus, Search, Warning, InfoFilled, QuestionFilled, Select } from '@element-plus/icons-vue'
    import { ElMessage, ElMessageBox } from 'element-plus'
    import { systemApi } from '../../api/index.js'

    const store = useStore()
    const userList = ref([])
    const currentPage = ref(1)
    const pageSize = ref(10)
    const total = ref(0)
    const roleOptions = ref([])
    const menuOptions = ref([])
    const allMenus = ref([]) // 存储所有菜单数据，用于权限检查
    const loading = ref(false)
    const submitting = ref(false)
    const dialogVisible = ref(false)
    const formRef = ref(null)
    const hasPermissionWarning = ref(false) // 是否有权限冲突警告

    // 家目录选择器的占位符文本
    const homePathPlaceholder = computed(() => {
        if (form.value.roleIds.length === 0) {
            return '请先选择角色'
        } else if (menuOptions.value.length === 0) {
            return '所选角色无任何可访问的家目录'
        } else {
            return `请选择家目录（共 ${menuOptions.value.length} 个可选项）`
        }
    })

    const displayUsers = computed(() => {
        const start = (currentPage.value - 1) * pageSize.value
        const end = start + pageSize.value
        return userList.value.slice(start, end)
    })

    const queryParams = reactive({
        username: '',
        status: ''
    })

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

    const rules = {
        username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
        password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
        nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
        email: [
            { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
        ],
        roleIds: [{ required: true, message: '请至少选择一个角色', trigger: 'change', type: 'array' }]
    }

    const fetchMenus = async () => {
        try {
            const res = await systemApi.getMenus()
            if (res.code === 200) {
                // 保存所有菜单数据用于权限检查
                allMenus.value = res.data
                // 初始化家目录选项（显示所有一级菜单）
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

                // 前端模拟搜索过滤
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
                    roleNames: (user.roleIds || []).map(rid => roles.find(r => r.id === rid)?.name).filter(Boolean)
                }))
                total.value = userList.value.length
            }
        } catch (error) {
            console.error('Error fetching users:', error)
        } finally {
            loading.value = false
        }
    }

    const handleAdd = () => {
        // 重置表单为初始状态，不预填任何信息
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
        // 清空家目录选项和警告状态
        menuOptions.value = []
        hasPermissionWarning.value = false
        dialogVisible.value = true
        nextTick(() => formRef.value?.clearValidate())
    }

    // 获取用户有权限访问的菜单ID集合
    const getUserAccessibleMenuIds = (selectedRoleIds) => {
        if (!selectedRoleIds || selectedRoleIds.length === 0) return new Set()

        const menuIds = new Set()
        selectedRoleIds.forEach(roleId => {
            const role = roleOptions.value.find(r => r.id === roleId)
            if (role && role.menuIds) {
                // role.menuIds 是 JSON 字符串，需要解析
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

    // 检查指定角色是否有权访问某路径（不依赖当前登录用户权限）
    const hasRolePathPermission = (path, selectedRoleIds) => {
        if (!selectedRoleIds || selectedRoleIds.length === 0) return false

        // 获取这些角色关联的所有菜单权限
        const accessibleMenuIds = getUserAccessibleMenuIds(selectedRoleIds)

        // 如果没有任何菜单权限，返回false
        if (accessibleMenuIds.size === 0) return false

        // 查找对应的菜单（递归查找）
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

    // 处理角色选择变化
    const handleRoleIdsChange = (newRoleIds) => {
        if (!newRoleIds || newRoleIds.length === 0) {
            // 没有选择角色时，清空家目录选项并清空当前选择
            menuOptions.value = []
            form.value.homePath = ''
            hasPermissionWarning.value = false
            return
        }

        // 获取用户有权限访问的菜单
        const accessibleMenuIds = getUserAccessibleMenuIds(newRoleIds)

        // 过滤出用户有权限的一级菜单
        const accessibleMenus = allMenus.value.filter(menu => {
            if (accessibleMenuIds.has(menu.id)) {
                // 只显示一级菜单（没有父菜单或父菜单ID为0）
                return !menu.parentId || menu.parentId === 0
            }
            return false
        })

        if (accessibleMenus.length === 0) {
            // 如果没有任何可访问的菜单，清空选项并警告
            menuOptions.value = []
            form.value.homePath = ''
            hasPermissionWarning.value = true
            return
        }

        // 更新家目录选项
        menuOptions.value = accessibleMenus.map(item => ({
            label: item.name,
            value: item.path
        }))

        // 检查当前家目录是否在权限范围内
        const currentHomePath = form.value.homePath
        if (currentHomePath && currentHomePath !== '/') {
            const hasPermission = hasRolePathPermission(currentHomePath, newRoleIds)
            hasPermissionWarning.value = !hasPermission

            if (!hasPermission) {
                // 如果当前选择的家目录无权限，清空让用户重新选择
                form.value.homePath = ''
            }
        } else {
            // 如果当前没有选择家目录，不自动选择，让用户自己决定
            hasPermissionWarning.value = false
        }
    }

    // 处理家目录变化
    const handleHomePathChange = (newHomePath) => {
        if (!newHomePath || newHomePath === '/') {
            // 根路径需要角色权限验证
            if (form.value.roleIds && form.value.roleIds.length > 0) {
                hasPermissionWarning.value = true
            } else {
                hasPermissionWarning.value = false
            }
            return
        }

        if (form.value.roleIds && form.value.roleIds.length > 0) {
            const hasPermission = hasRolePathPermission(newHomePath, form.value.roleIds)
            hasPermissionWarning.value = !hasPermission
        } else {
            // 没有选择角色时，任何家目录都无效
            hasPermissionWarning.value = true
        }
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

        // 如果用户有角色，触发角色变化逻辑以更新可用家目录选项
        if (form.value.roleIds.length > 0) {
            handleRoleIdsChange(form.value.roleIds)
            // 重新设置用户原有的家目录（如果有权限的话）
            const originalHomePath = row.homePath || ''
            if (originalHomePath && hasRolePathPermission(originalHomePath, form.value.roleIds)) {
                form.value.homePath = originalHomePath
            } else if (originalHomePath) {
                // 如果原有家目录无权限，清空让用户重新选择
                form.value.homePath = ''
            }
        } else {
            // 没有角色时，清空家目录选项
            menuOptions.value = []
        }

        dialogVisible.value = true
        nextTick(() => formRef.value?.clearValidate())
    }

    const handleResetPwd = (row) => {
        ElMessageBox.prompt(`请输入用户 "${row.username}" 的新密码`, '重置密码', {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            inputPattern: /^.{6,20}$/,
            inputErrorMessage: '密码长度需在 6-20 位之间'
        }).then(async ({ value }) => {
            try {
                // 模拟重置密码接口
                await systemApi.updateUser(row.id, { password: value })
                ElMessage.success('密码重置成功')
            } catch (error) {
                ElMessage.error('重置失败')
            }
        })
    }

    const submitForm = async () => {
        if (!formRef.value) return
        await formRef.value.validate(async (valid) => {
            if (valid) {
                submitting.value = true
                try {
                    let res
                    if (form.value.id) {
                        res = await systemApi.updateUser(form.value.id, form.value)
                    } else {
                        res = await systemApi.addUser(form.value)
                    }

                    if (res.code === 200) {
                        ElMessage.success(form.value.id ? '更新成功' : '添加成功')
                        dialogVisible.value = false
                        fetchUsers()
                        // 如果修改的是当前用户，刷新菜单和权限
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
        })
    }

    const handleStatusChange = async (row, val) => {
        try {
            await systemApi.updateUser(row.id, { status: val })
            ElMessage.success(`${val === 'active' ? '启用' : '禁用'}成功`)
            // 如果修改的是当前用户，刷新状态
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

    const handleSizeChange = (val) => {
        pageSize.value = val
        fetchUsers()
    }

    const handleCurrentChange = (val) => {
        currentPage.value = val
        fetchUsers()
    }

    onMounted(() => {
        fetchUsers()
        fetchRoles()
        fetchMenus()
    })
</script>

<style scoped>
    .text-accent {
      color: var(--accent) !important;
    }

    .text-accent:hover {
      color: var(--accent-hover) !important;
      text-decoration: underline;
    }

    .users-container {
      display: flex;
      flex-direction: column;
    }

    .card-header-content {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .page-title {
      margin: 0;
      font-size: 18px;
      font-weight: 600;
    }

    .page-subtitle {
      margin: 4px 0 0;
      font-size: 13px;
      color: var(--text-tertiary);
    }

    .table-card {
      border: 1px solid var(--border);
      border-radius: 8px;
    }

    :deep(.el-card__header) {
      padding: 12px 20px;
      border-bottom: 1px solid var(--border);
      background-color: var(--bg-primary);
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

    .pagination-container {
      margin-top: 20px;
      display: flex;
      justify-content: flex-end;
      padding: 0 20px 20px;
    }

    /* 角色权限提示样式 */
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

    /* 家目录标签图标样式 */
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
</style>
