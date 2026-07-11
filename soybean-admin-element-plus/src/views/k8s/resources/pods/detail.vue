<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElTag } from 'element-plus';
import yaml from 'js-yaml';
import { fetchK8sEvents, fetchK8sPodLogs, getK8sPod, updateK8sPod } from '@/service/api/k8s';
import { formatAnnotations, formatLabels, formatSelectors } from '@/utils/k8s-formatters';
import PodTerminal from '../../terminal/PodTerminal.vue';
import K8sResourceActionBar from '@/components/k8s/K8sResourceActionBar.vue';

defineOptions({ name: 'K8sPodDetail' });

const route = useRoute();
const router = useRouter();

const resource = ref<any>(null);
const events = ref<any[]>([]);
const loading = ref(true);
const eventsLoading = ref(false);
const activeTab = ref('containers');

const clusterId = ref<number>(0);
const namespace = ref('');
const resourceName = ref('');

// 终端相关
const showTerminal = ref(false);
const terminalProps = ref({ clusterId: 0, namespace: '', podName: '', containerName: '' });

// 日志相关
const showLogs = ref(false);
const logContent = ref('');
const logPodName = ref('');
const logContainerName = ref('');

// YAML 编辑相关
const showYamlEditor = ref(false);
const yamlContent = ref('');
const yamlEditingContent = ref('');
const yamlSaving = ref(false);

// 返回列表页的路径 - 直接返回工作负载汇总页，状态由 store 管理
const backPath = '/k8s/workloads';

const getPhaseTag = computed(() => {
  if (!resource.value) return { type: 'info', text: '未知' };
  switch (resource.value.phase) {
    case 'Running':
      return { type: 'success', text: '运行中' };
    case 'Succeeded':
      return { type: 'info', text: '已完成' };
    case 'Failed':
      return { type: 'danger', text: '失败' };
    case 'Pending':
      return { type: 'warning', text: '等待中' };
    default:
      return { type: 'info', text: resource.value.phase || '未知' };
  }
});

// 获取标签颜色
const getTagType = (key: string) => {
  const keyLower = key.toLowerCase();
  if (keyLower.includes('app') || keyLower.includes('name')) return 'primary';
  if (keyLower.includes('env') || keyLower.includes('environment')) return 'success';
  if (keyLower.includes('version') || keyLower.includes('ver')) return 'warning';
  if (keyLower.includes('component')) return 'info';
  return 'default';
};

// 获取注解标签颜色
const getAnnotationTagType = (key: string) => {
  const keyLower = key.toLowerCase();
  if (keyLower.includes('last-applied')) return 'warning';
  if (keyLower.includes('revision')) return 'success';
  if (keyLower.includes('project')) return 'primary';
  if (keyLower.includes('version')) return 'info';
  if (keyLower.includes('kubernetes.io/') || keyLower.includes('k8s.io/')) return 'info';
  return 'default';
};

// 获取截断的注解键名
const getShortAnnotationKey = (key: string) => {
  if (key.length > 25) {
    if (key.startsWith('kubectl.kubernetes.io/')) {
      return 'kubectl.../' + key.split('/').pop();
    }
    if (key.startsWith('deployment.kubernetes.io/')) {
      return 'dep.../' + key.split('/').pop();
    }
    return key.substring(0, 10) + '...' + key.substring(key.length - 10);
  }
  return key;
};

// 获取截断的注解值
const getShortAnnotationValue = (value: string) => {
  if (value.length > 20) {
    return value.substring(0, 20) + '...';
  }
  return value;
};

// 格式化注解数据
const formattedAnnotations = computed(() => {
  if (!resource.value?.annotations) return [];
  const result = formatAnnotations(resource.value.annotations);
  return [...result.user, ...result.system];
});

// 格式化选择器为列表
const formatSelectorsList = (selectors: Record<string, string> | undefined) => {
  if (!selectors || Object.keys(selectors).length === 0) return [];
  return Object.entries(selectors).map(([key, value]) => ({ key, value }));
};

