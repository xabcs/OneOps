<script setup lang="ts">
  import { computed, onMounted, ref } from 'vue';
  import { useDebounceFn } from '@vueuse/core';
  import { ElMessage, ElMessageBox } from 'element-plus';
  import { useBoolean } from '@sa/hooks';
  import { Plus, Refresh, Search, User } from '@element-plus/icons-vue';
  import { fetchDeleteUserGroup, fetchGetUserGroupList, fetchUpdateUserGroup } from '@/service/api';
  import { useThemeStore } from '@/store/modules/theme';
  import { executeWithPermission } from '@/hooks/business/auth';
  import UserGroupOperateDrawer from './modules/user-group-operate-drawer.vue';
  import GroupMemberModal from './modules/group-member-modal.vue';

  defineOptions({ name: 'UserGroupManage' });

  const themeStore = useThemeStore();

  // Hero区域显示状态
  const heroVisible = computed(() => themeStore.content.hero.visible !== false);

  const searchParams = ref({
    page: 1,
    pageSize: 10,
    keyword: ''
  });

  const loading = ref(false);
  const data = ref<Api.SystemManage.UserGroup[]>([]);
  const total = ref(0);

  async function getData() {
    loading.value = true;
    try {
      const { data: res, error } = await fetchGetUserGroupList(searchParams.value);
      if (!error && res) {
        data.value = res.list || [];
        total.value = res.total || 0;
      }
    } finally {
      loading.value = false;
    }
  }

  function handlePageChange(page: number) {
    searchParams.value.page = page;
    getData();
  }

  function handleSizeChange(size: number) {
    searchParams.value.pageSize = size;
    searchParams.value.page = 1;
    getData();
  }

  // 抽屉（新增/编辑）
  const drawerVisible = ref(false);
  const operateType = ref<UI.TableOperateType>('add');
  const editingData = ref<Api.SystemManage.UserGroup | null>(null);

  function handleAdd() {
    operateType.value = 'add';
    editingData.value = null;
    drawerVisible.value = true;
  }

  async function handleEditClick(id: number) {
    await executeWithPermission('system.user.update', async () => {
      operateType.value = 'edit';
      editingData.value = data.value.find(g => g.id === id) || null;
      drawerVisible.value = true;
    });
  }

  async function handleAddClick() {
    await executeWithPermission('system.user.create', async () => {
      handleAdd();
    });
  }

  // 成员管理弹窗
  const { bool: memberModalVisible, setTrue: openMemberModal } = useBoolean();
  const currentGroup = ref<Api.SystemManage.UserGroup | null>(null);

  function handleManageMembers(id: number) {
    const group = data.value.find(g => g.id === id);
    if (group) {
      currentGroup.value = group;
      openMemberModal();
    }
  }

  async function handleDelete(id: number) {
    const group = data.value.find(g => g.id === id);
    await ElMessageBox.confirm(`确认删除用户组 "${group?.name ?? id}" 吗？将级联删除其成员与集群组绑定。`, '提示', {
      type: 'warning'
    });
    await executeWithPermission('system.user.delete', async () => {
      const { error } = await fetchDeleteUserGroup(id);
      if (!error) {
        ElMessage.success('删除成功');
        await getData();
      }
    });
  }

  async function handleStatusChange(row: Api.SystemManage.UserGroup, val: number) {
    await executeWithPermission('system.user.update', async () => {
      const { error } = await fetchUpdateUserGroup(row.id, {
        name: row.name,
        description: row.description,
        status: val
      });
      if (!error) {
        window.$message?.success(`${val === 1 ? '启用' : '禁用'}成功`);
        await getData();
      }
    });
  }

  function resetSearchParams() {
    searchParams.value = { page: 1, pageSize: 10, keyword: '' };
    getData();
  }

  // 搜索输入处理（防抖 300ms）
  const handleSearchInput = useDebounceFn(() => {
    searchParams.value.page = 1;
    getData();
  }, 300);

  function handleSearch() {
    searchParams.value.page = 1;
    getData();
  }

  onMounted(() => {
    getData();
  });
