<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { fetchK8sPodLogs } from '@/service/api/k8s';

  defineOptions({ name: 'PodLogDialog' });

  interface Props {
    clusterId: number | null;
    namespace: string;
    podName: string;
    containerName?: string;
  }

  const props = defineProps<Props>();

  const visible = defineModel<boolean>('visible', { default: false });

  const logContent = ref('');
  const logTailLines = ref(100);

  async function loadLogs() {
    if (!props.clusterId || !props.podName) return;
    logContent.value = '加载中...';
    try {
      const { data } = await fetchK8sPodLogs(props.clusterId, props.namespace, props.podName, {
        container: props.containerName || '',
        tailLines: logTailLines.value
      });
      logContent.value = data?.logs || '暂无日志';
    } catch (error: unknown) {
      const err = error as Error;
      logContent.value = `日志加载失败: ${err.message || '未知错误'}`;
    }
  }

  watch(visible, val => {
    if (val) {
      logContent.value = '';
      loadLogs();
    }
  });
</script>

<template>
  <ElDialog v-model="visible" :title="`日志: ${podName}`" width="900px" top="5vh">
    <div class="mb-4 flex items-center gap-4">
      <span class="text-sm text-gray-600">行数:</span>
      <ElInputNumber v-model="logTailLines" :min="10" :max="10000" :step="100" />
      <ElButton size="small" @click="loadLogs">刷新</ElButton>
    </div>
    <div
      class="max-h-[600px] overflow-auto whitespace-pre-wrap rounded bg-black p-4 text-sm text-green-400 font-mono"
    >
      {{ logContent }}
    </div>
  </ElDialog>
</template>
