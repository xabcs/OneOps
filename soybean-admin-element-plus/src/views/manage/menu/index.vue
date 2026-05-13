<script setup lang="tsx">
import { ref, computed, onMounted } from 'vue';
import { useBoolean } from '@sa/hooks';
import { fetchGetMenuTree, fetchDeleteMenu, fetchUpdateMenu } from '@/service/api';
import { defaultTransform, useTableOperate, useUIPaginatedTable } from '@/hooks/common/table';
import { $t } from '@/locales';
import { Icon } from '@iconify/vue';
import { useAuthStore } from '@/store/modules/auth';
import { useRouteStore } from '@/store/modules/route';
import { jsonClone } from '@sa/utils';
import MenuOperateDrawer from './modules/menu-operate-drawer.vue';
import { ElNotification } from 'element-plus';
import { Top, Bottom, Plus } from '@element-plus/icons-vue';

defineOptions({ name: 'MenuManage' });

const authStore = useAuthStore();
const routeStore = useRouteStore();

const filterText = ref('');
const flatMenuList = ref<Api.SystemManage.Menu[]>([]);
const originalTreeData = ref<Api.SystemManage.Menu[]>([]);
const tableKey = ref(0); // 用于强制刷新表格

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
  // 先对每个层级的子节点按 sort 排序
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
          sort: item.sort,
          status: item.status
        };

        if (item.children && Array.isArray(item.children) && item.children.length > 0) {
          cleaned.children = sortTree(item.children);
        }

        return cleaned;
      });
  };

  // 分配层级序号
  const assignHierarchyIndex = (
    list: Api.SystemManage.Menu[],
    prefix: string = ''
  ): Api.SystemManage.Menu[] => {
    let siblingIndex = 1;

    return list.map(item => {
      const currentItem = { ...item };
      // 生成当前节点的层级序号
      currentItem.hierarchyIndex = prefix ? `${prefix}.${siblingIndex}` : `${siblingIndex}`;

      // 递归处理子节点
      if (item.children && Array.isArray(item.children) && item.children.length > 0) {
        currentItem.children = assignHierarchyIndex(
          item.children,
          currentItem.hierarchyIndex
        );
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

// 查找兄弟节点（只在同级中查找）
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

// 查找子菜单
function findChildren(parentId: number): Api.SystemManage.Menu[] {
  const findInList = (list: Api.SystemManage.Menu[]): Api.SystemManage.Menu[] => {
    for (const item of list) {
      if (item.id === parentId) {
        return item.children || [];
      }
      if (item.children && item.children.length > 0) {
        const found = findInList(item.children);
        if (found.length > 0) return found;
      }
    }
    return [];
  };
  return findInList(originalTreeData.value);
}

// 获取下一个排序号
function getNextSort(parentId: number | null): number {
  const siblings = parentId ? findChildren(parentId) : flatMenuList.value;
  if (!siblings || siblings.length === 0) return 1;
  const maxSort = Math.max(...siblings.map(item => item.sort || 0));
  return maxSort + 1;
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

const { columns, columnChecks, data, loading, getData, getDataByPage } = useUIPaginatedTable({
  api: async () => {
    const { error, data } = await fetchGetMenuTree();
    if (!error && data && Array.isArray(data)) {
      const tree = data as Api.SystemManage.Menu[];

      // 清理数据
      const cleanedTree = cleanMenuTree(tree);
      originalTreeData.value = cleanedTree;

      // 转换为扁平列表
      flatMenuList.value = flattenMenuTree(cleanedTree);

      // 根据搜索文本决定是否过滤
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
    {
      prop: 'name',
      label: '菜单名称',
      width: 200,
      align: 'center',
      labelClassName: 'custom-header-center',
      className: 'menu-name-column'
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
          return <ElTag size="small" type="info">{row.permission}</ElTag>;
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
          <ElButton
            type="primary"
            plain
            size="small"
            onClick={() => handleEdit(row.id)}
          >
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
  handleEdit: _handleEdit,
  checkedRowKeys,
  onBatchDeleted,
  onDeleted
} = useTableOperate(data, 'id', getData);

const isAddingChild = ref(false);
const parentMenuName = ref('');

// 重写 handleEdit，支持树形数据查找
function handleEdit(id: number) {
  console.log('🎯 [菜单管理] handleEdit', { id, flatMenuList: flatMenuList.value });

  isAddingChild.value = false;
  parentMenuName.value = '';
  operateType.value = 'edit';

  // 在扁平列表中查找（包含所有层级的菜单）
  const foundItem = flatMenuList.value.find(item => item.id === id);
  editingData.value = foundItem ? jsonClone(foundItem) : null;

  console.log('✅ [菜单管理] handleEdit 完成', {
    operateType: operateType.value,
    editingData: editingData.value,
    foundItem
  });

  drawerVisible.value = true;
}

// 重写 handleAdd
function handleAdd() {
  isAddingChild.value = false;
  parentMenuName.value = '';
  _handleAdd();
}

// 移动菜单（同级排序）
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

  // 验证是否为同级节点
  if (row.parentId !== targetRow.parentId) {
    ElNotification({
      title: '错误',
      message: '只能在同级菜单之间排序',
      type: 'error',
      duration: 3000
    });
    return;
  }

  const originalSort = row.sort;
  const targetOriginalSort = targetRow.sort;

  console.log('🔄 菜单排序:', {
    direction,
    current: { id: row.id, name: row.name, sort: originalSort, hierarchyIndex: row.hierarchyIndex },
    target: { id: targetRow.id, name: targetRow.name, sort: targetOriginalSort, hierarchyIndex: targetRow.hierarchyIndex },
    parentId: row.parentId
  });

  try {
    await Promise.all([
      fetchUpdateMenu(row.id, { sort: targetOriginalSort }),
      fetchUpdateMenu(targetRow.id, { sort: originalSort })
    ]);

    ElNotification({
      title: '成功',
      message: '排序更新成功',
      type: 'success',
      duration: 3000
    });

    // 重新获取数据，这会触发重新计算层级序号
    await getData();

    // 强制刷新表格
    tableKey.value++;

    // 刷新用户信息和左侧菜单栏
    await authStore.getUserInfo();
    routeStore.rebuildMenus();
  } catch (error) {
    console.error('排序更新失败:', error);
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
    // 刷新左侧菜单栏
    await authStore.getUserInfo();
    routeStore.rebuildMenus();
  } catch (error) {
    console.error('状态更新失败:', error);
    ElNotification({
      title: '错误',
      message: '操作失败',
      type: 'error',
      duration: 3000
    });
    // 恢复原状态
    row.status = val === 1 ? 0 : 1;
  }
}

// 添加子菜单
function handleAddChild(row: Api.SystemManage.Menu) {
  operateType.value = 'add';
  isAddingChild.value = true;
  parentMenuName.value = row.name;
  editingData.value = {
    id: 0,
    name: '',
    icon: '',
    path: '',
    permission: '',
    parentId: row.id,
    sort: getNextSort(row.id),
    status: 1
  } as Api.SystemManage.Menu;
  drawerVisible.value = true;
}

// 批量删除
async function handleBatchDelete() {
  if (checkedRowKeys.value.length === 0) {
    ElNotification({
      title: '提示',
      message: '请选择要删除的菜单',
      type: 'warning',
      duration: 3000
    });
    return;
  }

  // TODO: 实现批量删除
  ElNotification({
    title: '提示',
    message: '批量删除功能待实现',
    type: 'warning',
    duration: 3000
  });
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
    // 刷新左侧菜单栏
    await authStore.getUserInfo();
    routeStore.rebuildMenus();
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
  // 刷新表格数据
  await getDataByPage();
  // 刷新左侧菜单栏
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
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <ElCard class="card-wrapper sm:flex-1-hidden">
      <template #header>
        <div class="flex items-center justify-between">
          <p>{{ $t('page.manage.menu.title') }}</p>
          <div class="flex items-center gap-12px">
            <ElInput
              v-model="filterText"
              placeholder="搜索菜单名称"
              clearable
              style="width: 200px"
              @input="handleFilterChange"
              @clear="handleFilterChange"
            >
              <template #prefix>
                <SvgIcon icon="ri:search-line" />
              </template>
            </ElInput>
            <ElButton type="primary" :icon="Plus" @click="handleAdd">
              新增菜单
            </ElButton>
            <ElButton :icon="Top" @click="getData">
              刷新
            </ElButton>
          </div>
        </div>
      </template>
      <div class="h-[calc(100%-52px)]">
        <ElTable
          v-loading="loading"
          height="100%"
          border
          class="sm:h-full"
          :data="data"
          :key="tableKey"
          row-key="id"
          :tree-props="{ children: 'children', indent: 20 }"
        >
          <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
        </ElTable>
      </div>
      <MenuOperateDrawer
        v-model:visible="drawerVisible"
        :operate-type="operateType"
        :row-data="editingData"
        :is-adding-child="isAddingChild"
        :parent-menu-name="parentMenuName"
        @submitted="handleMenuSubmitted"
      />
    </ElCard>
  </div>
</template>

<style lang="scss" scoped>
:deep(.el-card) {
  .ht50 {
    height: calc(100% - 50px);
  }
}

/* 自定义表头居中 */
:deep(.custom-header-center) {
  .cell {
    text-align: center !important;
  }
}

/* 菜单名称列左边距 */
:deep(.menu-name-column) {
  .cell {
    padding-left: 8px !important;
  }
}

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

/* 层级序号样式 */
.hierarchy-index {
  font-family: 'Courier New', monospace;
  font-weight: 600;
  color: var(--el-color-primary);
  font-size: 13px;
}
</style>
