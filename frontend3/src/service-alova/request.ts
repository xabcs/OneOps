import { createAlovaRequest } from '@sa/alova';
import adapterFetch from 'alova/fetch';
import { getServiceBaseURL } from '@/utils/service';

const { baseURL } = getServiceBaseURL();

/**
 * alova 实例（供 src/views/alova 演示页面使用）
 *
 * 该模块此前随 service-alova 目录被误删导致构建失败，此处为最小恢复。
 */
export const alova = createAlovaRequest(
  {
    baseURL,
    requestAdapter: adapterFetch()
  },
  {
    isBackendSuccess: () => true,
    transformBackendResponse: async (response: Response) => response.json()
  }
);
