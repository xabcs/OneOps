<script setup lang="tsx">
import { computed, onMounted, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Refresh, Plus, Delete } from '@element-plus/icons-vue';
import {
  fetchGetAllRoles,
  fetchGetAllPolicies,
  fetchGetCasbinRolePermissions,
  fetchSyncCommonAPIs
} from '@/service/api';
import { useThemeStore } from '@/store/modules/theme';
import { $t } from '@/locales';

defineOptions({ name: 'APIPermissionManage' });

const themeStore = useThemeStore();

// Hero区域显示状态
const heroVisible = computed(() => themeStore.contentTheme2.heroSection.visible !== false);

// 数据状态
const loading = ref(false);
const selectedRoleCode = ref('');
const allRoles = ref<Api.SystemManage.AllRole[]>([]);

// 统计数据
const stats = ref({
  totalAPIs: 0,
  totalPolicies: 0,
  roleStats: [] as Array<{ roleCode: string; permissionCount: number }>
});

// API权限列表数据
const apiList = ref<Array<{
  id: number;
  path: string;
  method: string;
  name: string;
  description: string;
  module: string;
  allowed: boolean;
}>>([]);

// 分页配置
const pagination = ref({
  current: 1,
  size: 20,
  total: 0
});

// 角色选项
const roleOptions = computed(() => {
  return allRoles.value.map(role => ({
    label: role.name,
    value: role.code
  }));
});

// 获取所有角色
async function getAllRoles() {
  try {
    const { error, data } = await fetchGetAllRoles();
    if (!error && data) {
      if (Array.isArray(data)) {
        allRoles.value = data;
      } else if (data && Array.isArray(data.records)) {
        allRoles.value = data.records;
      } else {
        allRoles.value = [];
      }
    } else {
      allRoles.value = [];
    }
  } catch (error) {
    console.error('获取角色列表失败:', error);
    allRoles.value = [];
  }
}

// 获取API权限统计
async function fetchStats() {
  try {
    loading.value = true;
    const { error, data } = await fetchGetAllPolicies();
    if (!error && data) {
      // 计算统计数据
      const policies = data || [];
      stats.value.totalPolicies = policies.length;

      // 统计唯一的API数量
      const uniqueAPIs = new Set(policies.map(p => `${p.object}:${p.action}`));
      stats.value.totalAPIs = uniqueAPIs.size;

      // 按角色统计
      const roleMap = new Map<string, number>();
      policies.forEach(p => {
        const roleCode = p.subject;
        roleMap.set(roleCode, (roleMap.get(roleCode) || 0) + 1);
      });

      stats.value.roleStats = Array.from(roleMap.entries())
        .map(([roleCode, permissionCount]) => ({
          roleCode,
          permissionCount
        }))
        .sort((a, b) => b.permissionCount - a.permissionCount); // 按权限数量排序
    }
  } catch (error) {
    console.error('获取统计数据失败:', error);
    ElMessage.error('获取统计数据失败');
  } finally {
    loading.value = false;
  }
}

// 获取角色权限列表
async function fetchRolePermissions() {
  if (!selectedRoleCode.value) {
    ElMessage.warning('请先选择角色');
    return;
  }

  loading.value = true;
  try {
    const { error, data } = await fetchGetCasbinRolePermissions(selectedRoleCode.value);
    if (!error && data) {
      apiList.value = data.map((api, index) => ({
        id: index + 1,
        path: api.path,
        method: api.method,
        name: api.name || api.path,
        description: api.description || '-',
        module: api.module || '未分类',
        allowed: true
      }));
      pagination.value.total = apiList.value.length;
    } else {
      apiList.value = [];
      pagination.value.total = 0;
    }
  } catch (error) {
    console.error('获取角色权限失败:', error);
    ElMessage.error('获取角色权限失败');
    apiList.value = [];
  } finally {
    loading.value = false;
  }
}

// 处理角色变更
function handleRoleChange() {
  if (selectedRoleCode.value) {
    fetchRolePermissions();
  } else {
    apiList.value = [];
    pagination.value.total = 0;
  }
}

// 刷新数据
async function handleRefresh() {
  await Promise.all([
    fetchStats(),
    selectedRoleCode.value ? fetchRolePermissions() : Promise.resolve()
  ]);
  ElMessage.success('刷新成功');
}

// 批量分配权限
async function handleBatchAssign() {
  if (!selectedRoleCode.value) {
    ElMessage.warning('请先选择角色');
    return;
  }

  try {
    await ElMessageBox.confirm(
      '此功能将打开批量权限分配界面，是否继续？',
      '批量分配权限',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'info'
      }
    );

    // TODO: 打开批量分配对话框
    ElMessage.info('批量分配功能开发中...');
  } catch {
    // 用户取消
  }
}

