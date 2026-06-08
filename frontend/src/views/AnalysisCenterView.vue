<template>
  <div class="analysis-center app-page">
    <header class="page-header">
      <div>
        <p class="page-kicker">Analysis Center</p>
        <h1 class="page-title">分析报告中心</h1>
        <p class="page-subtitle">查看报告进度、重试失败任务，并进入已完成报告。</p>
      </div>
      <el-button type="primary" :icon="Plus" @click="newAnalysisDialogVisible = true">
        发起新分析
      </el-button>
    </header>

    <section class="stat-strip">
      <div v-for="card in statusCards" :key="card.label" class="stat-card">
        <div>
          <div class="stat-label">{{ card.label }}</div>
          <div class="stat-value">{{ card.value }}</div>
        </div>
        <span class="stat-icon"><el-icon><component :is="card.icon" /></el-icon></span>
      </div>
    </section>

    <div class="toolbar-card">
      <div class="toolbar-right">
        <el-input
          v-model="searchQuery"
          placeholder="按报告名称搜索"
          clearable
          :prefix-icon="Search"
          class="search-input"
          @change="handleSearchChange"
        />
        <el-select
          v-model="statusFilter"
          placeholder="按状态筛选"
          clearable
          class="status-select"
        >
          <el-option label="分析完成" value="completed"></el-option>
          <el-option label="分析中" value="processing"></el-option>
          <el-option label="排队中" value="submitted"></el-option>
          <el-option label="分析失败" value="failed"></el-option>
        </el-select>
        <el-button @click="() => fetchReports()" :icon="Refresh" :loading="isLoading">刷新列表</el-button>
      </div>
    </div>

    <el-table :data="allReports" stripe v-loading="isLoading" class="report-table panel-card" row-key="id" empty-text="暂无分析报告">
      <el-table-column prop="report_name" label="报告名称" min-width="250">
        <template #default="{ row }">
            <span>{{ row.report_name }}</span>
            <el-tag v-if="row.exam && row.exam.status === 'draft'" type="warning" size="small" effect="light" style="margin-left: 8px;">
                成绩可编辑
            </el-tag>
          </template>
      </el-table-column>

      <el-table-column prop="status" label="状态" width="120">
        <template #default="{ row }">
          <el-tooltip
            :content="row.error_message"
            placement="top"
            :disabled="row.status !== 'failed' || !row.error_message"
          >
            <el-tag :type="getStatusType(row.status)" effect="light" round>
              {{ getStatusText(row.status) }}
            </el-tag>
          </el-tooltip>
        </template>
      </el-table-column>

      <el-table-column prop="created_at" label="创建时间" width="200">
        <template #default="{ row }">
          {{ formatDateTime(row.created_at) }}
        </template>
      </el-table-column>

      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button
            type="primary"
            link
            :icon="View"
            :disabled="row.status !== 'completed'"
            @click="viewReport(row.id)"
          >
            查看报告
          </el-button>

          <el-button
            v-if="row.status === 'failed'"
            type="warning"
            link
            :icon="Refresh"
            @click="handleRetry(row.id)"
          >
            重试
          </el-button>

          <el-button
            type="danger"
            link
            :icon="DeleteIcon"
            @click="handleDelete(row.id, row.report_name)"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-container">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :layout="paginationLayout"
        :small="isCompactViewport"
        :total="totalReports"
      />
    </div>

    <NewAnalysisDialog v-model="newAnalysisDialogVisible" @submitted="onAnalysisSubmitted" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onBeforeUnmount, onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import { ElIcon, ElTable, ElTableColumn, ElButton, ElTag, ElInput, ElSelect, ElOption, ElPagination, ElMessage, ElMessageBox, ElTooltip, vLoading } from 'element-plus';
import { CircleCheckFilled, Clock, Delete as DeleteIcon, Refresh, Plus, Search, View, WarningFilled } from '@element-plus/icons-vue';
import {
  getReports,
  deleteReport as apiDeleteReport,
  retryAnalysis as apiRetryReport, type ISubmitResponse, type IGetReportsParams,
} from '@/api/analysisApi';
import { startReportPolling } from '@/utils/pollingService';
import NewAnalysisDialog from '@/components/dialogs/NewAnalysisDialog.vue';
import type {IAnalysisReport} from "@/types/dataModels.ts";

const router = useRouter();
const allReports = ref<IAnalysisReport[]>([]);
const isLoading = ref(true);
const newAnalysisDialogVisible = ref(false);

const searchQuery = ref('');
const statusFilter = ref('');
const currentPage = ref(1);
const pageSize = ref(10);
const totalReports = ref(0);
const isCompactViewport = ref(false);
let compactQuery: MediaQueryList | null = null;

