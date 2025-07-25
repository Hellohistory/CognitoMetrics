<template>
  <div class="home-container">
    <div class="welcome-card">
      <div class="welcome-header">
        <img src="/favicon.ico" alt="EduInsight Logo" class="logo" v-if="false" />
        <h1>欢迎使用 EduInsight 学情分析系统</h1>
      </div>

      <p class="tagline">
        系统准备就绪，请选择一项操作开始您的数据分析之旅。
      </p>

      <el-divider />

      <div class="action-prompt">
        <p>请通过上方导航开始您的工作。</p>
      </div>

      <div class="welcome-footer">
        <p>本项目为开源软件，您可以在以下平台找到本项目：</p>
        <div class="social-links">
          <a href="https://github.com/Hellohistory/EduInsight" target="_blank" rel="noopener noreferrer" title="GitHub">
            <img src="/github-mark.svg" alt="GitHub" />
          </a>
          <a href="https://gitee.com/Hellohistory/EduInsight" target="_blank" rel="noopener noreferrer" title="Gitee">
            <img src="/gitee-svgrepo-com.svg" alt="Gitee" />
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ElDivider } from 'element-plus';
</script>

<style scoped>
.home-container {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  min-height: calc(100vh - 120px);
  background-color: #f7f8fa;
  padding: 2rem;
  box-sizing: border-box;
}

.welcome-card {
  max-width: 800px;
  width: 100%;
  padding: 3rem 4rem;
  background-color: #ffffff;
  border-radius: 12px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.05);
  text-align: center;
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.welcome-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.08);
}

.welcome-header {
  margin-bottom: 1.5rem;
}

.logo {
  height: 60px;
  margin-bottom: 1rem;
}

h1 {
  font-size: 2.2rem;
  font-weight: 600;
  color: #303133;
  letter-spacing: 1px;
}

.tagline {
  font-size: 1.2rem;
  color: #606266;
  margin-bottom: 2rem;
}

.action-prompt {
  margin: 2rem 0;
  font-size: 1rem;
  color: #909399;
}

.quick-actions {
  margin-top: 1rem;
}

.welcome-footer {
  margin-top: 2rem;
  padding-top: 1.5rem;
  border-top: 1px solid #e4e7ed;
  font-size: 0.9rem;
  color: #a9acb3;
}

.social-links {
  margin-top: 1rem;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 2rem;
}

.social-links a {
  display: inline-block;
}

.social-links img {
  height: 32px;
  width: 32px;
  opacity: 0.6;
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.social-links a:hover img {
  opacity: 1;
  transform: scale(1.1);
}
</style>