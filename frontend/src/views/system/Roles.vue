<template>
    <div class="roles-container">
        <el-card shadow="never" class="table-card" v-loading="loading">
            <template #header>
                <div class="card-header-content">
                    <div class="header-left">
                        <div style="display: flex; align-items: center; gap: 12px">
                            <h2 class="page-title">角色管理</h2>
                            <span class="accent-dot"></span>
                        </div>
                        <p class="page-subtitle">管理系统用户角色及权限分配</p>
                    </div>
                    <div class="header-right">
                        <el-space>
                            <el-input
                                v-model="queryParams.name"
                                placeholder="搜索角色名称/标识"
                                :prefix-icon="Search"
                                clearable
                                style="width: 220px"
                                @keyup.enter="handleSearch"
                            />
                            <el-button type="accent" :icon="Plus" @click="handleAdd">新增角色</el-button>
                        </el-space>
                    </div>
                </div>
            </template>
            <el-table :data="displayRoles" border style="width: 100%">
                <el-table-column prop="name" label="角色名称" width="150" />
                <el-table-column prop="code" label="角色标识" width="150">
                    <template #default="{ row }">
                        <el-tag size="small" class="permission-tag">{{ row.code }}</el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="userNames" label="关联用户" min-width="200">
                    <template #default="{ row }">
                        <div class="user-tags">
                            <el-tag 
                                v-for="userName in row.userNames" 
                                :key="userName" 
                                size="small" 
                                type="warning"
                                class="user-tag"
                            >
                                {{ userName }}
                            </el-tag>
                            <span v-if="!row.userNames || row.userNames.length === 0" class="text-tertiary">暂无用户</span>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column prop="description" label="描述" min-width="200" />
                <el-table-column prop="status" label="状态" width="100" align="center">
                    <template #default="{ row }">
                        <el-switch
                            v-model="row.status"
                            :active-value="1"
                            :inactive-value="0"
                            :disabled="row.code === 'admin'"
                            @change="(val) => handleStatusChange(row, val)"
                        />
                    </template>
                </el-table-column>
                <el-table-column label="操作" width="250" align="center">
                    <template #default="{ row }">
                        <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
                        <el-button link class="text-accent" @click="handlePermission(row)">权限设置</el-button>
                        <el-button 
                            link 
                            type="danger" 
                            :disabled="row.code === 'admin'"
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

        <!-- Role Form Dialog -->
        <el-dialog
            v-model="dialogVisible"
            :title="form.id ? '编辑角色' : '新增角色'"
            width="500px"
        >
            <el-form :model="form" :rules="rules" ref="formRef" label-width="80px" label-position="top">
                <el-form-item label="角色名称" prop="name">
                    <el-input v-model="form.name" placeholder="请输入角色名称" />
                </el-form-item>
                <el-form-item label="角色标识" prop="code">
                    <el-input v-model="form.code" placeholder="请输入角色标识" :disabled="!!form.id" />
                </el-form-item>
                <el-form-item label="描述" prop="description">
                    <el-input v-model="form.description" type="textarea" placeholder="请输入角色描述" />
                </el-form-item>
                <el-form-item label="状态" prop="status">
                    <el-radio-group v-model="form.status" :disabled="form.code === 'admin'">
                        <el-radio :label="1">启用</el-radio>
                        <el-radio :label="0">禁用</el-radio>
                    </el-radio-group>
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="dialogVisible = false">取消</el-button>
                <el-button type="primary" @click="submitForm" :loading="submitting">确定</el-button>
            </template>
        </el-dialog>

        <!-- Permission Dialog -->
        <el-dialog
            v-model="permissionDialogVisible"
            title="权限设置"
            width="400px"
        >
            <div class="permission-tree-container">
                <el-tree
                    ref="treeRef"
                    :data="menuTree"
                    show-checkbox
                    node-key="id"
                    default-expand-all
                    :props="{ label: 'name', children: 'children' }"
                />
            </div>
            <template #footer>
                <el-button @click="permissionDialogVisible = false">取消</el-button>
                <el-button type="primary" @click="submitPermission" :loading="submitting">保存</el-button>
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
    const roleList = ref([])
    const currentPage = ref(1)
    const pageSize = ref(10)
    const total = ref(0)
    const menuTree = ref([])
    const loading = ref(false)
    const submitting = ref(false)
    const dialogVisible = ref(false)
    const permissionDialogVisible = ref(false)
    const formRef = ref(null)
    const treeRef = ref(null)
    const displayRoles = computed(() => {
        const start = (currentPage.value - 1) * pageSize.value
        const end = start + pageSize.value
        return roleList.value.slice(start, end)
    })

    const queryParams = reactive({
        name: ''
    })

    const form = ref({
        id: null,
        name: '',
        code: '',
        description: '',
        status: 1
    })

    const rules = {
        name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
        code: [{ required: true, message: '请输入角色标识', trigger: 'blur' }]
    }

    const currentRole = ref(null)

    const fetchRoles = async () => {
        loading.value = true
        try {
            const [roleRes, userRes] = await Promise.all([
                systemApi.getRoles(),
                systemApi.getUsers()
            ])
            
            if (roleRes.code === 200 && userRes.code === 200) {
                const users = userRes.data || []
                let data = roleRes.data
                
                // 前端模拟搜索过滤
                if (queryParams.name) {
                    const keyword = queryParams.name.toLowerCase()
                    data = data.filter(r => 
                        r.name.toLowerCase().includes(keyword) || 
                        r.code.toLowerCase().includes(keyword)
                    )
                }

                roleList.value = data.map(role => ({
                    ...role,
                    userNames: users
                        .filter(u => u.roleIds && u.roleIds.includes(role.id))
                        .map(u => u.nickname || u.username)
                }))
                total.value = roleList.value.length
            }
        } catch (error) {
            console.error('Error fetching roles:', error)
        } finally {
            loading.value = false
        }
    }

    const handleSearch = () => {
        currentPage.value = 1
        fetchRoles()
    }

    const fetchMenus = async () => {
        try {
            const res = await systemApi.getMenus()
            if (res.code === 200) {
                menuTree.value = res.data
            }
        } catch (error) {
            console.error('Error fetching menus:', error)
        }
    }

    const handleAdd = () => {
        form.value = { id: null, name: '', code: '', description: '', status: 1 }
        dialogVisible.value = true
        nextTick(() => formRef.value?.clearValidate())
    }

    const handleEdit = (row) => {
        form.value = {
            id: row.id,
            name: row.name,
            code: row.code,
            description: row.description,
            status: row.status ?? 1
        }
        dialogVisible.value = true
        nextTick(() => formRef.value?.clearValidate())
    }

    const handleStatusChange = async (row, val) => {
        try {
            await systemApi.updateRole(row.id, { status: val })
            ElMessage.success(`${val === 1 ? '启用' : '禁用'}成功`)
        } catch (error) {
            row.status = val === 1 ? 0 : 1
            ElMessage.error('状态更新失败')
        }
    }

    const submitForm = async () => {
        if (!formRef.value) return
        await formRef.value.validate(async (valid) => {
            if (valid) {
                submitting.value = true
                try {
                    if (form.value.id) {
                        await systemApi.updateRole(form.value.id, form.value)
                        ElMessage.success('更新成功')
                    } else {
                        await systemApi.addRole(form.value)
                        ElMessage.success('添加成功')
                    }
                    dialogVisible.value = false
                    fetchRoles()
                } catch (error) {
                    ElMessage.error('操作失败')
                } finally {
                    submitting.value = false
                }
            }
        })
    }

    const handleDelete = (row) => {
        ElMessageBox.confirm(`确定要删除角色 "${row.name}" 吗？`, '提示', {
            type: 'warning'
        }).then(async () => {
            try {
                await systemApi.deleteRole(row.id)
                ElMessage.success('删除成功')
                fetchRoles()
            } catch (error) {
                ElMessage.error('删除失败')
            }
        })
    }

    const handleSizeChange = (val) => {
        pageSize.value = val
        fetchRoles()
    }

    const handleCurrentChange = (val) => {
        currentPage.value = val
        fetchRoles()
    }

    const handlePermission = (row) => {
        currentRole.value = row
        permissionDialogVisible.value = true
        nextTick(() => {
            // 设置已选中的节点
            // 注意：Element Plus 的 tree 组件设置选中时，如果父节点被选中，子节点会自动选中
            // 这里我们通常只设置叶子节点，或者根据业务需求设置
            treeRef.value?.setCheckedKeys(row.menuIds || [])
        })
    }

    const submitPermission = async () => {
        if (!currentRole.value) return
        submitting.value = true
        try {
            const checkedKeys = treeRef.value?.getCheckedKeys() || []
            const halfCheckedKeys = treeRef.value?.getHalfCheckedKeys() || []
            const allKeys = [...checkedKeys, ...halfCheckedKeys]
            
            await systemApi.updateRole(currentRole.value.id, { menuIds: allKeys })
            ElMessage.success('权限更新成功')
            permissionDialogVisible.value = false
            fetchRoles()
            // 如果修改的是当前用户的角色权限，刷新菜单
            store.dispatch('fetchUserInfo')
        } catch (error) {
            ElMessage.error('权限更新失败')
        } finally {
            submitting.value = false
        }
    }

    onMounted(() => {
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

    .roles-container {
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

    .permission-tree-container {
        max-height: 400px;
        overflow-y: auto;
        padding: 8px;
        border: 1px solid var(--border);
        border-radius: 4px;
    }

    .user-tags {
        display: flex;
        flex-wrap: wrap;
        gap: 4px;
    }

    .user-tag {
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
