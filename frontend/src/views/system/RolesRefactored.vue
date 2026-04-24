<template>
  <div class="roles-container">
    <TablePage
      title="角色管理"
      subtitle="管理系统用户角色及权限分配"
      :loading="loading"
      :total="total"
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      @size-change="fetchRoles"
      @current-change="fetchRoles"
    >
      <template #actions>
        <el-space>
          <el-input
            v-model="queryParams.name"
            placeholder="搜索角色名称/标识"
            :prefix-icon="Search"
            clearable
            style="width: 220px"
            @keyup.enter="fetchRoles"
          />
          <el-button type="primary" :icon="Plus" @click="handleAdd">
            新增角色
          </el-button>
        </el-space>
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
              <span v-if="!row.userNames?.length" class="text-tertiary">暂无用户</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="description" label="描述" min-width="200" />

        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <SwitchStatus
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              :readonly="row.code === 'admin'"
              @change="(val) => handleStatusChange(row, val)"
            />
          </template>
        </el-table-column>

        <el-table-column label="操作" width="250" align="center">
          <template #default="{ row }">
            <TableActions
              :row="row"
              :buttons="[
                { text: '编辑', click: handleEdit },
                { text: '权限设置', click: handlePermission, class: 'text-accent' },
                {
                  text: '删除',
                  type: 'danger',
                  click: handleDelete,
                  disabled: (row) => row.code === 'admin'
                }
              ]"
            />
          </template>
        </el-table-column>
      </el-table>
    </TablePage>

    <!-- Role Form Dialog -->
    <FormDialog
      v-model="dialogVisible"
      :title="form.id ? '编辑角色' : '新增角色'"
      :form-model="form"
      :rules="rules"
      :submitting="submitting"
      @submit="submitForm"
    >
      <el-form-item label="角色名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入角色名称" />
      </el-form-item>

      <el-form-item label="角色标识" prop="code">
        <el-input
          v-model="form.code"
          placeholder="请输入角色标识"
          :disabled="!!form.id"
        />
      </el-form-item>

      <el-form-item label="描述" prop="description">
        <el-input
          v-model="form.description"
          type="textarea"
          placeholder="请输入角色描述"
        />
      </el-form-item>

      <el-form-item label="状态" prop="status">
        <el-radio-group v-model="form.status" :disabled="form.code === 'admin'">
          <el-radio :label="1">启用</el-radio>
          <el-radio :label="0">禁用</el-radio>
        </el-radio-group>
      </el-form-item>
    </FormDialog>

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
          :check-strictly="false"
        />
      </div>
      <template #footer>
        <el-button @click="permissionDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitPermission">
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { useStore } from 'vuex'
import { Plus, Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { systemApi } from '@/api'
import { TablePage, FormDialog, TableActions, SwitchStatus } from '@/components'

const store = useStore()

// 状态管理
const roleList = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const menuTree = ref([])
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const permissionDialogVisible = ref(false)
const treeRef = ref(null)
const currentRole = ref(null)

// 查询参数
const queryParams = reactive({
  name: ''
})

// 表单数据
const form = ref({
  id: null,
  name: '',
  code: '',
  description: '',
  status: 1
})

// 表单验证规则
const rules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入角色标识', trigger: 'blur' }]
}

// 计算属性
const displayRoles = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return roleList.value.slice(start, end)
})

// 数据获取
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

      // 前端过滤
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

// 事件处理
const handleAdd = () => {
  form.value = { id: null, name: '', code: '', description: '', status: 1 }
  dialogVisible.value = true
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

const handleDelete = (row) => {
  ElMessageBox.confirm(`确定要删除角色 "${row.name}" 吗？`, '删除角色', {
    confirmButtonText: '确定删除',
    cancelButtonText: '取消',
    type: 'warning',
    distinguishCancelAndClose: true
  }).then(async () => {
    try {
      const res = await systemApi.deleteRole(row.id)
      if (res.code === 200) {
        ElMessage.success(res.message || '删除成功')
        fetchRoles()
      } else {
        ElMessage.error(res.message || '删除失败')
      }
    } catch (error) {
      if (error.response?.data) {
        const errorMsg = error.response.data.message || error.response.data.error || '删除失败'
        ElMessage.error(errorMsg)
      } else {
        ElMessage.error('删除失败：网络错误')
      }
    }
  })
}

const handlePermission = (row) => {
  currentRole.value = row
  permissionDialogVisible.value = true
  nextTick(() => {
    let menuIds = []
    try {
      if (row.menuIds) {
        menuIds = JSON.parse(row.menuIds)
      }
    } catch (e) {
      console.error('解析 menuIds 失败:', e)
      menuIds = []
    }

    const leafKeys = getLeafKeysFromTree(menuTree.value, menuIds)
    treeRef.value?.setCheckedKeys(leafKeys)
  })
}

const getLeafKeysFromTree = (tree, selectedIds) => {
  const leafKeys = []

  const traverse = (nodes) => {
    nodes.forEach(node => {
      const isParent = node.children && node.children.length > 0
      const isSelected = selectedIds.includes(node.id)

      if (isParent) {
        traverse(node.children)
      } else if (isSelected) {
        leafKeys.push(node.id)
      }
    })
  }

  traverse(tree)
  return leafKeys
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
    store.dispatch('fetchUserInfo')
  } catch (error) {
    ElMessage.error('权限更新失败')
  } finally {
    submitting.value = false
  }
}

// 初始化
onMounted(() => {
  fetchRoles()
  fetchMenus()
})
</script>

<style scoped>
.roles-container {
  display: flex;
  flex-direction: column;
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

.permission-tag {
  font-family: 'Monaco', 'Consolas', monospace;
  font-size: 12px;
}

.text-accent {
  color: var(--accent) !important;
}

.text-accent:hover {
  color: var(--accent-hover) !important;
  text-decoration: underline;
}
</style>
