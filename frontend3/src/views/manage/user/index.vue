<script setup lang="tsx">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { Delete, Plus, Refresh, Search, User } from '@element-plus/icons-vue';
import { fetchDeleteUser, fetchGetAllRoles, fetchGetUserList } from '@/service/api';
import { useThemeStore } from '@/store/modules/theme';
import { defaultTransform, useTableOperate, useUIPaginatedTable } from '@/hooks/common/table';
import { $t } from '@/locales';
import { useUnifiedPermission } from '@/composables/useUnifiedPermission';
import UserOperateDrawer from './modules/user-operate-drawer.vue';
import ResetPasswordModal from './modules/reset-password-modal.vue';

defineOptions({ name: 'UserManage' });

const themeStore = useThemeStore();

// 使用统一权限检查
const { executeWithPermission } = useUnifiedPermission();

// Hero区域显示状态
const heroVisible = computed(() => themeStore.contentTheme2.heroSection.visible !== false);

// 用户统计数据
const userStats = ref({
  total: 0,
  active: 0,
  inactive: 0
});

// 角色列表和映射
const allRoles = ref<Api.SystemManage.AllRole[]>([]);
const roleMap = computed(() => {
  const map = new Map<number, Api.SystemManage.AllRole>();
  // 添加安全检查，避免在路由切换时出错
  if (allRoles.value && Array.isArray(allRoles.value)) {
    allRoles.value.forEach(role => {
      if (role && role.id) {
        map.set(role.id, role);
      }
    });
  }
  return map;
});

// 预定义的颜色列表（涵盖多种颜色，确保不同角色编码显示不同颜色）
// 包含 Element Plus 主题色 + 自定义颜色
const COLOR_PALETTE: Array<{ type: UI.ThemeColor; customClass?: string }> = [
  { type: 'primary' }, // 蓝色
  { type: 'success' }, // 绿色
  { type: 'warning' }, // 橙色
  { type: 'danger' }, // 红色
  { type: 'info' } // 灰色
];

// 字符串哈希函数
function stringHash(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = (hash << 5) - hash + char;
    hash &= hash; // 转换为32位整数
  }
  return Math.abs(hash);
}

// 根据角色代码获取标签颜色（每个不同编码显示不同颜色）
function getRoleTagType(roleCode: string): UI.ThemeColor {
  // 计算角色编码的哈希值
  const hash = stringHash(roleCode);

  // 根据哈希值选择颜色，确保不同编码大概率显示不同颜色
  const colorIndex = hash % COLOR_PALETTE.length;

  return COLOR_PALETTE[colorIndex].type;
}

// 获取角色颜色的自定义样式类（可选，用于更丰富的颜色）
function getRoleTagClass(roleCode: string): string {
  const hash = stringHash(roleCode);

  // 可以根据哈希值添加自定义样式
  return `role-tag-${hash % 10}`; // 10种自定义样式变体
}

// 获取所有角色
async function getAllRoles() {
  const { error, data } = await fetchGetAllRoles();
  if (!error && data) {
    // fetchGetAllRoles 返回的是分页对象，需要提取 records 数组
    if (Array.isArray(data)) {
      // 如果直接返回数组（兼容旧格式）
      allRoles.value = data;
    } else if (data && Array.isArray(data.list)) {
      // 如果返回分页对象（新格式）
      allRoles.value = data.list;
    } else {
      allRoles.value = [];
    }
  } else {
    allRoles.value = [];
  }
}

// 更新用户统计数据
async function updateUserStats() {
  const { error, data } = await fetchGetUserList({
    page: 1,
    pageSize: 1,
    status: undefined,
    username: undefined,
    nickname: undefined,
    email: undefined
  });

  if (!error && data) {
    userStats.value.total = data.total || 0;

    // 获取启用用户数
    const activeResult = await fetchGetUserList({
      page: 1,
      pageSize: 1,
      status: 'active',
      username: undefined,
      nickname: undefined,
      email: undefined
    });

    if (!activeResult.error && activeResult.data) {
      userStats.value.active = activeResult.data.total || 0;
    }

    // 计算禁用用户数
    userStats.value.inactive = userStats.value.total - userStats.value.active;
  }
}

