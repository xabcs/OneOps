<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import {
  ElButton,
  ElMessage,
  ElOption,
  ElSelect,
  ElSpace,
  ElTable,
  ElTableColumn,
  ElTag,
  ElTooltip
} from 'element-plus';
import {
  fetchK8sActiveTerminalSessions,
  fetchK8sClusters,
  terminateK8sTerminalSession
} from '@/service/api/k8s';

defineOptions({ name: 'K8sSessions' });

const message = ElMessage;

const loading = ref(false);
const dataSource = ref<any[]>([]);
const activeSessions = ref<any[]>([]);

// 当前选中的集群
const selectedCluster = ref<number | null>(null);
const selectedClusterLabel = ref('');

// 可用的集群列表
const clusters = ref<any[]>([]);

const filters = reactive({
  clusterId: null as number | null,
  status: 'all'
});

const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0
});

// 会话时长格式化
const formatDuration = (seconds: number) => {
  if (seconds < 60) return `${seconds}秒`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}分${seconds % 60}秒`;
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return `${hours}小时${minutes}分`;
};

// 获取会话时长（实时计算活跃会话）
const getSessionDuration = (session: any) => {
  if (session.status === 'closed' && session.duration) {
    return formatDuration(session.duration);
  }
  // 对于活跃会话，计算从开始到现在的时间
  const startedAt = new Date(session.startedAt || session.last_activity);
  const now = new Date();
  const seconds = Math.floor((now.getTime() - startedAt.getTime()) / 1000);
  return formatDuration(seconds);
};

// 状态标签
const getStatusTag = (status: string) => {
  switch (status) {
    case 'active':
      return { type: 'success', text: '活跃' };
    case 'closed':
      return { type: 'info', text: '已关闭' };
    case 'error':
      return { type: 'danger', text: '异常' };
    default:
      return { type: 'info', text: status };
  }
};

// 加载集群列表
const loadClusters = async () => {
  try {
    const res = await fetchK8sClusters();
    clusters.value = res.data || [];
  } catch (error: any) {
    message.error(error.message || '加载集群列表失败');
  }
};

// 加载活跃会话
const loadActiveSessions = async () => {
  loading.value = true;
  try {
    const res = await fetchK8sActiveTerminalSessions();
    activeSessions.value = res || [];
  } catch (error: any) {
    message.error(error.message || '加载活跃会话失败');
  } finally {
    loading.value = false;
  }
};

// 刷新数据
const handleRefresh = () => {
  loadActiveSessions();
};

// 终止会话
const handleTerminate = async (row: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要终止该终端会话吗？\n\nPod: ${row.pod_name}\n命名空间: ${row.namespace}`,
      '确认终止',
      {
        type: 'warning',
        confirmButtonText: '终止',
        cancelButtonText: '取消'
      }
    );

    await terminateK8sTerminalSession(row.session_id);
    message.success('会话已终止');
    loadActiveSessions();
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || '终止会话失败');
    }
  }
};

// 定时刷新活跃会话
let refreshTimer: NodeJS.Timeout | null = null;

onMounted(() => {
  loadClusters();
  loadActiveSessions();

  // 每 10 秒刷新一次活跃会话
  refreshTimer = setInterval(() => {
    loadActiveSessions();
  }, 10000);
});

// 组件卸载时清除定时器
onBeforeUnmount(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer);
  }
});
</script>

<template>
  <div class="p-4">
    <!-- 页面标题 -->
    <div class="mb-4">
      <h2 class="text-xl font-bold">K8s 终端会话审计</h2>
      <p class="text-sm text-gray-500 mt-1">查看和管理 Pod 终端会话，记录所有终端操作</p>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
      <div class="bg-card p-4 rounded-lg border-l-4 border-green-500">
        <div class="text-sm text-gray-600">当前活跃会话</div>
        <div class="text-2xl font-bold text-green-600">{{ activeSessions.length }}</div>
      </div>
      <div class="bg-card p-4 rounded-lg border-l-4 border-blue-500">
        <div class="text-sm text-gray-600">总集群数</div>
        <div class="text-2xl font-bold text-blue-600">{{ clusters.length }}</div>
      </div>
      <div class="bg-card p-4 rounded-lg border-l-4 border-amber-500">
        <div class="text-sm text-gray-600">最后更新</div>
        <div class="text-lg font-semibold text-amber-600">
          {{ new Date().toLocaleTimeString('zh-CN') }}
        </div>
      </div>
    </div>

    <!-- 筛选栏 -->
    <div class="mb-4 flex items-center gap-4 bg-card p-4 rounded-lg">
      <div class="flex items-center gap-2">
        <span class="text-sm font-medium">状态:</span>
        <el-select v-model="filters.status" placeholder="请选择" style="width: 150px">
          <el-option label="全部" value="all" />
          <el-option label="活跃" value="active" />
          <el-option label="已关闭" value="closed" />
        </el-select>
      </div>

      <div class="flex-1" />

      <el-button type="primary" :loading="loading" @click="handleRefresh">
        刷新
      </el-button>
    </div>

    <!-- 活跃会话表格 -->
    <div class="mb-6">
      <h3 class="text-lg font-semibold mb-3 flex items-center gap-2">
        <span class="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
        活跃会话 ({{ activeSessions.length }})
      </h3>
      <el-table
        v-loading="loading"
        :data="activeSessions"
        stripe
        :empty-text="loading ? '加载中...' : '暂无活跃会话'"
      >
        <el-table-column prop="session_id" label="会话 ID" width="80" />
        <el-table-column prop="cluster_id" label="集群 ID" width="100" />
        <el-table-column prop="namespace" label="命名空间" width="150" />
        <el-table-column prop="pod_name" label="Pod 名称" min-width="200">
          <template #default="{ row }">
            <el-tooltip :content="`容器: ${row.container || '默认'}`" placement="top">
              <span>{{ row.pod_name }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="container" label="容器" width="150" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusTag('active').type" size="small">
              {{ getStatusTag('active').text }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="时长" width="120">
          <template #default="{ row }">
            {{ getSessionDuration(row) }}
          </template>
        </el-table-column>
        <el-table-column prop="last_activity" label="最后活动" width="160" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="danger" @click="handleTerminate(row)">
              终止
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 说明信息 -->
    <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
      <div class="flex items-start gap-3">
        <svg class="w-6 h-6 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <div class="flex-1">
          <h4 class="font-semibold text-blue-800 mb-2">审计说明</h4>
          <ul class="text-sm text-blue-700 space-y-1">
            <li>• 所有终端会话都会记录，包括会话开始/结束时间、用户信息、Pod 信息等</li>
            <li>• 活跃会话列表每 10 秒自动刷新一次</li>
            <li>• 管理员可以强制终止任何活跃的终端会话</li>
            <li>• 历史会话记录和命令审计功能正在开发中</li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { onBeforeUnmount } from 'vue';
export default {
  setup() {
    // onBeforeUnmount 已经在 script setup 中处理
  }
};
</script>
