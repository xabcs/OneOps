<script setup lang="ts">
import { ElButton, ElDialog } from 'element-plus';

const showDialog = defineModel<boolean>('show', { required: true });

defineProps<{
  title: string;
  content: string;
}>();
</script>

<template>
  <ElDialog v-model="showDialog" :title="title" width="900px" top="5vh">
    <div class="log-content">{{ content }}</div>
    <template #footer>
      <ElButton @click="showDialog = false">关闭</ElButton>
    </template>
  </ElDialog>
</template>

<style scoped>
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
