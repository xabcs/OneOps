<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import {
  fetchCheckConnectPermission,
  fetchConnectServer,
  fetchGetServerById,
  fetchGetServerSessions
} from '@/service/api';

defineOptions({ name: 'CMDBServerDetail' });

const route = useRoute();
const router = useRouter();

const server = ref<CMDB.Server | null>(null);
const loading = ref(true);
const activeTab = ref('basic');
const sessions = ref<Bastion.BastionSession[]>([]);
const sessionsLoading = ref(false);
const hasPermission = ref(false);
const availableCredentials = ref<CMDB.SSHCredential[]>([]);

const serverStatusText = computed(() => {
  if (!server.value) return '-';
  const statusMap: Record<string, string> = {
    online: '在线',
    offline: '离线',
    unknown: '未知'
  };
  return statusMap[server.value.status] || '未知';
});

async function loadServerDetail() {
  const serverId = (route.query.id as string) || (route.params.id as string);
  if (!serverId) {
    router.push('/cmdb/servers');
    return;
  }

  try {
    const res = await fetchGetServerById(Number(serverId));
    server.value = res.data;
    await checkConnectPermission(Number(serverId));
  } catch (error) {
    console.error('获取服务器详情失败:', error);
    window.$message?.error('获取服务器详情失败');
  } finally {
    loading.value = false;
  }
}

async function checkConnectPermission(serverId: number) {
  try {
    const res = await fetchCheckConnectPermission(serverId);
    hasPermission.value = res.data.hasPermission;
    availableCredentials.value = res.data.credentials || [];
  } catch (error) {
    console.error('检查连接权限失败:', error);
  }
}

async function loadSessions() {
  if (!server.value) return;
  sessionsLoading.value = true;
  try {
    const res = await fetchGetServerSessions(server.value.id, { page: 1, pageSize: 10 });
    sessions.value = res.data.list || [];
  } catch (error) {
    console.error('获取会话列表失败:', error);
  } finally {
    sessionsLoading.value = false;
  }
}

function handleTabChange(tab: string) {
  activeTab.value = tab;
  if (tab === 'session' && server.value && sessions.value.length === 0) {
    loadSessions();
  }
}

async function handleConnect() {
  if (!server.value || availableCredentials.value.length === 0) {
    window.$message?.warning('没有可用的SSH凭证');
    return;
  }

  const credentialId = availableCredentials.value[0].id;
  try {
    const res = await fetchConnectServer(server.value.id, {
      protocol: 'ssh',
      credentialId
    });

    const { sessionId } = res.data;
    router.push({
      path: '/terminal/workbench',
      query: { sessionId: String(sessionId) }
    });
  } catch (error: any) {
    console.error('连接失败:', error);
    window.$message?.error(error.message || '连接失败');
  }
}

onMounted(() => {
  loadServerDetail();
});
</script>

