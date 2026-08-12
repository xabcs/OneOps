<script setup lang="ts">
  import { computed, onMounted, ref, watch } from 'vue';
  import { useRoute } from 'vue-router';
  import {
    CircleCheck,
    CircleClose,
    Clock,
    CopyDocument,
    Download,
    Refresh,
    Search,
    VideoPlay,
    Warning
  } from '@element-plus/icons-vue';
  import { fetchK8sClusterNamespaces, fetchK8sClusters } from '@/service/api/k8s';
  import {
    executeDiagnostic,
    fetchDiagnosticCommands,
    fetchDiagnosticHistory,
    fetchJavaPods
  } from '@/service/api/diagnostic';
  import HistoryDialog from './modules/HistoryDialog.vue';

  defineOptions({ name: 'K8sDiagnostic' });

  // 数据状态
  const loading = ref(false);
  const executing = ref(false);
  const historyDialogVisible = ref(false);

  // 过滤条件
  const filters = ref({
    clusterId: '',
    namespace: '',
    search: ''
  });

  // 数据列表
  const clusters = ref([]);
  const namespaces = ref([]);
  const pods = ref([]);
  const diagnosticCommands = ref([]);

  // 选择状态
  const selectedPod = ref(null);
  const selectedCommand = ref(null);
  const commandArgs = ref({});

  // 诊断结果
  const diagnosticResult = ref(null);
  const diagnosticHistory = ref([]);

  // 计算属性
  const diagnosablePods = computed(() => pods.value.filter(pod => pod.hasAgent));

  const filteredPods = computed(() => {
    let result = pods.value;

    if (filters.value.search) {
      const search = filters.value.search.toLowerCase();
      result = result.filter(
        pod => pod.podName.toLowerCase().includes(search) || pod.namespace.toLowerCase().includes(search)
      );
    }

    return result;
  });

  const hasHistory = computed(() => diagnosticHistory.value.length > 0);

  // 生命周期
  onMounted(() => {
    initializeData();
  });

  // 方法
  const initializeData = async () => {
    await Promise.all([loadClusters(), loadNamespaces(), loadDiagnosticCommands()]);

    // 从URL参数获取初始值
    const route = useRoute();
    if (route.query.cluster) {
      filters.value.clusterId = route.query.cluster as string;
    }
    if (route.query.namespace) {
      filters.value.namespace = route.query.namespace as string;
    }

    // 加载Pod列表
    await loadPods();
  };

  const loadClusters = async () => {
    try {
      const response = await fetchK8sClusters();
      clusters.value = response.data || [];

      // 自动选择第一个集群
      if (!filters.value.clusterId && clusters.value.length > 0) {
        filters.value.clusterId = clusters.value[0].id;
      }
    } catch (error) {
      console.error('加载集群失败:', error);
    }
  };

  const loadNamespaces = async () => {
    if (!filters.value.clusterId) return;

    try {
      const response = await fetchK8sClusterNamespaces(filters.value.clusterId);
      namespaces.value = response.data || [];

      // 自动选择default命名空间
      if (!filters.value.namespace && namespaces.value.length > 0) {
        const defaultNs = namespaces.value.find(ns => ns.name === 'default');
        filters.value.namespace = defaultNs ? defaultNs.name : namespaces.value[0].name;
      }
    } catch (error) {
      console.error('加载命名空间失败:', error);
    }
  };

  const loadPods = async () => {
    if (!filters.value.clusterId || !filters.value.namespace) return;

    loading.value = true;
    try {
      const response = await fetchJavaPods(filters.value.clusterId, filters.value.namespace);
      pods.value = response.data || [];
    } catch (error) {
      console.error('加载Pod列表失败:', error);
    } finally {
      loading.value = false;
    }
  };

  const loadDiagnosticCommands = async () => {
    try {
      const response = await fetchDiagnosticCommands();
      diagnosticCommands.value = response.data.commands || [];
    } catch (error) {
      console.error('加载诊断命令失败:', error);
    }
  };

  // 事件处理
  const onClusterChange = () => {
    filters.value.namespace = '';
    selectedPod.value = null;
    loadNamespaces();
    loadPods();
  };

  const onNamespaceChange = () => {
    selectedPod.value = null;
    loadPods();
  };

  const onSearchChange = () => {
    // 搜索由计算属性自动处理
  };

  const refreshPods = () => {
    loadPods();
  };

  const selectPod = pod => {
    selectedPod.value = pod;
    diagnosticResult.value = null;
  };

  const selectCommand = command => {
    selectedCommand.value = command;

    // 初始化参数默认值
    commandArgs.value = {};
    command.args.forEach(arg => {
      commandArgs.value[arg.name] = arg.defaultValue;
    });
  };

  const executeDiagnostic = async () => {
    if (!selectedPod.value || !selectedCommand.value) return;

    executing.value = true;
    try {
      const response = await executeDiagnostic({
        clusterId: filters.value.clusterId,
        namespace: selectedPod.value.namespace,
        podName: selectedPod.value.podName,
        command: selectedCommand.value.id,
        args: commandArgs.value,
        timeout: 60
      });

      diagnosticResult.value = response.data;
    } catch (error) {
      console.error('执行诊断失败:', error);
      diagnosticResult.value = {
        status: 'error',
        error: error.message || '执行诊断失败',
        output: '',
        timestamp: Date.now(),
        duration: 0
      };
    } finally {
      executing.value = false;
    }
  };

  const showHistory = async () => {
    try {
      const response = await fetchDiagnosticHistory({
        clusterId: filters.value.clusterId,
        namespace: selectedPod.value.namespace,
        podName: selectedPod.value.podName
      });
      diagnosticHistory.value = response.data.list || [];
      historyDialogVisible.value = true;
    } catch (error) {
      console.error('加载历史记录失败:', error);
    }
  };

  // 辅助方法
  const getPodStatusColor = status => {
    const colors = {
      Running: 'success',
      Pending: 'warning',
      Failed: 'danger',
      Succeeded: 'info'
    };
    return colors[status] || 'info';
  };

  const getPodStatusIcon = status => {
    const icons = {
      Running: CircleCheck,
      Pending: Warning,
      Failed: CircleClose,
      Succeeded: CircleCheck
    };
    return icons[status] || Warning;
  };

  const getPodStatusType = status => {
    const types = {
      Running: 'success',
      Pending: 'warning',
      Failed: 'danger',
      Succeeded: 'info'
    };
    return types[status] || 'info';
  };

  const getCommandIcon = category => {
    // 根据类别返回不同图标
    const icons = {
      performance: 'TrendCharts',
      memory: 'Coin',
      system: 'Setting',
      monitoring: 'Monitor',
      logging: 'Document'
    };
    return icons[category] || 'Grid';
  };

  const copyResult = () => {
    if (diagnosticResult.value?.output) {
      navigator.clipboard.writeText(diagnosticResult.value.output);
      // 显示复制成功提示
    }
  };

  const downloadResult = () => {
    if (diagnosticResult.value?.output) {
      const blob = new Blob([diagnosticResult.value.output], { type: 'text/plain' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `diagnostic-${selectedPod.value.podName}-${Date.now()}.txt`;
      a.click();
      URL.revokeObjectURL(url);
    }
  };

  const formatTime = timestamp => {
    return new Date(timestamp).toLocaleString('zh-CN');
  };

  // 监听过滤器变化
  watch(
    () => filters.value.clusterId,
    () => {
      if (filters.value.clusterId) {
        loadNamespaces();
        loadPods();
      }
    }
  );

  watch(
    () => filters.value.namespace,
    () => {
      if (filters.value.namespace) {
        loadPods();
      }
    }
  );
</script>

<template>
  <div class="diagnostic-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <h1>K8s 诊断中心</h1>
      <p class="description">对集群中的Java应用进行实时诊断分析</p>
    </div>

    <!-- 集群和命名空间选择 -->
    <div class="filter-section">
      <ElForm :inline="true" :model="filters" class="filter-form">
        <ElFormItem label="集群">
          <ElSelect v-model="filters.clusterId" placeholder="选择集群" style="width: 200px" @change="onClusterChange">
            <ElOption v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
          </ElSelect>
        </ElFormItem>

        <ElFormItem label="命名空间">
          <ElSelect
            v-model="filters.namespace"
            placeholder="选择命名空间"
            style="width: 180px"
            @change="onNamespaceChange"
          >
            <ElOption v-for="ns in namespaces" :key="ns.name" :label="ns.name" :value="ns.name" />
          </ElSelect>
        </ElFormItem>

        <ElFormItem label="应用过滤">
          <ElInput
            v-model="filters.search"
            placeholder="搜索应用名称"
            style="width: 200px"
            clearable
            @input="onSearchChange"
          >
            <template #prefix>
              <ElIcon><Search /></ElIcon>
            </template>
          </ElInput>
        </ElFormItem>

        <ElFormItem>
          <ElButton type="primary" :icon="Refresh" :loading="loading" @click="refreshPods">刷新</ElButton>
        </ElFormItem>
      </ElForm>
    </div>

    <!-- 主要内容区域 -->
    <div class="main-content">
      <!-- 左侧：Pod列表 -->
      <div class="pod-list-section">
        <div class="section-header">
          <h3>Java应用列表</h3>
          <ElBadge :value="diagnosablePods.length" :max="99" class="badge">
            <ElTag size="small">可诊断</ElTag>
          </ElBadge>
        </div>

        <!-- Pod列表 -->
        <div v-loading="loading" class="pod-list">
          <div
            v-for="pod in filteredPods"
            :key="pod.podName"
            class="pod-item"
            :class="{ active: selectedPod?.podName === pod.podName }"
            @click="selectPod(pod)"
          >
            <div class="pod-status">
              <ElIcon :color="getPodStatusColor(pod.phase)" :size="12">
                <component :is="getPodStatusIcon(pod.phase)" />
              </ElIcon>
            </div>

            <div class="pod-info">
              <div class="pod-name">{{ pod.podName }}</div>
              <div class="pod-meta">
                <span class="pod-namespace">{{ pod.namespace }}</span>
                <ElDivider direction="vertical" />
                <span class="pod-ip">{{ pod.ip }}</span>
              </div>
            </div>

            <div class="pod-diagnostic">
              <ElTag v-if="pod.hasAgent" type="success" size="small" effect="plain">可诊断</ElTag>
              <ElTag v-else type="info" size="small" effect="plain">不可诊断</ElTag>
            </div>
          </div>

          <ElEmpty v-if="!loading && filteredPods.length === 0" description="暂无可诊断的Java应用" :image-size="100" />
        </div>
      </div>

      <!-- 右侧：诊断面板 -->
      <div class="diagnostic-panel-section">
        <div v-if="!selectedPod" class="empty-state">
          <ElEmpty description="请选择一个Pod进行诊断" />
        </div>

        <div v-else class="diagnostic-panel">
          <!-- Pod信息卡片 -->
          <div class="pod-info-card">
            <div class="pod-header">
              <h3>{{ selectedPod.podName }}</h3>
              <ElTag :type="getPodStatusType(selectedPod.phase)">
                {{ selectedPod.phase }}
              </ElTag>
            </div>
            <div class="pod-details">
              <ElDescriptions :column="2" size="small">
                <ElDescriptionsItem label="命名空间">
                  {{ selectedPod.namespace }}
                </ElDescriptionsItem>
                <ElDescriptionsItem label="Pod IP">
                  {{ selectedPod.ip }}
                </ElDescriptionsItem>
                <ElDescriptionsItem label="节点">
                  {{ selectedPod.nodeName }}
                </ElDescriptionsItem>
                <ElDescriptionsItem label="诊断方式">
                  <ElTag v-if="selectedPod.hasAgent" type="success" size="small">可诊断</ElTag>
                  <ElTag v-else type="info" size="small">不可用</ElTag>
                </ElDescriptionsItem>
              </ElDescriptions>
            </div>
          </div>

          <!-- 诊断命令选择 -->
          <div class="command-selector">
            <h4>选择诊断命令</h4>
            <ElRow :gutter="12">
              <ElCol v-for="cmd in diagnosticCommands" :key="cmd.id" :span="8">
                <ElCard
                  class="command-card"
                  :class="{ active: selectedCommand?.id === cmd.id }"
                  shadow="hover"
                  @click="selectCommand(cmd)"
                >
                  <div class="command-header">
                    <ElIcon class="command-icon">
                      <component :is="getCommandIcon(cmd.category)" />
                    </ElIcon>
                    <span class="command-name">{{ cmd.name }}</span>
                  </div>
                  <div class="command-description">
                    {{ cmd.description }}
                  </div>
                </ElCard>
              </ElCol>
            </ElRow>
          </div>

          <!-- 命令参数配置 -->
          <div v-if="selectedCommand" class="command-config">
            <h4>参数配置</h4>
            <ElForm ref="commandFormRef" :model="commandArgs" label-width="120px" size="small">
              <ElFormItem v-for="arg in selectedCommand.args" :key="arg.name" :label="arg.name">
                <ElInput v-if="arg.name === 'level'" v-model="commandArgs[arg.name]" placeholder="请选择日志级别">
                  <template #suffix>
                    <ElSelect v-model="commandArgs[arg.name]" placeholder="选择级别" style="width: 100px">
                      <ElOption label="DEBUG" value="DEBUG" />
                      <ElOption label="INFO" value="INFO" />
                      <ElOption label="WARN" value="WARN" />
                      <ElOption label="ERROR" value="ERROR" />
                    </ElSelect>
                  </template>
                </ElInput>

                <ElSwitch v-else-if="typeof arg.defaultValue === 'boolean'" v-model="commandArgs[arg.name]" />

                <ElInput v-else v-model="commandArgs[arg.name]" :placeholder="arg.description">
                  <template v-if="arg.defaultValue" #append>
                    <ElButton @click="commandArgs[arg.name] = arg.defaultValue">重置</ElButton>
                  </template>
                </ElInput>

                <div class="arg-description">{{ arg.description }}</div>
              </ElFormItem>
            </ElForm>

            <div class="action-buttons">
              <ElButton
                type="primary"
                :icon="VideoPlay"
                :loading="executing"
                :disabled="!selectedPod.hasAgent"
                @click="executeDiagnostic"
              >
                执行诊断
              </ElButton>

              <ElButton :icon="Clock" :disabled="!hasHistory" @click="showHistory">历史记录</ElButton>
            </div>
          </div>

          <!-- 诊断结果展示 -->
          <div v-if="diagnosticResult" class="result-section">
            <div class="result-header">
              <h4>诊断结果</h4>
              <div class="result-actions">
                <ElButton :icon="CopyDocument" size="small" @click="copyResult">复制</ElButton>
                <ElButton :icon="Download" size="small" @click="downloadResult">下载</ElButton>
              </div>
            </div>

            <div class="result-content">
              <ElAlert v-if="diagnosticResult.status === 'error'" type="error" :closable="false" show-icon>
                {{ diagnosticResult.error }}
              </ElAlert>

              <pre v-else class="result-output">{{ diagnosticResult.output }}</pre>
            </div>

            <div class="result-meta">
              <ElTag size="small">耗时: {{ diagnosticResult.duration }}ms</ElTag>
              <ElTag size="small" type="info">时间: {{ formatTime(diagnosticResult.timestamp) }}</ElTag>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 历史记录对话框 -->
    <HistoryDialog v-model:visible="historyDialogVisible" :history="diagnosticHistory" />
  </div>
</template>

<style scoped lang="scss">
  .diagnostic-page {
    padding: 20px;

    .page-header {
      margin-bottom: 20px;

      h1 {
        font-size: 24px;
        margin: 0 0 8px 0;
      }

      .description {
        color: #606266;
        margin: 0;
      }
    }

    .filter-section {
      background: #fff;
      padding: 16px;
      border-radius: 4px;
      margin-bottom: 20px;
      box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    }

    .main-content {
      display: flex;
      gap: 20px;
      height: calc(100vh - 200px);

      .pod-list-section {
        width: 350px;
        background: #fff;
        border-radius: 4px;
        box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        display: flex;
        flex-direction: column;

        .section-header {
          padding: 16px;
          border-bottom: 1px solid #ebeef5;
          display: flex;
          justify-content: space-between;
          align-items: center;

          h3 {
            margin: 0;
            font-size: 16px;
          }
        }

        .pod-list {
          flex: 1;
          overflow-y: auto;
          padding: 8px;

          .pod-item {
            display: flex;
            align-items: center;
            padding: 12px;
            margin-bottom: 8px;
            border: 1px solid #ebeef5;
            border-radius: 4px;
            cursor: pointer;
            transition: all 0.3s;

            &:hover {
              background-color: #f5f7fa;
              border-color: #409eff;
            }

            &.active {
              background-color: #ecf5ff;
              border-color: #409eff;
            }

            .pod-status {
              margin-right: 12px;
            }

            .pod-info {
              flex: 1;

              .pod-name {
                font-weight: 500;
                margin-bottom: 4px;
              }

              .pod-meta {
                font-size: 12px;
                color: #909399;

                .el-divider {
                  margin: 0 8px;
                }
              }
            }

            .pod-diagnostic {
              margin-left: 8px;
            }
          }
        }
      }

      .diagnostic-panel-section {
        flex: 1;
        background: #fff;
        border-radius: 4px;
        box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        overflow-y: auto;

        .empty-state {
          display: flex;
          align-items: center;
          justify-content: center;
          height: 100%;
        }

        .diagnostic-panel {
          padding: 20px;

          .pod-info-card {
            margin-bottom: 20px;
            padding: 16px;
            background: #f5f7fa;
            border-radius: 4px;

            .pod-header {
              display: flex;
              justify-content: space-between;
              align-items: center;
              margin-bottom: 12px;

              h3 {
                margin: 0;
                font-size: 16px;
              }
            }
          }

          .command-selector {
            margin-bottom: 20px;

            h4 {
              margin: 0 0 12px 0;
              font-size: 14px;
            }

            .command-card {
              margin-bottom: 12px;
              cursor: pointer;
              transition: all 0.3s;

              &:hover {
                transform: translateY(-2px);
              }

              &.active {
                border-color: #409eff;
                background-color: #ecf5ff;
              }

              .command-header {
                display: flex;
                align-items: center;
                margin-bottom: 8px;

                .command-icon {
                  margin-right: 8px;
                  color: #409eff;
                }

                .command-name {
                  font-weight: 500;
                }
              }

              .command-description {
                font-size: 12px;
                color: #606266;
              }
            }
          }

          .command-config {
            margin-bottom: 20px;
            padding: 16px;
            background: #f5f7fa;
            border-radius: 4px;

            h4 {
              margin: 0 0 12px 0;
              font-size: 14px;
            }

            .action-buttons {
              margin-top: 16px;
              display: flex;
              gap: 12px;
            }

            .arg-description {
              font-size: 12px;
              color: #909399;
              margin-top: 4px;
            }
          }

          .result-section {
            margin-top: 20px;

            .result-header {
              display: flex;
              justify-content: space-between;
              align-items: center;
              margin-bottom: 12px;

              h4 {
                margin: 0;
                font-size: 14px;
              }

              .result-actions {
                display: flex;
                gap: 8px;
              }
            }

            .result-content {
              margin-bottom: 12px;

              .result-output {
                background: #1e1e1e;
                color: #d4d4d4;
                padding: 16px;
                border-radius: 4px;
                max-height: 400px;
                overflow-y: auto;
                font-family: 'Courier New', monospace;
                font-size: 12px;
                line-height: 1.6;
              }
            }

            .result-meta {
              display: flex;
              gap: 8px;
            }
          }
        }
      }
    }

    .history-item {
      .history-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 8px;
      }

      .history-args {
        color: #606266;
        font-size: 14px;
        margin-bottom: 4px;
      }

      .history-user {
        color: #909399;
        font-size: 12px;
      }
    }
  }
</style>
