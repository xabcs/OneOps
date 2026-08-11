<script setup lang="tsx">
import { computed, ref } from 'vue';
import { ElNotification } from 'element-plus';
import { Icon } from '@iconify/vue';
import { jsonClone } from '@sa/utils';
import { Bottom, Menu as MenuIcon, Plus, Refresh, Search, Top } from '@element-plus/icons-vue';
import { fetchDeleteMenu, fetchGetMenuTree, fetchUpdateMenu } from '@/service/api';
import { useAuthStore } from '@/store/modules/auth';
import { useRouteStore } from '@/store/modules/route';
import { useThemeStore } from '@/store/modules/theme';
import { useTableOperate, useUIPaginatedTable } from '@/hooks/common/table';
import { $t } from '@/locales';
import MenuOperateDrawer from './modules/menu-operate-drawer.vue';

defineOptions({ name: 'MenuManage' });

const authStore = useAuthStore();
const routeStore = useRouteStore();
const themeStore = useThemeStore();

// Hero区域显示状态
const heroVisible = computed(() => themeStore.contentTheme2.heroSection.visible !== false);

const filterText = ref('');
const flatMenuList = ref<Api.SystemManage.Menu[]>([]);
const originalTreeData = ref<Api.SystemManage.Menu[]>([]);
const tableKey = ref(0);

// 将树形结构转换为扁平列表
function flattenMenuTree(tree: Api.SystemManage.Menu[]): Api.SystemManage.Menu[] {
  const result: Api.SystemManage.Menu[] = [];

  function traverse(list: Api.SystemManage.Menu[]) {
    list.forEach(item => {
      result.push({
        id: item.id,
        parentId: item.parentId,
        name: item.name,
        path: item.path,
        icon: item.icon,
        permission: item.permission,
        menuType: item.menuType,
        sort: item.sort,
        status: item.status,
        hierarchyIndex: (item as any).hierarchyIndex
      });

      if (item.children && Array.isArray(item.children) && item.children.length > 0) {
        traverse(item.children);
      }
    });
  }

  traverse(tree);
  return result;
}

// 清理菜单树并添加层级序号
function cleanMenuTree(tree: Api.SystemManage.Menu[]): Api.SystemManage.Menu[] {
  const sortTree = (list: Api.SystemManage.Menu[]): Api.SystemManage.Menu[] => {
    return list
      .sort((a, b) => (a.sort || 0) - (b.sort || 0))
      .map(item => {
        const cleaned: any = {
          id: item.id,
          parentId: item.parentId,
          name: item.name,
          path: item.path,
          icon: item.icon,
          permission: item.permission,
          menuType: item.menuType,
          sort: item.sort,
          status: item.status
        };

        if (item.children && Array.isArray(item.children) && item.children.length > 0) {
          cleaned.children = sortTree(item.children);
        }

        return cleaned;
      });
  };

  const assignHierarchyIndex = (list: Api.SystemManage.Menu[], prefix: string = ''): Api.SystemManage.Menu[] => {
    let siblingIndex = 1;

    return list.map(item => {
      const currentItem = { ...item };
      currentItem.hierarchyIndex = prefix ? `${prefix}.${siblingIndex}` : `${siblingIndex}`;

      if (item.children && Array.isArray(item.children) && item.children.length > 0) {
        currentItem.children = assignHierarchyIndex(item.children, currentItem.hierarchyIndex);
      }

      siblingIndex++;
      return currentItem;
    });
  };

  const sortedTree = sortTree(tree);
  return assignHierarchyIndex(sortedTree);
}

// 过滤菜单树
function filterMenuTree(list: Api.SystemManage.Menu[]): Api.SystemManage.Menu[] {
  if (!filterText.value) return list;

  const result: Api.SystemManage.Menu[] = [];

  list.forEach(item => {
    const nameMatch = item.name.toLowerCase().includes(filterText.value.toLowerCase());
    const pathMatch = item.path && item.path.toLowerCase().includes(filterText.value.toLowerCase());
    const permissionMatch = item.permission && item.permission.toLowerCase().includes(filterText.value.toLowerCase());

    const children = item.children && item.children.length > 0 ? filterMenuTree(item.children) : [];
    const childrenMatch = children.length > 0;

    if (nameMatch || pathMatch || permissionMatch || childrenMatch) {
      const newItem = { ...item } as any;
      if (children.length > 0) {
        newItem.children = children;
      }
      result.push(newItem);
    }
  });

  return result;
}

