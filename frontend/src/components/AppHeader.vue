<!--src/components/AppHeader.vue-->
<template>
  <header class="app-header">
    <div class="logo">EduInsight 分析系统</div>
    <nav>
      <router-link to="/">首页</router-link>
      <router-link to="/students">学生管理</router-link>
      <router-link to="/scores">成绩录入</router-link>
      <router-link to="/analysis">分析中心</router-link>
      </nav>
  </header>
</template>

<style scoped>
.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 2rem;
  height: 60px;
  background-color: #fff;
  border-bottom: 1px solid #e0e0e0;
  flex-shrink: 0;
}
.logo {
  font-size: 1.5rem;
  font-weight: bold;
}
nav a {
  margin-left: 1.5rem;
  text-decoration: none;
  color: #333;
  font-size: 1rem;
}
</style>