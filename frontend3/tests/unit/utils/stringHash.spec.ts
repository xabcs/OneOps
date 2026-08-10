import { describe, it, expect } from 'vitest'
import { stringHash } from '@/views/manage/user/index.vue'

describe('工具函数测试', () => {
  describe('stringHash', () => {
    it('能够为相同字符串生成一致的哈希值', () => {
      const hash1 = stringHash('test')
      const hash2 = stringHash('test')
      expect(hash1).toBe(hash2)
    })

    it('为不同字符串生成不同的哈希值', () => {
      const hash1 = stringHash('admin')
      const hash2 = stringHash('user')
      expect(hash1).not.toBe(hash2)
    })

    it('返回的哈希值是正数', () => {
      const hash = stringHash('test')
      expect(hash).toBeGreaterThanOrEqual(0)
    })
  })
})
