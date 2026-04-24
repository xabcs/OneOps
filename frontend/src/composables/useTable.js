/**
 * 表格页面 composable
 * 通用的表格数据管理逻辑
 */

import { ref, reactive, computed } from 'vue'

export function useTable(fetchFn, options = {}) {
  const {
    pageSize = 15,
    defaultQuery = {}
  } = options

  // 状态
  const loading = ref(false)
  const data = ref([])
  const total = ref(0)

  // 查询参数
  const queryParams = reactive({
    page: 1,
    pageSize,
    ...defaultQuery
  })

  // 计算属性
  const currentPage = computed({
    get: () => queryParams.page,
    set: (val) => {
      queryParams.page = val
      fetchData()
    }
  })

  /**
   * 获取数据
   */
  async function fetchData() {
    loading.value = true
    try {
      const result = await fetchFn(queryParams)
      data.value = result.data || result.items || []
      total.value = result.total || result.count || 0
    } catch (error) {
      console.error('Failed to fetch data:', error)
      data.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }

  /**
   * 搜索
   */
  function search() {
    queryParams.page = 1
    fetchData()
  }

  /**
   * 重置
   */
  function reset() {
    Object.assign(queryParams, {
      page: 1,
      pageSize,
      ...defaultQuery
    })
    fetchData()
  }

  /**
   * 删除项
   */
  async function deleteItem(deleteFn, id) {
    try {
      await deleteFn(id)
      // 刷新当前页
      fetchData()
      return true
    } catch (error) {
      console.error('Failed to delete item:', error)
      return false
    }
  }

  // 初始加载数据
  fetchData()

  return {
    // 状态
    loading,
    data,
    total,
    queryParams,

    // 计算属性
    currentPage,

    // 方法
    fetchData,
    search,
    reset,
    deleteItem
  }
}
