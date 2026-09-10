<script setup lang="ts">
  import { computed, onMounted, ref } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { ElMessage, ElMessageBox } from 'element-plus';
  import { ArrowLeft } from '@element-plus/icons-vue';
  import {
    deleteK8sCronJob,
    fetchK8sEvents,
    getK8sCronJob,
    getK8sCronJobPods,
    suspendK8sCronJob,
    updateK8sCronJob
  } from '@/service/api/k8s';
  import { formatLabels, getTagType, parseManifest } from '@/views/k8s/shared/k8s-formatters';
  import { useYamlEdit } from '@/views/k8s/composables/useYamlEdit';
  import YamlEditor from '@/components/k8s/YamlEditor.vue';
  import K8sPodsTable from '@/components/k8s/K8sPodsTable.vue';
  import K8sEventsTable from '@/components/k8s/K8sEventsTable.vue';

  defineOptions({ name: 'K8sCronJobDetail' });

  const route = useRoute();
  const router = useRouter();
  const message = ElMessage;

  const clusterId = computed(() => Number(route.query.clusterId));
  const namespace = computed(() => route.query.namespace as string);
  const resourceName = computed(() => route.query.name as string);

  const resource = ref<K8s.CronJob | null>(null);
  const pods = ref<K8s.Pod[]>([]);
  const events = ref<K8s.Event[]>([]);
  const loading = ref(true);
  const podsLoading = ref(false);
  const eventsLoading = ref(false);
  const activeTab = ref('basic');
  const showYamlEditor = ref(false);
  const yamlContent = ref('');

  async function loadData() {
    loading.value = true;
    try {
      const { data, error } = await getK8sCronJob(clusterId.value, namespace.value, resourceName.value);
      if (!error && data) {
        resource.value = data;
        yamlContent.value = parseManifest(data.manifest || '');
      }
    } catch (error: unknown) {
      const err = error as Error;
      message.error(err.message || '加载失败');
    } finally {
      loading.value = false;
    }
  }

  async function loadEvents() {
    eventsLoading.value = true;
    try {
      const { data: eventsData, error: eventsError } = await fetchK8sEvents(
        clusterId.value,
        namespace.value,
        `involvedObject.name=${resourceName.value}`
      );
      events.value = !eventsError && eventsData ? eventsData : [];
    } catch (error: unknown) {
      const err = error as Error;
      message.error(err.message || '加载事件失败');
    } finally {
      eventsLoading.value = false;
    }
  }

  async function loadPods() {
    if (!clusterId.value) return;
    podsLoading.value = true;
    try {
      const { data: podsData, error: podsError } = await getK8sCronJobPods(
        clusterId.value,
        namespace.value,
        resourceName.value
      );
      pods.value = !podsError && podsData ? podsData : [];
    } catch (error: unknown) {
      const err = error as Error;
      message.error(err.message || '加载容器组失败');
    } finally {
      podsLoading.value = false;
    }
  }

  // 返回列表页的路径
  const backPath = computed(() => {
    const savedState = sessionStorage.getItem('k8s_cronjobs_list_state');

    if (savedState) {
      try {
        const state = JSON.parse(savedState);
        const path = state.listPath || '/k8s/workloads';
        return path;
      } catch (e) {
        console.error('[CronJob详情] 解析状态失败:', e);
      }
    }
    return '/k8s/workloads';
  });

  function goBack() {
    router.push(backPath.value);
  }

  async function handleToggleSuspend() {
    try {
      await suspendK8sCronJob(clusterId.value, {
        namespace: namespace.value,
        name: resourceName.value
      });
      message.success(resource.value?.suspend ? '恢复成功' : '暂停成功');
      await loadData();
    } catch (error: unknown) {
      const err = error as Error;
      message.error(err.message || '操作失败');
    }
  }

  async function handleDelete() {
    try {
      await ElMessageBox.confirm(`确定要删除 CronJob "${resourceName.value}" 吗？此操作不可恢复！`, '确认删除', {
        type: 'warning'
      });

      await deleteK8sCronJob(clusterId.value, {
        namespace: namespace.value,
        name: resourceName.value
      });

      message.success('删除成功');
      router.push(backPath);
    } catch (error: unknown) {
      if (error !== 'cancel') {
        const err = error as Error;
        message.error(err.message || '删除失败');
      }
    }
  }

  // YAML 编辑保存：统一走 useYamlEdit（YAML 校验 → 更新 API → 重载详情）
  const { yamlSaving, handleYamlApply } = useYamlEdit({
    getClusterId: () => clusterId.value,
    getNamespace: () => namespace.value,
    updateFn: updateK8sCronJob,
    reload: loadData,
    successMessage: '更新成功'
  });

  async function handleTabChange(tab: string) {
    activeTab.value = tab;
    if (tab === 'pods' && pods.value.length === 0) {
      await loadPods();
    } else if (tab === 'events' && events.value.length === 0) {
      await loadEvents();
    }
  }

  onMounted(() => {
    loadData();
    loadPods();
    loadEvents();
  });
