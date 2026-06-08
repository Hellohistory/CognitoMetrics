<template>
  <div class="home-page app-page">
    <section class="home-hero page-header">
      <div>
        <p class="page-kicker">EduInsight Dashboard</p>
        <h1 class="page-title">学情数据工作台</h1>
        <p class="page-subtitle">
          汇总组织结构、考试录入和分析报告状态，进入系统后可以直接从当前任务继续。
        </p>
      </div>
      <div class="hero-actions">
        <el-button type="primary" :icon="EditPen" @click="router.push('/scores')">录入成绩</el-button>
        <el-button :icon="DataAnalysis" @click="router.push('/analysis')">查看报告</el-button>
      </div>
    </section>

    <section class="stat-strip">
      <div class="stat-card">
        <div>
          <div class="stat-label">年级</div>
          <div class="stat-value">{{ gradeCount }}</div>
        </div>
        <span class="stat-icon"><el-icon><CollectionTag /></el-icon></span>
      </div>
      <div class="stat-card">
        <div>
          <div class="stat-label">班级</div>
          <div class="stat-value">{{ classCount }}</div>
        </div>
        <span class="stat-icon"><el-icon><Grid /></el-icon></span>
      </div>
      <div class="stat-card">
        <div>
          <div class="stat-label">学生</div>
          <div class="stat-value">{{ studentCount }}</div>
        </div>
        <span class="stat-icon"><el-icon><UserFilled /></el-icon></span>
      </div>
      <div class="stat-card">
        <div>
          <div class="stat-label">考试</div>
          <div class="stat-value">{{ examStore.examList.length }}</div>
        </div>
        <span class="stat-icon"><el-icon><Memo /></el-icon></span>
      </div>
    </section>

    <section class="home-grid">
      <article
        v-for="action in quickActions"
        :key="action.route"
        class="quick-card panel-card"
        @click="router.push(action.route)"
      >
        <span class="quick-icon"><el-icon><component :is="action.icon" /></el-icon></span>
        <div>
          <h2>{{ action.title }}</h2>
          <p>{{ action.detail }}</p>
        </div>
      </article>

      <article class="panel-card recent-panel">
        <div class="panel-heading">
          <div>
            <p class="page-kicker">Recent Exams</p>
            <h2>最近考试</h2>
          </div>
          <el-button link type="primary" @click="router.push('/scores')">进入录入</el-button>
        </div>
        <div v-if="recentExams.length" class="recent-list">
          <div v-for="exam in recentExams" :key="exam.id" class="recent-row">
            <div>
              <strong>{{ exam.name }}</strong>
              <span>{{ formatDate(exam.exam_date) }}</span>
            </div>
            <el-tag :type="getExamTagType(exam.status)" effect="light">{{ getExamStatusText(exam.status) }}</el-tag>
          </div>
        </div>
        <el-empty v-else description="暂无考试" :image-size="88" />
      </article>

      <article class="panel-card recent-panel">
        <div class="panel-heading">
          <div>
            <p class="page-kicker">Reports</p>
            <h2>分析报告</h2>
          </div>
          <el-button link type="primary" @click="router.push('/analysis')">进入中心</el-button>
        </div>
        <div class="report-status-grid">
          <div v-for="item in reportStatusCards" :key="item.label" class="report-status">
            <span>{{ item.label }}</span>
            <strong>{{ item.value }}</strong>
          </div>
        </div>
      </article>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { ElButton, ElEmpty, ElIcon, ElTag } from 'element-plus';
import {
  CollectionTag,
  DataAnalysis,
  EditPen,
  Grid,
  Memo,
  TrendCharts,
  UserFilled,
} from '@element-plus/icons-vue';
import { useClassStore } from '@/stores/classStore';
import { useExamStore } from '@/stores/examStore';
import { getReports } from '@/api/analysisApi';
import type { IAnalysisReport, IExam } from '@/types/dataModels';

const router = useRouter();
const classStore = useClassStore();
const examStore = useExamStore();
const reports = ref<IAnalysisReport[]>([]);

const gradeCount = computed(() => classStore.classTree.length);
const classCount = computed(() => classStore.classTree.reduce((sum, grade) => sum + grade.classes.length, 0));
const studentCount = computed(() =>
  classStore.classTree.reduce((sum, grade) => sum + grade.classes.reduce((inner, cls) => inner + cls.student_count, 0), 0)
);
const recentExams = computed(() => examStore.examList.slice(0, 4));

const reportStatusCards = computed(() => [
  { label: '已完成', value: reports.value.filter(r => r.status === 'completed').length },
  { label: '处理中', value: reports.value.filter(r => r.status === 'processing').length },
  { label: '失败', value: reports.value.filter(r => r.status === 'failed').length },
]);

