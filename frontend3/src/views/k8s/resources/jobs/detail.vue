<script setup lang="ts">
    import { computed, onMounted, ref } from 'vue';
    import { useRoute, useRouter } from 'vue-router';
    import { ElMessage, ElMessageBox } from 'element-plus';
    import { ArrowLeft } from '@element-plus/icons-vue';
    import yaml from 'js-yaml';
    import { deleteK8sJob, fetchK8sEvents, getK8sJob, getK8sJobPods, updateK8sJob } from '@/service/api/k8s';
    import YamlEditor from '@/components/YamlEditor.vue';
    import K8sPodsTable from '@/components/k8s/K8sPodsTable.vue';

    defineOptions({ name: 'K8sJobDetail' });

    const route = useRoute();
    const router = useRouter();
    const message = ElMessage;

    const clusterId = computed(() => Number(route.query.clusterId));
    const namespace = computed(() => route.query.namespace as string);
    const resourceName = computed(() => route.query.name as string);

    const resource = ref<K8s.Job | null>(null);
    const pods = ref<K8s.Pod[]>([]);
    const events = ref<K8s.Event[]>([]);
    const loading = ref(true);
    const podsLoading = ref(false);
    const eventsLoading = ref(false);
    const activeTab = ref('basic');
    const showYamlEditor = ref(false);
    const yamlContent = ref('');
    const yamlSaving = ref(false);

    const statusType = computed(() => {
      if (!resource.value) return 'info';
      return 'success';
    });

    const statusText = computed(() => {
      return 'Job';
    });

    // 解析manifest为YAML
    function parseManifest(manifestStr: string): string {
      if (!manifestStr) return '';
      try {
        const obj = JSON.parse(manifestStr);
        delete obj.managedFields;
        return yaml.dump(obj, { indent: 2, lineWidth: 120, noRefs: true });
      } catch (e) {
        return manifestStr;
      }
    }

    async function loadData() {
      loading.value = true;
      try {
        const response = await getK8sJob(clusterId.value, namespace.value, resourceName.value);
        const data = response.data || response;
        resource.value = data;
        yamlContent.value = parseManifest(data.manifest || '');
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
        const response = await fetchK8sEvents(
          clusterId.value,
          namespace.value,
          `involvedObject.name=${resourceName.value}`
        );
        events.value = response.data || response || [];
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
        const response = await getK8sJobPods(clusterId.value, namespace.value, resourceName.value);
        pods.value = response.data || response || [];
      } catch (error: unknown) {
        const err = error as Error;
        message.error(err.message || '加载容器组失败');
      } finally {
        podsLoading.value = false;
      }
    }

    // 返回列表页的路径 - 直接返回工作负载汇总页，状态由 store 管理
    const backPath = '/k8s/workloads';

    function goBack() {
      router.push(backPath);
    }

    async function handleDelete() {
      try {
        await ElMessageBox.confirm(`确定要删除 Job "${resourceName.value}" 吗？此操作不可恢复！`, '确认删除', {
          type: 'warning'
        });

        await deleteK8sJob(clusterId.value, {
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

    async function handleYamlApply(yamlStr: string) {
      yamlSaving.value = true;
      try {
        const manifest = yaml.load(yamlStr);
        await updateK8sJob(clusterId.value, {
          namespace: namespace.value,
          manifest
        });
        message.success('更新成功');
        await loadData();
      } catch (error: unknown) {
        const err = error as Error;
        throw new Error(err.message || '更新失败');
      } finally {
        yamlSaving.value = false;
      }
    }

    async function handleTabChange(tab: string) {
      activeTab.value = tab;
      if (tab === 'pods' && pods.value.length === 0) {
        await loadPods();
      } else if (tab === 'events' && events.value.length === 0) {
        await loadEvents();
      }
    }

    // 获取标签颜色
    const getTagType = (key: string) => {
      const keyLower = key.toLowerCase();
      if (keyLower.includes('app') || keyLower.includes('name')) return 'primary';
      if (keyLower.includes('env') || keyLower.includes('environment')) return 'success';
      if (keyLower.includes('version') || keyLower.includes('ver')) return 'warning';
      if (keyLower.includes('component')) return 'info';
      return 'default';
    };

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
                <ElTag :type="statusType">{{ statusText }}</ElTag>
            </div>
            <div class="header-actions">
                <ElButton type="primary" @click="showYamlEditor = true">编辑YAML</ElButton>
                <ElButton type="danger" @click="handleDelete">删除</ElButton>
            </div>
        </div>

        <ElTabs v-model="activeTab" class="detail-tabs" @tab-change="handleTabChange">
            <ElTabPane label="基本信息" name="basic">
                <ElDescriptions :column="2" border>
                    <ElDescriptionsItem label="名称">{{ resource?.name }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="命名空间">{{ resource?.namespace }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="创建时间">{{ resource?.age }}</ElDescriptionsItem>
                    <ElDescriptionsItem v-if="resource?.labels && Object.keys(resource?.labels).length > 0" label="标签">
                        <div class="vertical-tags-list">
                            <ElTag v-for="(tag, idx) in formatLabels(resource?.labels)" :key="idx" :type="getTagType(tag.key)" size="small" class="vertical-tag-item">
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
                    <ElTable v-loading="eventsLoading" :data="events" class="jobs-events-table" :header-cell-style="{
              background: '#f5f7fa',
              color: '#303133',
              fontWeight: '600',
              paddingLeft: '16px',
              paddingRight: '16px'
            }" :row-style="{ backgroundColor: 'transparent' }" :cell-style="{ backgroundColor: 'transparent', padding: '8px 16px' }">
                        <ElTableColumn type="index" label="序号" width="60" />
                        <ElTableColumn prop="type" label="类型" width="100" />
                        <ElTableColumn prop="reason" label="原因" width="150" />
                        <ElTableColumn prop="message" label="消息" />
                        <ElTableColumn prop="count" label="次数" width="80" />
                        <ElTableColumn prop="lastTimestamp" label="最后时间" width="180" />
                    </ElTable>
                </div>
            </ElTabPane>
        </ElTabs>

        <!-- YAML编辑器 -->
        <YamlEditor v-model="showYamlEditor" :title="`编辑 ${resourceName}`" :yaml="yamlContent" :can-edit="true" :on-apply="handleYamlApply" />
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
      font-family: 'Monaco', 'Menlo', monospace;
      font-size: 13px;
      line-height: 1.6;
      white-space: pre-wrap;
      word-break: break-all;
    }

    /* 表格透明样式 - 完全覆盖 Element Plus 默认样式 */
    .jobs-events-table,
    .jobs-events-table.el-table,
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