</script>

<template>
  <div class="detail-page">
    <div class="page-header">
      <div class="header-left">
        <ElButton :icon="ArrowLeft" @click="goBack">返回</ElButton>
        <h2 class="title">{{ resourceName }}</h2>
        <ElTag :type="resource?.suspend ? 'warning' : 'success'">
          {{ resource?.suspend ? '已暂停' : '运行中' }}
        </ElTag>
      </div>
      <div class="header-actions">
        <PermissionButton code="k8s.resource.update" type="primary" @click="handleToggleSuspend">
          {{ resource?.suspend ? '恢复' : '暂停' }}
        </PermissionButton>
        <PermissionButton code="k8s.resource.update" type="primary" @click="showYamlEditor = true">
          编辑YAML
        </PermissionButton>
        <PermissionButton code="k8s.resource.delete" type="danger" @click="handleDelete">删除</PermissionButton>
      </div>
    </div>

    <ElTabs v-model="activeTab" class="detail-tabs" @tab-change="handleTabChange">
      <ElTabPane label="基本信息" name="basic">
        <ElDescriptions :column="2" border>
          <ElDescriptionsItem label="名称">{{ resource?.name }}</ElDescriptionsItem>
          <ElDescriptionsItem label="命名空间">{{ resource?.namespace }}</ElDescriptionsItem>
          <ElDescriptionsItem label="调度规则">{{ resource?.schedule }}</ElDescriptionsItem>
          <ElDescriptionsItem label="状态">{{ resource?.suspend ? '已暂停' : '运行中' }}</ElDescriptionsItem>
          <ElDescriptionsItem label="活跃Job数">{{ resource?.active }}</ElDescriptionsItem>
          <ElDescriptionsItem label="上次执行">{{ resource?.lastSchedule || '-' }}</ElDescriptionsItem>
          <ElDescriptionsItem label="创建时间">{{ resource?.age }}</ElDescriptionsItem>
          <ElDescriptionsItem v-if="resource?.labels && Object.keys(resource?.labels).length > 0" label="标签">
            <div class="vertical-tags-list">
              <ElTag
                v-for="(tag, idx) in formatLabels(resource?.labels)"
                :key="idx"
                :type="getTagType(tag.key)"
                size="small"
                class="vertical-tag-item"
              >
                {{ tag.key }}: {{ tag.value }}
              </ElTag>
            </div>
          </ElDescriptionsItem>
        </ElDescriptions>
      </ElTabPane>

      <ElTabPane :label="`容器组 (${pods.length})`" name="pods">
        <template #label>
          <div class="tab-pane-header">
            <span>容器组</span>
            <ElTag size="small" class="count-tag">{{ pods.length }}</ElTag>
          </div>
        </template>
        <div class="tab-content">
          <K8sPodsTable :pods="pods" :loading="podsLoading" :cluster-id="clusterId" :namespace="namespace" />
        </div>
      </ElTabPane>

      <ElTabPane :label="`事件 (${events.length})`" name="events">
        <template #label>
          <div class="tab-pane-header">
            <span>事件</span>
            <ElTag size="small" class="count-tag">{{ events.length }}</ElTag>
          </div>
        </template>
        <div class="tab-content">
          <K8sEventsTable :events="events" :loading="eventsLoading" />
        </div>
      </ElTabPane>
    </ElTabs>

    <!-- YAML编辑器 -->
    <YamlEditor
      v-model="showYamlEditor"
      :title="`编辑 ${resourceName}`"
      :yaml="yamlContent"
      :can-edit="true"
      :on-apply="handleYamlApply"
    />
  </div>
</template>

<style scoped>
  .detail-page {
    padding: 20px;
  }

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .title {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
  }

  .header-actions {
    display: flex;
    gap: 8px;
  }

  /* 标签页样式 - 优化与tab的衔接 */
  .detail-tabs {
    background: var(--el-bg-color);
    border-radius: 8px;
    padding: 16px;
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

  /* 标签内容容器 - 移除多余的padding */
  .tab-content {
    padding: 12px 0;
  }

  .yaml-viewer {
    background: #1e1e1e;
    border-radius: 4px;
    padding: 16px;
    max-height: 600px;
    overflow: auto;
  }

  .yaml-viewer pre {
    margin: 0;
    color: #d4d4d4;
    font-family: Monaco, Menlo, monospace;
    font-size: 13px;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-all;
  }

  /* 垂直标签列表样式 */
  .vertical-tags-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
    align-items: flex-start;
    width: 100%;
  }

  .vertical-tag-item {
    display: flex !important;
    align-items: center;
    width: 100%;
    font-family: 'Courier New', Courier, monospace;
    margin: 0;
    padding: 4px 8px !important;
    border: none !important;
  }
</style>