const quickActions = [
  {
    title: '学生管理',
    detail: '维护年级、班级和学生名单',
    route: '/students',
    icon: UserFilled,
  },
  {
    title: '成绩录入',
    detail: '创建考试、录入分数并定稿',
    route: '/scores',
    icon: EditPen,
  },
  {
    title: '分析中心',
    detail: '发起分析并查看报告结果',
    route: '/analysis',
    icon: TrendCharts,
  },
];

onMounted(async () => {
  await Promise.all([
    classStore.fetchClassTree(),
    examStore.fetchExams(),
    getReports({ page: 1, page_size: 50 }).then(result => {
      reports.value = result.items || [];
    }).catch(() => {
      reports.value = [];
    }),
  ]);
});

const getExamStatusText = (status: IExam['status']) => {
  const map: Record<IExam['status'], string> = {
    draft: '草稿',
    submitted: '排队中',
    processing: '处理中',
    completed: '已定稿',
    failed: '失败',
  };
  return map[status] || '未知';
};

const getExamTagType = (status: IExam['status']): 'success' | 'primary' | 'warning' | 'danger' | 'info' => {
  if (status === 'completed') return 'success';
  if (status === 'draft') return 'warning';
  if (status === 'failed') return 'danger';
  if (status === 'processing') return 'primary';
  return 'info';
};

const formatDate = (date?: string) => {
  if (!date) return '';
  return new Date(date).toLocaleDateString('zh-CN');
};
</script>

<style scoped>
.home-page {
  width: 100%;
  max-width: 1240px;
  margin: 0 auto;
}

.home-hero {
  align-items: center;
  overflow: hidden;
  position: relative;
}

.home-hero > div:first-child {
  min-width: 0;
  z-index: 1;
}

.home-hero::after {
  content: "";
  position: absolute;
  right: 28px;
  bottom: 18px;
  width: 220px;
  height: 120px;
  opacity: 0.11;
  background:
    linear-gradient(90deg, var(--app-primary) 1px, transparent 1px) 0 0 / 22px 22px,
    linear-gradient(0deg, var(--app-primary) 1px, transparent 1px) 0 0 / 22px 22px;
  pointer-events: none;
}

.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  z-index: 1;
}

.hero-actions :deep(.el-button + .el-button) {
  margin-left: 0;
}

.home-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

.quick-card {
  display: flex;
  gap: 14px;
  min-width: 0;
  min-height: 132px;
  padding: 18px;
  cursor: pointer;
  transition: transform 0.18s ease, border-color 0.18s ease, box-shadow 0.18s ease;
}

.quick-card:hover {
  transform: translateY(-2px);
  border-color: var(--app-border-strong);
  box-shadow: var(--app-shadow);
}

.quick-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 44px;
  width: 44px;
  height: 44px;
  border-radius: 8px;
  color: #fff;
  background: var(--app-primary);
}

.quick-card h2,
.recent-panel h2 {
  margin: 0;
  color: var(--app-text);
  font-size: 18px;
}

.quick-card p {
  margin: 8px 0 0;
  color: var(--app-text-muted);
  font-size: 14px;
}

.recent-panel {
  grid-column: span 3;
  padding: 18px;
}

.panel-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
}

.recent-list {
  display: grid;
  gap: 10px;
}

.recent-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  border: 1px solid var(--app-border);
  border-radius: 7px;
  background: var(--app-surface-soft);
}
.recent-row > div {
  min-width: 0;
}

.recent-row strong,
.recent-row span {
  display: block;
  overflow-wrap: anywhere;
}

.recent-row span {
  margin-top: 2px;
  color: var(--app-text-muted);
  font-size: 13px;
}

.report-status-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}

.report-status {
  padding: 14px;
  border: 1px solid var(--app-border);
  border-radius: 7px;
  background: var(--app-surface-soft);
}

.report-status span,
.report-status strong {
  display: block;
}

.report-status span {
  color: var(--app-text-muted);
  font-size: 13px;
}

.report-status strong {
  margin-top: 6px;
  font-size: 24px;
}

@media (max-width: 980px) {
  .home-grid {
    grid-template-columns: 1fr;
  }
  .recent-panel {
    grid-column: auto;
  }
}

@media (max-width: 640px) {
  .home-hero {
    align-items: stretch;
  }

  .home-hero::after {
    right: -56px;
    bottom: 16px;
    width: 200px;
  }

  .hero-actions,
  .report-status-grid {
    display: grid;
    grid-template-columns: 1fr;
  }
  .hero-actions {
    justify-items: center;
  }
  .hero-actions .el-button {
    width: min(160px, 100%);
  }
  .recent-row {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