onMounted(() => {
  getAllRoles();
  updateUserStats();
});

const searchParams = ref(getInitSearchParams());

function getInitSearchParams(): Api.SystemManage.UserSearchParams {
  return {
    page: 1,
    pageSize: 10,
    status: undefined,
    username: undefined,
    nickname: undefined,
    email: undefined
  };
}

const { columns, columnChecks, data, getData, getDataByPage, loading, mobilePagination } = useUIPaginatedTable({
  paginationProps: {
    currentPage: searchParams.value.page,
    pageSize: searchParams.value.pageSize
  },
  api: () => fetchGetUserList(searchParams.value),
  transform: response => {
    return defaultTransform(response);
  },
  onPaginationParamsChange: params => {
    searchParams.value.page = params.currentPage;
    searchParams.value.pageSize = params.pageSize;
  },
  columns: () => [
    { prop: 'selection', type: 'selection', width: 48 },
    { prop: 'index', type: 'index', label: $t('common.index'), width: 64 },
    { prop: 'username', label: $t('page.manage.user.userName'), minWidth: 100 },
    { prop: 'nickname', label: $t('page.manage.user.nickName'), minWidth: 100 },
    { prop: 'email', label: $t('page.manage.user.userEmail'), minWidth: 200 },
    {
      prop: 'roleIds',
      label: '分配角色',
      minWidth: 150,
      formatter: row => {
        if (!row.roleIds || row.roleIds.length === 0) {
          return <span class="pl-12px text-gray">-</span>;
        }

        // 将 roleIds 数组转换为角色标签
        const roleTags = row.roleIds
          .map((id: number) => {
            const role = roleMap.value ? roleMap.value.get(id) : null;
            if (!role) return null;

            // 根据角色编码（code）获取标签颜色，更严谨
            const tagType = getRoleTagType(role.code);

            return (
              <ElTag key={id} size="small" type={tagType}>
                {role.name}
              </ElTag>
            );
          })
          .filter(Boolean);

        if (roleTags.length === 0) {
          return <span class="pl-12px text-gray">-</span>;
        }

        return <div class="flex flex-wrap items-center gap-4px pl-12px">{roleTags}</div>;
      }
    },
    {
      prop: 'status',
      label: $t('page.manage.user.userStatus'),
      align: 'center',
      width: 100,
      formatter: row => {
        if (row.status === undefined) {
          return '';
        }

        const statusMap: Record<string, UI.ThemeColor> = {
          active: 'success',
          inactive: 'warning'
        };

        const label = row.status === 'active' ? '启用' : '禁用';

        return <ElTag type={statusMap[row.status] || 'info'}>{label}</ElTag>;
      }
    },
    {
      prop: 'homePath',
      label: '家目录',
      align: 'center',
      width: 150,
      formatter: row => {
        const path = row.homePath || '/';
        const pathMap: Record<string, string> = {
          '/': '首页',
          '/servers': '资产管理',
          '/tasks': '自动化任务',
          '/monitoring': '监控中心',
          '/system': '系统管理'
        };
        return <span class="text-primary">{pathMap[path] || path}</span>;
      }
    },
    {
      prop: 'operate',
      label: $t('common.operate'),
      align: 'center',
      width: 260,
      formatter: row => (
        <div class="flex-center gap-8px">
          <ElButton type="primary" plain size="small" onClick={() => edit(row.id)}>
            {$t('common.edit')}
          </ElButton>
          <ElButton type="warning" plain size="small" onClick={() => openResetPassword(row)}>
            重置密码
          </ElButton>
          <ElPopconfirm title={$t('common.confirmDelete')} onConfirm={() => handleDelete(row.id)}>
            {{
              reference: () => (
                <ElButton type="danger" plain size="small" disabled={row.username === 'admin'}>
                  {$t('common.delete')}
                </ElButton>
              )
            }}
          </ElPopconfirm>
        </div>
      )
    }
  ]
});

