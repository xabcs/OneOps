<script setup lang="ts">
import { computed } from 'vue';

defineOptions({ name: 'DarkModeContainer' });

interface Props {
  inverted?: boolean;
  customColor?: string;
}

const props = withDefaults(defineProps<Props>(), {
  inverted: false,
  customColor: ''
});

const containerStyle = computed(() => {
  if (props.inverted && props.customColor) {
    return {
      backgroundColor: props.customColor
    };
  }
  return {};
});
</script>

<template>
  <div
    class="bg-container text-base-text transition-300"
    :class="{ 'bg-inverted text-#1f1f1f': inverted && !customColor }"
    :style="containerStyle"
  >
    <slot></slot>
  </div>
</template>

<style scoped></style>