async function loadDetail() {
  const qClusterId = route.query.clusterId as string;
  const qNamespace = route.query.namespace as string;
  const qName = route.query.name as string;

  if (!qClusterId || !qNamespace || !qName) {
    ElMessage.error('参数不完整');
    router.push('/k8s/workloads');
    return;
  }

  clusterId.value = Number(qClusterId);
  namespace.value = qNamespace;
  resourceName.value = qName;

  loading.value = true;
  try {
    const res = await getK8sPod(clusterId.value, namespace.value, resourceName.value);
    resource.value = res.data || res;
  } catch (error: any) {
    ElMessage.error(error.message || '获取 Pod 详情失败');
  } finally {
    loading.value = false;
  }
}

async function loadEvents() {
  if (!clusterId.value) return;
  eventsLoading.value = true;
  try {
    const fieldSelector = `involvedObject.name=${resourceName.value},involvedObject.kind=Pod`;
    const res = await fetchK8sEvents(clusterId.value, namespace.value, fieldSelector);
    const eventData = res.data || res;
    events.value = Array.isArray(eventData) ? eventData : [];
  } catch (error: any) {
    ElMessage.error(error.message || '获取 Events 失败');
    events.value = [];
  } finally {
    eventsLoading.value = false;
  }
}

async function handleTabChange(tab: string) {
  activeTab.value = tab;
  if (tab === 'events' && events.value.length === 0) await loadEvents();
}

// 终端
function handleTerminal(containerName?: string) {
  terminalProps.value = {
    clusterId: clusterId.value,
    namespace: namespace.value,
    podName: resourceName.value,
    containerName: containerName || resource.value?.containers?.[0]?.name || ''
  };
  showTerminal.value = true;
}

// 关闭终端
function closeTerminal() {
  showTerminal.value = false;
}

// 日志
async function handleLogs(containerName?: string) {
  const cName = containerName || resource.value?.containers?.[0]?.name || '';
  logPodName.value = resourceName.value;
  logContainerName.value = cName;
  logContent.value = '加载中...';
  showLogs.value = true;

  try {
    const res = await fetchK8sPodLogs(clusterId.value, namespace.value, resourceName.value, {
      container: cName,
      tailLines: 200
    });
    logContent.value = res.data?.logs || '暂无日志';
  } catch (error: any) {
    logContent.value = `日志加载失败: ${error.message || '未知错误'}`;
  }
}

// 关闭日志
function closeLogs() {
  showLogs.value = false;
  logContent.value = '';
}

// 打开 YAML 编辑器
function handleEditYaml() {
  try {
    // 后端返回的是 JSON 字符串，需要先解析为对象
    const manifestStr = resource.value?.manifest || resource.value;
    const manifestObj = typeof manifestStr === 'string' ? JSON.parse(manifestStr) : manifestStr;
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
    // 调用更新 API
    await updateK8sPod(clusterId.value, {
      namespace: namespace.value,
      manifest: manifestObj
    });
    ElMessage.success('保存成功');
    closeYamlEditor();
    // 刷新数据
    await loadDetail();
  } catch (error: any) {
    console.error('保存 YAML 失败:', error);
    ElMessage.error(`保存失败: ${error.message || 'YAML 格式错误'}`);
  } finally {
    yamlSaving.value = false;
  }
}

onMounted(() => {
  console.log('[Pod详情] onMounted');
  console.log('[Pod详情] 当前路由参数:', {
    clusterId: route.query.clusterId,
    namespace: route.query.namespace,
    name: route.query.name
  });

  loadDetail();
});
</script>

