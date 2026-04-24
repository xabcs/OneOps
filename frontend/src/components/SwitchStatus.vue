<template>
  <el-switch
    v-model="internalValue"
    :active-value="activeValue"
    :inactive-value="inactiveValue"
    :disabled="disabled || readonly"
    @change="handleChange"
  />
</template>

<script setup>
import { computed } from 'vue'
import { ElMessageBox } from 'element-plus'

const props = defineProps({
  modelValue: {
    type: [String, Number, Boolean],
    required: true
  },
  activeValue: {
    type: [String, Number, Boolean],
    default: true
  },
  inactiveValue: {
    type: [String, Number, Boolean],
    default: false
  },
  disabled: {
    type: Boolean,
    default: false
  },
  readonly: {
    type: Boolean,
    default: false
  },
  confirmMessage: {
    type: [String, Function],
    default: null
  }
})

const emit = defineEmits(['update:modelValue', 'change'])

const internalValue = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const handleChange = async (val) => {
  if (props.confirmMessage) {
    const message = typeof props.confirmMessage === 'function'
      ? props.confirmMessage(val)
      : props.confirmMessage

    try {
      await ElMessageBox.confirm(
        message,
        '确认操作',
        {
          type: 'warning',
          confirmButtonText: '确定',
          cancelButtonText: '取消'
        }
      )
      emit('change', val)
    } catch {
      // 用户取消，恢复原值
      internalValue.value = val === props.activeValue ? props.inactiveValue : props.activeValue
    }
  } else {
    emit('change', val)
  }
}
</script>
