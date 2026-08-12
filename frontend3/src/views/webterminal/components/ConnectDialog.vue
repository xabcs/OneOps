<script setup lang="ts">
  import { ElButton, ElDialog, ElOption, ElSelect, ElTag } from 'element-plus';
  import { Icon } from '@iconify/vue';

  interface Props {
    modelValue: boolean;
    connectingServer: CMDB.Server | null;
    selectedCredentialId: number | null;
  }

  interface Emits {
    (e: 'update:modelValue', value: boolean): void;
    (e: 'update:selectedCredentialId', value: number | null): void;
    (e: 'connect'): void;
    (e: 'cancel'): void;
  }

  const props = defineProps<Props>();
  const emit = defineEmits<Emits>();

  function handleClose() {
    emit('cancel');
  }

  function handleCredentialChange(val: number | null) {
    emit('update:selectedCredentialId', val);
  }
</script>

<template>
  <ElDialog
    :model-value="modelValue"
    title="连接主机"
    width="520px"
    :close-on-click-modal="true"
    :append-to-body="false"
    class="terminal-connect-dialog"
    @update:model-value="(val: boolean) => $emit('update:modelValue', val)"
    @close="handleClose"
  >
    <div v-if="connectingServer" class="connect-dialog-content">
      <!-- 主机信息 -->
      <div class="connect-section">
        <div class="connect-section-title">主机信息</div>
        <div class="connect-info-grid">
          <div class="connect-info-item">
            <span class="connect-info-label">主机名称</span>
            <span class="connect-info-value">{{ connectingServer.hostname || '未知' }}</span>
          </div>
          <div class="connect-info-item">
            <span class="connect-info-label">IP地址</span>
            <span class="connect-info-value">{{ connectingServer.ip || '未知' }}</span>
          </div>
          <div class="connect-info-item">
            <span class="connect-info-label">环境</span>
            <span class="connect-info-value">{{ connectingServer.env || 'unknown' }}</span>
          </div>
          <div class="connect-info-item">
            <span class="connect-info-label">Agent状态</span>
            <span class="connect-info-value">
              <ElTag v-if="connectingServer.agentStatus === 'running'" type="success" size="small">在线</ElTag>
              <ElTag v-else-if="connectingServer.agentStatus === 'offline'" type="warning" size="small">离线</ElTag>
              <ElTag v-else type="info" size="small">未安装</ElTag>
            </span>
          </div>
        </div>
      </div>

      <!-- 安全上下文 -->
      <div class="connect-section">
        <div class="connect-section-title">安全上下文</div>
        <div class="connect-info-grid">
          <div class="connect-info-item">
            <span class="connect-info-label">认证方式</span>
            <span class="connect-info-value">
              <Icon icon="lucide:shield-check" style="width: 14px; height: 14px; margin-right: 4px; color: #4ec9b0" />
              密钥认证
            </span>
          </div>
          <div class="connect-info-item">
            <span class="connect-info-label">安全策略</span>
            <span class="connect-info-value">
              <ElTag type="success" size="small" effect="plain">允许访问</ElTag>
            </span>
          </div>
          <div class="connect-info-item">
            <span class="connect-info-label">会话记录</span>
            <span class="connect-info-value">
              <Icon icon="lucide:check-circle-2" style="width: 14px; height: 14px; margin-right: 4px; color: #4ec9b0" />
              已启用
            </span>
          </div>
        </div>
      </div>

      <!-- 凭证选择 -->
      <div class="connect-section">
        <div class="connect-section-title">选择连接凭证</div>
        <div v-if="connectingServer.credentials && connectingServer.credentials.length > 0" class="credential-selector">
          <ElSelect
            :model-value="selectedCredentialId"
            placeholder="请选择凭证"
            popper-class="terminal-select-dropdown"
            style="width: 100%"
            @update:model-value="handleCredentialChange"
          >
            <ElOption
              v-for="cred in connectingServer.credentials"
              :key="cred.id"
              :label="`${cred.name} - ${cred.username}`"
              :value="cred.id"
            >
              <span style="display: flex; align-items: center; gap: 8px">
                <Icon icon="lucide:key" style="width: 14px; height: 14px; color: #858585" />
                <span>{{ cred.name }}</span>
                <ElTag size="small" effect="plain" style="margin-left: auto">{{ cred.username }}</ElTag>
              </span>
            </ElOption>
          </ElSelect>
          <div class="credential-hint">
            <Icon icon="lucide:info" style="width: 14px; height: 14px; margin-right: 4px; color: #858585" />
            <span>选择的凭证将决定登录账号</span>
          </div>
        </div>
        <div v-else class="credential-empty">
          <Icon icon="lucide:alert-circle" style="width: 16px; height: 16px; color: #f14c4c" />
          <span>该服务器没有可用的连接凭证</span>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="connect-actions">
        <ElButton @click="$emit('cancel')">取消</ElButton>
        <ElButton type="primary" :disabled="!selectedCredentialId" @click="$emit('connect')">
          <Icon icon="lucide:terminal" style="margin-right: 6px; width: 14px; height: 14px" />
          连接
        </ElButton>
      </div>
    </div>
  </ElDialog>
