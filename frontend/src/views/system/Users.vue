<template>
    <div class="users-container">
        <header class="page-header">
            <div class="header-content">
                <h2 class="page-title">用户管理</h2>
                <p class="page-subtitle">管理系统登录账号及所属角色</p>
            </div>
            <div class="header-actions">
                <el-button type="primary" :icon="Plus" @click="handleAdd">新增用户</el-button>
            </div>
        </header>

        <el-card shadow="never" class="table-card" v-loading="loading">
            <el-table :data="userList" border style="width: 100%">
                <el-table-column label="用户名" width="180">
                    <template #default="{ row }">
                        <div style="display: flex; align-items: center; gap: 8px">
                            <el-avatar :size="32" :src="row.avatar">{{ row.nickname?.charAt(0).toUpperCase() }}</el-avatar>
                            <span>{{ row.username }}</span>
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
                <el-table-column prop="status" label="状态" width="100" align="center">
                    <template #default="{ row }">
                        <el-switch
                            v-model="row.status"
                            active-value="active"
                            inactive-value="disabled"
                            @change="(val) => handleStatusChange(row, val)"
                        />
                    </template>
                </el-table-column>
                <el-table-column prop="createdAt" label="创建时间" width="180" />
                <el-table-column label="操作" width="150" align="center">
                    <template #default="{ row }">
                        <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
                        <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
                    </template>
                </el-table-column>
            </el-table>
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
    import { ref, onMounted, nextTick } from 'vue'
    import { useStore } from 'vuex'
    import { Plus } from '@element-plus/icons-vue'
    import { ElMessage, ElMessageBox } from 'element-plus'
    import { systemApi } from '../../api/index.js'

    const store = useStore()
    const userList = ref([])
    const roleOptions = ref([])
    const loading = ref(false)
    const submitting = ref(false)
    const dialogVisible = ref(false)
    const formRef = ref(null)

    const form = ref({
        id: null,
        username: '',
        nickname: '',
        email: '',
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
                userList.value = userRes.data.map(user => ({
                    ...user,
                    roleNames: (user.roleIds || []).map(rid => roles.find(r => r.id === rid)?.name).filter(Boolean)
                }))
            }
        } catch (error) {
            console.error('Error fetching users:', error)
        } finally {
            loading.value = false
        }
    }

    const handleAdd = () => {
        form.value = { id: null, username: '', nickname: '', email: '', roleIds: [], status: 'active' }
        dialogVisible.value = true
        nextTick(() => formRef.value?.clearValidate())
    }

    const handleEdit = (row) => {
        form.value = {
            id: row.id,
            username: row.username,
            nickname: row.nickname,
            email: row.email,
            roleIds: [...(row.roleIds || [])],
            status: row.status
        }
        dialogVisible.value = true
        nextTick(() => formRef.value?.clearValidate())
    }

    const submitForm = async () => {
        if (!formRef.value) return
        await formRef.value.validate(async (valid) => {
            if (valid) {
                submitting.value = true
                try {
                    if (form.value.id) {
                        await systemApi.updateUser(form.value.id, form.value)
                        ElMessage.success('更新成功')
                    } else {
                        await systemApi.addUser(form.value)
                        ElMessage.success('添加成功')
                    }
                    dialogVisible.value = false
                    fetchUsers()
                    // 如果修改的是当前用户，刷新菜单和权限
                    if (form.value.id === store.state.user?.id) {
                        store.dispatch('fetchUserInfo')
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

    onMounted(() => {
        fetchUsers()
        fetchRoles()
    })
</script>

<style scoped>
    .users-container {
        display: flex;
        flex-direction: column;
        gap: 16px;
    }

    .page-header {
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        margin-bottom: 4px;
    }

    .table-card {
        border: none;
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
</style>
