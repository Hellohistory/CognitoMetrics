/**
 * @file src/api/apiClient.ts
 * @description Axios 实例封装，配置请求基础路径、统一请求头、超时与全局错误处理
 * @author Hellohistory
 */

/**
 * @module api/apiClient
 * 高内聚：专注于 HTTP 请求实例的创建与拦截处理
 * 低耦合：与具体业务逻辑分离，仅提供基础网络能力
 */

import axios from 'axios';
import { ElMessage } from 'element-plus';

/**
 * 创建 Axios 实例
 */
const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL, // 接口基础路径，从环境变量获取
  headers: {
    'Content-Type': 'application/json', // 默认请求内容类型
  },
  timeout: 10000, // 请求超时设置（毫秒）
});

/**
 * 响应拦截器：统一处理成功或失败的 HTTP 响应
 */
apiClient.interceptors.response.use(
  /**
   * 成功响应处理
   * @param response - Axios 响应对象
   * @returns 原始响应对象，供调用处继续使用
   */
  response => {
    return response;
  },
  /**
   * 错误响应处理
   * @param error - 错误对象
   * @returns 拒绝的 Promise，以便调用端通过 .catch 继续处理
   */
  error => {
    if (error.response) {
      // 服务端返回非 2xx 状态码
      const errorMessage =
        error.response.data.detail || '操作失败，请检查网络或联系管理员';
      ElMessage.error(errorMessage);
    } else if (error.request) {
      // 请求已发出，但未收到响应
      ElMessage.error('网络错误，无法连接到服务器');
    } else {
      // 请求配置时发生错误
      ElMessage.error('请求发送失败，请检查您的网络设置');
    }
    return Promise.reject(error);
  }
);

export default apiClient;
