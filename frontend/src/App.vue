<!--src/App.vue-->
<template>
  <div id="app-container">
    <app-header />
    <main class="main-content">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import AppHeader from '@/components/AppHeader.vue';
</script>

<style scoped>
#app-container {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: transparent;
}
.main-content {
  flex-grow: 1;
  width: 100%;
  max-width: 100vw;
  min-width: 0;
  padding: 24px;
  overflow-x: hidden;
  overflow-y: auto;
}

@media (max-width: 860px) {
  .main-content {
    padding: 14px;
  }
}
</style>
