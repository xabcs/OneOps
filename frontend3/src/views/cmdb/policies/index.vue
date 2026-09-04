<script setup lang="ts">
  import { onMounted, ref } from 'vue';
  import { ElMessageBox } from 'element-plus';
  import {
    fetchDeleteAccessPolicy,
    fetchGetAccessPolicies,
    fetchServerOptions,
    fetchUpdateAccessPolicy
  } from '@/service/api/cmdb';
  import {
    fetchGetBusinessUnits,
    fetchGetServerGroups,
    fetchGetServerTags,
    fetchRoleOptions,
    fetchUserOptions
  } from '@/service/api';
  import PolicyOperateDrawer from './modules/policy-operate-drawer.vue';

  defineOptions({
    name: 'CMDBAccessPolicies'
  });

  const loading = ref(false);
  const policies = ref<Bastion.AccessPolicy[]>([]);
  const total = ref(0);

  // 分页
  const pagination = ref({
    page: 1,
    pageSize: 20
  });

  // Drawer 状态
  const drawerVisible = ref(false);
  const operateType = ref<UI.TableOperateType>('add');
  const editingData = ref<Bastion.AccessPolicy | null>(null);

  // ========== 资产数据（用于列表展示名称解析）==========
  const users = ref<{ id: number; username: string; nickname: string }[]>([]);
  const roles = ref<{ id: number; name: string; code: string }[]>([]);
  const serverGroups = ref<CMDB.ServerGroup[]>([]);
  const businessUnits = ref<CMDB.BusinessUnit[]>([]);
  const serverTags = ref<CMDB.ServerTag[]>([]);
  const servers = ref<{ id: number; hostname: string; ip: string }[]>([]);

  // 获取策略列表
  async function getPolicies() {
    loading.value = true;
    try {
      const { data } = await fetchGetAccessPolicies({
        page: pagination.value.page,
        pageSize: pagination.value.pageSize
      });
      policies.value = data?.list || [];
      total.value = data?.total || 0;
    } catch (error) {
      window.$message?.error('获取策略列表失败');
    } finally {
      loading.value = false;
    }
  }

  // 新增策略
  function handleCreate() {
    operateType.value = 'add';
    editingData.value = null;
    drawerVisible.value = true;
  }

  // 编辑策略
  function handleEdit(policy: Bastion.AccessPolicy) {
    operateType.value = 'edit';
    editingData.value = policy;
    drawerVisible.value = true;
  }

  // 删除策略
  async function handleDelete(policy: Bastion.AccessPolicy) {
    try {
      await ElMessageBox.confirm(`确定要删除策略"${policy.name}"吗？`, '确认删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      });
    } catch {
      return;
    }

    try {
      await fetchDeleteAccessPolicy(policy.id);
      window.$message?.success('删除成功');
      getPolicies();
    } catch (error: unknown) {
      window.$message?.error(error instanceof Error ? error.message : '删除失败');
    }
  }

  // 切换策略状态
  async function handleToggleStatus(policy: Bastion.AccessPolicy) {
    const newStatus = policy.status === 1 ? 0 : 1;
    try {
      await fetchUpdateAccessPolicy(policy.id, { status: newStatus });
      window.$message?.success('状态更新成功');
      getPolicies();
    } catch (error: unknown) {
      window.$message?.error(error instanceof Error ? error.message : '状态更新失败');
    }
  }

  // 格式化时间窗口
  function formatTimeWindow(timeWindow?: Bastion.AccessPolicy['timeWindow']): string {
    if (!timeWindow?.start || !timeWindow?.end) return '不限';
    const days = ['周一', '周二', '周三', '周四', '周五', '周六', '周日'];
    const dayNames = (timeWindow.days || [])
      .map(d => days[d - 1])
      .filter(Boolean)
      .join('、');
    return `${timeWindow.start}-${timeWindow.end} (${dayNames || '每天'})`;
  }

  // 获取状态标签
  function getStatusTag(policy: Bastion.AccessPolicy) {
    return policy.status === 1 ? '启用' : '禁用';
  }

  // ========== 获取资产数据 ==========
  async function getAssetData() {
    try {
      const [usersRes, rolesRes, groupsRes, businessRes, tagsRes, serversRes] = await Promise.allSettled([
        fetchUserOptions(),
        fetchRoleOptions(),
        fetchGetServerGroups(),
        fetchGetBusinessUnits(),
        fetchGetServerTags(),
        fetchServerOptions()
      ]);

      if (usersRes.status === 'fulfilled') users.value = usersRes.value.data || [];
      if (rolesRes.status === 'fulfilled') roles.value = rolesRes.value.data || [];
      if (groupsRes.status === 'fulfilled') serverGroups.value = groupsRes.value.data?.groups || [];
      if (businessRes.status === 'fulfilled') businessUnits.value = businessRes.value.data || [];
      if (tagsRes.status === 'fulfilled') serverTags.value = tagsRes.value.data || [];
      if (serversRes.status === 'fulfilled') servers.value = serversRes.value.data || [];
    } catch (error) {
      console.error('获取资产数据失败:', error);
    }
  }

  // ========== 辅助函数 ==========
  function getSubjectName(type: string, id: number): string {
    if (type === 'user') {
      const user = users.value.find(u => u.id === id);
      return user ? user.username : `ID: ${id}`;
    }
    if (type === 'role') {
      const role = roles.value.find(r => r.id === id);
      return role ? role.name : `ID: ${id}`;
    }
    return `ID: ${id}`;
  }

  function getAssetScopeTypeName(type: string): string {
    const typeMap: Record<string, string> = {
      all: '全部资产',
      server: '单台服务器',
      group: '主机分组',
      business: '业务系统',
      tag: '标签'
    };
    return typeMap[type] || type;
  }

  function getAssetScopeName(type: string, id: number): string {
    if (type === 'server') {
      const server = servers.value.find(s => s.id === id);
      return server ? `${server.hostname} (${server.ip})` : `ID: ${id}`;
    }
    if (type === 'group') {
      const findGroup = (groups: CMDB.ServerGroup[], targetId: number): string => {
        for (const group of groups) {
          if (group.id === targetId) return group.name;
          if (group.children) {
            const found = findGroup(group.children, targetId);
            if (found) return found;
          }
        }
        return '';
      };
      const name = findGroup(serverGroups.value, id);
      return name || `ID: ${id}`;
    }
    if (type === 'business') {
      const findBusiness = (units: CMDB.BusinessUnit[], targetId: number): string => {
        for (const unit of units) {
          if (unit.id === targetId) return unit.name;
          if (unit.children) {
            const found = findBusiness(unit.children, targetId);
            if (found) return found;
          }
        }
        return '';
      };
      const name = findBusiness(businessUnits.value, id);
      return name || `ID: ${id}`;
    }
    if (type === 'tag') {
      const tag = serverTags.value.find(t => t.id === id);
      return tag ? tag.name : `ID: ${id}`;
    }
    return `ID: ${id}`;
  }

  onMounted(() => {
    getPolicies();
    getAssetData();
  });
</script>

<template>
  <div class="access-policies-page">
    <ElCard shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">访问策略管理</span>
          <PermissionButton code="cmdb.access_policy.create" type="primary" @click="handleCreate">
            新增策略
          </PermissionButton>
        </div>
      </template>

      <!-- 策略列表 -->
      <ElTable v-loading="loading" :data="policies" stripe style="width: 100%">
        <ElTableColumn prop="id" label="ID" width="60" />

        <ElTableColumn prop="name" label="策略名称" min-width="150" />

        <ElTableColumn label="授权对象" width="200">
          <template #default="{ row }">
            <ElTag size="small" type="primary">
              {{ row.subjectType === 'user' ? '用户' : row.subjectType === 'role' ? '角色' : '用户组' }}
            </ElTag>
            <span style="margin-left: 8px">
              {{ getSubjectName(row.subjectType, row.subjectId) }}
            </span>
          </template>
        </ElTableColumn>

        <ElTableColumn label="资产范围" width="200">
          <template #default="{ row }">
            <ElTag size="small" type="success">
              {{ getAssetScopeTypeName(row.assetScopeType) }}
            </ElTag>
            <span v-if="row.assetScopeType !== 'all'" style="margin-left: 8px">
              {{ getAssetScopeName(row.assetScopeType, row.assetScopeId) }}
            </span>
          </template>
        </ElTableColumn>

        <ElTableColumn label="允许账号" width="200">
          <template #default="{ row }">
            <ElTag
              v-for="(account, idx) in (row.loginAccounts || []).slice(0, 2)"
              :key="idx"
              size="small"
              style="margin-right: 4px"
            >
              {{ account }}
            </ElTag>
            <span v-if="(row.loginAccounts || []).length > 2" style="font-size: 12px; color: #909399">
              +{{ (row.loginAccounts || []).length - 2 }}
            </span>
          </template>
        </ElTableColumn>

        <ElTableColumn label="允许协议" width="120">
          <template #default="{ row }">
            <ElTag
              v-for="(protocol, idx) in row.protocols || []"
              :key="idx"
              size="small"
              :type="protocol === 'ssh' ? 'primary' : 'success'"
              style="margin-right: 4px"
            >
              {{ protocol.toUpperCase() }}
            </ElTag>
          </template>
        </ElTableColumn>

        <ElTableColumn label="时间窗口" width="180">
          <template #default="{ row }">
            {{ formatTimeWindow(row.timeWindow) }}
          </template>
        </ElTableColumn>

        <ElTableColumn label="状态" width="80">
          <template #default="{ row }">
            <ElTag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ getStatusTag(row) }}
            </ElTag>
          </template>
        </ElTableColumn>

        <ElTableColumn label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <PermissionButton code="cmdb.access_policy.update" type="primary" size="small" @click="handleEdit(row)">
              编辑
            </PermissionButton>
            <PermissionButton
              code="cmdb.access_policy.update"
              :type="row.status === 1 ? 'warning' : 'success'"
              size="small"
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 1 ? '禁用' : '启用' }}
            </PermissionButton>
            <PermissionButton code="cmdb.access_policy.delete" type="danger" size="small" @click="handleDelete(row)">
              删除
            </PermissionButton>
          </template>
        </ElTableColumn>
      </ElTable>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <ElPagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="getPolicies"
          @current-change="getPolicies"
        />
      </div>
    </ElCard>

    <PolicyOperateDrawer
      v-model:visible="drawerVisible"
      :operate-type="operateType"
      :row-data="editingData"
      @submitted="getPolicies"
    />
  </div>
</template>

<style scoped>
  .access-policies-page {
    padding: 16px;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .title {
    font-size: 16px;
    font-weight: 500;
  }

  .pagination-wrapper {
    display: flex;
    justify-content: center;
    margin-top: 16px;
  }
</style>
