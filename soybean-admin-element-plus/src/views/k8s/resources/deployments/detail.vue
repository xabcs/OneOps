<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import {
  ElButton,
  ElDialog,
  ElForm,
  ElFormItem,
  ElInput,
  ElInputNumber,
  ElMessage,
  ElMessageBox,
  ElTabPane,
  ElTable,
  ElTableColumn,
  ElTabs,
  ElTag
} from 'element-plus';
import * as yaml from 'js-yaml';
import {
  deleteK8sDeployment,
  fetchK8sEvents,
  fetchK8sPodLogs,
  getK8sDeployment,
  getK8sDeploymentPods,
  restartK8sDeployment,
  scaleK8sDeployment
} from '@/service/api/k8s';
import { formatConditions, formatLabels } from '@/utils/k8s-formatters';
import PodTerminal from '../../terminal/PodTerminal.vue';

defineOptions({ name: 'K8sDeploymentDetail' });

const route = useRoute();
const router = useRouter();

const deployment = ref<any>(null);
const pods = ref<any[]>([]);
const events = ref<any[]>([]);
const loading = ref(true);
const podsLoading = ref(false);
const eventsLoading = ref(false);
const activeTab = ref('pods');

const clusterId = ref<number>(0);
const namespace = ref('');
const deploymentName = ref('');

// 终端相关
const showTerminal = ref(false);
const terminalProps = ref({
  clusterId: 0,
  namespace: '',
  podName: '',
  containerName: ''
});

// 日志相关
const showLogs = ref(false);
const logContent = ref('');
const logPodName = ref('');

// 缩放相关
const showScaleModal = ref(false);
const scaleData = ref({ replicas: 1 });
const scaleFormRef = ref<any>(null);

// YAML 编辑相关
const showYamlEditor = ref(false);
const yamlContent = ref('');
const yamlEditingContent = ref('');
const yamlSaving = ref(false);

// 状态显示
const getStatusTag = computed(() => {
  if (!deployment.value) return { type: 'info', text: '未知' };

  const ready = deployment.value.ready || 0;
  const total = deployment.value.replicas || 0;

  if (total === 0) {
    return { type: 'info', text: '未就绪' };
  }

  if (ready === total) {
    return { type: 'success', text: `运行中 (${ready}/${total})` };
  } else if (ready > 0) {
    return { type: 'warning', text: `部分就绪 (${ready}/${total})` };
  }
  return { type: 'danger', text: '未就绪' };
});

// Pod 状态标签
const getPodStatusTag = (pod: any) => {
  const phase = pod.phase || 'Unknown';
  switch (phase) {
    case 'Running':
      return { type: 'success', text: '运行中' };
    case 'Succeeded':
      return { type: 'info', text: '已完成' };
    case 'Failed':
      return { type: 'danger', text: '失败' };
    case 'Pending':
      return { type: 'warning', text: '等待中' };
    default:
      return { type: 'info', text: '未知' };
  }
};

// 加载 Deployment 详情
async function loadDeploymentDetail() {
  const queryClusterId = route.query.clusterId as string;
  const queryNamespace = route.query.namespace as string;
  const queryName = route.query.name as string;

  if (!queryClusterId || !queryNamespace || !queryName) {
    ElMessage.error('参数不完整');
    router.push('/k8s/workloads');
    return;
  }

  clusterId.value = Number(queryClusterId);
  namespace.value = queryNamespace;
  deploymentName.value = queryName;

  loading.value = true;
  try {
    const res = await getK8sDeployment(clusterId.value, namespace.value, deploymentName.value);
    deployment.value = res.data || res;

    // 默认加载容器组
    if (activeTab.value === 'pods') {
      await loadPods();
    }
  } catch (error: any) {
    console.error('获取 Deployment 详情失败:', error);
    ElMessage.error(error.message || '获取 Deployment 详情失败');
  } finally {
    loading.value = false;
  }
}

// 加载关联的 Pods
async function loadPods() {
  if (!clusterId.value) return;

  podsLoading.value = true;
  try {
    const res = await getK8sDeploymentPods(clusterId.value, namespace.value, deploymentName.value);
    const podData = res.data || res;
    pods.value = Array.isArray(podData) ? podData : [];
  } catch (error: any) {
    console.error('获取 Pods 失败:', error);
    ElMessage.error(error.message || '获取 Pods 失败');
    pods.value = [];
  } finally {
    podsLoading.value = false;
  }
}

// 加载事件
async function loadEvents() {
  if (!clusterId.value) return;

  eventsLoading.value = true;
  try {
    // 使用 fieldSelector 过滤与当前 Deployment 相关的事件
    const fieldSelector = `involvedObject.name=${deploymentName.value},involvedObject.kind=Deployment`;
    const res = await fetchK8sEvents(clusterId.value, namespace.value, fieldSelector);
    const eventData = res.data || res;
    events.value = Array.isArray(eventData) ? eventData : [];
  } catch (error: any) {
    console.error('获取 Events 失败:', error);
    ElMessage.error(error.message || '获取 Events 失败');
    events.value = [];
  } finally {
    eventsLoading.value = false;
  }
}

