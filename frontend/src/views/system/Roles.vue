<template>
    <div class="roles-container">
        <header class="page-header">
            <div class="header-content">
                <h2 class="page-title">角色管理</h2>
                <p class="page-subtitle">管理系统用户角色及权限分配</p>
            </div>
            <div class="header-actions">
                <el-button type="primary" :icon="Plus" @click="handleAdd">新增角色</el-button>
            </div>
        </header>

        <el-card shadow="never" class="table-card" v-loading="loading">
            <el-table :data="roleList" border style="width: 100%">
                <el-table-column prop="name" label="角色名称" width="150" />
                <el-table-column prop="code" label="角色标识" width="150">
                    <template #default="{ row }">
                        <el-tag size="small">{{ row.code }}</el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="userNames" label="关联用户" min-width="200">
                    <template #default="{ row }">
                        <div class="user-tags">
                            <el-tag 
                                v-for="userName in row.userNames" 
                                :key="userName" 
                                size="small" 
                                type="info"
                                class="user-tag"
                            >
                                {{ userName }}
                            </el-tag>
                            <span v-if="!row.userNames || row.userNames.length === 0" class="text-tertiary">暂无用户</span>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column prop="description" label="描述" min-width="200" />
                <el-table-column label="操作" width="250" align="center">
                    <template #default="{ row }">
                        <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
                        <el-button link type="primary" @click="handlePermission(row)">权限设置</el-button>
                        <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
                    </template>
                </el-table-column>
            </el-table>
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
    import { ref, onMounted, nextTick } from 'vue'
    import { useStore } from 'vuex'
    import { Plus } from '@element-plus/icons-vue'
    import { ElMessage, ElMessageBox } from 'element-plus'
    import { systemApi } from '../../api/index.js'

    const store = useStore()
    const roleList = ref([])
    const menuTree = ref([])
    const loading = ref(false)
    const submitting = ref(false)
    const dialogVisible = ref(false)
    const permissionDialogVisible = ref(false)
    const formRef = ref(null)
    const treeRef = ref(null)

    const form = ref({
        id: null,
        name: '',
        code: '',
        description: ''
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
                roleList.value = roleRes.data.map(role => ({
                    ...role,
                    userNames: users
                        .filter(u => u.roleIds && u.roleIds.includes(role.id))
                        .map(u => u.nickname || u.username)
                }))
            }
        } catch (error) {
            console.error('Error fetching roles:', error)
        } finally {
            loading.value = false
        }
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
        form.value = { id: null, name: '', code: '', description: '' }
        dialogVisible.value = true
        nextTick(() => formRef.value?.clearValidate())
    }

    const handleEdit = (row) => {
        form.value = {
            id: row.id,
            name: row.name,
            code: row.code,
            description: row.description
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
    .roles-container {
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
</style>
