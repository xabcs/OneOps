<script setup lang="ts">
  import { ref } from 'vue';
  import { ElButton, ElDialog, ElMessage, ElTable, ElTableColumn, ElTag } from 'element-plus';
  import { fetchK8sPodLogs } from '@/service/api/k8s';
  import { formatImages } from '@/views/k8s/shared/k8s-formatters';

  interface Pod {
    name: string;
    phase?: string;
    containers?: Array<{ name: string; image: string }>;
    ip?: string;
    node?: string;
    restarts?: number;
    age?: string;
  }

  interface Props {
    pods: Pod[];
    loading?: boolean;
    clusterId?: number;
    namespace?: string;
  }

  const props = withDefaults(defineProps<Props>(), {
    loading: false
  });

  const emit = defineEmits<{
    terminal: [pod: Pod];
    logs: [pod: Pod];
  }>();

  const getPodStatusTag = (pod: Pod) => {
    const phase = pod.phase || 'Unknown';
    switch (phase) {
      case 'Running':
        return { type: 'success', text: '运行中' };
      case 'Succeeded':
        return { type: 'info', text: '已完成' };
      case 'Failed':
        return { type: 'danger', text: '失败' };
      case 'Pending':
        return { type: 'warning', text: '等待中' };
      default:
        return { type: 'info', text: '未知' };
    }
  };

  // 在新标签页打开 Pod 终端
  const handleTerminal = (row: Pod) => {
    if (!props.clusterId || !props.namespace) {
      ElMessage.error('缺少必要参数：clusterId 或 namespace');
      return;
    }

    const containerName = row.containers?.[0]?.name || '';
    const baseUrl = window.location.origin;
    const terminalUrl = `${baseUrl}/k8s/terminal?clusterId=${props.clusterId}&namespace=${props.namespace}&podName=${row.name}&containerName=${containerName}`;

    const newWindow = window.open(terminalUrl, '_blank');

    if (!newWindow) {
      ElMessage.warning('浏览器阻止了新标签页打开，请检查浏览器设置允许弹窗');
    } else {
      newWindow.focus();
    }
  };

  // 在对话框中显示 Pod 日志
  const showLogs = ref(false);
  const logContent = ref('');
  const logPodName = ref('');
  const logContainerName = ref('');

  const handleLogs = async (row: Pod) => {
    if (!props.clusterId || !props.namespace) {
      ElMessage.error('缺少必要参数：clusterId 或 namespace');
      return;
    }

    const containerName = row.containers?.[0]?.name || '';
    logPodName.value = row.name;
    logContainerName.value = containerName;
    logContent.value = '加载中...';
    showLogs.value = true;

    try {
      const res = await fetchK8sPodLogs(props.clusterId, props.namespace, row.name, {
        container: containerName,
        tailLines: 100
      });
      logContent.value = res.data?.logs || '暂无日志';
    } catch (error: unknown) {
      logContent.value = `日志加载失败: ${(error as Error).message || '未知错误'}`;
      ElMessage.error(`日志加载失败: ${(error as Error).message}`);
    }
  };

  const closeLogs = () => {
    showLogs.value = false;
    logContent.value = '';
    logPodName.value = '';
    logContainerName.value = '';
  };
</script>

<template>
  <ElTable
    :data="pods"
    :loading="loading"
    stripe
    size="small"
    class="k8s-pods-table"
    :header-cell-style="{
      background: '#f5f7fa',
      color: '#303133',
      fontWeight: '600',
      paddingLeft: '16px',
      paddingRight: '16px'
    }"
    :row-style="{ backgroundColor: 'transparent' }"
    :cell-style="{ backgroundColor: 'transparent', padding: '8px 16px' }"
  >
    <ElTableColumn prop="name" label="Pod 名称" min-width="200" show-overflow-tooltip align="left" />
    <ElTableColumn label="状态" width="120" align="left">
      <template #default="{ row }">
        <ElTag :type="getPodStatusTag(row).type" size="small">
          {{ getPodStatusTag(row).text }}
        </ElTag>
      </template>
    </ElTableColumn>
    <ElTableColumn label="镜像" min-width="200" show-overflow-tooltip align="left">
      <template #default="{ row }">
        <span class="whitespace-pre-line">{{ formatImages(row) }}</span>
      </template>
    </ElTableColumn>
    <ElTableColumn prop="ip" label="IP 地址" width="140" align="left" />
    <ElTableColumn prop="node" label="节点" width="150" show-overflow-tooltip align="left" />
    <ElTableColumn label="重启次数" width="100" align="center">
      <template #default="{ row }">
        {{ row.restarts || 0 }}
      </template>
    </ElTableColumn>
    <ElTableColumn label="创建时间" width="140" align="left">
      <template #default="{ row }">
        {{ row.age || '-' }}
      </template>
    </ElTableColumn>
    <ElTableColumn label="操作" width="150" fixed="right" align="center" class-name="msre-table-actions">
      <template #default="{ row }">
        <ElButton link type="primary" size="small" @click="handleTerminal(row)">终端</ElButton>
        <ElButton link type="primary" size="small" @click="handleLogs(row)">日志</ElButton>
      </template>
    </ElTableColumn>
  </ElTable>

  <!-- 日志对话框 -->
  <ElDialog
    v-model="showLogs"
    :title="`日志: ${logPodName}${logContainerName ? ' (' + logContainerName + ')' : ''}`"
    width="900px"
    top="5vh"
  >
    <div class="log-content">{{ logContent }}</div>
    <template #footer>
      <ElButton @click="closeLogs">关闭</ElButton>
    </template>
  </ElDialog>
</template>

<style scoped>
  .whitespace-pre-line {
    white-space: pre-line;
    word-break: break-all;
  }

  .log-content {
    background: #1e1e1e;
    color: #4ec9b0;
    padding: 16px;
    border-radius: 4px;
    font-family: 'Courier New', Courier, monospace;
    font-size: 13px;
    white-space: pre-wrap;
    word-break: break-all;
    max-height: 600px;
    overflow: auto;
  }
</style>
