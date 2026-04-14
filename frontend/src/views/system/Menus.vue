<template>
    <div class="menus-container">
        <header class="page-header">
            <div class="header-content">
                <h2 class="page-title">菜单管理</h2>
                <p class="page-subtitle">配置系统导航菜单及权限标识。</p>
            </div>
            <div class="header-actions">
                <el-button type="primary" :icon="Plus" @click="handleAdd">新增菜单</el-button>
            </div>
        </header>

        <el-card shadow="never" class="table-card" v-loading="loading">
            <el-table
                :data="menuList"
                row-key="id"
                border
                default-expand-all
                :tree-props="{ children: 'children' }"
            >
                <el-table-column prop="name" label="菜单名称" min-width="180" />
                <el-table-column prop="icon" label="图标" width="80" align="center">
                    <template #default="{ row }">
                        <el-icon v-if="row.icon"><component :is="getIcon(row.icon)" /></el-icon>
                    </template>
                </el-table-column>
                <el-table-column prop="sort" label="排序" width="120" align="center">
                    <template #default="{ row }">
                        <div class="sort-actions">
                            <el-button 
                                link 
                                :icon="Top" 
                                @click="handleMove(row, 'up')"
                                title="上移"
                            />
                            <span class="sort-value">{{ row.sort }}</span>
                            <el-button 
                                link 
                                :icon="Bottom" 
                                @click="handleMove(row, 'down')"
                                title="下移"
                            />
                        </div>
                    </template>
                </el-table-column>
                <el-table-column prop="status" label="状态" width="100" align="center">
                    <template #default="{ row }">
                        <el-switch
                            v-model="row.status"
                            :active-value="1"
                            :inactive-value="0"
                            @change="(val) => handleStatusChange(row, val)"
                        />
                    </template>
                </el-table-column>
                <el-table-column prop="path" label="路由路径" min-width="180" />
                <el-table-column prop="permission" label="权限标识" min-width="180">
                    <template #default="{ row }">
                        <el-tag size="small" type="info">{{ row.permission }}</el-tag>
                    </template>
                </el-table-column>
                <el-table-column label="操作" width="150" align="center">
                    <template #default="{ row }">
                        <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
                        <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
                    </template>
                </el-table-column>
            </el-table>
        </el-card>

        <!-- Edit Menu Dialog -->
        <el-dialog
            v-model="dialogVisible"
            :title="form.id ? '编辑菜单' : '新增菜单'"
            width="500px"
            destroy-on-close
        >
            <el-form
                ref="formRef"
                :model="form"
                :rules="rules"
                label-width="80px"
                label-position="top"
            >
                <el-form-item label="菜单名称" prop="name">
                    <el-input v-model="form.name" placeholder="请输入菜单名称" />
                </el-form-item>
                <el-row :gutter="20">
                    <el-col :span="12">
                        <el-form-item label="菜单图标" prop="icon">
                            <el-select v-model="form.icon" placeholder="请选择图标" style="width: 100%" filterable>
                                <el-option
                                    v-for="(item, key) in ElementPlusIconsVue"
                                    :key="key"
                                    :label="key"
                                    :value="key"
                                >
                                    <div style="display: flex; align-items: center; gap: 8px">
                                        <el-icon><component :is="item" /></el-icon>
                                        <span>{{ key }}</span>
                                    </div>
                                </el-option>
                            </el-select>
                        </el-form-item>
                    </el-col>
                    <el-col :span="12">
                        <el-form-item label="显示排序" prop="sort">
                            <el-input-number v-model="form.sort" :min="1" style="width: 100%" />
                        </el-form-item>
                    </el-col>
                </el-row>
                <el-form-item label="路由路径" prop="path">
                    <el-input v-model="form.path" placeholder="例如: /system/menus" />
                </el-form-item>
                <el-form-item label="权限标识" prop="permission">
                    <el-input v-model="form.permission" placeholder="例如: menu:system:menus" />
                </el-form-item>
                <el-form-item label="菜单状态">
                    <el-radio-group v-model="form.status">
                        <el-radio :label="1">启用</el-radio>
                        <el-radio :label="0">停用</el-radio>
                    </el-radio-group>
                </el-form-item>
            </el-form>
            <template #footer>
                <div class="dialog-footer">
                    <el-button @click="dialogVisible = false">取消</el-button>
                    <el-button type="primary" :loading="submitting" @click="submitForm">确定</el-button>
                </div>
            </template>
        </el-dialog>
    </div>
</template>

