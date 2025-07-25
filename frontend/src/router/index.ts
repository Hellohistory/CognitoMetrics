// src/router/index.ts
import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import StudentManagementView from '../views/StudentManagementView.vue'
import ScoreEntryView from '../views/ScoreEntryView.vue'
import ReportView from '../views/ReportView.vue'
import AnalysisCenterView from '../views/AnalysisCenterView.vue'
// 【新增】导入学生详情视图
import StudentDetailView from '../views/StudentDetailView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView
    },
    {
      path: '/students',
      name: 'student-management',
      component: StudentManagementView
    },
    {
        path: '/students/:id',
        name: 'student-detail',
        component: StudentDetailView,
        props: true, // 将路由参数:id作为props传递给组件
    },
    {
      path: '/scores',
      name: 'score-entry',
      component: ScoreEntryView
    },
    {
      path: '/analysis',
      name: 'analysis-center',
      component: AnalysisCenterView,
    },
    {
      path: '/reports/:id',
      name: 'report-view',
      component: ReportView,
      props: true
    }
  ]
})

export default router