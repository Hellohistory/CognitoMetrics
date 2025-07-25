// src/utils/pollingService.ts

import { getReportDetails } from '@/api/analysisApi';
import type { IAnalysisReportDetail } from '@/types/dataModels';

interface ReportPollingOptions {
  reportId: number;
  onSuccess: (report: IAnalysisReportDetail) => void;
  onFailure: (error?: any) => void;
  onUpdate?: (report: IAnalysisReportDetail) => void;
  interval?: number;
  timeout?: number;
}

export function startReportPolling(options: ReportPollingOptions): () => void {
  const { reportId, onSuccess, onFailure, onUpdate, interval = 3000, timeout = 120000 } = options;
  let pollingHandle: number | undefined;
  const startTime = Date.now();
  const poll = async () => {
    if (Date.now() - startTime > timeout) {
      clearInterval(pollingHandle);
      onFailure(new Error(`轮询超时 (超过 ${timeout / 1000} 秒)`));
      return;
    }
    try {
      const report = await getReportDetails(reportId);
      if (onUpdate) onUpdate(report);
      if (report.status === 'completed') {
        clearInterval(pollingHandle);
        onSuccess(report);
      } else if (report.status === 'failed') {
        clearInterval(pollingHandle);
        onFailure(report);
      }
    } catch (error) {
      clearInterval(pollingHandle);
      onFailure(error);
    }
  };
  poll();
  pollingHandle = window.setInterval(poll, interval);
  return () => { if (pollingHandle) clearInterval(pollingHandle); };
}


// --- 【新增】AI分析轮询服务 ---
interface AiPollingOptions extends Omit<ReportPollingOptions, 'onUpdate'> {}

/**
 * 【新增】启动一个专门针对AI分析状态的轮询任务。
 * @param options - 轮询配置
 * @returns 一个可以手动停止轮询的函数
 */
export function startAiAnalysisPolling(options: AiPollingOptions): () => void {
  const {
    reportId,
    onSuccess,
    onFailure,
    interval = 5000,
    timeout = 1000000,
  } = options;

  let pollingHandle: number | undefined;
  const startTime = Date.now();

  const poll = async () => {
    // 检查是否超时
    if (Date.now() - startTime > timeout) {
      clearInterval(pollingHandle);
      onFailure(new Error(`AI分析轮询超时 (超过 ${timeout / 1000} 秒)`));
      return;
    }

    try {
      const report = await getReportDetails(reportId);

      // 检查 ai_analysis_status 字段
      if (report.ai_analysis_status === 'completed') {
        clearInterval(pollingHandle);
        onSuccess(report);
      } else if (report.ai_analysis_status === 'failed') {
        clearInterval(pollingHandle);
        onFailure(report); // 可以传递整个report对象，让调用者获取错误信息
      }
      // 如果状态是 'processing' 或 'not_started'，则继续轮询
    } catch (error) {
      clearInterval(pollingHandle);
      onFailure(error);
    }
  };

  // 立即执行一次，然后开始定时轮询
  poll();
  pollingHandle = window.setInterval(poll, interval);

  // 返回一个停止函数
  return () => {
    if (pollingHandle) {
      clearInterval(pollingHandle);
    }
  };
}