<template>
  <div v-loading="loading" class="detail-page bg-layout">
    <!-- 顶部操作栏 - 使用新组件 -->
    <K8sResourceActionBar
      :name="resource?.name || '-'"
      :namespace="resource?.namespace || '-'"
      :status-tag="getPhaseTag"
      :back-path="backPath"
      :actions="[
        ...(resource?.phase === 'Running' ? [{ label: '终端', type: 'primary', handler: () => handleTerminal(), tooltip: '打开容器终端' }] : []),
        { label: '日志', handler: () => handleLogs(), tooltip: '查看容器日志' },
        { label: '编辑YAML', handler: handleEditYaml, tooltip: '编辑 YAML 配置' }
      ]"
    />

    <!-- 基本信息区域 -->
    <div class="basic-info-section">
      <!-- 基本信息 -->
      <div class="info-section">
        <h3 class="section-title">基本信息</h3>
        <div class="desc-grid">
          <div class="desc-row">
            <div class="desc-item">
              <div class="item-label">名称:</div>
              <div class="item-content">{{ resource?.name || '-' }}</div>
            </div>
            <div class="desc-item">
              <div class="item-label">命名空间:</div>
              <div class="item-content">{{ resource?.namespace || '-' }}</div>
            </div>
          </div>
          <div class="desc-row">
            <div class="desc-item">
              <div class="item-label">创建时间:</div>
              <div class="item-content">{{ resource?.age || '-' }}</div>
            </div>
            <div class="desc-item">
              <div class="item-label">状态:</div>
              <div class="item-content">
                <ElTag :type="getPhaseTag.type" size="small">{{ getPhaseTag.text }}</ElTag>
              </div>
            </div>
          </div>
          <div class="desc-row">
            <div class="desc-item">
              <div class="item-label">IP 地址:</div>
              <div class="item-content">{{ resource?.ip || '-' }}</div>
            </div>
            <div class="desc-item">
              <div class="item-label">所在节点:</div>
              <div class="item-content">{{ resource?.node || '-' }}</div>
            </div>
          </div>
          <div class="desc-row">
            <div class="desc-item">
              <div class="item-label">重启次数:</div>
              <div class="item-content">{{ resource?.restarts ?? 0 }}</div>
            </div>
          </div>
          <div v-if="resource?.selectors && Object.keys(resource.selectors).length > 0" class="desc-row">
            <div class="desc-item desc-item-full">
              <div class="item-label-vertical">选择器:</div>
              <div class="item-content-vertical selectors-list">
                <ElTag
                  v-for="(selector, idx) in formatSelectorsList(resource.selectors)"
                  :key="idx"
                  type="primary"
                  size="small"
                  class="selector-tag"
                >
                  {{ selector.key }}: {{ selector.value }}
                </ElTag>
              </div>
            </div>
          </div>
          <div v-if="resource?.annotations && Object.keys(resource.annotations).length > 0" class="desc-row">
            <div class="desc-item desc-item-full">
              <div class="item-label-vertical">注解:</div>
              <div class="item-content-vertical annotations-list">
                <div v-for="(item, idx) in formattedAnnotations" :key="idx" class="annotation-item">
                  <ElTag :type="getAnnotationTagType(item.key)" size="small" class="annotation-tag-vertical">
                    <span v-if="item.isLong" class="annotation-key-short">{{ getShortAnnotationKey(item.key) }}</span>
                    <span v-else class="annotation-key-full">{{ item.key }}</span>
                    <span v-if="!item.isLong" class="annotation-sep">:</span>
                    <span v-if="item.isLong" class="expand-hint">展开</span>
                    <span v-else class="annotation-value-short">{{ getShortAnnotationValue(item.value) }}</span>
                  </ElTag>
                </div>
              </div>
            </div>
          </div>
          <div v-if="resource?.labels && Object.keys(resource.labels).length > 0" class="desc-row">
            <div class="desc-item desc-item-full">
              <div class="item-label">标签:</div>
              <div class="item-content tags-list">
                <ElTag
                  v-for="(tag, idx) in formatLabels(resource.labels)"
                  :key="idx"
                  :type="getTagType(tag.key)"
                  size="small"
                  class="tag-item"
                >
                  {{ tag.key }}: {{ tag.value }}
                </ElTag>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 镜像信息 -->
      <div v-if="resource?.images && resource.images.length > 0" class="info-section">
        <h3 class="section-title">镜像信息</h3>
        <div class="desc-grid">
          <div class="desc-row">
            <div v-for="(image, index) in resource.images" :key="index" class="desc-item desc-item-full">
              <div class="item-label">镜像 {{ index + 1 }}</div>
              <div class="item-content font-mono">{{ image }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 标签切换 -->
    <ElTabs v-model="activeTab" class="detail-tabs" @tab-change="handleTabChange">
      <!-- 容器组 -->
      <ElTabPane :label="`容器 (${resource?.containers?.length || 0})`" name="containers">
        <template #label>
          <div class="tab-pane-header">
            <span>容器</span>
            <ElTag size="small" class="count-tag">{{ resource?.containers?.length || 0 }}</ElTag>
          </div>
        </template>
        <div class="tab-content">
          <div v-if="resource?.containers && resource.containers.length > 0">
            <ElTable
              :data="resource.containers"
              stripe
              size="small"
              class="pods-containers-table"
              :header-cell-style="{ background: '#f5f7fa', color: '#303133', fontWeight: '600', paddingLeft: '16px', paddingRight: '16px' }"
              :row-style="{ backgroundColor: 'transparent' }"
              :cell-style="{ backgroundColor: 'transparent', padding: '8px 16px' }"
            >
              <ElTableColumn prop="name" label="容器名称" min-width="180" />
              <ElTableColumn prop="image" label="镜像" min-width="250" show-overflow-tooltip />
              <ElTableColumn label="操作" width="180" fixed="right">
                <template #default="{ row }">
                  <ElButton size="small" type="primary" link @click="handleTerminal(row.name)">终端</ElButton>
                  <ElButton size="small" link @click="handleLogs(row.name)">日志</ElButton>
                </template>
              </ElTableColumn>
            </ElTable>
          </div>
          <div v-else class="py-8 text-center text-gray-500">暂无容器组信息</div>
        </div>
      </ElTabPane>

      <!-- 事件 -->
      <ElTabPane :label="`事件 (${events.length})`" name="events">
        <div class="tab-toolbar">
          <ElButton size="small" @click="loadEvents">刷新</ElButton>
        </div>
        <div class="tab-content">
          <ElTable
            v-loading="eventsLoading"
            :data="events"
            stripe
            size="small"
            class="pods-events-table"
            :header-cell-style="{ background: '#f5f7fa', color: '#303133', fontWeight: '600' }"
            :row-style="{ backgroundColor: 'transparent' }"
            :cell-style="{ backgroundColor: 'transparent', padding: '8px 0' }"
          >
            <ElTableColumn label="类型" width="120">
              <template #default="{ row }">
                <ElTag :type="row.type === 'Normal' ? 'success' : 'warning'" size="small">{{ row.type }}</ElTag>
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

    <!-- 终端对话框 -->
    <ElDialog v-model="showTerminal" title="Pod 终端" width="900px" fullscreen @close="closeTerminal">
      <PodTerminal
        v-if="showTerminal"
        :cluster-id="terminalProps.clusterId"
        :namespace="terminalProps.namespace"
        :pod-name="terminalProps.podName"
        :container-name="terminalProps.containerName"
      />
    </ElDialog>

    <!-- 日志对话框 -->
    <ElDialog
      v-model="showLogs"
      :title="`日志: ${logPodName}${logContainerName ? ' (' + logContainerName + ')' : ''}`"
      width="900px"
      top="5vh"
    >
      <div class="log-content">{{ logContent }}</div>
      <template #footer>
        <ElButton @click="closeLogs">关闭</ElButton>
      </template>
    </ElDialog>

    <!-- YAML 编辑对话框 -->
    <ElDialog v-model="showYamlEditor" title="编辑 YAML 配置" width="900px" top="5vh" @close="closeYamlEditor">
      <div class="yaml-editor-container">
        <textarea v-model="yamlEditingContent" class="yaml-editor" spellcheck="false"></textarea>
      </div>
      <template #footer>
        <ElButton @click="closeYamlEditor">取消</ElButton>
        <ElButton type="primary" :loading="yamlSaving" @click="handleSaveYaml">保存</ElButton>
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

/* 标签页工具栏 */
.tab-toolbar {
  display: flex;
  justify-content: flex-end;
  padding: 0 16px 8px 16px;
}

/* 基本信息区域 */
.basic-info-section {
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

.tab-content {
  padding: 16px;
}

.detail-tabs {
  margin-top: 16px;
}

/* 移除 tab内容默认的padding，让内容更紧凑 */
.detail-tabs :deep(.el-tabs__content) {
  padding: 0;
}

/* tab标题栏 - 自定义样式 */
.tab-pane-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 8px;
  line-height: 1;
}

/* 计数标签 */
.tab-pane-header .count-tag {
  background: #f0f2f5;
  color: #606266;
  border: none;
  font-size: 12px;
  height: 18px;
  line-height: 18px;
  padding: 0 6px;
  font-weight: 500;
}

/* 描述列表样式 */
.desc-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.desc-row {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 24px;
}

.desc-item {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.desc-item-full {
  grid-column: 1 / -1;
}

.item-label {
  font-size: 13px;
  color: #606266;
  white-space: nowrap;
}

.item-content {
  font-size: 13px;
  color: #303133;
}

.tags-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: flex-start;
  width: 100%;
}

.tag-item {
  display: flex;
  align-items: center;
  width: 100%;
  gap: 4px;
}

/* 注解列表样式 */
.annotations-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: flex-start;
  width: 100%;
}

