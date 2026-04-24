<template>
  <el-space :size="12" class="search-bar">
    <slot name="filters"></slot>
    <el-button
      v-if="showSearch"
      type="primary"
      :icon="Search"
      @click="handleSearch"
    >
      搜索
    </el-button>
    <el-button
      v-if="showReset"
      :icon="RefreshRight"
      @click="handleReset"
    >
      重置
    </el-button>
  </el-space>
</template>

<script setup>
import { Search, RefreshRight } from '@element-plus/icons-vue'

defineProps({
  showSearch: {
    type: Boolean,
    default: true
  },
  showReset: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits(['search', 'reset'])

const handleSearch = () => {
  emit('search')
}

const handleReset = () => {
  emit('reset')
}
</script>

<style scoped>
.search-bar {
  display: flex;
  flex-wrap: wrap;
}
</style>
