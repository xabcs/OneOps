<template>
  <div class="user-management">
    <!-- 工具栏 - 所有用户看到相同的界面 -->
    <div class="toolbar">
      <el-button type="primary" @click="handleCreate">
        <el-icon><plus /></el-icon>
        新增用户
      </el-button>

      <el-button type="danger" :disabled="selectedUsers.length === 0" @click="handleBatchDelete">
        <el-icon><delete /></el-icon>
        批量删除
      </el-button>

      <el-button @click="handleExport">
        <el-icon><download /></el-icon>
        导出用户
      </el-button>

      <el-button @click="handleImport">
        <el-icon><upload /></el-icon>
        导入用户
      </el-button>

      <el-button @click="handleAdvancedFeature">
        <el-icon><setting /></el-icon>
        高级功能
      </el-button>
    </div>

    <!-- 用户表格 -->
    <el-table :data="users" @selection-change="handleSelectionChange">
      <el-table-column type="selection" width="55" />
      <el-table-column prop="username" label="用户名" />
      <el-table-column prop="email" label="邮箱" />
      <el-table-column prop="role" label="角色" />
      <el-table-column prop="status" label="状态" />

      <!-- 操作列 - 所有用户都看到相同的操作按钮 -->
      <el-table-column label="操作" width="350">
        <template #default="{ row }">
          <el-button type="info" size="small" @click="handleView(row)">
            查看
          </el-button>

          <el-button type="primary" size="small" @click="handleEdit(row)">
            编辑
          </el-button>

          <el-button type="danger" size="small" @click="handleDelete(row)">
            删除
          </el-button>

          <el-dropdown @command="(cmd) => handleMoreCommand(cmd, row)">
            <el-button type="info" size="small">
              更多
              <el-icon><arrow-down /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="reset_password">
                  重置密码
                </el-dropdown-item>
                <el-dropdown-item command="lock">
                  锁定用户
                </el-dropdown-item>
                <el-dropdown-item command="unlock">
                  解锁用户
                </el-dropdown-item>
                <el-dropdown-item command="assign_role">
                  分配角色
                </el-dropdown-item>
                <el-dropdown-item command="view_history">
                  查看历史
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useUserActionExecutors } from '@/composables/useUnifiedPermission'

// 使用统一的权限检查执行器
const {
  executeCreate,
  executeUpdate,
  executeDelete,
  executeBatchDelete,
  executeResetPassword,
  executeLockUser,
  executeUnlockUser,
  executeAssignRole,
  executeExport,
  executeImport
} = useUserActionExecutors()

// 用户数据
const users = ref([])
const selectedUsers = ref([])

/**
 * 新增用户 - 所有用户都能看到按钮，操作时才检查权限
 */
const handleCreate = async () => {
  await executeCreate(async () => {
    // 打开创建用户对话框
    showCreateDialog()

    // 或者直接调用API
    // await createUserAPI(userData)
    // ElMessage.success('用户创建成功')
  })
}

/**
 * 编辑用户
 */
const handleEdit = async (user: any) => {
  await executeUpdate(async () => {
    // 打开编辑对话框
    showEditDialog(user)

    // 或者直接调用API
    // await updateUserAPI(user.id, userData)
    // ElMessage.success('用户更新成功')
  })
}

/**
 * 删除用户
 */
const handleDelete = async (user: any) => {
  await executeDelete(async () => {
    // 确认删除
    await ElMessageBox.confirm('确定要删除该用户吗？', '确认删除', {
      type: 'warning'
    })

    // 调用删除API
    // await deleteUserAPI(user.id)
    // ElMessage.success('用户删除成功')

    // 刷新列表
    await loadUsers()
  })
}

/**
 * 批量删除
 */
