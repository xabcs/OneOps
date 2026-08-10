import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { ElCollapse, ElCollapseItem, ElForm, ElFormItem, ElInput, ElButton } from 'element-plus'
import UserSearch from '@/views/auth/users/modules/user-search.vue'

describe('UserSearch 组件测试', () => {
  it('组件能够正常渲染', () => {
    const wrapper = mount(UserSearch, {
      global: {
        components: {
          ElCollapse,
          ElCollapseItem,
          ElForm,
          ElFormItem,
          ElInput,
          ElButton
        }
      }
    })
    expect(wrapper.exists()).toBe(true)
  })

  it('包含折叠面板组件', () => {
    const wrapper = mount(UserSearch, {
      global: {
        components: {
          ElCollapse,
          ElCollapseItem,
          ElForm,
          ElFormItem,
          ElInput,
          ElButton
        }
      }
    })
    expect(wrapper.findComponent(ElCollapse).exists()).toBe(true)
  })

  it('包含搜索表单', () => {
    const wrapper = mount(UserSearch, {
      global: {
        components: {
          ElCollapse,
          ElCollapseItem,
          ElForm,
          ElFormItem,
          ElInput,
          ElButton
        }
      }
    })
    expect(wrapper.findComponent(ElForm).exists()).toBe(true)
  })
})
