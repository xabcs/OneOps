<script setup lang="ts">
  import { computed, onMounted, ref } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { ElMessage, ElMessageBox } from 'element-plus';
  import { ArrowLeft } from '@element-plus/icons-vue';
  import { deleteK8sConfigMap, fetchK8sEvents, getK8sConfigMap, updateK8sConfigMap } from '@/service/api/k8s';
  import { parseManifest } from '@/views/k8s/shared/k8s-formatters';
  import { useYamlEdit } from '@/views/k8s/composables/useYamlEdit';
  import YamlEditor from '@/components/k8s/YamlEditor.vue';

  defineOptions({ name: 'K8sConfigMapDetail' });

  const route = useRoute();
  const router = useRouter();
  const message = ElMessage;

  const clusterId = computed(() => Number(route.query.clusterId));
  const namespace = computed(() => route.query.namespace as string);
  const resourceName = computed(() => route.query.name as string);

  const resource = ref<K8s.ConfigMap | null>(null);
  const events = ref<K8s.Event[]>([]);
  const dataKeys = ref<{ key: string; value: string; editValue: string; editing: boolean; originalValue: string }[]>(
    []
  );
  const loading = ref(true);
  const eventsLoading = ref(false);
  const activeTab = ref('basic');
  const showYamlEditor = ref(false);
  const yamlContent = ref('');

  // 解析数据项
  function parseDataKeys(data: Record<string, string>) {
    if (!data) return [];
    return Object.entries(data).map(([key, value]) => ({
      key,
      value,
      editValue: value,
      editing: false,
      originalValue: value
    }));
  }

  async function loadData() {
    loading.value = true;
    try {
      const { data, error } = await getK8sConfigMap(clusterId.value, namespace.value, resourceName.value);
      if (!error && data) {
        resource.value = data;
        yamlContent.value = parseManifest(data.manifest || '');
        dataKeys.value = parseDataKeys(data.data || {});
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

  function goBack() {
    router.back();
  }

  function cancelEdit(row: { editing: boolean; editValue: string; originalValue: string }, _index: number) {
    row.editing = false;
    row.editValue = row.originalValue;
  }

  function saveDataItem(row: { editing: boolean; editValue: string; originalValue: string }) {
    message.info('数据项编辑功能需要实现 patch 逻辑');
    row.editing = false;
    row.originalValue = row.editValue;
  }

  async function handleDelete() {
    try {
      await ElMessageBox.confirm(`确定要删除 ConfigMap "${resourceName.value}" 吗？`, '确认删除', { type: 'warning' });

      await deleteK8sConfigMap(clusterId.value, {
        namespace: namespace.value,
        name: resourceName.value
      });

      message.success('删除成功');
      router.back();
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
    updateFn: updateK8sConfigMap,
    reload: loadData,
    successMessage: '更新成功'
  });

  onMounted(() => {
    loadData();
    loadEvents();
  });
</script>

<template>
  <div class="detail-page">
    <div class="page-header">
      <div class="header-left">
        <ElButton :icon="ArrowLeft" @click="goBack">返回</ElButton>
        <h2 class="title">{{ resourceName }}</h2>
        <ElTag type="info">ConfigMap</ElTag>
      </div>
      <div class="header-actions">
        <PermissionButton code="k8s.resource.update" type="primary" @click="showYamlEditor = true">
          编辑YAML
        </PermissionButton>
        <PermissionButton code="k8s.resource.delete" type="danger" @click="handleDelete">删除</PermissionButton>
      </div>
    </div>

    <ElTabs v-model="activeTab" class="detail-tabs">
      <ElTabPane label="基本信息" name="basic">
        <ElDescriptions :column="2" border>
          <ElDescriptionsItem label="名称">{{ resource?.name }}</ElDescriptionsItem>
          <ElDescriptionsItem label="命名空间">{{ resource?.namespace }}</ElDescriptionsItem>
          <ElDescriptionsItem label="创建时间">{{ resource?.age }}</ElDescriptionsItem>
          <ElDescriptionsItem label="标签">
            <ElTag v-for="(value, key) in resource?.labels" :key="key" size="small" style="margin-right: 4px">
              {{ key }}: {{ value }}
            </ElTag>
          </ElDescriptionsItem>
        </ElDescriptions>

        <h4 style="margin-top: 20px">数据项</h4>
        <ElTable :data="dataKeys" border>
          <ElTableColumn type="index" label="序号" width="60" />
          <ElTableColumn prop="key" label="键" />
          <ElTableColumn prop="value" label="值">
            <template #default="{ row }">
              <ElInput v-if="row.editing" v-model="row.editValue" type="textarea" :rows="3" />
              <span v-else class="value-cell">{{ row.value }}</span>
            </template>
          </ElTableColumn>
          <ElTableColumn label="操作" align="center" width="120" class-name="msre-table-actions">
            <template #default="{ row, $index }">
              <ElButton link type="primary" size="small" v-if="!row.editing" @click="row.editing = true">编辑</ElButton>
              <ElButton link type="primary" size="small" v-if="row.editing" @click="saveDataItem(row)">保存</ElButton>
              <ElButton link type="primary" size="small" v-if="row.editing" @click="cancelEdit(row, $index)">
                取消
              </ElButton>
            </template>
          </ElTableColumn>
        </ElTable>
      </ElTabPane>

      <ElTabPane label="YAML" name="yaml">
        <div class="yaml-viewer">
          <pre>{{ yamlContent }}</pre>
        </div>
      </ElTabPane>

      <ElTabPane label="事件" name="events">
        <ElTable v-loading="eventsLoading" :data="events">
          <ElTableColumn type="index" label="序号" width="60" />
          <ElTableColumn prop="type" label="类型" width="100" />
          <ElTableColumn prop="reason" label="原因" width="150" />
          <ElTableColumn prop="message" label="消息" />
          <ElTableColumn prop="count" label="次数" width="80" />
          <ElTableColumn prop="lastTimestamp" label="最后时间" width="180" />
        </ElTable>
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

  .detail-tabs {
    background: var(--el-bg-color);
    border-radius: 8px;
    padding: 16px;
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

  .value-cell {
    word-break: break-all;
    font-family: monospace;
    font-size: 12px;
  }
</style>