// 标签页切换
async function handleTabChange(tab: string) {
  activeTab.value = tab;
  if (tab === 'pods' && pods.value.length === 0) {
    await loadPods();
  } else if (tab === 'events' && events.value.length === 0) {
    await loadEvents();
  }
}

// 刷新数据
function handleRefresh() {
  loadDeploymentDetail();
}

// 刷新 Pods
function handleRefreshPods() {
  loadPods();
}

// 缩放 Deployment
function handleScale() {
  if (!deployment.value) return;
  scaleData.value = { replicas: deployment.value.replicas || 1 };
  showScaleModal.value = true;
}

async function handleScaleSubmit() {
  if (!scaleFormRef.value) return;

  try {
    await scaleFormRef.value.validate();
    await scaleK8sDeployment(clusterId.value, {
      namespace: namespace.value,
      name: deploymentName.value,
      replicas: scaleData.value.replicas
    });
    ElMessage.success('缩放成功');
    showScaleModal.value = false;
    loadDeploymentDetail();
  } catch (error: any) {
    ElMessage.error(error.message || '缩放失败');
  }
}

// 重启 Deployment
async function handleRestart() {
  try {
    await ElMessageBox.confirm(`确定要重启 Deployment "${deploymentName.value}" 吗？`, '确认重启', {
      type: 'warning'
    });

    await restartK8sDeployment(clusterId.value, {
      namespace: namespace.value,
      name: deploymentName.value
    });
    ElMessage.success('重启成功');
    loadDeploymentDetail();
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '重启失败');
    }
  }
}

// 删除 Deployment
async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      `确定要删除 Deployment "${deploymentName.value}" 吗？此操作危险，请谨慎操作。`,
      '确认删除',
      {
        type: 'warning',
        confirmButtonText: '危险操作确认',
        cancelButtonText: '取消'
      }
    );

    await deleteK8sDeployment(clusterId.value, {
      namespace: namespace.value,
      name: deploymentName.value
    });
    ElMessage.success('删除成功');
    router.push('/k8s/workloads');
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败');
    }
  }
}

// Pod 终端
function handlePodTerminal(row: any) {
  terminalProps.value = {
    clusterId: clusterId.value,
    namespace: namespace.value,
    podName: row.name,
    containerName: row.containers?.[0]?.name || ''
  };
  showTerminal.value = true;
}

function closeTerminal() {
  showTerminal.value = false;
}

// Pod 日志
async function handlePodLogs(row: any) {
  logPodName.value = row.name;
  logContent.value = '加载中...';
  showLogs.value = true;

  try {
    const containerName = row.containers?.[0]?.name || '';
    const res = await fetchK8sPodLogs(clusterId.value, namespace.value, row.name, {
      container: containerName,
      tailLines: 100
    });
    logContent.value = res.data?.logs || '暂无日志';
  } catch (error: any) {
    logContent.value = `日志加载失败: ${error.message || '未知错误'}`;
  }
}

function closeLogs() {
  showLogs.value = false;
  logContent.value = '';
}

// 打开 YAML 编辑器
function handleEditYaml() {
  try {
    // 后端返回的是 JSON 字符串，需要先解析为对象
    const manifestObj = deployment.value?.manifest ? JSON.parse(deployment.value.manifest) : {};
    // 转换为格式化的 YAML
    yamlEditingContent.value = yaml.dump(manifestObj, {
      indent: 2,
      lineWidth: -1,
      noRefs: true
    });
    yamlContent.value = yamlEditingContent.value;
    showYamlEditor.value = true;
  } catch (error) {
    console.error('解析 manifest 失败:', error);
    ElMessage.error('解析 YAML 失败');
  }
}

// 关闭 YAML 编辑器
function closeYamlEditor() {
  showYamlEditor.value = false;
  yamlEditingContent.value = '';
}

// 保存 YAML
async function handleSaveYaml() {
  yamlSaving.value = true;
  try {
    // 将 YAML 转换回 JSON 对象
    const manifestObj = yaml.load(yamlEditingContent.value);
    // 调用更新 API，将 manifestObj 传递给后端
    await updateK8sDeployment(clusterId.value, {
      namespace: namespace.value,
      manifest: manifestObj
    });
    ElMessage.success('保存成功');
    closeYamlEditor();
    // 刷新数据
    await loadDeploymentDetail();
  } catch (error: any) {
    console.error('保存 YAML 失败:', error);
    ElMessage.error(`保存失败: ${error.message || 'YAML 格式错误'}`);
  } finally {
    yamlSaving.value = false;
  }
}

