/**
 * 虚拟滚动 Hook
 *
 * 用于处理大列表的性能优化，只渲染可见区域的项目
 *
 * @example
 * const { list, containerProps, wrapperProps } = useVirtualList({
 *   data: largeDataSource,
 *   itemHeight: 50
 * })
 */

import { type Ref, computed, onMounted, onUnmounted, ref } from 'vue';

export interface VirtualListOptions<T> {
  /** 数据源 */
  data: Ref<T[]>;
  /** 每个项目的高度 */
  itemHeight: number;
  /** 容器高度 */
  containerHeight?: number;
  /** 缓冲区大小（渲染额外项目） */
  bufferSize?: number;
}

export interface VirtualListReturn {
  /** 可见列表数据 */
  list: Ref<unknown[]>;
  /** 容器属性 */
  containerProps: {
    ref: Ref<HTMLElement | undefined>;
    style: Record<string, string>;
  };
  /** 包装器属性 */
  wrapperProps: {
    style: Record<string, string>;
  };
  /** 滚动到指定索引 */
  scrollToIndex: (index: number) => void;
}

export function useVirtualList<T>(options: VirtualListOptions<T>): VirtualListReturn {
  const { data, itemHeight, containerHeight = 600, bufferSize = 3 } = options;

  const scrollTop = ref(0);
  const containerRef = ref<HTMLElement>();

  // 计算可见范围
  const visibleRange = computed(() => {
    const start = Math.floor(scrollTop.value / itemHeight);
    const visibleCount = Math.ceil(containerHeight / itemHeight);
    const end = start + visibleCount;

    return {
      start: Math.max(0, start - bufferSize),
      end: Math.min(data.value.length, end + bufferSize)
    };
  });

  // 可见列表数据
  const list = computed(() => {
    const { start, end } = visibleRange.value;
    return data.value.slice(start, end).map((item, index) => ({
      data: item,
      index: start + index
    }));
  });

  // 容器总高度
  const totalHeight = computed(() => data.value.length * itemHeight);

  // 偏移量
  const offsetY = computed(() => visibleRange.value.start * itemHeight);

  // 处理滚动事件
  const handleScroll = (e: Event) => {
    scrollTop.value = (e.target as HTMLElement).scrollTop;
  };

  // 滚动到指定索引
  const scrollToIndex = (index: number) => {
    if (containerRef.value) {
      containerRef.value.scrollTop = index * itemHeight;
    }
  };

  // 容器属性
  const containerProps = {
    ref: containerRef,
    style: {
      height: `${containerHeight}px`,
      overflow: 'auto'
    }
  };

  // 包装器属性
  const wrapperProps = {
    style: computed(() => ({
      height: `${totalHeight.value}px`,
      position: 'relative'
    }))
  };

  onMounted(() => {
    if (containerRef.value) {
      containerRef.value.addEventListener('scroll', handleScroll);
    }
  });

  onUnmounted(() => {
    if (containerRef.value) {
      containerRef.value.removeEventListener('scroll', handleScroll);
    }
  });

  return {
    list,
    containerProps,
    wrapperProps,
    scrollToIndex
  };
}
