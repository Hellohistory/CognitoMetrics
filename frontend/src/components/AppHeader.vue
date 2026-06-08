<!--src/components/AppHeader.vue-->
<template>
  <header class="app-header">
    <router-link to="/" class="brand" aria-label="EduInsight 首页">
      <span class="brand-mark">
        <el-icon><TrendCharts /></el-icon>
      </span>
      <span class="brand-copy">
        <strong>EduInsight</strong>
        <span>学情分析系统</span>
      </span>
    </router-link>

    <nav class="nav-links" aria-label="主导航">
      <router-link v-for="item in navItems" :key="item.to" :to="item.to" class="nav-link">
        <el-icon><component :is="item.icon" /></el-icon>
        <span class="nav-label-full">{{ item.label }}</span>
        <span class="nav-label-short">{{ item.shortLabel }}</span>
      </router-link>
    </nav>
  </header>
</template>

<script setup lang="ts">
import { DataAnalysis, EditPen, HomeFilled, TrendCharts, UserFilled } from '@element-plus/icons-vue';

const navItems = [
  { to: '/', label: '首页', shortLabel: '首页', icon: HomeFilled },
  { to: '/students', label: '学生管理', shortLabel: '学生', icon: UserFilled },
  { to: '/scores', label: '成绩录入', shortLabel: '成绩', icon: EditPen },
  { to: '/analysis', label: '分析中心', shortLabel: '分析', icon: DataAnalysis },
];
</script>

<style scoped>
.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
  width: 100%;
  max-width: 100vw;
  min-height: var(--app-header-height);
  padding: 10px 24px;
  background: rgba(255, 255, 255, 0.9);
  border-bottom: 1px solid var(--app-border);
  box-shadow: 0 8px 24px rgba(37, 51, 44, 0.05);
  backdrop-filter: blur(10px);
  flex-shrink: 0;
  position: sticky;
  top: 0;
  z-index: 20;
}
.brand {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  min-width: 220px;
  color: var(--app-text);
  text-decoration: none;
}
.brand-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border: 1px solid var(--app-border-strong);
  border-radius: 8px;
  color: #fff;
  background: linear-gradient(135deg, var(--app-primary), var(--app-success));
  box-shadow: 0 10px 22px rgba(36, 95, 115, 0.18);
}
.brand-copy {
  display: flex;
  flex-direction: column;
  line-height: 1.15;
}
.brand-copy strong {
  font-family: "Noto Serif SC", "Source Han Sans SC", serif;
  font-size: 20px;
  letter-spacing: 0;
}
.brand-copy span {
  color: var(--app-text-muted);
  font-size: 12px;
  font-weight: 600;
}
.nav-links {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px;
  border: 1px solid var(--app-border);
  border-radius: 10px;
  background: var(--app-surface-soft);
}
.nav-link {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  min-height: 36px;
  padding: 0 13px;
  border-radius: 7px;
  color: var(--app-text-muted);
  text-decoration: none;
  font-size: 14px;
  font-weight: 700;
  transition: color 0.18s ease, background 0.18s ease, box-shadow 0.18s ease;
}
.nav-link:hover {
  color: var(--app-primary);
  background: #fff;
}
.nav-link.router-link-active {
  color: var(--app-primary-strong);
  background: #fff;
  box-shadow: 0 6px 16px rgba(37, 51, 44, 0.08);
}
.nav-label-short {
  display: none;
}

@media (max-width: 860px) {
  .app-header {
    align-items: stretch;
    flex-direction: column;
    padding: 10px 14px;
    overflow: hidden;
  }
  .brand {
    min-width: 0;
  }
  .nav-links {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    width: 100%;
    max-width: 100%;
    gap: 3px;
    overflow: hidden;
  }
  .nav-link {
    justify-content: center;
    width: 100%;
    min-width: 0;
    gap: 3px;
    padding: 0 2px;
    overflow: hidden;
    font-size: 12px;
  }
  .nav-link .el-icon {
    flex: 0 0 auto;
  }
  .nav-link span {
    min-width: 0;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: clip;
  }
  .nav-label-full {
    display: none;
  }
  .nav-label-short {
    display: inline;
  }
}
</style>