const handleBatchDelete = async () => {
  await executeBatchDelete(async () => {
    if (selectedUsers.value.length === 0) {
      ElMessage.warning('请先选择要删除的用户')
      return
    }

    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedUsers.value.length} 个用户吗？`,
      '确认批量删除',
      { type: 'warning' }
    )

    // 调用批量删除API
    // await batchDeleteUsersAPI(selectedUsers.value.map(u => u.id))
    // ElMessage.success('批量删除成功')

    // 刷新列表
    await loadUsers()
  })
}

/**
 * 导出用户
 */
const handleExport = async () => {
  await executeExport(async () => {
    // 调用导出API
    // await exportUsersAPI()
    // ElMessage.success('用户数据导出成功')
  })
}

/**
 * 导入用户
 */
const handleImport = async () => {
  await executeImport(async () => {
    // 打开导入对话框
    showImportDialog()

    // 或者直接处理文件上传
    // await importUsersAPI(file)
    // ElMessage.success('用户数据导入成功')
  })
}

/**
 * 高级功能 - 弹窗提醒
 */
const handleAdvancedFeature = async () => {
  const { executeWithPermission } = useUnifiedPermission()

  await executeWithPermission(
    'system.user.advanced',
    async () => {
      // 执行高级功能
      ElMessage.success('高级功能执行成功')
    },
    {
      message: '您需要【高级用户】权限才能使用此功能\n\n该功能包括：批量操作、数据分析、系统配置等高级特性',
      useNotification: true  // 使用通知方式提醒
    }
  )
}

/**
 * 更多操作命令处理
 */
const handleMoreCommand = async (command: string, user: any) => {
  switch (command) {
    case 'reset_password':
      await executeResetPassword(async () => {
        await ElMessageBox.prompt('请输入新密码', '重置密码', {
          inputPattern: /.+/,
          inputErrorMessage: '密码不能为空'
        })
        // await resetPasswordAPI(user.id, newPassword)
        ElMessage.success('密码重置成功')
      })
      break

    case 'lock':
      await executeLockUser(async () => {
        await ElMessageBox.confirm(`确定要锁定用户 ${user.username} 吗？`, '锁定用户', {
          type: 'warning'
        })
        // await lockUserAPI(user.id)
        ElMessage.success('用户已锁定')
      })
      break

    case 'unlock':
      await executeUnlockUser(async () => {
        await ElMessageBox.confirm(`确定要解锁用户 ${user.username} 吗？`, '解锁用户', {
          type: 'success'
        })
        // await unlockUserAPI(user.id)
        ElMessage.success('用户已解锁')
      })
      break

    case 'assign_role':
      await executeAssignRole(async () => {
        // 打开角色分配对话框
        showRoleAssignDialog(user)
      })
      break

    case 'view_history':
      const { executeView } = useUserActionExecutors()
      await executeView(async () => {
        // 显示用户操作历史
        showUserHistory(user)
      })
      break
  }
}

/**
 * 选择变化处理
 */
const handleSelectionChange = (selection: any[]) => {
  selectedUsers.value = selection
}

/**
 * 显示创建对话框
 */
const showCreateDialog = () => {
  // 实现创建对话框逻辑
}

/**
 * 显示编辑对话框
 */
const showEditDialog = (user: any) => {
  // 实现编辑对话框逻辑
}

/**
 * 显示导入对话框
 */
const showImportDialog = () => {
  // 实现导入对话框逻辑
}

/**
 * 显示角色分配对话框
 */
const showRoleAssignDialog = (user: any) => {
  // 实现角色分配对话框逻辑
}

/**
 * 显示用户历史
 */
const showUserHistory = (user: any) => {
  // 实现用户历史显示逻辑
}

/**
 * 加载用户列表
 */
const loadUsers = async () => {
  // 加载用户数据
  // users.value = await fetchUsersAPI()
}
</script>

<style scoped>
.user-management {
  padding: 20px;
}

.toolbar {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.toolbar .el-button {
  display: flex;
  align-items: center;
  gap: 5px;
}
</style>