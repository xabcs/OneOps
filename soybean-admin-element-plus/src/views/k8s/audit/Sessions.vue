<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
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
import { fetchK8sActiveTerminalSessions, fetchK8sClusters, terminateK8sTerminalSession } from '@/service/api/k8s';
defineOptions({
  setup() {
    // onBeforeUnmount 已经在 script setup 中处理
  }
});
</script>

<script>

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
      <p class="mt-1 text-sm text-gray-500">查看和管理 Pod 终端会话，记录所有终端操作</p>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 mb-4 gap-4 md:grid-cols-3">
      <div class="bg-card border-l-4 border-green-500 rounded-lg p-4">
        <div class="text-sm text-gray-600">当前活跃会话</div>
        <div class="text-2xl text-green-600 font-bold">{{ activeSessions.length }}</div>
      </div>
      <div class="bg-card border-l-4 border-blue-500 rounded-lg p-4">
        <div class="text-sm text-gray-600">总集群数</div>
        <div class="text-2xl text-blue-600 font-bold">{{ clusters.length }}</div>
      </div>
      <div class="bg-card border-l-4 border-amber-500 rounded-lg p-4">
        <div class="text-sm text-gray-600">最后更新</div>
        <div class="text-lg text-amber-600 font-semibold">
          {{ new Date().toLocaleTimeString('zh-CN') }}
        </div>
      </div>
    </div>

    <!-- 筛选栏 -->
    <div class="bg-card mb-4 flex items-center gap-4 rounded-lg p-4">
      <div class="flex items-center gap-2">
        <span class="text-sm font-medium">状态:</span>
        <ElSelect v-model="filters.status" placeholder="请选择" style="width: 150px">
          <ElOption label="全部" value="all" />
          <ElOption label="活跃" value="active" />
          <ElOption label="已关闭" value="closed" />
        </ElSelect>
      </div>

      <div class="flex-1" />

      <ElButton type="primary" :loading="loading" @click="handleRefresh">刷新</ElButton>
    </div>

    <!-- 活跃会话表格 -->
    <div class="mb-6">
      <h3 class="mb-3 flex items-center gap-2 text-lg font-semibold">
        <span class="h-2 w-2 animate-pulse rounded-full bg-green-500"></span>
        活跃会话 ({{ activeSessions.length }})
      </h3>
      <ElTable v-loading="loading" :data="activeSessions" stripe :empty-text="loading ? '加载中...' : '暂无活跃会话'">
        <ElTableColumn prop="session_id" label="会话 ID" width="80" />
        <ElTableColumn prop="cluster_id" label="集群 ID" width="100" />
        <ElTableColumn prop="namespace" label="命名空间" width="150" />
        <ElTableColumn prop="pod_name" label="Pod 名称" min-width="200">
          <template #default="{ row }">
            <ElTooltip :content="`容器组: ${row.container || '默认'}`" placement="top">
              <span>{{ row.pod_name }}</span>
            </ElTooltip>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="container" label="容器组" width="150" />
        <ElTableColumn label="状态" width="100">
          <template #default="{ row }">
            <ElTag :type="getStatusTag('active').type" size="small">
              {{ getStatusTag('active').text }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="时长" width="120">
          <template #default="{ row }">
            {{ getSessionDuration(row) }}
          </template>
        </ElTableColumn>
        <ElTableColumn prop="last_activity" label="最后活动" width="160" />
        <ElTableColumn label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <ElButton size="small" type="danger" @click="handleTerminate(row)">终止</ElButton>
          </template>
        </ElTableColumn>
      </ElTable>
    </div>

    <!-- 说明信息 -->
    <div class="border border-blue-200 rounded-lg bg-blue-50 p-4">
      <div class="flex items-start gap-3">
        <svg class="h-6 w-6 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <div class="flex-1">
          <h4 class="mb-2 text-blue-800 font-semibold">审计说明</h4>
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
