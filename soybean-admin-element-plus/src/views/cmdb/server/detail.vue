<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage, ElNotification } from 'element-plus';
import { ArrowLeft, Refresh } from '@element-plus/icons-vue';
import {
  fetchGetServerById,
  fetchConnectServer,
  fetchCheckConnectPermission,
  fetchGetServerAttributes
} from '@/service/api';
import { $t } from '@/locales';

defineOptions({ name: 'CmdbServerDetail' });

const route = useRoute();
const router = useRouter();

const serverId = computed(() => parseInt(route.query.id as string));
const loading = ref(false);
const server = ref<CMDB.Server | null>(null);
const attributes = ref<Record<string, string>>({});

// 环境显示信息
const envDisplayMap: Record<string, { text: string; type: '' | 'success' | 'warning' | 'info' | 'danger' }> = {
  prod: { text: '生产', type: 'danger' },
  test: { text: '测试', type: 'warning' },
  dev: { text: '开发', type: 'info' }
};

function getEnvDisplayInfo(env: string) {
  return envDisplayMap[env] || { text: env, type: 'info' };
}

// 状态显示信息
const statusDisplayMap: Record<string, { text: string; type: '' | 'success' | 'warning' | 'info' | 'danger' }> = {
  online: { text: '在线', type: 'success' },
  offline: { text: '离线', type: 'danger' },
  unknown: { text: '未知', type: 'info' }
};

function getStatusDisplayInfo(status: string) {
  return statusDisplayMap[status] || { text: status, type: 'info' };
}

// 加载服务器详情
async function loadServerDetail() {
  if (!serverId.value) {
    ElMessage.error('缺少服务器ID');
    return;
  }

  loading.value = true;
  try {
    const { data, error } = await fetchGetServerById(serverId.value);
    if (error || !data) {
      ElNotification({
        title: '错误',
        message: '加载服务器详情失败',
        type: 'error',
        duration: 3000
      });
      return;
    }
    server.value = data;
  } catch (err) {
    console.error('加载服务器详情失败:', err);
    ElNotification({
      title: '错误',
      message: '加载服务器详情失败',
      type: 'error',
      duration: 3000
    });
  } finally {
    loading.value = false;
  }
}

// 加载服务器属性
async function loadServerAttributes() {
  if (!serverId.value) return;

  try {
    const { data } = await fetchGetServerAttributes(serverId.value);
    if (data && typeof data === 'object') {
      attributes.value = data;
    }
  } catch (err) {
    console.error('加载服务器属性失败:', err);
  }
}

// 连接服务器
async function handleConnect() {
  if (!server.value) return;

  // 检查连接权限
  const { data: permissionData } = await fetchCheckConnectPermission(server.value.id);
  if (!permissionData?.allowed) {
    ElNotification({
      title: '权限不足',
      message: permissionData?.reason || '您没有连接此服务器的权限',
      type: 'warning',
      duration: 3000
    });
    return;
  }

  // 检查SSH凭证
  if (!server.value.sshCredentialId) {
    ElNotification({
      title: '凭证缺失',
      message: '服务器未配置SSH凭证，无法连接',
      type: 'warning',
      duration: 3000
    });
    return;
  }

  try {
    const { data, error } = await fetchConnectServer(server.value.id);
    if (error || !data?.sessionId) {
      ElNotification({
        title: '连接失败',
        message: '无法建立SSH连接',
        type: 'error',
        duration: 3000
      });
      return;
    }

    // 跳转到Web终端页面
    router.push({
      path: '/webterminal',
      query: { sessionId: data.sessionId }
    });
  } catch (err) {
    console.error('连接服务器失败:', err);
    ElNotification({
      title: '连接失败',
      message: '连接服务器时发生错误',
      type: 'error',
      duration: 3000
    });
  }
}

// 返回列表
function handleBack() {
  router.back();
}

// 刷新数据
function handleRefresh() {
  loadServerDetail();
  loadServerAttributes();
}

onMounted(() => {
  loadServerDetail();
  loadServerAttributes();
});
</script>

