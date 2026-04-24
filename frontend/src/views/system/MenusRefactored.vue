<template>
  <div class="menus-container">
    <TablePage
      title="菜单管理"
      subtitle="配置系统导航菜单及权限标识"
      :loading="loading"
      :total="total"
      :show-pagination="false"
    >
      <template #actions>
        <el-space>
          <el-input
            v-model="filterText"
            placeholder="搜索菜单名称"
            :prefix-icon="Search"
            clearable
            style="width: 200px"
          />
          <el-button type="primary" :icon="Plus" @click="handleAdd">
            新增菜单
          </el-button>
        </el-space>
      </template>

      <el-table
        :data="filteredMenuList"
        row-key="id"
        border
        default-expand-all
        :tree-props="{ children: 'children' }"
      >
        <el-table-column prop="name" label="菜单名称" width="160">
          <template #default="{ row }">
            <span class="menu-name">{{ row.name }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="icon" label="图标" width="70" align="center">
          <template #default="{ row }">
            <el-icon v-if="row.icon">
              <component :is="getIcon(row.icon)" />
            </el-icon>
          </template>
        </el-table-column>

        <el-table-column prop="sort" label="排序" width="100" align="center">
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

        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <SwitchStatus
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              @change="(val) => handleStatusChange(row, val)"
            />
          </template>
        </el-table-column>

        <el-table-column prop="path" label="路由路径" min-width="150" />

        <el-table-column prop="permission" label="权限标识" min-width="150">
          <template #default="{ row }">
            <el-tag v-if="row.permission" size="small" class="permission-tag">
              {{ row.permission }}
            </el-tag>
            <span v-else class="text-tertiary">-</span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="150" align="center">
          <template #default="{ row }">
            <TableActions
              :row="row"
              :buttons="[
                { text: '编辑', click: handleEdit },
                { text: '删除', type: 'danger', click: handleDelete }
              ]"
            />
          </template>
        </el-table-column>
      </el-table>
    </TablePage>

    <!-- Edit Menu Dialog -->
    <FormDialog
      v-model="dialogVisible"
      :title="form.id ? '编辑菜单' : '新增菜单'"
      :form-model="form"
      :rules="rules"
      :submitting="submitting"
      @submit="submitForm"
      width="500px"
    >
      <el-form-item label="菜单名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入菜单名称" />
      </el-form-item>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="菜单图标" prop="icon">
            <el-select
              v-model="form.icon"
              placeholder="请选择图标"
              style="width: 100%"
              filterable
            >
              <el-option
                v-for="icon in iconList"
                :key="icon.key"
                :label="icon.key"
                :value="icon.key"
              >
                <div style="display: flex; align-items: center; gap: 8px">
                  <el-icon><component :is="icon.component" /></el-icon>
                  <span>{{ icon.key }}</span>
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
    </FormDialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, markRaw } from 'vue'
import { useStore } from 'vuex'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { Plus, Top, Bottom, Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { systemApi, loginApi } from '@/api'
import { TablePage, FormDialog, TableActions, SwitchStatus } from '@/components'

const store = useStore()

// 图标列表
const iconList = Object.entries(ElementPlusIconsVue).map(([key, component]) => ({
  key,
  component: markRaw(component)
}))

// 状态管理
const menuList = ref([])
const filterText = ref('')
const loading = ref(false)
const dialogVisible = ref(false)
const submitting = ref(false)

// 表单数据
const form = ref({
  id: null,
  name: '',
  icon: '',
  path: '',
  permission: '',
  sort: 1,
  status: 1
})

// 表单验证规则
const rules = {
  name: [{ required: true, message: '请输入菜单名称', trigger: 'blur' }],
  path: [{ required: true, message: '请输入路由路径', trigger: 'blur' }],
  permission: [{ required: true, message: '请输入权限标识', trigger: 'blur' }]
}

// 计算属性
const total = computed(() => menuList.value.length)

const filteredMenuList = computed(() => {
  if (!filterText.value) return menuList.value

  const filterTree = (list) => {
    const result = []
    list.forEach(item => {
      const nameMatch = item.name.toLowerCase().includes(filterText.value.toLowerCase())
      const children = item.children ? filterTree(item.children) : []
      const childrenMatch = children.length > 0

      if (nameMatch || childrenMatch) {
        const newItem = { ...item, children }
        result.push(newItem)
      }
    })
    return result
  }

  return filterTree(menuList.value)
})

// 数据获取
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

// 事件处理
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

const handleEdit = (row) => {
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
  ElMessageBox.confirm(`确定要删除菜单 "${row.name}" 吗？`, '警告', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await systemApi.deleteMenu(row.id)
      ElMessage.success('删除成功')
      fetchMenus()
      await refreshUserMenu()
    } catch (error) {
      ElMessage.error('删除失败')
    }
  })
}

const handleStatusChange = async (row, val) => {
  try {
    await systemApi.updateMenu(row.id, { status: val })
    ElMessage.success(`${val === 1 ? '启用' : '停用'}成功`)
    await refreshUserMenu()
  } catch (error) {
    row.status = val === 1 ? 0 : 1
    ElMessage.error('操作失败')
  }
}

const submitForm = async () => {
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
    await refreshUserMenu()
  } catch (error) {
    console.error('Error updating menu:', error)
    ElMessage.error('保存失败')
  } finally {
    submitting.value = false
  }
}

const handleMove = async (row, direction) => {
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

  try {
    const originalSort = row.sort
    const targetOriginalSort = targetRow.sort

    await Promise.all([
      systemApi.updateMenu(row.id, { sort: targetOriginalSort }),
      systemApi.updateMenu(targetRow.id, { sort: originalSort })
    ])

    ElMessage.success('排序更新成功')
    fetchMenus()
    await refreshUserMenu()
  } catch (error) {
    ElMessage.error('排序更新失败')
    fetchMenus()
  }
}

const refreshUserMenu = async () => {
  try {
    const res = await loginApi.getUserInfo()
    if (res.code === 200) {
      store.commit('SET_MENU_TREE', res.data.menuTree)
      store.commit('SET_PERMISSIONS', res.data.permissions)
    }
  } catch (error) {
    console.error('Error refreshing user menu:', error)
  }
}

const getIcon = (iconName) => {
  return ElementPlusIconsVue[iconName] || iconName
}

// 初始化
onMounted(fetchMenus)
</script>

<style scoped>
.menus-container {
  display: flex;
  flex-direction: column;
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

.menu-name {
  font-weight: 500;
}

.permission-tag {
  font-family: 'Monaco', 'Consolas', monospace;
  font-size: 12px;
}

.text-tertiary {
  color: var(--text-tertiary);
  font-size: 12px;
}
</style>
