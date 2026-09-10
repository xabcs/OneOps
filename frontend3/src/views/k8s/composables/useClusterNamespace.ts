import type { Ref } from 'vue';
import { ref } from 'vue';
import { ElMessage } from 'element-plus';
import { fetchK8sClusterNamespaces, fetchK8sClusters } from '@/service/api/k8s';

/**
 * 提取 flat 请求 error 上的提示信息，缺省使用兜底文案
 * flat 封装恒返回 {data, error} 不 reject，失败信息挂在 error.message 上
 */
function errMsg(error: unknown, fallback: string): string {
  return (error as { message?: string } | null)?.message || fallback;
}

/** useClusterNamespace 配置项 */
export interface UseClusterNamespaceOptions {
  /** 集群/命名空间就绪后的回调（加载资源列表） */
  onDataReady?: (clusterId: number, namespace: string) => void;
  /** 集群变化回调（可选，默认触发 onDataReady） */
  onClusterChange?: (clusterId: number, namespace: string) => void;
  /** sessionStorage 恢复 key（配合 useListStateRestore，可选） */
  storageKey?: string;
  /** 集群列表查询参数（部分页面需要一次拉取更大的分页容量） */
  clusterParams?: K8s.ClusterQuery;
}

/** useClusterNamespace 返回的上下文（useListStateRestore 依赖此结构） */
export interface ClusterNamespaceContext {
  /** 可用的集群列表 */
  clusters: Ref<K8s.Cluster[]>;
  /** 可用的命名空间名称列表 */
  namespaces: Ref<string[]>;
  /** 当前选中的集群 */
  selectedCluster: Ref<number | null>;
  /** 当前选中的命名空间 */
  selectedNamespace: Ref<string>;
  /** 列表状态恢复 key（透传自 options.storageKey） */
  storageKey?: string;
  /** 仅加载集群列表（无选中时默认选第一个，不加载命名空间） */
  loadClusters: () => Promise<void>;
  /** 完整初始化：集群列表 → 命名空间列表 → onDataReady */
  loadAll: () => Promise<void>;
  /** 加载当前选中集群的命名空间列表（选中项不在列表时自动回填第一个） */
  loadNamespaces: () => Promise<void>;
  /** 集群切换处理器（模板 @change 直接绑定） */
  handleClusterChange: () => Promise<void>;
  /** 命名空间切换处理器（模板 @change 直接绑定） */
  handleNamespaceChange: () => void;
}

/**
 * k8s 页面通用的「集群/命名空间初始化 + 联动」composable
 *
 * 内部统一流程：
 * 1. fetchK8sClusters → res.data?.list（data 为 {list,total} 分页对象）
 * 2. 无选中集群时默认选择第一个
 * 3. fetchK8sClusterNamespaces → data（Namespace[] 数组）→ map(name)
 * 4. 选中的命名空间不在列表中时默认选择第一个
 * 5. 就绪后触发 onDataReady 加载资源列表
 *
 * 错误处理统一 ElMessage.error
 */
export function useClusterNamespace(options: UseClusterNamespaceOptions = {}): ClusterNamespaceContext {
  // 可用的集群列表
  const clusters = ref<K8s.Cluster[]>([]);
  // 可用的命名空间列表（仅名称）
  const namespaces = ref<string[]>([]);
  // 当前选中的集群和命名空间
  const selectedCluster = ref<number | null>(null);
  const selectedNamespace = ref('default');

  // 加载集群列表：flat 请求恒返回 {data, error} 不 reject，data 为 {list,total} 分页对象
  async function loadClusters() {
    const { data, error } = await fetchK8sClusters(options.clusterParams);
    if (error) {
      ElMessage.error(errMsg(error, '加载集群列表失败'));
      return;
    }
    clusters.value = data?.list || [];

    // 如果有集群且当前未选中，默认选择第一个
    if (clusters.value.length > 0 && !selectedCluster.value) {
      selectedCluster.value = clusters.value[0].id;
    }
  }

  // 加载命名空间列表：data 为 Namespace[] 数组，映射为名称数组
  async function loadNamespaces() {
    if (!selectedCluster.value) return;
    const { data, error } = await fetchK8sClusterNamespaces(selectedCluster.value);
    if (error) {
      ElMessage.error(errMsg(error, '加载命名空间列表失败'));
      return;
    }
    namespaces.value = (data || []).map((ns: K8s.Namespace) => ns.name);

    // 当前命名空间不在列表中时默认选择第一个
    if (namespaces.value.length > 0 && !namespaces.value.includes(selectedNamespace.value)) {
      selectedNamespace.value = namespaces.value[0];
    }
  }

  // 完整初始化：集群列表 → 默认选中 → 命名空间列表 → 默认选中 → 触发 onDataReady
  async function loadAll() {
    await loadClusters();
    if (!selectedCluster.value) return;
    await loadNamespaces();
    options.onDataReady?.(selectedCluster.value, selectedNamespace.value);
  }

  // 集群切换：重载命名空间（选中项失效时自动纠正）后触发回调
  async function handleClusterChange() {
    await loadNamespaces();
    if (!selectedCluster.value) return;
    if (options.onClusterChange) {
      options.onClusterChange(selectedCluster.value, selectedNamespace.value);
    } else {
      options.onDataReady?.(selectedCluster.value, selectedNamespace.value);
    }
  }

  // 命名空间切换：直接触发 onDataReady 加载列表
  function handleNamespaceChange() {
    if (!selectedCluster.value) return;
    options.onDataReady?.(selectedCluster.value, selectedNamespace.value);
  }

  return {
    clusters,
    namespaces,
    selectedCluster,
    selectedNamespace,
    storageKey: options.storageKey,
    loadClusters,
    loadAll,
    loadNamespaces,
    handleClusterChange,
    handleNamespaceChange
  };
}