<template>
  <div class="server-detail p-4">
    <!-- 头部操作栏 -->
    <div class="header-actions mb-4 flex items-center justify-between">
      <div class="flex items-center gap-2">
        <ElButton :icon="ArrowLeft" @click="handleBack">返回</ElButton>
        <h2 class="text-xl font-bold">{{ server?.hostname || '服务器详情' }}</h2>
      </div>
      <ElButton :icon="Refresh" :loading="loading" @click="handleRefresh">刷新</ElButton>
    </div>

    <!-- 加载状态 -->
    <ElSkeleton v-if="loading" :rows="10" animated />

    <!-- 详情内容 -->
    <div v-else-if="server" class="detail-content">
      <ElRow :gutter="20">
        <!-- 基本信息卡片 -->
        <ElCol :span="12">
          <ElCard header="基本信息" shadow="hover">
            <div class="info-grid">
              <div class="info-item">
                <span class="label">主机名:</span>
                <span class="value">{{ server.hostname }}</span>
              </div>
              <div class="info-item">
                <span class="label">IP地址:</span>
                <span class="value">{{ server.ip }}</span>
              </div>
              <div class="info-item" v-if="server.innerIp">
                <span class="label">内网IP:</span>
                <span class="value">{{ server.innerIp }}</span>
              </div>
              <div class="info-item">
                <span class="label">SSH端口:</span>
                <span class="value">{{ server.sshPort }}</span>
              </div>
              <div class="info-item">
                <span class="label">SSH用户:</span>
                <span class="value">{{ server.sshUser }}</span>
              </div>
              <div class="info-item">
                <span class="label">环境:</span>
                <ElTag :type="getEnvDisplayInfo(server.env).type" size="small">
                  {{ getEnvDisplayInfo(server.env).text }}
                </ElTag>
              </div>
              <div class="info-item">
                <span class="label">状态:</span>
                <ElTag :type="getStatusDisplayInfo(server.status).type" size="small">
                  {{ getStatusDisplayInfo(server.status).text }}
                </ElTag>
              </div>
              <div class="info-item">
                <span class="label">提供商:</span>
                <span class="value">{{ server.provider }}</span>
              </div>
              <div class="info-item">
                <span class="label">服务器类型:</span>
                <span class="value">{{ server.serverType }}</span>
              </div>
            </div>
          </ElCard>
        </ElCol>

        <!-- 硬件信息卡片 -->
        <ElCol :span="12">
          <ElCard header="硬件配置" shadow="hover">
            <div class="info-grid">
              <div class="info-item">
                <span class="label">CPU:</span>
                <span class="value">{{ server.cpu }} 核</span>
              </div>
              <div class="info-item">
                <span class="label">内存:</span>
                <span class="value">{{ server.memory }} GB</span>
              </div>
              <div class="info-item">
                <span class="label">磁盘:</span>
                <span class="value">{{ server.disk }} GB</span>
              </div>
              <div class="info-item" v-if="server.os">
                <span class="label">操作系统:</span>
                <span class="value">{{ server.os }}</span>
              </div>
              <div class="info-item" v-if="server.osVersion">
                <span class="label">系统版本:</span>
                <span class="value">{{ server.osVersion }}</span>
              </div>
              <div class="info-item" v-if="server.arch">
                <span class="label">架构:</span>
                <span class="value">{{ server.arch }}</span>
              </div>
              <div class="info-item" v-if="server.manufacturer">
                <span class="label">制造商:</span>
                <span class="value">{{ server.manufacturer }}</span>
              </div>
              <div class="info-item" v-if="server.model">
                <span class="label">型号:</span>
                <span class="value">{{ server.model }}</span>
              </div>
            </div>
          </ElCard>
        </ElCol>

        <!-- 资产信息卡片 -->
        <ElCol :span="12" v-if="server.sn || server.assetNumber || server.purchaseDate">
          <ElCard header="资产信息" shadow="hover">
            <div class="info-grid">
              <div class="info-item" v-if="server.sn">
                <span class="label">序列号:</span>
                <span class="value">{{ server.sn }}</span>
              </div>
              <div class="info-item" v-if="server.assetNumber">
                <span class="label">资产编号:</span>
                <span class="value">{{ server.assetNumber }}</span>
              </div>
              <div class="info-item" v-if="server.purchaseDate">
                <span class="label">采购日期:</span>
                <span class="value">{{ server.purchaseDate }}</span>
              </div>
              <div class="info-item" v-if="server.expireWarranty">
                <span class="label">保修到期:</span>
                <span class="value">{{ server.expireWarranty }}</span>
              </div>
            </div>
          </ElCard>
        </ElCol>

        <!-- 运行状态卡片 -->
        <ElCol :span="12">
          <ElCard header="运行状态" shadow="hover">
            <div class="info-grid">
              <div class="info-item" v-if="server.cpuUsage !== undefined">
                <span class="label">CPU使用率:</span>
                <span class="value">{{ server.cpuUsage?.toFixed(1) }}%</span>
              </div>
              <div class="info-item" v-if="server.memoryUsage !== undefined">
                <span class="label">内存使用率:</span>
                <span class="value">{{ server.memoryUsage?.toFixed(1) }}%</span>
              </div>
              <div class="info-item" v-if="server.diskUsage !== undefined">
                <span class="label">磁盘使用率:</span>
                <span class="value">{{ server.diskUsage?.toFixed(1) }}%</span>
              </div>
              <div class="info-item" v-if="server.lastCheckTime">
                <span class="label">最后检查:</span>
                <span class="value">{{ server.lastCheckTime }}</span>
              </div>
              <div class="info-item" v-if="server.lastConnectTime">
                <span class="label">最后连接:</span>
                <span class="value">{{ server.lastConnectTime }}</span>
              </div>
            </div>
          </ElCard>
        </ElCol>

        <!-- 机房信息卡片 -->
        <ElCol :span="12" v-if="server.cabinetId || server.uPosition">
          <ElCard header="机房信息" shadow="hover">
            <div class="info-grid">
              <div class="info-item" v-if="server.cabinetId">
                <span class="label">机柜ID:</span>
                <span class="value">{{ server.cabinetId }}</span>
              </div>
              <div class="info-item" v-if="server.uPosition">
                <span class="label">U位:</span>
                <span class="value">U{{ server.uPosition }}</span>
              </div>
            </div>
          </ElCard>
        </ElCol>

        <!-- 业务信息卡片 -->
        <ElCol :span="12" v-if="server.business">
          <ElCard header="业务信息" shadow="hover">
            <div class="info-grid">
              <div class="info-item">
                <span class="label">业务单元:</span>
                <span class="value">{{ server.business.name }}</span>
              </div>
              <div class="info-item" v-if="server.business.code">
                <span class="label">业务代码:</span>
                <span class="value">{{ server.business.code }}</span>
              </div>
            </div>
          </ElCard>
        </ElCol>

        <!-- 自定义属性卡片 -->
        <ElCol :span="12" v-if="Object.keys(attributes).length > 0">
          <ElCard header="自定义属性" shadow="hover">
            <div class="info-grid">
              <div class="info-item" v-for="(value, key) in attributes" :key="key">
                <span class="label">{{ key }}:</span>
                <span class="value">{{ value }}</span>
              </div>
            </div>
          </ElCard>
        </ElCol>

        <!-- 备注信息卡片 -->
        <ElCol :span="24" v-if="server.remarks">
          <ElCard header="备注" shadow="hover">
            <p class="whitespace-pre-wrap">{{ server.remarks }}</p>
          </ElCard>
        </ElCol>
      </ElRow>

      <!-- 操作按钮区 -->
      <div class="mt-4 flex gap-2">
        <ElButton type="primary" @click="handleConnect">连接服务器</ElButton>
        <ElButton @click="router.push({ path: '/cmdb/servers' })">返回列表</ElButton>
      </div>
    </div>

    <!-- 无数据提示 -->
    <ElEmpty v-else description="未找到服务器信息" />
  </div>
</template>

<style scoped>
.server-detail {
  max-width: 1400px;
  margin: 0 auto;
}

.header-actions {
  padding: 16px 0;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-item .label {
  font-weight: 500;
  color: var(--el-text-color-secondary);
  min-width: 80px;
}

.info-item .value {
  color: var(--el-text-color-primary);
}

.whitespace-pre-wrap {
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
