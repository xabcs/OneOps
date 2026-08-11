<script setup lang="ts">
/**
 * 连接主机确认对话框
 * 从 index.vue 拆分：SSH 连接确认弹窗
 */

const props = defineProps<{
  visible: boolean;
  connectingServer: CMDB.Server | null;
}>();

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void;
  (e: 'confirm'): void;
  (e: 'cancel'): void;
}>();

function onCancel() {
  emit('update:visible', false);
  emit('cancel');
}
</script>

<template>
  <ElDialog :model-value="visible" title="连接主机" width="480px" :close-on-click-modal="true" @close="onCancel">
    <div v-if="connectingServer" class="connect-dialog-content">
      <ElDescriptions :column="1" border>
        <ElDescriptionsItem label="主机名">{{ connectingServer.hostname }}</ElDescriptionsItem>
        <ElDescriptionsItem label="IP地址">{{ connectingServer.ip }}</ElDescriptionsItem>
        <ElDescriptionsItem label="环境">{{ connectingServer.env || 'unknown' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="Agent状态">
          <ElTag v-if="connectingServer.agentStatus === 'running'" type="success">在线</ElTag>
          <ElTag v-else-if="connectingServer.agentStatus === 'offline'" type="warning">离线</ElTag>
          <ElTag v-else type="info">未安装</ElTag>
        </ElDescriptionsItem>
      </ElDescriptions>
    </div>
    <template #footer>
      <span class="dialog-footer">
        <ElButton @click="onCancel">取消</ElButton>
        <ElButton type="primary" @click="emit('confirm')">连接</ElButton>
      </span>
    </template>
  </ElDialog>
</template>

<style scoped>
.connect-dialog-content { padding: 16px 0; }
.dialog-footer { display: flex; justify-content: flex-end; gap: 12px; }
</style>
