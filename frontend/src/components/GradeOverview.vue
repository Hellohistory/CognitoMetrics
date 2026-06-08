<template>
  <div class="grade-overview" v-if="grade">
    <header class="overview-header">
      <div>
        <p class="page-kicker">Grade Overview</p>
        <h2>{{ grade.name }} 年级概览</h2>
        <p>班级规模、入学年份和学生数量集中展示。</p>
      </div>
    </header>

    <section class="stat-strip overview-stats">
      <div class="stat-card">
        <div>
          <div class="stat-label">班级总数</div>
          <div class="stat-value">{{ grade.classes.length }}</div>
        </div>
        <span class="stat-icon"><el-icon><Grid /></el-icon></span>
      </div>
      <div class="stat-card">
        <div>
          <div class="stat-label">学生总人数</div>
          <div class="stat-value">{{ totalStudents }}</div>
        </div>
        <span class="stat-icon"><el-icon><UserFilled /></el-icon></span>
      </div>
    </section>

    <el-table :data="grade.classes" stripe class="class-table" empty-text="该年级下暂无班级">
      <el-table-column type="index" label="#" width="80" />
      <el-table-column prop="name" label="班级名称" sortable />
      <el-table-column prop="student_count" label="学生人数" sortable width="150" />
      <el-table-column prop="enrollment_year" label="入学年份" sortable width="150" />
      <el-table-column label="操作" width="150" align="center">
        <template #default="{ row }">
          <el-button type="primary" link @click="selectClass(row)">查看详情</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
  <div v-else class="empty-overview">
    <el-empty description="选择左侧年级或班级后查看数据" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { ElButton, ElEmpty, ElIcon, ElTable, ElTableColumn } from 'element-plus';
import { Grid, UserFilled } from '@element-plus/icons-vue';
import type { IGradeNode, IClassNode } from '@/types/dataModels';
import { useClassStore } from '@/stores/classStore';

const props = defineProps<{
  grade: IGradeNode | null;
}>();

const classStore = useClassStore();

const totalStudents = computed(() => {
  if (!props.grade) return 0;
  return props.grade.classes.reduce((sum, cls) => sum + cls.student_count, 0);
});

const selectClass = (classNode: unknown) => {
  // 调用 store 中新的、无歧义的 selectNode 方法，直接传递班级节点对象
  classStore.selectNode(classNode as IClassNode);
};
</script>

<style scoped>
.grade-overview {
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
  padding: 20px;
  overflow: auto;
}

.overview-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 18px;
  border: 1px solid var(--app-border);
  border-radius: var(--app-radius);
  background: var(--app-surface);
}

.overview-header h2 {
  margin: 0;
  color: var(--app-text);
  font-size: 24px;
}

.overview-header p:not(.page-kicker) {
  margin: 6px 0 0;
  color: var(--app-text-muted);
}

.overview-stats {
  grid-template-columns: repeat(2, minmax(180px, 1fr));
}

.class-table {
  flex: 1;
  min-height: 280px;
  border: 1px solid var(--app-border);
}

.empty-overview {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  min-height: 420px;
}

@media (max-width: 720px) {
  .overview-stats {
    grid-template-columns: 1fr;
  }
}
</style>