// ==================== 列表状态保存/恢复（配合 sessionStorage） ====================

/** 列表页保存到 sessionStorage 的状态（页面可扩展搜索条件等字段） */
export interface K8sListState {
  clusterId: number | null;
  namespace: string;
  /** 返回列表时的目标路径（历史字段，保留兼容） */
  listPath?: string;
  [key: string]: unknown;
}

/** 保存列表状态到 sessionStorage（跳转详情页前调用） */
export function saveListState(key: string, state: K8sListState): void {
  try {
    sessionStorage.setItem(key, JSON.stringify(state));
  } catch (e) {
    console.error('[K8s列表] sessionStorage 保存失败:', e);
  }
}

/** useListStateRestore 配置项 */
export interface UseListStateRestoreOptions {
  /** sessionStorage key，缺省使用 useClusterNamespace 传入的 storageKey */
  key?: string;
  /** 集群/命名空间恢复完成（loadNamespaces 之后）加载资源列表 */
  onLoadList: (clusterId: number, namespace: string) => void | Promise<void>;
  /** 恢复页面扩展状态（搜索条件等），在写回集群/命名空间选中值之后调用 */
  onRestored?: (state: K8sListState) => void;
}

/**
 * 列表页 sessionStorage 状态恢复 composable
 *
 * 封装「getItem → JSON.parse → 恢复 selected/搜索条件 → removeItem → loadNamespaces → 加载列表」流程：
 * - 有存档：恢复出的集群/命名空间写回 useClusterNamespace 的选中值，再加载命名空间与列表
 * - 无存档：走 loadAll 正常初始化（内部触发 onDataReady 加载列表）
 */
export function useListStateRestore(cn: ClusterNamespaceContext, options: UseListStateRestoreOptions) {
  const storageKey = options.key ?? cn.storageKey;

  /** 恢复列表状态（onMounted / 从详情页返回时调用） */
  async function restoreState(): Promise<void> {
    if (storageKey) {
      const raw = sessionStorage.getItem(storageKey);
      if (raw) {
        try {
          const state = JSON.parse(raw) as K8sListState;

          // 恢复出的集群/命名空间写回 useClusterNamespace 的选中值
          if (state.clusterId != null) {
            cn.selectedCluster.value = state.clusterId;
          }
          if (state.namespace) {
            cn.selectedNamespace.value = state.namespace;
          }

          // 恢复页面扩展的搜索条件等
          options.onRestored?.(state);

          // 消费后立即清除，避免刷新页面后误恢复
          sessionStorage.removeItem(storageKey);

          if (cn.selectedCluster.value) {
            await cn.loadNamespaces();
            await options.onLoadList(cn.selectedCluster.value, cn.selectedNamespace.value);
            return;
          }
        } catch (e) {
          console.error('[K8s列表] 恢复状态失败:', e);
        }
      }
    }

    // 无存档时走正常初始化流程
    await cn.loadAll();
  }

  /** 绑定 storageKey 的保存函数（跳转详情页前调用） */
  function saveState(state: K8sListState): void {
    if (!storageKey) return;
    saveListState(storageKey, state);
  }

  return { restoreState, saveState };
}