onMounted(() => {
  loadDeploymentDetail();
});
</script>

<template>
  <div v-loading="loading" class="detail-page bg-layout">
    <!-- 顶部操作栏 -->
    <div class="instance-bar">
      <div class="instance-left">
        <span class="back-arrow" @click="router.push('/k8s/workloads')">←</span>
        <span class="instance-name">{{ deployment?.name || '-' }}</span>
        <span class="instance-meta">命名空间: {{ deployment?.namespace || '-' }}</span>
        <span class="instance-meta">副本数: {{ deployment?.ready || 0 }} / {{ deployment?.replicas || 0 }}</span>
        <ElTag v-if="deployment" :type="getStatusTag.type" size="small">
          {{ getStatusTag.text }}
        </ElTag>
      </div>
      <div class="instance-actions">
        <ElButton size="small" @click="handleEditYaml">YAML 编辑</ElButton>
        <ElButton size="small" type="primary" @click="handleScale">缩放</ElButton>
        <ElButton size="small" type="warning" @click="handleRestart">重启</ElButton>
        <ElButton size="small" type="danger" @click="handleDelete">删除</ElButton>
        <ElButton size="small" @click="router.push('/k8s/workloads')">返回列表</ElButton>
      </div>
    </div>

    <!-- 基本信息区域 -->
    <div class="basic-info-section">
      <!-- 基本信息 -->
      <div class="info-section">
        <h3 class="section-title">基本信息</h3>
        <div class="desc-grid">
          <div class="desc-row">
            <div class="desc-item">
              <div class="item-label">名称</div>
              <div class="item-content">
                <span class="value-text">{{ deployment?.name || '-' }}</span>
              </div>
            </div>
            <div class="desc-item">
              <div class="item-label">命名空间</div>
              <div class="item-content">
                <span class="value-text">{{ deployment?.namespace || '-' }}</span>
              </div>
            </div>
            <div class="desc-item">
              <div class="item-label">创建时间</div>
              <div class="item-content">
                <span class="value-text">{{ deployment?.age || '-' }}</span>
              </div>
            </div>
          </div>
          <div class="desc-row">
            <div class="desc-item">
              <div class="item-label">副本数</div>
              <div class="item-content">
                <span class="value-text">{{ deployment?.ready || 0 }} / {{ deployment?.replicas || 0 }}</span>
              </div>
            </div>
            <div class="desc-item">
              <div class="item-label">可用副本</div>
              <div class="item-content">
                <span class="value-text">{{ deployment?.available || 0 }}</span>
              </div>
            </div>
            <div class="desc-item">
              <div class="item-label">最新副本</div>
              <div class="item-content">
                <span class="value-text">{{ deployment?.upToDate || 0 }}</span>
              </div>
            </div>
          </div>
          <div v-if="deployment?.labels" class="desc-row">
            <div class="desc-item desc-item-full">
              <div class="item-label">标签</div>
              <div class="item-content">
                <span class="value-text">{{ formatLabels(deployment.labels) }}</span>
              </div>
            </div>
          </div>
          <div v-if="deployment?.conditions && deployment.conditions.length > 0" class="desc-row">
            <div class="desc-item desc-item-full">
              <div class="item-label">状态条件</div>
              <div class="item-content">
                <span class="value-text">{{ formatConditions(deployment.conditions) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 标签切换 -->
    <ElTabs v-model="activeTab" class="detail-tabs" @tab-change="handleTabChange">
      <!-- 容器组 -->
      <ElTabPane label="容器组" name="pods">
        <div class="tab-content">
          <div class="section-header">
            <span class="section-title">共 {{ pods.length }} 个容器组</span>
            <ElButton size="small" @click="handleRefreshPods">刷新</ElButton>
          </div>
          <ElTable v-loading="podsLoading" :data="pods" stripe size="small">
            <ElTableColumn prop="name" label="Pod 名称" min-width="200" show-overflow-tooltip />
            <ElTableColumn label="状态" width="120">
              <template #default="{ row }">
                <ElTag :type="getPodStatusTag(row).type" size="small">
                  {{ getPodStatusTag(row).text }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="ip" label="IP 地址" width="140" />
            <ElTableColumn prop="node" label="节点" width="150" show-overflow-tooltip />
            <ElTableColumn label="重启次数" width="100" align="center">
              <template #default="{ row }">
                {{ row.restarts || 0 }}
              </template>
            </ElTableColumn>
            <ElTableColumn prop="age" label="年龄" width="140" />
            <ElTableColumn label="操作" width="150" fixed="right">
              <template #default="{ row }">
                <ElButton size="small" type="primary" link @click="handlePodTerminal(row)">终端</ElButton>
                <ElButton size="small" link @click="handlePodLogs(row)">日志</ElButton>
              </template>
            </ElTableColumn>
          </ElTable>
          <div v-if="!podsLoading && pods.length === 0" class="py-8 text-center text-gray-500">暂无容器组</div>
        </div>
      </ElTabPane>

      <!-- 事件 -->
      <ElTabPane label="事件" name="events">
        <div class="tab-content">
          <div class="section-header">
            <span class="section-title">共 {{ events.length }} 个事件</span>
            <ElButton size="small" @click="loadEvents">刷新</ElButton>
          </div>
          <ElTable v-loading="eventsLoading" :data="events" stripe size="small">
            <ElTableColumn prop="type" label="类型" width="120">
              <template #default="{ row }">
                <ElTag :type="row.type === 'Normal' ? 'success' : 'warning'" size="small">
                  {{ row.type }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="reason" label="原因" width="150" show-overflow-tooltip />
            <ElTableColumn prop="message" label="消息" min-width="300" show-overflow-tooltip />
            <ElTableColumn prop="source" label="来源" width="150" show-overflow-tooltip />
            <ElTableColumn prop="count" label="次数" width="80" align="center" />
            <ElTableColumn prop="lastTimestamp" label="最后时间" width="160" />
          </ElTable>
          <div v-if="!eventsLoading && events.length === 0" class="py-8 text-center text-gray-500">暂无事件</div>
        </div>
      </ElTabPane>
    </ElTabs>

    <!-- 缩放对话框 -->
    <ElDialog v-model="showScaleModal" title="缩放 Deployment" width="500px">
      <ElForm ref="scaleFormRef" :model="scaleData" label-width="100px">
        <ElFormItem label="Deployment">
          <ElInput :value="deploymentName" disabled />
        </ElFormItem>
        <ElFormItem label="副本数" prop="replicas">
          <ElInputNumber v-model="scaleData.replicas" :min="0" :max="100" :step="1" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="showScaleModal = false">取消</ElButton>
        <ElButton type="primary" @click="handleScaleSubmit">确定</ElButton>
      </template>
    </ElDialog>

    <!-- YAML 编辑器对话框 -->
    <ElDialog v-model="showYamlEditor" title="编辑 YAML 配置" width="900px" top="5vh" @close="closeYamlEditor">
      <div class="yaml-editor-container">
        <textarea v-model="yamlEditingContent" class="yaml-editor" spellcheck="false"></textarea>
      </div>
      <template #footer>
        <ElButton @click="closeYamlEditor">取消</ElButton>
        <ElButton type="primary" :loading="yamlSaving" @click="handleSaveYaml">保存</ElButton>
      </template>
    </ElDialog>

    <!-- Pod 终端对话框 -->
    <ElDialog v-model="showTerminal" title="Pod 终端" width="900px" fullscreen @close="closeTerminal">
      <PodTerminal
        v-if="showTerminal"
        :cluster-id="terminalProps.clusterId"
        :namespace="terminalProps.namespace"
        :pod-name="terminalProps.podName"
        :container-name="terminalProps.containerName"
      />
    </ElDialog>

    <!-- Pod 日志对话框 -->
    <ElDialog v-model="showLogs" :title="`日志: ${logPodName}`" width="900px" top="5vh">
      <div class="log-content">
        {{ logContent }}
      </div>
      <template #footer>
        <ElButton @click="closeLogs">关闭</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<style scoped>
/* 页面容器 */
.detail-page {
  min-height: 100vh;
  padding: 24px;
}

/* 顶部操作栏 */
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
  opacity: 0.8;
}

.instance-name {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.instance-meta {
  font-size: 14px;
  color: #909399;
}

.instance-actions {
  display: flex;
  gap: 8px;
}

/* 基本信息区域 - 透明背景，无边框 */
.basic-info-section {
  background: transparent;
  margin-bottom: 24px;
}

.info-section {
  background: transparent;
  margin-bottom: 16px;
}

.section-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 12px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

/* 标签内容 */
.tab-content {
  padding: 16px;
}

/* 标签页样式 */
.detail-tabs {
  margin-top: 16px;
}

/* 描述网格 */
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
}

.item-content {
  font-size: 14px;
  color: #303133;
}

.value-text {
  color: #303133;
}

.font-mono {
  font-family: 'Courier New', Courier, monospace;
  word-break: break-all;
}

/* 标签列表 */
.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

/* YAML 编辑器 */
.yaml-editor-container {
  width: 100%;
  height: 500px;
}

.yaml-editor {
  width: 100%;
  height: 100%;
  padding: 12px;
  font-family: 'Courier New', Courier, monospace;
  font-size: 13px;
  line-height: 1.5;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background: #f5f7fa;
  color: #303133;
  resize: none;
  outline: none;
}

.yaml-editor:focus {
  border-color: #0052d9;
}

/* 日志内容 */
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
