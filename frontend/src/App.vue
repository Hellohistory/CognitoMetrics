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
  height: 100vh;
}
.main-content {
  flex-grow: 1;
  padding: 2rem;
  overflow-y: auto;
}
</style>