.annotation-item {
  width: 100%;
}

.annotation-tag-vertical {
  display: flex !important;
  align-items: center;
  width: 100%;
  font-family: 'Courier New', Courier, monospace;
  padding: 4px 8px !important;
}

.annotation-key-full {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 180px;
}

.annotation-key-short {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 120px;
  font-weight: 500;
}

.annotation-sep {
  flex-shrink: 0;
}

.annotation-value-short {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 80px;
  flex-shrink: 1;
}

.expand-hint {
  color: inherit;
  font-weight: 600;
  font-size: 11px;
  flex-shrink: 0;
}

/* 选择器列表样式 */
.selectors-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: flex-start;
  width: 100%;
}

.selector-tag {
  display: flex !important;
  align-items: center;
  width: 100%;
  font-family: 'Courier New', Courier, monospace;
  margin: 0;
  padding: 4px 8px !important;
  border: none !important;
}

/* 垂直布局样式 */
.item-label-vertical {
  font-size: 13px;
  color: #606266;
  white-space: nowrap;
  width: 80px;
  flex-shrink: 0;
  padding-top: 2px;
}

.item-content-vertical {
  flex: 1;
  font-size: 13px;
  color: #303133;
}

.font-mono {
  font-family: 'Courier New', Courier, monospace;
  word-break: break-all;
}

