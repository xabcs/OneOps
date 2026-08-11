<script setup lang="ts">
    import { computed, onMounted, ref } from 'vue';
    import { useRoute, useRouter } from 'vue-router';
    import { ElMessage, ElMessageBox } from 'element-plus';
    import { ArrowLeft } from '@element-plus/icons-vue';
    import yaml from 'js-yaml';
    import { deleteK8sSecret, fetchK8sEvents, getK8sSecret, updateK8sSecret } from '@/service/api/k8s';
    import YamlEditor from '@/components/YamlEditor.vue';

    defineOptions({ name: 'K8sSecretDetail' });

    const route = useRoute();
    const router = useRouter();
    const message = ElMessage;

    const clusterId = computed(() => Number(route.query.clusterId));
    const namespace = computed(() => route.query.namespace as string);
    const resourceName = computed(() => route.query.name as string);

    const resource = ref<K8s.Secret | null>(null);
    const events = ref<K8s.Event[]>([]);
    const dataKeys = ref<{ key: string; value: string; displayValue: string; isOpaque: boolean; type: string }[]>([]);
    const loading = ref(true);
    const eventsLoading = ref(false);
    const activeTab = ref('basic');
    const showYamlEditor = ref(false);
    const yamlContent = ref('');
    const yamlSaving = ref(false);

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

    // 解析数据项
    function parseDataKeys(data: Record<string, string>, type: string) {
      if (!data) return [];
      const isOpaque = type === 'Opaque';
      return Object.keys(data).map(key => {
        const value = data[key];
        // 如果是Opaque类型，尝试base64解码显示
        let displayValue = value;
        if (!isOpaque) {
          try {
            displayValue = atob(value);
          } catch {
            displayValue = value;
          }
        }
        return {
          key,
          value,
          displayValue,
          isOpaque,
          type: isOpaque ? 'Opaque' : type
        };
      });
    }

    async function loadData() {
      loading.value = true;
      try {
        const response = await getK8sSecret(clusterId.value, namespace.value, resourceName.value);
        const data = response.data || response;
        resource.value = data;
        yamlContent.value = parseManifest(data.manifest || '');
        dataKeys.value = parseDataKeys(data.data || {}, data.type || 'Opaque');
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

    function goBack() {
      router.back();
    }

    async function handleDelete() {
      try {
        await ElMessageBox.confirm(`确定要删除 Secret "${resourceName.value}" 吗？`, '确认删除', { type: 'warning' });

        await deleteK8sSecret(clusterId.value, {
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

    async function handleYamlApply(yamlStr: string) {
      yamlSaving.value = true;
      try {
        const manifest = yaml.load(yamlStr);
        await updateK8sSecret(clusterId.value, {
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
                <ElTag type="info">Secret</ElTag>
            </div>
            <div class="header-actions">
                <ElButton type="primary" @click="showYamlEditor = true">编辑YAML</ElButton>
                <ElButton type="danger" @click="handleDelete">删除</ElButton>
            </div>
        </div>

        <ElTabs v-model="activeTab" class="detail-tabs">
            <ElTabPane label="基本信息" name="basic">
                <ElDescriptions :column="2" border>
                    <ElDescriptionsItem label="名称">{{ resource?.name }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="命名空间">{{ resource?.namespace }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="类型">{{ resource?.type }}</ElDescriptionsItem>
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
                            <span v-if="row.isOpaque" class="opaque-value">OPAQUE</span>
                            <span v-else class="value-cell">{{ row.displayValue }}</span>
                        </template>
                    </ElTableColumn>
                    <ElTableColumn prop="type" label="类型" width="100" />
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
      font-family: 'Monaco', 'Menlo', monospace;
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

    .opaque-value {
      color: #909399;
      font-style: italic;
    }
</style>
