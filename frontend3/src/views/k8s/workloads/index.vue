<script setup lang="ts">
    import { computed } from 'vue';
    import { useRouter } from 'vue-router';
    import { ElButton, ElTabPane, ElTabs } from 'element-plus';
    import YamlEditor from '@/components/YamlEditor.vue';
    import { useWorkloadData } from './composables/useWorkloadData';
    import { useWorkloadActions } from './composables/useWorkloadActions';
    import PodTab from './components/PodTab.vue';
    import DeploymentTab from './components/DeploymentTab.vue';
    import WorkloadTab from './components/WorkloadTab.vue';
    import ScaleDialog from './components/ScaleDialog.vue';
    import LogDialog from './components/LogDialog.vue';

    defineOptions({ name: 'K8sWorkloads' });

    const router = useRouter();

    const {
      loading,
      activeTab,
      selectedCluster,
      selectedNamespace,
      namespaces,
      clusters,
      deploymentsData,
      podsData,
      statefulSetsData,
      daemonSetsData,
      jobsData,
      cronJobsData,
      deploymentsPagination,
      podsPagination,
      statefulSetsPagination,
      daemonSetsPagination,
      jobsPagination,
      cronJobsPagination,
      loadNamespaces,
      loadCurrentData,
      loadDeployments,
      loadPods,
      loadStatefulSets,
      loadDaemonSets,
      loadJobs,
      loadCronJobs,
      handleTabChange,
      handlePageSizeChange
    } = useWorkloadData();

    const actions = useWorkloadActions(
      () => selectedCluster.value,
      () => activeTab.value
    );
    actions.setLoadFunctions({
      deployments: loadDeployments,
      pods: loadPods,
      statefulsets: loadStatefulSets,
      daemonsets: loadDaemonSets,
      jobs: loadJobs,
      cronjobs: loadCronJobs
    });

    const showBatchOperations = computed(() => {
      const map: Record<string, K8s.WorkloadRow[]> = {
        deployments: actions.selectedDeployments.value,
        pods: actions.selectedPods.value,
        statefulsets: actions.selectedStatefulSets.value,
        daemonsets: actions.selectedDaemonSets.value,
        jobs: actions.selectedJobs.value,
        cronjobs: actions.selectedCronJobs.value
      };
      return (map[activeTab.value] || []).length > 0;
    });

    function goToDetail(row: K8s.WorkloadRow) {
      const baseQuery = { clusterId: selectedCluster.value, namespace: row.namespace, name: row.name };
      const pathMap: Record<string, string> = {
        deployments: '/k8s/resources/deployments/detail',
        statefulsets: '/k8s/resources/statefulsets/detail',
        daemonsets: '/k8s/resources/daemonsets/detail',
        pods: '/k8s/resources/pods/detail',
        jobs: '/k8s/resources/jobs/detail',
        cronjobs: '/k8s/resources/cronjobs/detail'
      };
      const path = pathMap[activeTab.value];
      if (path) router.push({ path, query: baseQuery });
    }

    function onTabChange(tabName: string) {
      actions.clearSelection();
      handleTabChange(tabName);
    }
</script>

