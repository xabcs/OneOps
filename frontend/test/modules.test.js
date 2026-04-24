// 测试所有模块导入
import { describe, it, expect } from 'vitest'

describe('模块导入测试', () => {
  it('应该能正确导入 errorHandler', async () => {
    const errorHandler = await import('../src/utils/errorHandler.js')
    expect(errorHandler).toBeDefined()
    expect(errorHandler.isNetworkError).toBeDefined()
    expect(errorHandler.handleApiError).toBeDefined()
  })

  it('应该能正确导入 healthCheck', async () => {
    const healthCheck = await import('../src/services/healthCheck.js')
    expect(healthCheck).toBeDefined()
    expect(healthCheck.healthCheckService).toBeDefined()
  })

  it('应该能正确导入 API', async () => {
    const api = await import('../src/api/index.js')
    expect(api).toBeDefined()
    expect(api.loginApi).toBeDefined()
    expect(api.systemApi).toBeDefined()
  })

  it('应该能正确导入 store', async () => {
    const store = await import('../src/store/index.js')
    expect(store).toBeDefined()
    expect(store.default).toBeDefined()
  })

  it('应该能正确导入路由', async () => {
    const router = await import('../src/router/index.js')
    expect(router).toBeDefined()
    expect(router.default).toBeDefined()
  })
})