// 查找兄弟节点
function findSiblings(targetId: number): Api.SystemManage.Menu[] | null {
  const findInList = (list: Api.SystemManage.Menu[]): Api.SystemManage.Menu[] | null => {
    for (let i = 0; i < list.length; i++) {
      if (list[i].id === targetId) return list;
      if (list[i].children && list[i].children.length > 0) {
        const found = findInList(list[i].children);
        if (found) return found;
      }
    }
    return null;
  };
  return findInList(originalTreeData.value);
}

// 判断是否是第一个
function isFirst(id: number): boolean {
  const siblings = findSiblings(id);
  if (!siblings || siblings.length === 0) return true;
  return siblings[0].id === id;
}

// 判断是否是最后一个
function isLast(id: number): boolean {
  const siblings = findSiblings(id);
  if (!siblings || siblings.length === 0) return true;
  return siblings[siblings.length - 1].id === id;
}

const { columns, data, loading, getData, getDataByPage } = useUIPaginatedTable({
  api: async () => {
    const { error, data } = await fetchGetMenuTree();
    if (!error && data && Array.isArray(data)) {
      const tree = data as Api.SystemManage.Menu[];

      const cleanedTree = cleanMenuTree(tree);
      originalTreeData.value = cleanedTree;

      flatMenuList.value = flattenMenuTree(cleanedTree);

      const displayData = filterText.value ? filterMenuTree(cleanedTree) : cleanedTree;

      return {
        data: displayData,
        records: cleanedTree,
        current: 1,
        size: cleanedTree.length,
        total: cleanedTree.length
      };
    }
    return { data: [], records: [], current: 1, size: 0, total: 0 };
  },
  transform: response => response,
  columns: () => [
    { prop: 'id', label: 'ID', width: 80, align: 'center', fixed: 'left' },
    {
      prop: 'hierarchyIndex',
      label: '序号',
      width: 100,
      align: 'center',
      formatter: row => {
        return <span class="hierarchy-index">{row.hierarchyIndex || '-'}</span>;
      }
    },
    { prop: 'name', label: '菜单名称', width: 200, align: 'center', className: 'menu-name-column' },
    {
      prop: 'menuType',
      label: '菜单类型',
      width: 100,
      align: 'center',
      formatter: row => {
        const menuType = row.menuType || 'menu';
        const typeMap: Record<string, { text: string; type: any }> = {
          directory: { text: '目录', type: 'primary' },
          menu: { text: '菜单', type: 'success' }
        };
        const config = typeMap[menuType] || { text: '菜单', type: 'info' };
        return (
          <ElTag size="small" type={config.type}>
            {config.text}
          </ElTag>
        );
      }
    },
    {
      prop: 'icon',
      label: '图标',
      width: 80,
      align: 'center',
      formatter: row => {
        if (row.icon) {
          return (
            <div class="flex-center">
              <Icon icon={row.icon} style="font-size: 18px" />
            </div>
          );
        }
        return <span class="text-tertiary">-</span>;
      }
    },
    { prop: 'path', label: '路由路径', minWidth: 180, align: 'center' },
    {
      prop: 'permission',
      label: '权限标识',
      minWidth: 150,
      align: 'center',
      formatter: row => {
        if (row.permission) {
          return (
            <ElTag size="small" type="info">
              {row.permission}
            </ElTag>
          );
        }
        return <span class="text-tertiary">-</span>;
      }
    },
    {
      prop: 'sort',
      label: '排序',
      width: 120,
      align: 'center',
      formatter: row => (
        <div class="flex items-center justify-center gap-4px">
          <ElButton
            link
            type="primary"
            icon={Top}
            onClick={() => handleMove(row, 'up')}
            disabled={isFirst(row.id)}
            title="上移"
          />
          <span class="sort-value">{row.sort}</span>
          <ElButton
            link
            type="primary"
            icon={Bottom}
            onClick={() => handleMove(row, 'down')}
            disabled={isLast(row.id)}
            title="下移"
          />
        </div>
      )
    },
    {
      prop: 'status',
      label: '状态',
      width: 80,
      align: 'center',
      formatter: row => {
        if (row.status !== undefined) {
          return (
            <ElSwitch
              v-model={row.status}
              activeValue={1}
              inactiveValue={0}
              onChange={(val: number) => handleStatusChange(row, val)}
            />
          );
        }
        return null;
      }
    },
    {
      prop: 'operate',
      label: '操作',
      width: 280,
      align: 'center',
      fixed: 'right',
      formatter: row => (
        <div class="flex-center gap-8px">
          {row.menuType === 'directory' && (
            <ElButton
              type="primary"
              plain
              size="small"
              icon={Plus}
              onClick={() => handleAddChild(row)}
              title="添加子菜单"
            >
              添加子菜单
            </ElButton>
          )}
          <ElButton type="primary" plain size="small" onClick={() => handleEdit(row.id)}>
            {$t('common.edit')}
          </ElButton>
          <ElPopconfirm title={$t('common.confirmDelete')} onConfirm={() => handleDelete(row.id)}>
            {{
              reference: () => (
                <ElButton type="danger" plain size="small">
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

const {
  drawerVisible,
  operateType,
  editingData,
  handleAdd: _handleAdd,
  onDeleted
} = useTableOperate(data, 'id', getData);

const isAddingChild = ref(false);
const parentMenuName = ref('');

// 重写 handleEdit
function handleEdit(id: number) {
  isAddingChild.value = false;
  parentMenuName.value = '';
  operateType.value = 'edit';

  const foundItem = flatMenuList.value.find(item => item.id === id);
  editingData.value = foundItem ? jsonClone(foundItem) : null;

  drawerVisible.value = true;
}

// 重写 handleAdd
function handleAdd() {
  isAddingChild.value = false;
  parentMenuName.value = '';
  _handleAdd();
}

// 移动菜单
async function handleMove(row: Api.SystemManage.Menu, direction: 'up' | 'down') {
  const siblings = findSiblings(row.id);
  if (!siblings || siblings.length < 2) {
    ElNotification({
      title: '提示',
      message: '无法移动：同级菜单不足',
      type: 'warning',
      duration: 3000
    });
    return;
  }

  const currentIndex = siblings.findIndex(item => item.id === row.id);
  const targetIndex = direction === 'up' ? currentIndex - 1 : currentIndex + 1;

  if (targetIndex < 0 || targetIndex >= siblings.length) {
    ElNotification({
      title: '提示',
      message: direction === 'up' ? '已经是第一个了' : '已经是最后一个了',
      type: 'warning',
      duration: 3000
    });
    return;
  }

  const targetRow = siblings[targetIndex];

  if (row.parentId !== targetRow.parentId) {
    ElNotification({
      title: '错误',
      message: '只能在同级菜单之间排序',
      type: 'error',
      duration: 3000
    });
    return;
  }

  try {
    await Promise.all([
      fetchUpdateMenu(row.id, { sort: targetRow.sort }),
      fetchUpdateMenu(targetRow.id, { sort: row.sort })
    ]);

    ElNotification({
      title: '成功',
      message: '排序更新成功',
      type: 'success',
      duration: 3000
    });

    await getData();
    tableKey.value++;
    await authStore.getUserInfo();
    routeStore.rebuildRoutes();
  } catch (error) {
    ElNotification({
      title: '错误',
      message: '排序更新失败',
      type: 'error',
      duration: 3000
    });
  }
}

// 状态变更
async function handleStatusChange(row: Api.SystemManage.Menu, val: number) {
  try {
    await fetchUpdateMenu(row.id, { status: val });
    ElNotification({
      title: '成功',
      message: `${val === 1 ? '启用' : '停用'}成功`,
      type: 'success',
      duration: 3000
    });
    await getData();
    await authStore.getUserInfo();
    routeStore.rebuildRoutes();
  } catch (error) {
    ElNotification({
      title: '错误',
      message: '操作失败',
      type: 'error',
      duration: 3000
    });
    row.status = val === 1 ? 0 : 1;
  }
}

// 添加子菜单
function handleAddChild(row: Api.SystemManage.Menu) {
  operateType.value = 'add';
  isAddingChild.value = true;
  parentMenuName.value = row.name;
  editingData.value = jsonClone(row);
  drawerVisible.value = true;
}

// 删除菜单
async function handleDelete(id: number) {
  const { error } = await fetchDeleteMenu(id);

  if (!error) {
    ElNotification({
      title: '成功',
      message: '删除成功',
      type: 'success',
      duration: 3000
    });
    onDeleted();
    await authStore.getUserInfo();
    routeStore.rebuildRoutes();
  } else {
    ElNotification({
      title: '错误',
      message: error.msg || '删除失败',
      type: 'error',
      duration: 3000
    });
  }
}

// 菜单操作提交成功后的处理
async function handleMenuSubmitted() {
  await getDataByPage();
  await authStore.getUserInfo();
  routeStore.rebuildMenus();
}

// 监听搜索文本变化
async function handleFilterChange() {
  if (!filterText.value) {
    data.value = originalTreeData.value;
  } else {
    data.value = filterMenuTree(originalTreeData.value);
  }
}

// 刷新数据
async function refreshData() {
  await getDataByPage();
  await authStore.getUserInfo();
  routeStore.rebuildRoutes();
}
</script>

<template>
  <div class="page-container menu-management-page">
    <!-- Hero 区域 -->
    <section
      v-if="heroVisible"
      class="hero-section"
      :style="{
        background: 'var(--sx-hero-bg)',
        border: '1px solid var(--sx-hero-border)',
        borderRadius: 'var(--sx-hero-radius)',
        boxShadow: 'var(--sx-hero-shadow)',
        padding: 'var(--sx-hero-padding)'
      }"
    >
      <div class="hero-content">
        <div class="hero-title-row">
          <span class="hero-icon">
            <ElIcon><MenuIcon /></ElIcon>
          </span>
          <h2>{{ $t('page.manage.menu.title') }}</h2>
          <p class="hero-desc">管理系统菜单结构，支持树形层级展示与拖拽排序</p>
        </div>
      </div>
      <div class="hero-actions">
        <ElButton size="small" :loading="loading" @click="refreshData">
          <ElIcon><Refresh /></ElIcon>
          刷新
        </ElButton>
      </div>
    </section>

    <!-- 内容卡片 -->
    <div
      class="content-card menu-content-card"
      :style="{
        background: 'var(--sx-content-card-bg)',
        border: '1px solid var(--sx-content-card-border)',
        borderRadius: 'var(--sx-content-card-radius)',
        boxShadow: 'var(--sx-content-card-shadow)',
        padding: 'var(--sx-content-card-padding)'
      }"
    >
      <!-- 工具栏 -->
      <div class="card-toolbar">
        <div class="toolbar-head">
          <span class="toolbar-title">菜单列表</span>
          <span class="toolbar-desc">管理菜单层级结构、路由配置与权限标识</span>
        </div>
        <div class="toolbar-actions">
          <ElButton type="primary" size="small" @click="handleAdd">
            <ElIcon><Plus /></ElIcon>
            新增菜单
          </ElButton>
        </div>
      </div>

      <!-- 搜索工具栏 -->
      <div class="workbench-toolbar workbench-toolbar--history menus-toolbar">
        <div class="workbench-toolbar-left">
          <ElInput
            v-model="filterText"
            placeholder="搜索菜单名称 / 路径 / 权限"
            clearable
            style="width: 300px"
            @input="handleFilterChange"
            @clear="handleFilterChange"
          >
            <template #prefix>
              <ElIcon><Search /></ElIcon>
            </template>
          </ElInput>
        </div>
      </div>

      <!-- 数据表格 -->
      <div class="table-section">
        <ElTable
          :key="tableKey"
          v-loading="loading"
          :data="data"
          border
          class="data-table"
          row-key="id"
          :tree-props="{ children: 'children', indent: 20 }"
        >
          <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
        </ElTable>
      </div>
    </div>

    <!-- 抽屉 -->
    <MenuOperateDrawer
      v-model:visible="drawerVisible"
      :operate-type="operateType"
      :row-data="editingData"
      :is-adding-child="isAddingChild"
      :parent-menu-name="parentMenuName"
      @submitted="handleMenuSubmitted"
    />
  </div>
</template>

<style scoped lang="scss">
/* ============================================
	   1. 页面容器与布局
	   ============================================ */
.menu-management-page {
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
	   3. 内容卡片
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
	   4. 搜索工具栏
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
	   5. 数据表格
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

  --el-table-border-color: var(--sx-table-border);
  --el-table-header-bg-color: var(--sx-table-header-bg);
  --el-table-row-hover-bg-color: var(--sx-table-row-hover);

  :deep(th.el-table__cell) {
    background-color: var(--sx-table-header-bg);
    color: var(--sx-table-header-text);
    font-weight: 600;
    border-radius: 0;
  }

  :deep(td.el-table__cell) {
    border-radius: 0;
  }

  :deep(.el-checkbox__inner),
  :deep(.el-checkbox__inner::before),
  :deep(.el-checkbox__inner::after) {
    border-radius: 0;
  }

  &--striped :deep(.el-table__body tr.el-table__row--striped td) {
    background-color: var(--sx-table-striped-bg);
  }
}

/* ============================================
	   6. 自定义样式
	   ============================================ */
.text-tertiary {
  color: var(--el-text-color-placeholder);
}

.sort-value {
  min-width: 24px;
  text-align: center;
  font-size: 13px;
  font-weight: 500;
}

:deep(.el-button--link.is-disabled) {
  opacity: 0.3;
  cursor: not-allowed;
}

.hierarchy-index {
  font-family: 'Courier New', monospace;
  font-weight: 600;
  color: var(--el-color-primary);
  font-size: 13px;
}

:deep(.menu-name-column) {
  .cell {
    padding-left: 8px !important;
  }
}

/* ============================================
	   7. 响应式设计
	   ============================================ */
@media (max-width: 768px) {
  .hero-section {
    flex-direction: column;
    align-items: flex-start;
  }

  .card-toolbar {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }
}
</style>