// 同步常用API
async function handleSyncAPIs() {
  try {
    await ElMessageBox.confirm(
      '这将同步系统常用API到权限库，并为其添加业务信息。是否继续？',
      '同步常用API',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    );

    loading.value = true;
    const { error } = await fetchSyncCommonAPIs();
    if (!error) {
      ElMessage.success('同步成功');
      await handleRefresh();
    } else {
      ElMessage.error('同步失败');
    }
  } catch {
    // 用户取消
  } finally {
    loading.value = false;
  }
}

// 表格列配置
const columns = [
  { prop: 'index', type: 'index', label: $t('common.index'), width: 64 },
  {
    prop: 'method',
    label: 'HTTP方法',
    width: 100,
    formatter: (row: any) => {
      const methodColors: Record<string, string> = {
        GET: 'success',
        POST: 'primary',
        PUT: 'warning',
        DELETE: 'danger',
        PATCH: 'info'
      };
      return <el-tag type={methodColors[row.method] || 'info'} size="small">{row.method}</el-tag>;
    }
  },
  { prop: 'path', label: 'API路径', minWidth: 200 },
  { prop: 'name', label: 'API名称', minWidth: 150 },
  { prop: 'module', label: '所属模块', width: 120 },
  { prop: 'description', label: '描述', minWidth: 200 }
];

// 初始化
onMounted(() => {
  getAllRoles();
  fetchStats();
});
</script>

<template>
  <div class="api-permission-manage">
    <!-- Hero区域 -->
    <div v-if="heroVisible" class="mb-16px">
      <h3 class="text-18px font-semibold text-gray-900">API权限管理</h3>
      <div class="text-14px text-gray-600 mt-8px">
        Level 4 API级权限控制，直接管理API端点访问权限
      </div>
    </div>

    <!-- 权限统计卡片 -->
    <el-card v-loading="loading" shadow="never" class="mb-16px">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-16px font-semibold">权限统计</span>
          <div class="flex gap-8px">
            <el-button
              type="primary"
              :icon="Refresh"
              size="small"
              @click="handleRefresh"
            >
              刷新
            </el-button>
            <el-button
              size="small"
              @click="handleSyncAPIs"
            >
              同步常用API
            </el-button>
          </div>
        </div>
      </template>

      <el-row v-if="stats.totalPolicies > 0" :gutter="16">
        <el-col :span="6">
          <el-statistic title="系统API总数" :value="stats.totalAPIs" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="策略总数" :value="stats.totalPolicies" />
        </el-col>
        <el-col
          v-for="roleStat in stats.roleStats.slice(0, 2)"
          :key="roleStat.roleCode"
          :span="6"
        >
          <el-statistic
            :title="`${roleStat.roleCode}角色权限`"
            :value="roleStat.permissionCount"
          />
        </el-col>
      </el-row>
      <el-empty v-else description="暂无权限数据，请先同步常用API或为角色分配权限" />
    </el-card>

    <!-- 角色权限管理 -->
    <el-card shadow="never">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-16px font-semibold">角色权限管理</span>
          <el-button
            v-if="selectedRoleCode"
            type="primary"
            :icon="Plus"
            size="small"
            @click="handleBatchAssign"
          >
            批量分配权限
          </el-button>
        </div>
      </template>

      <!-- 角色选择 -->
      <div class="mb-16px">
        <el-select
          v-model="selectedRoleCode"
          placeholder="请选择角色查看权限"
          style="width: 200px"
          clearable
          filterable
          @change="handleRoleChange"
        >
          <el-option
            v-for="role in allRoles"
            :key="role.id"
            :label="role.name"
            :value="role.code"
          >
            <span>{{ role.name }}</span>
            <span style="float: right; color: var(--el-text-color-secondary); font-size: 12px">
              {{ role.code }}
            </span>
          </el-option>
        </el-select>
        <span v-if="!selectedRoleCode" class="ml-8px text-14px text-gray-500">
          选择角色后将显示该角色的API权限列表
        </span>
        <span v-else class="ml-8px text-14px text-primary">
          已选择角色，共 {{ apiList.length }} 个API权限
        </span>
      </div>

      <!-- API端点列表 -->
      <el-table
        v-loading="loading"
        :data="apiList"
        stripe
        border
        :max-height="600"
      >
        <el-table-column
          v-for="col in columns"
          :key="col.prop"
          :prop="col.prop"
          :type="col.type"
          :label="col.label"
          :width="col.width"
          :min-width="col.minWidth"
          :formatter="col.formatter"
        />

        <template #empty>
          <el-empty description="请选择角色查看权限列表" />
        </template>
      </el-table>
    </el-card>
  </div>
</template>

<style scoped>
.api-permission-manage {
  padding: 16px;
}
</style>