<script setup>
    import { ref, onMounted } from 'vue'
    import { useStore } from 'vuex'
    import * as ElementPlusIconsVue from '@element-plus/icons-vue'
    import { Plus, Top, Bottom } from '@element-plus/icons-vue'
    import { ElMessage, ElMessageBox } from 'element-plus'
    import { systemApi, loginApi } from '../../api/index.js'

    const store = useStore()
    const menuList = ref([])
    const loading = ref(false)
    const dialogVisible = ref(false)
    const submitting = ref(false)
    const formRef = ref(null)

    const form = ref({
        id: null,
        name: '',
        icon: '',
        path: '',
        permission: '',
        sort: 1,
        status: 1
    })

    const rules = {
        name: [{ required: true, message: '请输入菜单名称', trigger: 'blur' }],
        path: [{ required: true, message: '请输入路由路径', trigger: 'blur' }],
        permission: [{ required: true, message: '请输入权限标识', trigger: 'blur' }]
    }

    const handleAdd = () => {
        form.value = {
            id: null,
            name: '',
            icon: '',
            path: '',
            permission: '',
            sort: 1,
            status: 1
        }
        dialogVisible.value = true
    }

    const getIcon = (iconName) => {
        return ElementPlusIconsVue[iconName] || iconName
    }

    const fetchMenus = async () => {
        loading.value = true
        try {
            const res = await systemApi.getMenus()
            if (res.code === 200) {
                menuList.value = res.data
            }
        } catch (error) {
            console.error('Error fetching menus:', error)
        } finally {
            loading.value = false
        }
    }

    const handleEdit = (row) => {
        // 只提取需要的字段，避免带入 Vue 内部属性或循环引用
        form.value = {
            id: row.id,
            name: row.name,
            icon: row.icon,
            path: row.path,
            permission: row.permission,
            sort: row.sort,
            status: row.status
        }
        dialogVisible.value = true
    }

    const handleDelete = (row) => {
        ElMessageBox.confirm(
            `确定要删除菜单 "${row.name}" 吗？`,
            '警告',
            {
                confirmButtonText: '确定',
                cancelButtonText: '取消',
                type: 'warning',
            }
        ).then(async () => {
            try {
                await systemApi.deleteMenu(row.id)
                ElMessage.success('删除成功')
                fetchMenus()
                // 触发侧边栏更新
                const res = await loginApi.getUserInfo()
                if (res.code === 200) {
                    store.commit('SET_MENU_TREE', res.data.menuTree)
                    store.commit('SET_PERMISSIONS', res.data.permissions)
                }
            } catch (error) {
                ElMessage.error('删除失败')
            }
        }).catch(() => {})
    }

    const handleStatusChange = async (row, val) => {
        try {
            await systemApi.updateMenu(row.id, { status: val })
            ElMessage.success(`${val === 1 ? '启用' : '停用'}成功`)
            // 触发侧边栏更新
            const res = await loginApi.getUserInfo()
            if (res.code === 200) {
                store.commit('SET_MENU_TREE', res.data.menuTree)
                store.commit('SET_PERMISSIONS', res.data.permissions)
            }
        } catch (error) {
            row.status = val === 1 ? 0 : 1
            ElMessage.error('操作失败')
        }
    }

    const submitForm = async () => {
        if (!formRef.value) return
        await formRef.value.validate(async (valid) => {
            if (valid) {
                submitting.value = true
                try {
                    if (form.value.id) {
                        await systemApi.updateMenu(form.value.id, form.value)
                    } else {
                        await systemApi.addMenu(form.value)
                    }
                    ElMessage.success('保存成功')
                    dialogVisible.value = false
                    fetchMenus()
                    // 触发侧边栏更新
                    const res = await loginApi.getUserInfo()
                    if (res.code === 200) {
                        store.commit('SET_MENU_TREE', res.data.menuTree)
                        store.commit('SET_PERMISSIONS', res.data.permissions)
                    }
                } catch (error) {
                    console.error('Error updating menu:', error)
                    ElMessage.error('保存失败')
                } finally {
                    submitting.value = false
                }
            }
        })
    }

    const handleMove = async (row, direction) => {
        // 查找兄弟节点
        const findSiblings = (list, targetId) => {
            for (let i = 0; i < list.length; i++) {
                if (list[i].id === targetId) return list
                if (list[i].children) {
                    const found = findSiblings(list[i].children, targetId)
                    if (found) return found
                }
            }
            return null
        }

        const siblings = findSiblings(menuList.value, row.id)
        if (!siblings) return

        const index = siblings.findIndex(item => item.id === row.id)
        const targetIndex = direction === 'up' ? index - 1 : index + 1

        if (targetIndex < 0 || targetIndex >= siblings.length) {
            ElMessage.warning(direction === 'up' ? '已经是第一个了' : '已经是最后一个了')
            return
        }

        const targetRow = siblings[targetIndex]
        
        // 交换排序值
        const tempSort = row.sort
        try {
            // 乐观更新 UI
            const originalSort = row.sort
            const targetOriginalSort = targetRow.sort
            
            // 发送两次请求交换
            await Promise.all([
                systemApi.updateMenu(row.id, { sort: targetOriginalSort }),
                systemApi.updateMenu(targetRow.id, { sort: originalSort })
            ])
            
            ElMessage.success('排序更新成功')
            fetchMenus() // 重新获取排序后的列表
            // 触发侧边栏更新
            const res = await loginApi.getUserInfo()
            if (res.code === 200) {
                store.commit('SET_MENU_TREE', res.data.menuTree)
                store.commit('SET_PERMISSIONS', res.data.permissions)
            }
        } catch (error) {
            ElMessage.error('排序更新失败')
            fetchMenus()
        }
    }

    onMounted(fetchMenus)
</script>

<style scoped>
    .menus-container {
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

    .sort-actions {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 4px;
    }

    .sort-value {
        min-width: 24px;
        font-size: 12px;
        color: var(--text-secondary);
    }
</style>