const { drawerVisible, operateType, editingData, handleAdd, handleEdit, checkedRowKeys, onBatchDeleted, onDeleted } =
  useTableOperate(data, 'id', getData);

// 重置密码相关
const resetPasswordVisible = ref(false);
const resetPasswordUserId = ref(-1);
const resetPasswordUsername = ref('');

async function openResetPassword(row: Api.SystemManage.User) {
  await executeWithPermission('system.user.reset_password', async () => {
    resetPasswordUserId.value = row.id;
    resetPasswordUsername.value = row.username || '';
    resetPasswordVisible.value = true;
  });
}

function handleResetPasswordSubmitted() {
  // 重置密码成功后刷新列表
  getDataByPage();
}

async function handleBatchDelete() {
  await executeWithPermission('system.user.batch_delete', async () => {
    if (checkedRowKeys.value.length === 0) {
      window.$message?.warning('请选择要删除的用户');
      return;
    }

    // 批量删除（串行执行）
    let successCount = 0;
    let failCount = 0;

    for (const id of checkedRowKeys.value) {
      const { error } = await fetchDeleteUser(id as number);
      if (!error) {
        successCount++;
      } else {
        failCount++;
      }
    }

    if (successCount > 0) {
      window.$message?.success(`成功删除 ${successCount} 个用户`);
    }

    if (failCount > 0) {
      window.$message?.error(`${failCount} 个用户删除失败`);
    }

    onBatchDeleted();
    // 刷新角色列表和统计数据
    await getAllRoles();
    await updateUserStats();
  });
}

async function handleDelete(id: number) {
  await executeWithPermission('system.user.delete', async () => {
    const { error } = await fetchDeleteUser(id);

    if (!error) {
      window.$message?.success($t('common.deleteSuccess'));
      onDeleted();
      // 刷新角色列表和统计数据
      await getAllRoles();
      await updateUserStats();
    } else {
      window.$message?.error(error.msg || '删除失败');
    }
  }, { type: 'error' });
}

function resetSearchParams() {
  searchParams.value = getInitSearchParams();
}

async function edit(id: number) {
  await executeWithPermission('system.user.update', async () => {
    handleEdit(id);
  });
}

async function handleAddClick() {
  await executeWithPermission('system.user.create', async () => {
    handleAdd();
  });
}

// 刷新数据（包括统计和表格）
async function refreshData() {
  await updateUserStats();
  await getDataByPage();
}

// 搜索输入处理（防抖）
let searchTimeout: ReturnType<typeof setTimeout> | null = null;
function handleSearchInput() {
  if (searchTimeout) {
    clearTimeout(searchTimeout);
  }
  searchTimeout = setTimeout(() => {
    searchParams.value.page = 1;
    getDataByPage();
  }, 300);
}

// 手动搜索（点击搜索按钮）
function handleSearch() {
  searchParams.value.page = 1;
  getDataByPage();
}

// 组件卸载时清理资源
onUnmounted(() => {
  // 清理搜索定时器，避免内存泄漏
  if (searchTimeout) {
    clearTimeout(searchTimeout);
    searchTimeout = null;
  }
});
</script>