<template>
  <div v-loading="loading" class="detail-page">
    <div class="instance-bar">
      <div class="instance-left">
        <span class="back-arrow" @click="router.push('/cmdb/servers')">←</span>
        <span class="instance-name">{{ server?.hostname || '-' }}</span>
        <span class="instance-meta">ID: {{ server?.id || '-' }}</span>
        <span class="instance-meta">{{ server?.ip || '-' }}</span>
        <ElTag v-if="server" :type="server.status === 'online' ? 'success' : 'info'" size="small">
          {{ serverStatusText }}
        </ElTag>
      </div>
      <div class="instance-actions">
        <ElButton type="primary" :disabled="!hasPermission || availableCredentials.length === 0" @click="handleConnect">
          远程连接
        </ElButton>
        <ElButton @click="router.push('/cmdb/servers')">返回列表</ElButton>
      </div>
    </div>

    <div class="tab-nav">
      <div class="tab-item" :class="{ active: activeTab === 'basic' }" @click="handleTabChange('basic')">基本信息</div>
      <div class="tab-item" :class="{ active: activeTab === 'config' }" @click="handleTabChange('config')">
        配置信息
      </div>
      <div class="tab-item" :class="{ active: activeTab === 'network' }" @click="handleTabChange('network')">
        网络信息
      </div>
      <div class="tab-item" :class="{ active: activeTab === 'security' }" @click="handleTabChange('security')">
        安全信息
      </div>
      <div class="tab-item" :class="{ active: activeTab === 'session' }" @click="handleTabChange('session')">
        会话记录
      </div>
    </div>

    <div v-show="activeTab === 'basic'">
      <div class="info-card">
        <div class="card-head"><span class="card-title">基本信息</span></div>
        <div class="card-body">
          <div class="desc-grid">
            <div class="desc-row">
              <div class="desc-item">
                <div class="item-label">实例 ID</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.id || '-' }}</span>
                </div>
              </div>
              <div class="desc-item">
                <div class="item-label">主机名</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.hostname || '-' }}</span>
                </div>
              </div>
              <div class="desc-item">
                <div class="item-label">实例状态</div>
                <div class="item-content">
                  <div v-if="server" class="status-indicator" :class="`status-${server.status}`">
                    <span class="status-dot"></span>
                    <span>{{ serverStatusText }}</span>
                  </div>
                  <span v-else>-</span>
                </div>
              </div>
            </div>
            <div class="desc-row">
              <div class="desc-item">
                <div class="item-label">IP 地址</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.ip || '-' }}</span>
                </div>
              </div>
              <div class="desc-item">
                <div class="item-label">内网 IP</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.innerIp || '-' }}</span>
                </div>
              </div>
              <div class="desc-item">
                <div class="item-label">环境</div>
                <div class="item-content">
                  <span v-if="server" class="env-tag">{{ server.env }}</span>
                  <span v-else>-</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="info-card">
        <div class="card-head"><span class="card-title">硬件配置</span></div>
        <div class="card-body">
          <div class="desc-grid">
            <div class="desc-row">
              <div class="desc-item">
                <div class="item-label">CPU</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.cpu ? `${server.cpu} 核` : '-' }}</span>
                </div>
              </div>
              <div class="desc-item">
                <div class="item-label">内存</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.memory ? `${server.memory} GB` : '-' }}</span>
                </div>
              </div>
              <div class="desc-item">
                <div class="item-label">磁盘</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.disk ? `${server.disk} GB` : '-' }}</span>
                </div>
              </div>
            </div>
            <div class="desc-row">
              <div class="desc-item desc-item-full">
                <div class="item-label">架构</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.arch || '-' }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="info-card">
        <div class="card-head"><span class="card-title">系统信息</span></div>
        <div class="card-body">
          <div class="desc-grid">
            <div class="desc-row">
              <div class="desc-item">
                <div class="item-label">操作系统</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.os || '-' }}</span>
                </div>
              </div>
              <div class="desc-item">
                <div class="item-label">系统版本</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.osVersion || '-' }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-show="activeTab === 'config'">
      <div class="info-card">
        <div class="card-head"><span class="card-title">SSH 配置</span></div>
        <div class="card-body">
          <div class="desc-grid">
            <div class="desc-row">
              <div class="desc-item">
                <div class="item-label">SSH 端口</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.sshPort || '-' }}</span>
                </div>
              </div>
              <div class="desc-item">
                <div class="item-label">SSH 用户</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.sshCredential?.username || '-' }}</span>
                </div>
              </div>
              <div class="desc-item">
                <div class="item-label">用户凭证</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.sshCredential?.name || '-' }}</span>
                </div>
              </div>
            </div>
            <div class="desc-row">
              <div class="desc-item desc-item-full">
                <div class="item-label">系统凭证</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.systemCredential?.name || '-' }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="info-card">
        <div class="card-head"><span class="card-title">资产分组</span></div>
        <div class="card-body">
          <div class="desc-grid">
            <div class="desc-row">
              <div class="desc-item">
                <div class="item-label">业务系统</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.business?.name || '-' }}</span>
                </div>
              </div>
              <div class="desc-item desc-item-full">
                <div class="item-label">所属分组</div>
                <div class="item-content">
                  <template v-if="server?.groups && server.groups.length > 0">
                    <ElTag v-for="group in server.groups" :key="group.id" size="small" style="margin-right: 8px">
                      {{ group.name }}
                    </ElTag>
                  </template>
                  <span v-else>-</span>
                </div>
              </div>
            </div>
            <div class="desc-row">
              <div class="desc-item desc-item-full">
                <div class="item-label">标签</div>
                <div class="item-content">
                  <template v-if="server?.tags && server.tags.length > 0">
                    <ElTag
                      v-for="tag in server.tags"
                      :key="tag.id"
                      :color="tag.color"
                      size="small"
                      style="margin-right: 8px"
                    >
                      {{ tag.name }}
                    </ElTag>
                  </template>
                  <span v-else>-</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-show="activeTab === 'network'">
      <div class="info-card">
        <div class="card-head"><span class="card-title">网络配置</span></div>
        <div class="card-body">
          <div class="desc-grid">
            <div class="desc-row">
              <div class="desc-item">
                <div class="item-label">外网 IP</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.ip || '-' }}</span>
                </div>
              </div>
              <div class="desc-item">
                <div class="item-label">内网 IP</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.innerIp || '-' }}</span>
                </div>
              </div>
              <div class="desc-item">
                <div class="item-label">SSH 端口</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.sshPort || '-' }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="server?.cloudInfo" class="info-card">
        <div class="card-head"><span class="card-title">云平台信息</span></div>
        <div class="card-body">
          <div class="desc-grid">
            <div class="desc-row">
              <div class="desc-item">
                <div class="item-label">云平台</div>
                <div class="item-content">
                  <span class="value-text">{{ server.cloudInfo.provider || '-' }}</span>
                </div>
              </div>
              <div class="desc-item">
                <div class="item-label">实例 ID</div>
                <div class="item-content">
                  <span class="value-text">{{ server.cloudInfo.instanceId || '-' }}</span>
                </div>
              </div>
              <div class="desc-item">
                <div class="item-label">实例名称</div>
                <div class="item-content">
                  <span class="value-text">{{ server.cloudInfo.instanceName || '-' }}</span>
                </div>
              </div>
            </div>
            <div class="desc-row">
              <div class="desc-item">
                <div class="item-label">实例类型</div>
                <div class="item-content">
                  <span class="value-text">{{ server.cloudInfo.instanceType || '-' }}</span>
                </div>
              </div>
              <div class="desc-item">
                <div class="item-label">地域</div>
                <div class="item-content">
                  <span class="value-text">{{ server.cloudInfo.region || '-' }}</span>
                </div>
              </div>
              <div class="desc-item">
                <div class="item-label">可用区</div>
                <div class="item-content">
                  <span class="value-text">{{ server.cloudInfo.zone || '-' }}</span>
                </div>
              </div>
            </div>
            <div class="desc-row">
              <div class="desc-item">
                <div class="item-label">VPC</div>
                <div class="item-content">
                  <span class="value-text">{{ server.cloudInfo.vpcId || '-' }}</span>
                </div>
              </div>
              <div class="desc-item desc-item-full">
                <div class="item-label">子网</div>
                <div class="item-content">
                  <span class="value-text">{{ server.cloudInfo.subnetId || '-' }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-show="activeTab === 'security'">
      <div class="info-card">
        <div class="card-head"><span class="card-title">资产信息</span></div>
        <div class="card-body">
          <div class="desc-grid">
            <div class="desc-row">
              <div class="desc-item">
                <div class="item-label">序列号</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.sn || '-' }}</span>
                </div>
              </div>
              <div class="desc-item">
                <div class="item-label">资产编号</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.assetNumber || '-' }}</span>
                </div>
              </div>
              <div class="desc-item">
                <div class="item-label">制造商</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.manufacturer || '-' }}</span>
                </div>
              </div>
            </div>
            <div class="desc-row">
              <div class="desc-item desc-item-full">
                <div class="item-label">型号</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.model || '-' }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="info-card">
        <div class="card-head"><span class="card-title">机柜信息</span></div>
        <div class="card-body">
          <div class="desc-grid">
            <div class="desc-row">
              <div class="desc-item">
                <div class="item-label">所属机柜</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.cabinet?.name || '-' }}</span>
                </div>
              </div>
              <div class="desc-item">
                <div class="item-label">U 位置</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.uPosition ? `U${server.uPosition}` : '-' }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="info-card">
        <div class="card-head"><span class="card-title">时间信息</span></div>
        <div class="card-body">
          <div class="desc-grid">
            <div class="desc-row">
              <div class="desc-item">
                <div class="item-label">购买日期</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.purchaseDate || '-' }}</span>
                </div>
              </div>
              <div class="desc-item">
                <div class="item-label">保修到期</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.expireWarranty || '-' }}</span>
                </div>
              </div>
              <div class="desc-item">
                <div class="item-label">创建时间</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.createdAt || '-' }}</span>
                </div>
              </div>
            </div>
            <div class="desc-row">
              <div class="desc-item desc-item-full">
                <div class="item-label">更新时间</div>
                <div class="item-content">
                  <span class="value-text">{{ server?.updatedAt || '-' }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-show="activeTab === 'session'">
      <div class="info-card">
        <div class="card-head">
          <span class="card-title">最近会话</span>
          <ElButton type="primary" link :loading="sessionsLoading" @click="loadSessions">刷新</ElButton>
        </div>
        <div class="card-body">
          <div v-if="sessionsLoading && sessions.length === 0" class="loading-state"><span>加载中...</span></div>
          <div v-else-if="sessions.length === 0" class="empty-state"><ElEmpty description="暂无会话记录" /></div>
          <ElTable v-else :data="sessions" style="width: 100%">
            <ElTableColumn prop="id" label="会话ID" width="80" />
            <ElTableColumn prop="user.username" label="用户" width="120" />
            <ElTableColumn prop="loginAccount" label="登录账号" width="120" />
            <ElTableColumn prop="clientIp" label="客户端IP" width="140" />
            <ElTableColumn prop="protocol" label="协议" width="80" />
            <ElTableColumn label="状态" width="100">
              <template #default="{ row }">
                <ElTag :type="row.status === 'active' ? 'success' : 'info'" size="small">
                  {{ row.status === 'active' ? '活跃' : '已关闭' }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn label="时长" width="100">
              <template #default="{ row }">{{ Math.floor(row.duration / 60) }}分钟</template>
            </ElTableColumn>
            <ElTableColumn prop="startedAt" label="开始时间" width="160" />
            <ElTableColumn label="操作" width="80">
              <template #default="{ row }">
                <ElButton type="primary" link size="small" @click="router.push(`/cmdb/sessions/${row.id}`)">
                  详情
                </ElButton>
              </template>
            </ElTableColumn>
          </ElTable>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.detail-page {
  min-height: 100vh;
  background: #f5f7fa;
  padding: 16px 24px 24px;
}

.instance-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.instance-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.back-arrow {
  font-size: 20px;
  color: #0052d9;
  cursor: pointer;
  transition: opacity 0.2s;
}

.back-arrow:hover {
  opacity: 0.7;
}

.instance-name {
  font-size: 16px;
  font-weight: 500;
  color: #1d1d1f;
}

.instance-meta {
  font-size: 12px;
  color: #858e99;
}

.tab-nav {
  display: flex;
  background: #fff;
  border: 1px solid #e8e8e8;
  border-radius: 2px;
  margin-bottom: 16px;
  border-bottom: 1px solid #e8e8e8;
}

.tab-item {
  padding: 6px 16px;
  font-size: 14px;
  color: #606266;
  border-top: 2px solid transparent;
  border-bottom: 1px solid #e8e8e8;
  cursor: pointer;
  transition:
    color 0.2s,
    border-color 0.2s;
}

.tab-item:hover {
  color: #0052d9;
}

.tab-item.active {
  color: #0052d9;
  border-top-color: #0052d9;
  border-bottom-color: transparent;
  margin-bottom: -1px;
}

.info-card {
  background: #fff;
  border: 1px solid #e5e5e5;
  border-radius: 2px;
  margin-bottom: 16px;
}

.info-card:last-child {
  margin-bottom: 0;
}

.card-head {
  padding: 16px 16px 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.card-body {
  padding: 16px;
}

.desc-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.desc-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 24px;
}

.desc-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.desc-item-full {
  grid-column: 1 / -1;
}

.item-label {
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
}

.item-content {
  font-size: 14px;
  color: #303133;
  line-height: 1.5;
}

.value-text {
  color: #303133;
}

.status-indicator {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-online .status-dot {
  background-color: #52c41a;
}

.status-offline .status-dot {
  background-color: #d9d9d9;
}

.status-unknown .status-dot {
  background-color: #faad14;
}

.env-tag {
  display: inline-block;
  padding: 2px 8px;
  background-color: #f0f0f0;
  border-radius: 2px;
  font-size: 12px;
  color: #606266;
}

.loading-state,
.empty-state {
  padding: 40px 0;
  text-align: center;
  color: #909399;
}
</style>