</script>

<template>
  <div class="table-page">
    <!-- Hero 区域 -->
    <ElCard v-if="heroVisible" shadow="hover" class="card-static msre-hero">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-12px">
          <ElIcon :size="20">
            <User />
          </ElIcon>
          <div class="flex flex-col gap-2px">
            <h2 class="m-0 text-18px font-bold">用户组管理</h2>
            <p class="m-0 text-13px opacity-70">平台用户组用于集群等资源的批量授权，组内成员自动继承组的集群权限</p>
          </div>
        </div>
        <ElButton size="small" :loading="loading" @click="getData">
          <ElIcon>
            <Refresh />
          </ElIcon>
          刷新
        </ElButton>
      </div>
    </ElCard>

    <!-- 内容卡片 -->
    <ElCard shadow="hover">
      <template #header>
        <div class="flex items-center justify-between">
          <div class="flex flex-col gap-2px">
            <span class="text-16px font-bold">用户组列表</span>
            <span class="text-13px opacity-70">维护用户组与组成员，供集群授权时按组绑定</span>
          </div>
          <PermissionButton code="system.user.create" type="primary" size="small" @click="handleAddClick">
            <ElIcon>
              <Plus />
            </ElIcon>
            新增用户组
          </PermissionButton>
        </div>
      </template>

      <!-- 搜索工具栏 -->
      <ElSpace wrap class="mb-16px">
        <ElInput
          v-model="searchParams.keyword"
          placeholder="搜索名称/编码"
          clearable
          style="width: 220px"
          :prefix-icon="Search"
          @input="handleSearchInput"
        />
        <ElButton @click="resetSearchParams">
          <ElIcon>
            <Refresh />
          </ElIcon>
          重置
        </ElButton>
        <ElButton type="primary" @click="handleSearch">
          <ElIcon>
            <Search />
          </ElIcon>
          搜索
        </ElButton>
      </ElSpace>

      <!-- 数据表格 -->
      <div class="table-scroll-wrap">
        <ElTable v-loading="loading" :data="data" border stripe row-key="id" height="100%">
          <ElTableColumn prop="id" label="ID" width="70" />
          <ElTableColumn prop="code" label="编码" min-width="120" />
          <ElTableColumn prop="name" label="名称" min-width="120" />
          <ElTableColumn prop="memberCount" label="成员数" width="90" />
          <ElTableColumn prop="description" label="描述" min-width="140" show-overflow-tooltip />
          <ElTableColumn prop="createdAt" label="创建时间" width="170" />
          <ElTableColumn label="状态" width="90">
            <template #default="{ row }">
              <ElSwitch :model-value="row.status === 1" @change="(val: any) => handleStatusChange(row, val ? 1 : 0)" />
            </template>
          </ElTableColumn>
          <ElTableColumn label="操作" align="center" width="220" fixed="right" class-name="msre-table-actions">
            <template #default="{ row }">
              <ElButton link type="primary" size="small" @click="handleManageMembers(row.id)">成员管理</ElButton>
              <PermissionButton
                link
                type="primary"
                size="small"
                code="system.user.update"
                @click="handleEditClick(row.id)"
              >
                编辑
              </PermissionButton>
              <PermissionButton link type="danger" size="small" code="system.user.delete" @click="handleDelete(row.id)">
                删除
              </PermissionButton>
            </template>
          </ElTableColumn>
        </ElTable>
      </div>

      <!-- 分页 -->
      <div v-if="total" class="mt-16px flex justify-end">
        <ElPagination
          layout="total, sizes, prev, pager, next, jumper"
          :total="total"
          :current-page="searchParams.page"
          :page-size="searchParams.pageSize"
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </ElCard>

    <!-- 抽屉和弹窗 -->
    <UserGroupOperateDrawer
      v-model:visible="drawerVisible"
      :operate-type="operateType"
      :row-data="editingData"
      @submitted="getData"
    />
    <GroupMemberModal v-model:visible="memberModalVisible" :group-data="currentGroup" @submitted="getData" />
  </div>
</template>

<style scoped></style>