<template>
    <div class="page-container user-management-page">
        <!-- Hero 区域 -->
        <section v-if="heroVisible" class="hero-section" :style="{
        background: 'var(--sx-hero-bg)',
        border: '1px solid var(--sx-hero-border)',
        borderRadius: 'var(--sx-hero-radius)',
        boxShadow: 'var(--sx-hero-shadow)',
        padding: 'var(--sx-hero-padding)'
      }">
            <div class="hero-content">
                <div class="hero-title-row">
                    <span class="hero-icon">
                        <ElIcon>
                            <User />
                        </ElIcon>
                    </span>
                    <h2>{{ $t('page.manage.user.title') }}</h2>
                    <p class="hero-desc">统一维护用户、分配角色与权限，支持账号治理与安全策略管理</p>
                </div>
            </div>
            <div class="hero-actions">
                <ElButton size="small" :loading="loading" @click="refreshData">
                    <ElIcon>
                        <Refresh />
                    </ElIcon>
                    刷新
                </ElButton>
            </div>
        </section>

        <!-- 统计卡片网格 -->
        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-value">{{ userStats.total }}</div>
                <div class="stat-label">用户总数</div>
                <div class="stat-desc">系统注册用户总量</div>
            </div>
            <div class="stat-card success-card">
                <div class="stat-value">{{ userStats.active }}</div>
                <div class="stat-label">启用用户</div>
                <div class="stat-desc">当前状态为启用</div>
            </div>
            <div class="stat-card warning-card">
                <div class="stat-value">{{ userStats.inactive }}</div>
                <div class="stat-label">禁用用户</div>
                <div class="stat-desc">当前状态为禁用</div>
            </div>
        </div>

        <!-- 内容卡片 -->
        <div class="content-card user-content-card" :style="{
        background: 'var(--sx-content-card-bg)',
        border: '1px solid var(--sx-content-card-border)',
        borderRadius: 'var(--sx-content-card-radius)',
        boxShadow: 'var(--sx-content-card-shadow)',
        padding: 'var(--sx-content-card-padding)'
      }">
            <!-- 工具栏 -->
            <div class="card-toolbar">
                <div class="toolbar-head">
                    <span class="toolbar-title">用户列表</span>
                    <span class="toolbar-desc">管理系统用户账号、角色分配与状态控制</span>
                </div>
                <div class="toolbar-actions">
                    <ElButton type="primary" size="small" @click="handleAddClick">
                        <ElIcon>
                            <Plus />
                        </ElIcon>
                        新增用户
                    </ElButton>
                    <ElButton type="danger" size="small" :disabled="checkedRowKeys.length === 0" @click="handleBatchDelete">
                        <ElIcon>
                            <Delete />
                        </ElIcon>
                        批量删除
                    </ElButton>
                </div>
            </div>

            <!-- 搜索工具栏 -->
            <div class="workbench-toolbar workbench-toolbar--history users-toolbar">
                <div class="workbench-toolbar-left">
                    <ElInput v-model="searchParams.username" placeholder="搜索用户名" clearable style="width: 200px" @input="handleSearchInput" />
                    <ElInput v-model="searchParams.nickname" placeholder="搜索昵称" clearable style="width: 200px" @input="handleSearchInput" />
                    <ElInput v-model="searchParams.email" placeholder="搜索邮箱" clearable style="width: 240px" @input="handleSearchInput" />
                </div>
                <div class="workbench-toolbar-right">
                    <ElButton class="filter-refresh-btn" @click="resetSearchParams">
                        <ElIcon>
                            <Refresh />
                        </ElIcon>
                        重置
                    </ElButton>
                    <ElButton class="filter-refresh-btn" type="primary" @click="handleSearch">
                        <ElIcon>
                            <Search />
                        </ElIcon>
                        搜索
                    </ElButton>
                </div>
            </div>

            <!-- 数据表格 -->
            <div class="table-section">
                <ElTable v-loading="loading" :data="data" border stripe class="data-table" row-key="id" @selection-change="checkedRowKeys = $event">
                    <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
                </ElTable>
            </div>

            <!-- 分页 -->
            <div class="table-pagination">
                <ElPagination v-if="mobilePagination.total" layout="total, sizes, prev, pager, next" v-bind="mobilePagination" @current-change="mobilePagination['current-change']" @size-change="mobilePagination['size-change']" />
            </div>
        </div>

        <!-- 抽屉和模态框 -->
        <UserOperateDrawer v-model:visible="drawerVisible" :operate-type="operateType" :row-data="editingData" :all-roles="allRoles" @submitted="getDataByPage" />
        <ResetPasswordModal v-model:visible="resetPasswordVisible" :user-id="resetPasswordUserId" :username="resetPasswordUsername" @submitted="handleResetPasswordSubmitted" />
    </div>
