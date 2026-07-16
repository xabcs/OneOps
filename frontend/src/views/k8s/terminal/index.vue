<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import PodTerminal from './PodTerminal.vue';

defineOptions({ name: 'K8sTerminalPage' });

const route = useRoute();

// 从URL查询参数获取终端配置
const clusterId = computed(() => {
  const id = route.query.clusterId;
  return id ? Number(id) : 0;
});

const namespace = computed(() => {
  return (route.query.namespace as string) || '';
});

const podName = computed(() => {
  return (route.query.podName as string) || '';
});

const containerName = computed(() => {
  return (route.query.containerName as string) || undefined;
});
</script>

<template>
  <div class="h-screen w-screen overflow-hidden">
    <PodTerminal
      v-if="clusterId && namespace && podName"
      :cluster-id="clusterId"
      :namespace="namespace"
      :pod-name="podName"
      :container-name="containerName"
    />
    <div v-else class="flex h-full items-center justify-center bg-gray-900 text-white">
      <div class="text-center">
        <div class="mb-4 text-6xl">⚠️</div>
        <div class="text-xl">缺少必要参数</div>
        <div class="mt-2 text-sm text-gray-400">
          请从 Pod 详情页点击"终端"按钮打开此页面
        </div>
      </div>
    </div>
  </div>
</template>
