<template>
  <div class="action-buttons">
    <template v-for="btn in buttons" :key="btn.text">
      <el-button
        v-if="!btn.visible || btn.visible(row)"
        link
        :type="btn.type || 'primary'"
        :icon="btn.icon"
        :disabled="btn.disabled?.(row)"
        @click="() => btn.click(row)"
      >
        {{ btn.text }}
      </el-button>
    </template>
  </div>
</template>

<script setup>
const props = defineProps({
  row: {
    type: Object,
    required: true
  },
  buttons: {
    type: Array,
    required: true
    // 每个按钮的格式:
    // {
    //   text: string,
    //   type?: 'primary' | 'success' | 'warning' | 'danger' | 'info',
    //   icon?: any,
    //   disabled?: (row) => boolean,
    //   visible?: (row) => boolean,
    //   click: (row) => void
    // }
  }
})
</script>

<style scoped>
.action-buttons {
  display: flex;
  align-items: center;
  gap: 8px;
  justify-content: center;
}
</style>