</template>

<style scoped lang="scss">
    /* ============================================
               1. 页面容器与布局
               ============================================ */
    .user-management-page {
      padding: 16px 20px;
      background: var(--el-bg-color-page);
      min-height: 100vh;
      display: flex;
      flex-direction: column;
      gap: 6px;
    }

    /* ============================================
               2. Hero 区域
               ============================================ */
    .hero-section {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 8px;
    }

    .hero-content {
      display: flex;
      flex-direction: column;
      gap: 4px;
    }

    .hero-title-row {
      display: flex;
      align-items: center;
      gap: 4px;
      flex-wrap: wrap;
    }

    .hero-icon {
      width: 42px;
      height: 42px;
      border-radius: 14px;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      font-size: 20px;
      color: var(--sx-hero-icon-color);
      background: var(--sx-hero-icon-bg);
      border: 1px solid var(--sx-hero-icon-border);
      box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
    }

    .hero-title-row h2 {
      color: var(--sx-text-primary);
      font-size: 23px;
      font-weight: 700;
      margin: 0;
    }

    .hero-desc {
      color: var(--sx-text-secondary);
      font-size: 13px;
      line-height: 1.45;
      margin: 0;
      max-width: 600px;
    }

    .hero-actions {
      display: flex;
      gap: 4px;
    }

    /* ============================================
               3. 统计卡片网格
               ============================================ */
    .stats-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
      gap: 6px;
    }

    .stat-card {
      cursor: default;
      transition: all 0.3s ease;
      position: relative;
      overflow: hidden;
      text-align: center;
      background: var(--sx-stat-default-bg);
      border: 1px solid var(--sx-stat-default-border);
      border-radius: var(--sx-stat-radius);
      box-shadow: var(--sx-stat-shadow);
      padding: var(--sx-content-card-padding);
      display: flex;
      flex-direction: column;
      gap: 8px;

      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
      }

      /* 装饰性圆点 */
      &::after {
        content: '';
        position: absolute;
        right: -24px;
        bottom: -30px;
        width: 108px;
        height: 108px;
        border-radius: 50%;
        opacity: 0.16;
      }

      /* 状态变体 */
      &.success-card {
        background: var(--sx-stat-success-bg);
      }

      &.warning-card {
        background: var(--sx-stat-warning-bg);
      }
    }

    .stat-value {
      font-size: 28px;
      line-height: 1.05;
      color: var(--sx-text-primary);
      font-weight: 700;
      margin-bottom: 4px;
    }

    .stat-label {
      color: var(--sx-text-secondary);
      font-size: 14px;
      font-weight: 600;
      margin-bottom: 4px;
    }

    .stat-desc {
      color: var(--sx-text-muted);
      font-size: 12px;
      line-height: 1.4;
    }

    /* ============================================
               4. 内容卡片
               ============================================ */
    .content-card {
      display: flex;
      flex-direction: column;
      gap: 8px;
    }

    .card-toolbar {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      gap: 12px;
      padding-bottom: 12px;
      border-bottom: 1px solid var(--sx-border-soft);
    }

    .toolbar-head {
      display: flex;
      flex-direction: column;
      gap: 4px;
    }

    .toolbar-title {
      color: var(--sx-text-primary);
      font-size: 16px;
      font-weight: 700;
    }

    .toolbar-desc {
      color: var(--sx-text-secondary);
      font-size: 12px;
    }

    .toolbar-actions {
      display: flex;
      gap: 4px;
      align-items: center;
    }

    /* ============================================
               5. 搜索工具栏
               ============================================ */
    .workbench-toolbar {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
      flex-wrap: wrap;
      margin: 6px 0 8px;
    }

    .workbench-toolbar-left,
    .workbench-toolbar-right {
      display: flex;
      align-items: center;
      gap: 10px;
      flex-wrap: wrap;
    }

    /* 工具栏主题样式 */
    .workbench-toolbar.workbench-toolbar--history {
      background: var(--sx-toolbar-bg);
      border: 1px solid var(--sx-toolbar-border);
      border-radius: var(--sx-toolbar-radius);
      padding: var(--sx-toolbar-padding);
      box-shadow: var(--sx-toolbar-shadow);
      gap: 8px;
      margin: 6px 0;

      .workbench-toolbar-left,
      .workbench-toolbar-right {
        gap: 5px;
      }

      /* 输入框样式 */
      .el-input__wrapper {
        background: var(--sx-search-input-bg);
        border-radius: var(--sx-search-input-radius);
        box-shadow: 0 0 0 1px var(--sx-search-input-border) inset;
        transition: all 0.3s ease;

        &:hover {
          box-shadow: 0 0 0 1px var(--sx-search-input-hover-border) inset;
        }
      }

      .el-input.is-focus .el-input__wrapper {
        box-shadow: 0 0 0 1px var(--sx-search-input-focus-border) inset;
      }

      /* 按钮样式 */
      .el-button:not(.is-link) {
        background: var(--sx-search-button-bg);
        color: var(--sx-search-button-text);
        transition: all 0.3s ease;

        &:hover {
          background: var(--sx-search-button-hover);
        }
      }
    }

    /* ============================================
               6. 数据表格
               ============================================ */
    .table-section {
      flex: 1;
      display: flex;
      flex-direction: column;
    }

    .data-table {
      width: 100%;
      border-radius: var(--sx-table-radius);
      overflow: hidden;
      border: 1px solid var(--sx-table-border);
      border-collapse: separate;
      border-spacing: 0;

      /* Element Plus CSS 变量覆盖 */
      --el-table-border-color: var(--sx-table-border);
      --el-table-header-bg-color: var(--sx-table-header-bg);
      --el-table-row-hover-bg-color: var(--sx-table-row-hover);

      /* 表格单元格样式 - 统一使用 :deep() 深度选择器 */
      :deep(th.el-table__cell) {
        background-color: var(--sx-table-header-bg);
        color: var(--sx-table-header-text);
        font-weight: 600;
        border-radius: 0;
      }

      :deep(td.el-table__cell) {
        border-radius: 0;
      }

      /* 表格内部元素圆角重置（复选框、选择器等） */
      :deep(.el-checkbox__inner),
      :deep(.el-checkbox__inner::before),
      :deep(.el-checkbox__inner::after) {
        border-radius: 0;
      }

      /* 斑马纹行 */
      &--striped :deep(.el-table__body tr.el-table__row--striped td) {
        background-color: var(--sx-table-striped-bg);
      }
    }

    /* ============================================
               7. 分页组件
               ============================================ */
    .table-pagination {
      display: flex;
      justify-content: flex-end;
      padding-top: 12px;

      .el-pagination {
        .btn-next,
        .btn-prev,
        .el-pager li {
          background: var(--sx-pagination-button-bg);
          color: var(--sx-pagination-button-text);
          border-radius: var(--sx-pagination-radius);
          transition: all 0.3s ease;

          &:hover {
            background: var(--sx-pagination-button-hover);
          }

          &.is-active {
            background: var(--sx-pagination-active-bg);
            color: var(--sx-pagination-active-text);
          }
        }
      }
    }

    /* ============================================
               8. 响应式设计
               ============================================ */
    @media (max-width: 768px) {
      .hero-section {
        flex-direction: column;
        align-items: flex-start;
      }

      .stats-grid {
        grid-template-columns: 1fr;
      }

      .card-toolbar {
        flex-direction: column;
        align-items: stretch;
        gap: 8px;
      }
    }
</style>
