<script setup lang="ts">
import { ref, watch } from 'vue';
import { CircleClose } from '@element-plus/icons-vue';

interface Props {
  modelValue?: string | number;
  type?: string;
  placeholder?: string;
  disabled?: boolean;
  readonly?: boolean;
  maxlength?: number;
  clearable?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  type: 'text',
  disabled: false,
  readonly: false,
  clearable: false
});

const emit = defineEmits<{
  'update:modelValue': [value: string | number];
  input: [value: string | number];
  change: [value: string | number];
  focus: [event: FocusEvent];
  blur: [event: FocusEvent];
  clear: [];
}>();

const inputValue = ref(props.modelValue || '');

// 监听外部值变化
watch(
  () => props.modelValue,
  newValue => {
    inputValue.value = newValue || '';
  }
);

// 处理输入
const handleInput = (event: Event) => {
  const target = event.target as HTMLInputElement;
  inputValue.value = target.value;
  emit('update:modelValue', target.value);
  emit('input', target.value);
};

// 处理变化
const handleChange = (event: Event) => {
  const target = event.target as HTMLInputElement;
  emit('change', target.value);
};

// 处理焦点
const handleFocus = (event: FocusEvent) => {
  emit('focus', event);
};

// 处理失焦
const handleBlur = (event: FocusEvent) => {
  emit('blur', event);
};

// 处理清除
const handleClear = () => {
  inputValue.value = '';
  emit('update:modelValue', '');
  emit('clear');
};
</script>

<template>
  <div class="a-input-group">
    <div v-if="$slots.prepend" class="a-input-prepend">
      <slot name="prepend" />
    </div>

    <input
      v-model="inputValue"
      :type="type"
      :placeholder="placeholder"
      :disabled="disabled"
      :readonly="readonly"
      :maxlength="maxlength"
      class="a-input-inner"
      @input="handleInput"
      @change="handleChange"
      @focus="handleFocus"
      @blur="handleBlur"
    />

    <div v-if="$slots.append" class="a-input-append">
      <slot name="append" />
    </div>

    <div v-if="clearable && inputValue" class="a-input-clear" @click="handleClear">
      <ElIcon><CircleClose /></ElIcon>
    </div>
  </div>
</template>

<style scoped>
.a-input-group {
  display: flex;
  align-items: center;
  width: 100%;
  position: relative;
  box-sizing: border-box;
}

.a-input-prepend,
.a-input-append {
  display: flex;
  align-items: center;
  padding: 0 12px;
  background: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color);
  color: var(--el-text-color-regular);
  white-space: nowrap;
}

.a-input-prepend {
  border-right: none;
  border-radius: 4px 0 0 4px;
}

.a-input-append {
  border-left: none;
  border-radius: 0 4px 4px 0;
}

.a-input-inner {
  flex: 1;
  width: 100%;
  padding: 8px 12px;
  font-size: 14px;
  line-height: 1.5;
  color: var(--el-text-color-regular);
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  outline: none;
  transition: border-color 0.2s;
  box-sizing: border-box;
}

.a-input-inner:hover {
  border-color: var(--el-border-color-hover);
}

.a-input-inner:focus {
  border-color: var(--el-color-primary);
}

.a-input-inner:disabled {
  background: var(--el-fill-color-light);
  color: var(--el-text-color-placeholder);
  cursor: not-allowed;
}

.a-input-clear {
  position: absolute;
  right: 8px;
  cursor: pointer;
  color: var(--el-text-color-placeholder);
  display: flex;
  align-items: center;
}

.a-input-clear:hover {
  color: var(--el-text-color-regular);
}
</style>
