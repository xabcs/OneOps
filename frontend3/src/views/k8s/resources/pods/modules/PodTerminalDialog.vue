<script setup lang="ts">
  import PodTerminal from '@/views/k8s/terminal/PodTerminal.vue';

  defineOptions({ name: 'PodTerminalDialog' });

  interface Props {
    clusterId: number | null;
    namespace: string;
    podName: string;
    containerName?: string;
  }

  defineProps<Props>();

  const visible = defineModel<boolean>('visible', { default: false });
</script>

<template>
  <ElDialog v-model="visible" title="Pod 终端" width="900px" fullscreen>
    <PodTerminal
      v-if="visible && clusterId"
      :cluster-id="clusterId"
      :namespace="namespace"
      :pod-name="podName"
      :container-name="containerName"
    />
  </ElDialog>
</template>
