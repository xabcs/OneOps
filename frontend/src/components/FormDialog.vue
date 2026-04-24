<template>
  <el-dialog
    v-model="visible"
    :title="title"
    :width="width"
    :destroy-on-close="destroyOnClose"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formModel"
      :rules="rules"
      :label-width="labelWidth"
      :label-position="labelPosition"
    >
      <slot />
    </el-form>

    <template #footer v-if="showFooter">
      <el-button @click="handleClose" :disabled="submitting">
        {{ cancelText }}
      </el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">
        {{ confirmText }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed, watch, nextTick, ref } from 'vue'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  title: {
    type: String,
    default: '表单'
  },
  formModel: {
    type: Object,
    required: true
  },
  rules: {
    type: Object,
    default: () => ({})
  },
  width: {
    type: String,
    default: '500px'
  },
  labelWidth: {
    type: String,
    default: '80px'
  },
  labelPosition: {
    type: String,
    default: 'top'
  },
  submitting: {
    type: Boolean,
    default: false
  },
  cancelText: {
    type: String,
    default: '取消'
  },
  confirmText: {
    type: String,
    default: '确定'
  },
  showFooter: {
    type: Boolean,
    default: true
  },
  destroyOnClose: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits(['update:modelValue', 'submit', 'close'])

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const formRef = ref(null)

const handleClose = () => {
  visible.value = false
  emit('close')
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate((valid) => {
    if (valid) {
      emit('submit', props.formModel)
    }
  })
}

const clearValidate = () => {
  nextTick(() => {
    formRef.value?.clearValidate()
  })
}

const resetFields = () => {
  formRef.value?.resetFields()
}

watch(() => props.modelValue, (newVal) => {
  if (newVal) {
    clearValidate()
  }
})

defineExpose({
  clearValidate,
  resetFields,
  formRef
})
</script>
