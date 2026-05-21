<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { fetchConnectServer, fetchCheckConnectPermission } from '@/service/api/cmdb';
import { $t } from '@/locales';
import { useAuthStore } from '@/store/modules/auth';

interface Props {
  visible: boolean;
  serverId: number;
  serverName: string;
  serverIp: string;
  sshCredentialId?: number;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void;
  (e: 'connected', sessionId: number, websocketUrl: string): void;
}>();

const authStore = useAuthStore();

const loading = ref(false);
const connecting = ref(false);
const selectedAccount = ref('root');
const allowedAccounts = ref<string[]>([]);
const hasPermission = ref(true);
const permissionError = ref('');

const protocolOptions = [
  { label: 'SSH 终端', value: 'ssh' },
  { label: 'SFTP', value: 'sftp' }
];

const selectedProtocol = ref('ssh');

// 检查连接权限
async function checkPermission() {
  loading.value = true;
  try {
    const result = await fetchCheckConnectPermission(props.serverId);
    hasPermission.value = result.hasPermission;
    allowedAccounts.value = result.allowedAccounts || [];

    if (!hasPermission.value) {
      permissionError.value = '您没有连接此服务器的权限';
    } else if (allowedAccounts.value.length === 0) {
      // 使用默认账号
      allowedAccounts.value = ['root'];
    }

    // 设置默认选中的账号
    if (allowedAccounts.value.length > 0) {
      selectedAccount.value = allowedAccounts.value[0];
    }
  } catch (error: any) {
    hasPermission.value = false;
    permissionError.value = error.message || '检查权限失败';
  } finally {
    loading.value = false;
  }
}

// 连接服务器
async function handleConnect() {
  if (!hasPermission.value) {
    return;
  }

  connecting.value = true;
  try {
    const result = await fetchConnectServer(props.serverId, {
      protocol: selectedProtocol.value as 'ssh' | 'sftp',
      loginAccount: selectedAccount.value
    });

    emit('connected', result.sessionId, result.websocketUrl);
    handleClose();
  } catch (error: any) {
    window.$message?.error(error.message || '连接失败');
  } finally {
    connecting.value = false;
  }
}

// 关闭弹窗
function handleClose() {
  emit('update:visible', false);
}

// 监听 visible 变化
watch(() => props.visible, (visible) => {
  if (visible) {
    checkPermission();
  }
});
</script>

<template>
  <el-dialog
    :model-value="visible"
    :title="`连接服务器: ${serverName} (${serverIp})`"
    width="500px"
    :close-on-click-modal="false"
    @update:model-value="handleClose"
  >
    <el-skeleton v-if="loading" :rows="3" animated />

    <div v-else class="connect-dialog-content">
      <!-- 无权限提示 -->
      <el-alert
        v-if="!hasPermission"
        type="error"
        :closable="false"
        show-icon
        :title="permissionError"
        style="margin-bottom: 20px"
      />

      <!-- 连接配置 -->
      <el-form v-else label-width="100px" @submit.prevent="handleConnect">
        <el-form-item label="连接方式">
          <el-radio-group v-model="selectedProtocol">
            <el-radio-button
              v-for="option in protocolOptions"
              :key="option.value"
              :value="option.value"
              :label="option.label"
            />
          </el-radio-group>
        </el-form-item>

        <el-form-item label="登录账号">
          <el-select
            v-model="selectedAccount"
            placeholder="选择登录账号"
            style="width: 100%"
          >
            <el-option
              v-for="account in allowedAccounts"
              :key="account"
              :value="account"
              :label="account"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="服务器信息">
          <div class="server-info">
            <p><strong>服务器:</strong> {{ serverName }}</p>
            <p><strong>IP 地址:</strong> {{ serverIp }}</p>
            <p v-if="sshCredentialId">
              <strong>凭证:</strong> <span class="text-success">已配置</span>
            </p>
            <p v-else>
              <strong>凭证:</strong> <span class="text-warning">未配置</span>
            </p>
          </div>
        </el-form-item>

        <!-- 提示信息 -->
        <el-alert
          type="info"
          :closable="false"
          show-icon
          title="连接说明"
        >
          <ul class="connect-tips">
            <li>所有连接操作都会被记录和审计</li>
            <li>请勿执行危险操作，命令会被实时监控</li>
            <li>连接超时时间: 30 分钟无操作自动断开</li>
          </ul>
        </el-alert>
      </el-form>
    </div>

    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button
        type="primary"
        :loading="connecting"
        :disabled="!hasPermission"
        @click="handleConnect"
      >
        {{ connecting ? '连接中...' : '连接' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.connect-dialog-content {
  padding: 10px 0;
}

.server-info {
  background: #f5f7fa;
  padding: 12px;
  border-radius: 4px;
  font-size: 14px;
}

.server-info p {
  margin: 4px 0;
}

.text-success {
  color: #67c23a;
}

.text-warning {
  color: #e6a23c;
}

.connect-tips {
  margin: 8px 0 0 0;
  padding-left: 20px;
  font-size: 13px;
}

.connect-tips li {
  margin: 4px 0;
  color: #606266;
}
</style>
