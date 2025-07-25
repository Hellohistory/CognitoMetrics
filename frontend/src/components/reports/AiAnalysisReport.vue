<template>
  <div class="ai-analysis-container">
    <div v-if="isLoading" class="state-container loading-container">
      <el-progress type="circle" :percentage="100" status="success" :indeterminate="true" :duration="1" />
      <p>AI大脑正在高速运转，为您生成深度分析报告，请稍后...</p>
      <span>这通常需要1-3分钟，您可以先处理其他事务</span>
    </div>

    <div v-else-if="error" class="state-container error-container">
      <el-alert type="error" :title="error" show-icon :closable="false" />
      <el-button @click="startAnalysis" type="primary" style="margin-top: 20px;">
        <el-icon><Refresh /></el-icon>
        重试
      </el-button>
    </div>

    <div v-else-if="structuredResult" class="result-container">
      <el-tabs v-model="activeAiTab" type="border-card" class="ai-tabs">
        <el-tab-pane label="总体摘要" name="summary">
          <div v-html="renderMarkdown(structuredResult.summary)" class="markdown-body"></div>
        </el-tab-pane>
        <el-tab-pane label="横向对比" name="comparison">
           <div v-html="renderMarkdown(structuredResult.comparison)" class="markdown-body"></div>
        </el-tab-pane>
        <el-tab-pane label="各班诊断" name="diagnostics">
          <el-tabs v-model="activeClassTab" tab-position="left" class="class-diagnostic-tabs">
             <el-tab-pane
              v-for="(content, className) in structuredResult.diagnostics"
              :key="className"
              :label="className"
              :name="className"
             >
              <div v-html="renderMarkdown(content)" class="markdown-body"></div>
             </el-tab-pane>
          </el-tabs>
        </el-tab-pane>
      </el-tabs>
    </div>

    <div v-else class="state-container initial-state">
      <el-button @click="startAnalysis" type="primary" size="large" round>
        <el-icon style="margin-right: 8px;"><MagicStick /></el-icon>
        开始生成 AI 智能分析报告
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { ElButton, ElAlert, ElProgress, ElIcon, ElTabs, ElTabPane } from 'element-plus';
import { MagicStick, Refresh } from '@element-plus/icons-vue';
import { marked } from 'marked';
import DOMPurify from 'dompurify';
import { submitAiAnalysisTask, getReportDetails } from "@/api/analysisApi";
import { startAiAnalysisPolling } from '@/utils/pollingService';
import type { IAnalysisReportDetail } from '@/types/dataModels';

interface IStructuredAiResult {
  summary: string;
  comparison: string;
  diagnostics: Record<string, string>;
}

const props = defineProps<{
  reportId: number;
}>();

const isLoading = ref(false);
const error = ref<string | null>(null);
const structuredResult = ref<IStructuredAiResult | null>(null);
const activeAiTab = ref('summary');
const activeClassTab = ref('');

let stopPolling: (() => void) | null = null;

const renderMarkdown = (md: string) => {
  if (!md) return '';
  const rawHtml = marked.parse(md, { async: false }) as string;
  return DOMPurify.sanitize(rawHtml);
};

const parseAiResult = (cache: string | null): IStructuredAiResult | null => {
  if (!cache) return null;
  try {
    const parsed = JSON.parse(cache);
    if (parsed.summary && parsed.comparison && parsed.diagnostics) {
      return parsed;
    }
    return null;
  } catch (e) {
    console.error("AI result is not valid JSON, treating as plain text.", e);
    return { summary: cache, comparison: '无数据', diagnostics: {} };
  }
};

const handlePollingSuccess = (report: IAnalysisReportDetail) => {
  isLoading.value = false;
  structuredResult.value = parseAiResult(report.ai_analysis_cache);
  // 默认选中第一个班级
  if (structuredResult.value && Object.keys(structuredResult.value.diagnostics).length > 0) {
    activeClassTab.value = Object.keys(structuredResult.value.diagnostics)[0];
  }
};

const handlePollingFailure = (err?: any) => {
  isLoading.value = false;
  if (err && err.ai_analysis_status === 'failed') {
    error.value = 'AI分析任务执行失败，请检查报告或重试。';
  } else {
    error.value = err?.message || '获取AI分析结果时发生未知错误';
  }
};

async function startAnalysis() {
  isLoading.value = true;
  error.value = null;
  try {
    await submitAiAnalysisTask(props.reportId);
    stopPolling = startAiAnalysisPolling({
      reportId: props.reportId,
      onSuccess: handlePollingSuccess,
      onFailure: handlePollingFailure,
    });
  } catch (err: any) {
    isLoading.value = false;
    error.value = err.response?.data?.detail || '提交分析任务失败';
  }
}

onMounted(async () => {
  try {
    const report = await getReportDetails(props.reportId);
    if (report.ai_analysis_status === 'processing') {
      isLoading.value = true;
      stopPolling = startAiAnalysisPolling({
        reportId: props.reportId, onSuccess: handlePollingSuccess, onFailure: handlePollingFailure,
      });
    } else if (report.ai_analysis_status === 'completed') {
      structuredResult.value = parseAiResult(report.ai_analysis_cache);
      if (structuredResult.value && Object.keys(structuredResult.value.diagnostics).length > 0) {
         activeClassTab.value = Object.keys(structuredResult.value.diagnostics)[0];
      }
    } else if (report.ai_analysis_status === 'failed') {
      error.value = '此前的AI分析任务已失败，请重试。';
    }
  } catch (err: any) {
    error.value = err.response?.data?.detail || '加载报告状态失败';
  }
});

onUnmounted(() => {
  if (stopPolling) stopPolling();
});
</script>

<style scoped>
.ai-analysis-container {
  padding: 1rem 2rem;
  min-height: 60vh;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  background-color: #fff;
}
.state-container {
  text-align: center;
  margin-top: 10vh;
}
.loading-container p { margin-top: 1.5rem; color: #606266; font-size: 1.1rem; font-weight: 500; }
.loading-container span { margin-top: 0.5rem; color: #909399; font-size: 0.9rem; }

.result-container {
  width: 100%;
  text-align: left;
}
.ai-tabs {
  width: 100%;
  border: 1px solid #e4e7ed;
  box-shadow: none;
}
.class-diagnostic-tabs {
  min-height: 40vh;
}
.markdown-body {
  line-height: 1.8;
  font-size: 16px;
  color: #303133;
  padding: 0 1rem;
}
:deep(.markdown-body h1),:deep(.markdown-body h2),:deep(.markdown-body h3) { font-weight: 600; margin-top: 1.5em; margin-bottom: 0.8em; border-bottom: 1px solid #e4e7ed; padding-bottom: 0.5em; }
:deep(.markdown-body h1) { font-size: 1.8em; }
:deep(.markdown-body h2) { font-size: 1.5em; }
:deep(.markdown-body h3) { font-size: 1.25em; }
:deep(.markdown-body ul),:deep(.markdown-body ol) { padding-left: 2em; }
:deep(.markdown-body li) { margin-bottom: 0.5em; }
:deep(.markdown-body code) { background-color: rgba(64, 158, 255, 0.1); padding: 0.2em 0.4em; border-radius: 4px; font-size: 90%; color: #409eff; font-family: 'Courier New', Courier, monospace; }
:deep(.markdown-body blockquote) { border-left: 4px solid #409EFF; padding: 0.5em 1rem; color: #606266; margin-left: 0; background-color: #ecf5ff; }
</style>