.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

/* 日志内容样式 */
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

/* YAML 编辑器样式 */
.yaml-editor-container {
  width: 100%;
  height: 500px;
}

.yaml-editor {
  width: 100%;
  height: 100%;
  padding: 12px;
  font-family: 'Courier New', Courier, monospace;
  font-size: 14px;
  line-height: 1.5;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  resize: none;
  background-color: #1e1e1e;
  color: #d4d4d4;
}

.yaml-editor:focus {
  outline: none;
  border-color: #409eff;
}

/* 表格透明样式 - 完全覆盖 Element Plus 默认样式 */
.pods-containers-table,
.pods-events-table,
.pods-containers-table.el-table,
.pods-events-table.el-table,
:deep(.el-table),
:deep(.el-table__body),
:deep(.el-table__body-wrapper),
:deep(.el-table__inner-wrapper),
:deep(.el-table__header) {
  background-color: transparent !important;
}

/* 表头样式 */
:deep(.el-table__header-wrapper) {
  background-color: transparent !important;

  th.el-table__cell {
    background-color: #f5f7fa !important;
    color: #303133;
    font-weight: 600;
    text-align: left;
  }
}

/* 移除所有行的背景色 */
:deep(.el-table__body-wrapper) {
  background-color: transparent !important;
}

:deep(.el-table__body) {
  background-color: transparent !important;
}

:deep(.el-table__body tr) {
  background-color: transparent !important;
}

:deep(.el-table__body td.el-table__cell) {
  background-color: transparent !important;
}

/* 去掉斑马纹 */
:deep(.el-table--striped .el-table__body tr.el-table__row--striped) {
  background-color: transparent !important;

  td.el-table__cell {
    background-color: transparent !important;
  }
}

/* 移除 hover 效果的背景色 */
:deep(.el-table__body tr:hover > td.el-table__cell) {
  background-color: transparent !important;
}

/* 移除固定列的背景色 */
:deep(.el-table__fixed),
:deep(.el-table__fixed-body-wrapper) {
  background-color: transparent !important;

  .el-table__body tr {
    background-color: transparent !important;
  }

  .el-table__body td.el-table__cell {
    background-color: transparent !important;
  }
}

/* 移除表格容器的背景色 */
:deep(.el-table__inner-wrapper) {
  background-color: transparent !important;
}
</style>