const statusCards = computed(() => [
  { label: '已完成', value: allReports.value.filter(item => item.status === 'completed').length, icon: CircleCheckFilled },
  { label: '处理中', value: allReports.value.filter(item => item.status === 'processing').length, icon: Clock },
  { label: '失败', value: allReports.value.filter(item => item.status === 'failed').length, icon: WarningFilled },
]);
const paginationLayout = computed(() =>
  isCompactViewport.value ? 'prev, pager, next' : 'total, sizes, prev, pager, next, jumper'
);

const fetchReports = async () => {
  isLoading.value = true;
  const params: IGetReportsParams = {
    page: currentPage.value,
    page_size: pageSize.value,
    query: searchQuery.value || undefined,
    status: statusFilter.value || undefined,
  };

  try {
    const response = await getReports(params);
    allReports.value = response.items || [];
    totalReports.value = response.total;
  } catch (error) {
    console.error("获取报告列表失败:", error);
    ElMessage.error('获取报告列表失败');
  } finally {
    isLoading.value = false;
  }
};

watch([currentPage, pageSize], fetchReports);

const handleSearchChange = () => {
    if (currentPage.value !== 1) {
        currentPage.value = 1;
    } else {
        fetchReports();
    }
};

watch(statusFilter, handleSearchChange);

const startPollingForReport = (reportId: number) => {
  startReportPolling({
    reportId,
    onSuccess: (finalReport) => {
      ElMessage.success(`报告 #${finalReport.id} 分析完成！`);
      fetchReports();
    },
    onFailure: (errorReport) => {
      const message = errorReport?.error_message || `报告 #${reportId} 分析失败`;
      ElMessage.error({ message: message, duration: 5000 });
      fetchReports();
    },
  });
};

const onAnalysisSubmitted = (response: ISubmitResponse) => {
  ElMessage.success(response.message);
  fetchReports();
  startPollingForReport(response.report_id);
};

const handleRetry = async (id: number) => {
  const reportToUpdate = allReports.value.find(r => r.id === id);
  if (reportToUpdate) {
    reportToUpdate.status = 'processing';
  }

  try {
    await apiRetryReport(id);
    ElMessage.info(`报告 #${id} 已重新提交，开始监控状态...`);
    startPollingForReport(id);
  } catch (error: any) {
    ElMessage.error(error.message || '重试任务失败');
    if (reportToUpdate) {
      reportToUpdate.status = 'failed';
    }
  }
};

const handleDelete = async (id:number, name: string) => {
  try {
    await ElMessageBox.confirm(`确定要删除报告 "${name}" 吗？此操作不可撤销。`, '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    });
    await apiDeleteReport(id);
    ElMessage.success('报告已删除');
    fetchReports();
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败');
    }
  }
};

const updateCompactViewport = () => {
  isCompactViewport.value = Boolean(compactQuery?.matches);
};

onMounted(() => {
  compactQuery = window.matchMedia('(max-width: 720px)');
  updateCompactViewport();
  compactQuery.addEventListener('change', updateCompactViewport);
  fetchReports();
});

onBeforeUnmount(() => {
  compactQuery?.removeEventListener('change', updateCompactViewport);
});

const getStatusType = (status: string): 'success' | 'primary' | 'warning' | 'danger' | 'info' => {
  switch (status) {
    case 'completed': return 'success';
    case 'processing': return 'primary';
    case 'submitted': return 'warning';
    case 'failed': return 'danger';
    default: return 'info';
  }
};

const getStatusText = (status: string) => {
    const map: Record<string, string> = {
        completed: '分析完成',
        processing: '分析中',
        submitted: '排队中',
        failed: '分析失败'
    }
    return map[status] || '未知';
}

const formatDateTime = (dateTimeString: string) => {
  if (!dateTimeString) return '';
  return new Date(dateTimeString).toLocaleString('zh-CN', { hour12: false });
};

const viewReport = (id: number) => {
  router.push({ name: 'report-view', params: { id } });
};
</script>


<style scoped>
.analysis-center {
  max-width: 1240px;
  margin: 0 auto;
}
.toolbar-right {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-left: auto;
}
.search-input {
  width: 240px;
}
.status-select {
  width: 150px;
}
.report-table {
  width: 100%;
  overflow: hidden;
}
.pagination-container {
  display: flex;
  justify-content: flex-end;
  padding: 4px 2px;
}

@media (max-width: 720px) {
  .toolbar-right,
  .search-input,
  .status-select {
    width: 100%;
  }
  .pagination-container {
    justify-content: center;
  }
}
</style>
