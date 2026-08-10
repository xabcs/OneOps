import { ref } from 'vue';
import { defineStore } from 'pinia';
import { SetupStoreId } from '@/enum';

/**
 * K8s 工作负载筛选状态
 */
export interface K8sWorkloadFilterState {
  /** 集群 ID */
  clusterId: number | null;
  /** 命名空间 */
  namespace: string;
  /** 活动的标签页 */
  activeTab: string;
}

/**
 * K8s Store - 管理 K8s 相关的筛选和状态
 */
export const useK8sStore = defineStore(SetupStoreId.K8s, () => {
  /** 工作负载筛选状态 */
  const workloadFilterState = ref<K8sWorkloadFilterState>({
    clusterId: null,
    namespace: 'default',
    activeTab: 'deployments'
  });

  /**
   * 设置工作负载筛选状态
   */
  function setWorkloadFilterState(state: Partial<K8sWorkloadFilterState>) {
    workloadFilterState.value = {
      ...workloadFilterState.value,
      ...state
    };
  }

  /**
   * 重置工作负载筛选状态
   */
  function resetWorkloadFilterState() {
    workloadFilterState.value = {
      clusterId: null,
      namespace: 'default',
      activeTab: 'deployments'
    };
  }

  /**
   * 获取当前工作负载筛选状态
   */
  function getWorkloadFilterState(): K8sWorkloadFilterState {
    return workloadFilterState.value;
  }

  return {
    workloadFilterState,
    setWorkloadFilterState,
    resetWorkloadFilterState,
    getWorkloadFilterState
  };
});
