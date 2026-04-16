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
                            <el-input
                                v-model="queryParams.username"
                                placeholder="搜索用户名/昵称"
                                :prefix-icon="Search"
                                clearable
                                style="width: 200px"
                                @keyup.enter="handleSearch"
                            />
                            <el-select v-model="queryParams.status" placeholder="状态" clearable style="width: 100px" @change="handleSearch">
                                <el-option label="启用" value="active" />
                                <el-option label="禁用" value="disabled" />
                            </el-select>
                            <el-button type="accent" :icon="Plus" @click="handleAdd">新增用户</el-button>
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
                <el-table-column prop="roleNames" label="所属角色" min-width="200">
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
                            <span v-if="!row.roleNames || row.roleNames.length === 0" class="text-tertiary">未分配</span>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column prop="email" label="邮箱" min-width="200" />
                <el-table-column prop="homePath" label="家目录" width="150" />
                <el-table-column prop="status" label="状态" width="100" align="center">
                    <template #default="{ row }">
                        <el-switch
                            v-model="row.status"
                            active-value="active"
                            inactive-value="disabled"
                            :disabled="row.id === store.state.user?.id"
                            @change="(val) => handleStatusChange(row, val)"
                        />
                    </template>
                </el-table-column>
                <el-table-column prop="createdAt" label="创建时间" width="180" />
                <el-table-column label="操作" width="200" align="center">
                    <template #default="{ row }">
                        <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
                        <el-button link class="text-accent" @click="handleResetPwd(row)">重置密码</el-button>
                        <el-button 
                            link 
                            type="danger" 
                            :disabled="row.id === store.state.user?.id"
                            @click="handleDelete(row)"
                        >删除</el-button>
                    </template>
                </el-table-column>
            </el-table>
            
            <div class="pagination-container">
                <el-pagination
                    v-model:current-page="currentPage"
                    v-model:page-size="pageSize"
                    :page-sizes="[10, 20, 50, 100]"
                    layout="total, sizes, prev, pager, next, jumper"
                    :total="total"
                    @size-change="handleSizeChange"
                    @current-change="handleCurrentChange"
                />
            </div>
        </el-card>

        <!-- User Form Dialog -->
        <el-dialog
            v-model="dialogVisible"
            :title="form.id ? '编辑用户' : '新增用户'"
            width="500px"
        >
            <el-form :model="form" :rules="rules" ref="formRef" label-width="80px" label-position="top">
                <el-row :gutter="20">
                    <el-col :span="12">
                        <el-form-item label="用户名" prop="username">
                            <el-input v-model="form.username" placeholder="请输入用户名" :disabled="!!form.id" />
                        </el-form-item>
                    </el-col>
                    <el-col :span="12">
                        <el-form-item label="昵称" prop="nickname">
                            <el-input v-model="form.nickname" placeholder="请输入昵称" />
                        </el-form-item>
                    </el-col>
                </el-row>
                <el-form-item label="邮箱" prop="email">
                    <el-input v-model="form.email" placeholder="请输入邮箱" />
                </el-form-item>
                <el-form-item label="家目录" prop="homePath">
                    <el-select v-model="form.homePath" placeholder="请选择家目录" clearable style="width: 100%">
                        <el-option
                            v-for="item in menuOptions"
                            :key="item.value"
                            :label="item.label"
                            :value="item.value"
                        />
                    </el-select>
                </el-form-item>
                <el-form-item label="所属角色" prop="roleIds">
                    <el-select 
                        v-model="form.roleIds" 
                        placeholder="请选择角色" 
                        style="width: 100%"
                        multiple
                        collapse-tags
                        collapse-tags-tooltip
                    >
                        <el-option
                            v-for="role in roleOptions"
                            :key="role.id"
                            :label="role.name"
                            :value="role.id"
                        />
                    </el-select>
                </el-form-item>
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
    import { Plus, Search } from '@element-plus/icons-vue'
    import { ElMessage, ElMessageBox } from 'element-plus'
    import { systemApi } from '../../api/index.js'

    const store = useStore()
    const userList = ref([])
    const currentPage = ref(1)
    const pageSize = ref(10)
    const total = ref(0)
    const roleOptions = ref([])
    const menuOptions = ref([])
    const loading = ref(false)
    const submitting = ref(false)
    const dialogVisible = ref(false)
    const formRef = ref(null)
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
        nickname: '',
        email: '',
        homePath: '/',
        roleIds: [],
        status: 'active'
    })

    const rules = {
        username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
        nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
        email: [
            { required: true, message: '请输入邮箱', trigger: 'blur' },
            { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
        ],
        roleIds: [{ required: true, message: '请至少选择一个角色', trigger: 'change', type: 'array' }]
    }

    const fetchMenus = async () => {
        try {
            const res = await systemApi.getMenus()
            if (res.code === 200) {
                // 只取一级菜单
                menuOptions.value = res.data.map(item => ({
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

    const handleSearch = () => {
        currentPage.value = 1
        fetchUsers()
    }

    const handleAdd = () => {
        form.value = { id: null, username: '', nickname: '', email: '', homePath: '/', roleIds: [], status: 'active' }
        dialogVisible.value = true
        nextTick(() => formRef.value?.clearValidate())
    }

    const handleEdit = (row) => {
        form.value = {
            id: row.id,
            username: row.username,
            nickname: row.nickname,
            email: row.email,
            homePath: row.homePath || '/',
            roleIds: [...(row.roleIds || [])],
            status: row.status
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
</style>