</template>

<style lang="scss">
  /* 连接对话框样式 - 终端暗色主题 */
  .terminal-workbench :deep(.el-overlay) {
    background-color: rgba(0, 0, 0, 0.7) !important;
  }

  .terminal-workbench :deep(.el-dialog.terminal-connect-dialog) {
    background: #252526 !important;
    border: 1px solid #454545 !important;
    box-shadow: 0 4px 24px rgba(0, 0, 0, 0.6) !important;
  }

  .terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-dialog__header) {
    background: #2d2d2d !important;
    border-bottom: 1px solid #000000 !important;
    padding: 12px 16px !important;
  }

  .terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-dialog__title) {
    color: #cccccc !important;
    font-size: 13px !important;
    font-weight: 500 !important;
  }

  .terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-dialog__headerbtn .el-dialog__close) {
    color: #858585 !important;
  }

  .terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-dialog__headerbtn .el-dialog__close:hover) {
    color: #cccccc !important;
  }

  .terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-dialog__body) {
    background: #252526 !important;
    padding: 16px !important;
    color: #cccccc !important;
  }

  .terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-dialog__footer) {
    background: #2d2d2d !important;
    border-top: 1px solid #000000 !important;
    padding: 12px 16px !important;
  }

  /* 对话框中的标签样式 */
  .terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-tag) {
    background: rgba(78, 201, 176, 0.1) !important;
    border-color: transparent !important;
    color: #4ec9b0 !important;
  }

  .terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-tag.el-tag--success) {
    background: rgba(78, 201, 176, 0.1) !important;
    color: #4ec9b0 !important;
  }

  .terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-tag.el-tag--warning) {
    background: rgba(217, 119, 6, 0.1) !important;
    color: #d97706 !important;
  }

  .terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-tag.el-tag--info) {
    background: rgba(107, 114, 128, 0.1) !important;
    color: #6b7280 !important;
  }

  .terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-tag.el-tag--plain) {
    background: rgba(78, 201, 176, 0.15) !important;
    border: 1px solid rgba(78, 201, 176, 0.3) !important;
  }

  /* 对话框中的按钮样式 */
  .terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-button) {
    background: transparent !important;
    color: #cccccc !important;
    border: 1px solid #454545 !important;
  }

  .terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-button:hover) {
    background: rgba(0, 0, 0, 0.2) !important;
    border-color: #555 !important;
  }

  .terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-button--primary) {
    background: #007acc !important;
    border-color: #007acc !important;
    color: #ffffff !important;
  }

  .terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-button--primary:hover) {
    background: #0069b4 !important;
    border-color: #0069b4 !important;
  }

  .connect-dialog-content {
    padding: 0;
  }

  .connect-section {
    margin-bottom: 20px;
  }

  .connect-section:last-child {
    margin-bottom: 0;
  }

  .connect-section-title {
    font-size: 12px;
    font-weight: 500;
    color: #cccccc;
    margin-bottom: 12px;
    padding-left: 12px;
    position: relative;
  }

  .connect-section-title::before {
    content: '';
    position: absolute;
    left: 0;
    top: 50%;
    transform: translateY(-50%);
    width: 3px;
    height: 14px;
    background: #007acc;
    border-radius: 2px;
  }

  .connect-info-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 12px 16px;
    padding: 0 4px;
  }

  .connect-info-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .connect-info-label {
    font-size: 11px;
    color: #858585;
    line-height: 1.5;
  }

  .connect-info-value {
    font-size: 12px;
    color: #cccccc;
    line-height: 1.5;
    display: flex;
    align-items: center;
  }

  .connect-actions {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    margin-top: 24px;
    padding-top: 16px;
    border-top: 1px solid #333;
  }

  .credential-selector {
    padding: 0 4px;
  }

  .credential-hint {
    display: flex;
    align-items: center;
    margin-top: 8px;
    padding: 8px 12px;
    font-size: 11px;
    color: #858585;
    background: rgba(136, 85, 85, 0.1);
    border-radius: 2px;
  }

  .credential-empty {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px;
    color: #f14c4c;
    font-size: 12px;
    background: rgba(241, 76, 76, 0.1);
    border: 1px dashed #f14c4c;
    border-radius: 2px;
  }
</style>
