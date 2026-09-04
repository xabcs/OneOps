<script setup lang="ts">
  import { computed, nextTick, onActivated, onMounted, ref, watch } from 'vue';
  import { fetchCheckConnectPermission, fetchConnectServer, fetchGetSessions } from '@/service/api/cmdb';

  interface Props {
    visible: boolean;
    serverId: number;
    serverName: string;
    serverIp: string;
    serverEnv?: string;
    credentialId?: number;
  }

  const props = defineProps<Props>();
  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'connected', sessionId: number, websocketUrl: string, loginAccount: string): void;
  }>();

  const loading = ref(false);
  const connecting = ref(false);
  const selectedCredentialId = ref<number | undefined>(undefined);
  const availableCredentials = ref<CMDB.SSHCredential[]>([]);
  const hasPermission = ref(true);
  const permissionError = ref('');
  const connectReason = ref('');
  const recentSession = ref<CMDB.SSHSession | null>(null);

  const selectedProtocol = ref('ssh');

  // 当前选中的凭证对象
  const selectedCredential = computed(() => availableCredentials.value.find(c => c.id === selectedCredentialId.value));

  // 检查连接权限，获取可用凭证列表
  async function checkPermission() {
    loading.value = true;
    hasPermission.value = true;
    permissionError.value = '';
    availableCredentials.value = [];
    selectedCredentialId.value = undefined;

    try {
      const { data, error } = await fetchCheckConnectPermission(props.serverId);

      if (error) {
        hasPermission.value = false;
        permissionError.value = error.message || '检查权限失败';
        return;
      }

      if (data) {
        hasPermission.value = data.hasPermission;
        availableCredentials.value = data.credentials || [];
      } else {
        hasPermission.value = false;
        permissionError.value = '响应格式错误';
        return;
      }

      if (!hasPermission.value) {
        permissionError.value = '您没有连接此服务器的权限';
      } else if (availableCredentials.value.length === 0) {
        permissionError.value = '服务器未绑定任何凭证，请先在主机编辑页面绑定 SSH 凭证';
        hasPermission.value = false;
      } else {
        // 如果传入了凭证ID，优先选中该凭证，否则选中第一个凭证
        if (props.credentialId) {
          const targetCredential = availableCredentials.value.find(c => c.id === props.credentialId);
          if (targetCredential) {
            selectedCredentialId.value = targetCredential.id;
          } else {
            // 传入的凭证ID不在可用列表中，使用第一个凭证
            selectedCredentialId.value = availableCredentials.value[0].id;
          }
        } else {
          // 默认选中第一个凭证
          selectedCredentialId.value = availableCredentials.value[0].id;
        }
      }
    } catch (err: unknown) {
      hasPermission.value = false;
      const message = err instanceof Error ? err.message : '检查权限失败';
      permissionError.value = message;
      console.error('[ServerConnectDialog] checkPermission 异常:', err);
    } finally {
      loading.value = false;
    }
  }

  // 加载最近连接记录
  async function loadRecentSession() {
    try {
      const { data } = await fetchGetSessions({ serverId: props.serverId, pageSize: 1, page: 1 });
      recentSession.value = data?.list?.[0] || null;
    } catch {
      recentSession.value = null;
    }
  }

  // 连接服务器
  async function handleConnect() {
    if (!hasPermission.value || !selectedCredentialId.value) return;

    connecting.value = true;
    try {
      const { data, error } = await fetchConnectServer(props.serverId, {
        protocol: selectedProtocol.value as 'ssh' | 'sftp',
        credentialId: selectedCredentialId.value
      });

      if (error) {
        window.$message?.error(`连接失败: ${error.message || '未知错误'}`);
        return;
      }

      if (data?.sessionId) {
        const loginAccount = selectedCredential.value?.username || '';
        emit('connected', data.sessionId, data.websocketUrl, loginAccount);
        handleClose();
      } else {
        window.$message?.error('连接失败: 未获取到会话信息');
      }
    } catch (err: unknown) {
      window.$message?.error((err as Error).message || '连接失败');
    } finally {
      connecting.value = false;
    }
  }

  // 关闭弹窗
  function handleClose() {
    emit('update:visible', false);
  }

  // 组件挂载时检查权限
  onMounted(() => {
    if (props.visible) {
      checkPermission();
      loadRecentSession();
    }
  });

  // 组件从 keep-alive 激活时刷新数据
  onActivated(() => {
    if (props.visible) {
      // 从缓存恢复时刷新权限检查
      checkPermission();
    }
  });

  // 监听 visible 变化（优化：延迟加载非必要数据）
  watch(
    () => props.visible,
    visible => {
      if (visible) {
        connectReason.value = '';
        selectedProtocol.value = 'ssh';
        checkPermission();
        // 延迟加载最近会话，不阻塞主流程
        nextTick(() => loadRecentSession());
      }
    },
    { immediate: true }
  );
</script>