<template>
    <div class="workloads-page p-24px">
        <div class="mb-24px">
            <h1 class="text-28px text-primary font-bold">工作负载</h1>
            <p class="text-tertiary mt-8px text-14px">管理 Kubernetes 工作负载资源</p>
        </div>

        <div class="mb-16px flex items-center justify-between gap-12px">
            <div class="filter-inputs flex items-center gap-8px">
                <ElSelect v-model="selectedCluster" style="width: 200px" @change="loadNamespaces">
                    <template #prefix><span class="select-fixed-label">选择集群</span></template>
                    <ElOption v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
                </ElSelect>
                <ElSelect v-model="selectedNamespace" style="width: 180px" @change="loadCurrentData">
                    <template #prefix><span class="select-fixed-label">选择命名空间</span></template>
                    <ElOption v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
                </ElSelect>
            </div>
            <ElButton text @click="loadCurrentData">
                <icon-mdi-refresh class="text-18px" :class="{ 'animate-spin': loading }" />
            </ElButton>
        </div>

        <ElTabs v-model="activeTab" @tab-change="onTabChange">
            <ElTabPane label="容器组" name="pods">
                <PodTab :data="podsData" :loading="loading" :pagination="podsPagination" :selected-count="actions.selectedPods.value.length" @go-to-detail="goToDetail" @page-change="loadPods" @size-change="handlePageSizeChange" @selection-change="actions.handlePodSelectionChange" @select-all="actions.handlePodSelectAll" @clear-selection="actions.clearSelection" @pod-logs="actions.handlePodLogs" @pod-more-command="actions.handlePodMoreCommand" @batch-delete-pods="actions.handleBatchDeletePods" />
            </ElTabPane>

            <ElTabPane label="无状态" name="deployments">
                <DeploymentTab :data="deploymentsData" :loading="loading" :pagination="deploymentsPagination" :selected-count="actions.selectedDeployments.value.length" @go-to-detail="goToDetail" @page-change="loadDeployments" @size-change="handlePageSizeChange" @selection-change="actions.handleDeploymentSelectionChange" @select-all="actions.handleDeploymentSelectAll" @clear-selection="actions.clearSelection" @scale="actions.handleScale" @more-command="actions.handleMoreCommand" @batch-restart="actions.handleBatchRestart" @batch-delete="actions.handleBatchDelete" />
            </ElTabPane>

            <ElTabPane label="有状态" name="statefulsets">
                <WorkloadTab resource-type="statefulset" columns="statefulset" :data="statefulSetsData" :loading="loading" :pagination="statefulSetsPagination" :selected-count="actions.selectedStatefulSets.value.length" @go-to-detail="goToDetail" @page-change="loadStatefulSets" @size-change="handlePageSizeChange" @selection-change="actions.handleStatefulSetSelectionChange" @select-all="actions.handleStatefulSetSelectAll" @clear-selection="actions.clearSelection" @workload-command="(cmd, row) => actions.handleWorkloadCommand('statefulset', cmd, row)" @batch-delete-by-type="actions.handleBatchDeleteByType" />
            </ElTabPane>

            <ElTabPane label="守护进程集" name="daemonsets">
                <WorkloadTab resource-type="daemonset" columns="daemonset" :data="daemonSetsData" :loading="loading" :pagination="daemonSetsPagination" :selected-count="actions.selectedDaemonSets.value.length" @go-to-detail="goToDetail" @page-change="loadDaemonSets" @size-change="handlePageSizeChange" @selection-change="actions.handleDaemonSetSelectionChange" @select-all="actions.handleDaemonSetSelectAll" @clear-selection="actions.clearSelection" @workload-command="(cmd, row) => actions.handleWorkloadCommand('daemonset', cmd, row)" @batch-delete-by-type="actions.handleBatchDeleteByType" />
            </ElTabPane>

            <ElTabPane label="任务" name="jobs">
                <WorkloadTab resource-type="job" columns="job" :data="jobsData" :loading="loading" :pagination="jobsPagination" :selected-count="actions.selectedJobs.value.length" @go-to-detail="goToDetail" @page-change="loadJobs" @size-change="handlePageSizeChange" @selection-change="actions.handleJobSelectionChange" @select-all="actions.handleJobSelectAll" @clear-selection="actions.clearSelection" @workload-command="(cmd, row) => actions.handleWorkloadCommand('job', cmd, row)" @batch-delete-by-type="actions.handleBatchDeleteByType" />
            </ElTabPane>

            <ElTabPane label="定时任务" name="cronjobs">
                <WorkloadTab resource-type="cronjob" columns="cronjob" :data="cronJobsData" :loading="loading" :pagination="cronJobsPagination" :selected-count="actions.selectedCronJobs.value.length" @go-to-detail="goToDetail" @page-change="loadCronJobs" @size-change="handlePageSizeChange" @selection-change="actions.handleCronJobSelectionChange" @select-all="actions.handleCronJobSelectAll" @clear-selection="actions.clearSelection" @workload-command="(cmd, row) => actions.handleWorkloadCommand('cronjob', cmd, row)" @batch-delete-by-type="actions.handleBatchDeleteByType" />
            </ElTabPane>
        </ElTabs>

        <ScaleDialog v-model:show="actions.showScaleDialog.value" :deployment="actions.selectedDeployment.value" :scale-data="actions.scaleData" @submit="actions.handleScaleSubmit" />

        <LogDialog v-model:show="actions.showLogDialog.value" :title="`日志: ${actions.selectedResource.value?.name}${actions.selectedResource.value?.containers?.[0]?.name ? ' (' + actions.selectedResource.value.containers[0].name + ')' : ''}`" :content="actions.logDialogContent.value" />

        <YamlEditor v-model="actions.showYamlDialog.value" :title="actions.yamlDialogTitle.value" :yaml="actions.yamlContent.value" :can-edit="true" :on-apply="actions.handleYamlApply" />
    </div>
</template>

<style scoped>
    .workloads-page {
      height: calc(100vh - 80px);
      display: flex;
      flex-direction: column;
      overflow: hidden;
    }

    :deep(.el-tabs) {
      flex: 1;
      display: flex;
      flex-direction: column;
      overflow: hidden;
    }

    :deep(.el-tabs__content) {
      flex: 1;
      display: flex;
      flex-direction: column;
      overflow: hidden;
    }

    :deep(.el-tab-pane) {
      flex: 1;
      display: flex;
      flex-direction: column;
      overflow: hidden;
      height: 100%;
    }

    .filter-inputs {
      :deep(.el-select__wrapper) {
        border-radius: 0 !important;
        height: 30px;
        font-size: 12px;
        line-height: 30px;
      }

      :deep(.el-input__wrapper) {
        border-radius: 0 !important;
        height: 30px;
        font-size: 12px;
      }

      :deep(.el-select) {
        height: 30px;
        font-size: 12px;
      }

      :deep(.el-select .el-select__selection) {
        display: none;
      }

      :deep(.el-select .el-select__selected-item) {
        display: none;
      }

      :deep(.el-select .el-select__placeholder) {
        display: none;
      }

      .select-fixed-label {
        font-size: 12px;
        color: var(--el-text-color-regular);
        line-height: 30px;
        padding-left: 8px;
      }

      :deep(.el-select.has-value .el-select__prefix) {
        position: static;
        flex: none;
      }
    }
</style>
