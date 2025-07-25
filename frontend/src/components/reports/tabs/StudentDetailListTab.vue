<template>
  <div class="student-detail-list-tab">
    <div class="table-toolbar">
      <el-input
        v-model="searchQuery"
        placeholder="按姓名或班级搜索学生"
        clearable
        class="search-input"
        :prefix-icon="Search"
      />
    </div>

    <el-table
      :data="filteredStudents"
      stripe
      border
      height="70vh"
      class="student-table"
      v-loading="reportStore.isLoading"
      empty-text="没有找到符合条件的学生"
    >
      <el-table-column type="index" label="#" width="55" fixed />
      <el-table-column prop="studentName" label="姓名" sortable width="110" fixed>
        <template #default="{ row }">
          <router-link :to="`/students/${row.studentId}`" class="student-link" v-if="row.studentId">
            {{ row.studentName }}
          </router-link>
          <span v-else>{{ row.studentName }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="tableName" label="班级" sortable width="120" />
      <el-table-column prop="totalScore" label="总分" sortable width="100">
        <template #default="{ row }">
          {{ row.totalScore.toFixed(2) }}
        </template>
      </el-table-column>
      <el-table-column prop="ranks.totalScore.gradeRank" label="年级排名" sortable width="120" />
      <el-table-column prop="ranks.totalScore.classRank" label="班级排名" sortable width="120" />
      <el-table-column label="画像" prop="profile" width="120" />
      <el-table-column
        label="T-Score (总)"
        prop="scores.tScores.totalScore"
        sortable
        width="140"
      />
      <el-table-column label="均衡指数" prop="metrics.imbalanceIndex" sortable width="120">
        <template #default="{ row }">
          {{ row.metrics.imbalanceIndex.toFixed(2) }}
        </template>
      </el-table-column>
      <el-table-column
        v-for="subject in individualSubjectKeys"
        :key="subject"
        :prop="`scores.rawScores.${subject}`"
        :label="subject"
        sortable
        width="100"
      />
    </el-table>
     <div class="table-footer">
        <span>共 {{ filteredStudents.length }} 条记录</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { RouterLink } from 'vue-router';
import { ElTable, ElTableColumn, ElInput, vLoading } from 'element-plus';
import { Search } from '@element-plus/icons-vue';
import { useAnalysisReportStore } from '@/stores/analysisReportStore';
import type { IStudentReportData } from '@/types/dataModels';

const reportStore = useAnalysisReportStore();
const searchQuery = ref('');

const individualSubjectKeys = computed<string[]>(() => {
  const fullMarks = reportStore.fullReport?.fullMarks;
  if (!fullMarks) return [];
  return Object.keys(fullMarks).filter(key => key !== 'totalScore');
});

const filteredStudents = computed<IStudentReportData[]>(() => {
  const query = searchQuery.value.trim().toLowerCase();
  if (!query) {
    return reportStore.allStudents;
  }
  return reportStore.allStudents.filter(student =>
    student.studentName.toLowerCase().includes(query) ||
    student.tableName.toLowerCase().includes(query)
  );
});
</script>

<style scoped>
.student-detail-list-tab {
  margin-top: 1rem;
}

.table-toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 1rem;
}

.search-input {
  width: 300px;
}

.student-table {
  width: 100%;
}

.student-link {
  color: #409eff;
  text-decoration: none;
  font-weight: 500;
  transition: color 0.2s;
}

.student-link:hover {
  color: #79bbff;
  text-decoration: underline;
}

.table-footer {
    display: flex;
    justify-content: flex-start;
    align-items: center;
    margin-top: 1rem;
    padding: 0 10px;
    color: #909399;
    font-size: 14px;
}
</style>