<template>
  <ElDialog
    :model-value="visible"
    :title="`连接主机：${serverName}`"
    width="760px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <div v-loading="loading">
      <!-- 无权限提示 -->
      <ElAlert v-if="!hasPermission" type="error" :closable="false" style="margin-bottom: 16px">
        {{ permissionError || '您没有连接此主机的权限' }}
      </ElAlert>

      <ElRow :gutter="24">
        <!-- 左栏：连接配置 -->
        <ElCol :span="12">
          <div :style="{ padding: '4px 0' }">
            <div
              :style="{
                fontSize: '14px',
                fontWeight: 600,
                color: '#303133',
                marginBottom: '16px',
                paddingBottom: '8px',
                borderBottom: '1px solid #ebeef5'
              }"
            >
              连接配置
            </div>

            <ElForm label-position="top" size="default">
              <ElFormItem label="连接协议">
                <ElRadioGroup v-model="selectedProtocol" :disabled="!hasPermission">
                  <ElRadio value="ssh">SSH 终端</ElRadio>
                  <ElRadio value="sftp" disabled>SFTP 传输</ElRadio>
                </ElRadioGroup>
              </ElFormItem>

              <ElFormItem label="登录凭证">
                <ElSelect
                  v-model="selectedCredentialId"
                  placeholder="选择 SSH 凭证"
                  style="width: 100%"
                  :disabled="!hasPermission"
                >
                  <ElOption
                    v-for="cred in availableCredentials"
                    :key="cred.id"
                    :value="cred.id"
                    :label="`${cred.name}（${cred.username}）`"
                  />
                </ElSelect>
              </ElFormItem>

              <ElFormItem label="连接原因">
                <ElInput
                  v-model="connectReason"
                  placeholder="请输入连接原因（生产环境必填）"
                  type="textarea"
                  :rows="2"
                  :disabled="!hasPermission"
                />
              </ElFormItem>
            </ElForm>
          </div>
        </ElCol>

        <!-- 右栏：安全上下文 -->
        <ElCol :span="12">
          <div :style="{ padding: '4px 0' }">
            <div
              :style="{
                fontSize: '14px',
                fontWeight: 600,
                color: '#303133',
                marginBottom: '16px',
                paddingBottom: '8px',
                borderBottom: '1px solid #ebeef5'
              }"
            >
              安全上下文
            </div>

            <div :style="{ display: 'flex', alignItems: 'flex-start', marginBottom: '12px', gap: '8px' }">
              <span :style="{ fontSize: '13px', color: '#606266', minWidth: '70px', flexShrink: 0, paddingTop: '2px' }">
                主机环境
              </span>
              <ElTag
                :type="props.serverEnv === 'prod' ? 'danger' : props.serverEnv === 'test' ? 'warning' : 'info'"
                size="small"
              >
                {{ props.serverEnv === 'prod' ? '生产环境' : props.serverEnv === 'test' ? '测试环境' : '开发环境' }}
              </ElTag>
            </div>

            <div :style="{ display: 'flex', alignItems: 'flex-start', marginBottom: '12px', gap: '8px' }">
              <span :style="{ fontSize: '13px', color: '#606266', minWidth: '70px', flexShrink: 0, paddingTop: '2px' }">
                凭证状态
              </span>
              <ElTag :type="availableCredentials.length > 0 ? 'success' : 'warning'" size="small">
                {{ availableCredentials.length > 0 ? `${availableCredentials.length} 个可用` : '未配置' }}
              </ElTag>
            </div>

            <div :style="{ display: 'flex', alignItems: 'flex-start', marginBottom: '12px', gap: '8px' }">
              <span :style="{ fontSize: '13px', color: '#606266', minWidth: '70px', flexShrink: 0, paddingTop: '2px' }">
                最近连接
              </span>
              <span :style="{ fontSize: '13px', color: '#303133' }">
                <template v-if="recentSession">
                  {{ recentSession.username }} · {{ recentSession.startedAt?.substring(0, 16) }}
                </template>
                <template v-else>暂无记录</template>
              </span>
            </div>

            <div :style="{ display: 'flex', alignItems: 'flex-start', marginBottom: '12px', gap: '8px' }">
              <span :style="{ fontSize: '13px', color: '#606266', minWidth: '70px', flexShrink: 0, paddingTop: '2px' }">
                登录账号
              </span>
              <span :style="{ fontSize: '13px', color: '#303133' }">
                {{ selectedCredential?.username || '—' }}
              </span>
            </div>

            <ElAlert
              v-if="props.serverEnv === 'prod'"
              type="warning"
              :closable="false"
              title="生产环境操作将被完整审计，请谨慎操作"
              style="margin-top: 12px"
            />
            <ElAlert
              v-else
              type="info"
              :closable="false"
              title="本次连接操作将被记录，可在会话审计中查看"
              style="margin-top: 12px"
            />
          </div>
        </ElCol>
      </ElRow>
    </div>

    <template #footer>
      <ElButton @click="handleClose">取消</ElButton>
      <ElButton
        v-if="!loading && hasPermission && availableCredentials.length === 0"
        type="warning"
        @click="handleClose"
      >
        去绑定凭证
      </ElButton>
      <ElButton
        v-else
        type="primary"
        :disabled="!hasPermission || !selectedCredentialId"
        :loading="connecting"
        @click="handleConnect"
      >
        连接
      </ElButton>
    </template>
  </ElDialog>
</template>
