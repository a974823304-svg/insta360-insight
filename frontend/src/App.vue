<template>
  <!-- 整体布局:顶栏 + 左侧筛选 + 主内容 -->
  <div class="app-shell">
    <TopBar />
    <div class="app-body">
      <SideFilter />
      <main class="app-main">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup>
// 应用外壳,所有页面共享同一布局
// 路由只控制 <main> 内的内容
import TopBar from './components/TopBar.vue'
import SideFilter from './components/SideFilter.vue'
</script>

<style lang="scss" scoped>
.app-shell {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background:
    radial-gradient(1200px 600px at 80% -20%, rgba(61, 217, 235, 0.08), transparent 50%),
    radial-gradient(900px 500px at -10% 110%, rgba(160, 123, 255, 0.08), transparent 50%),
    var(--bg-base);
  overflow: hidden;
}

.app-body {
  display: flex;
  flex: 1;
  min-height: 0;
}

.app-main {
  flex: 1;
  min-width: 0;
  overflow: auto;
  padding: 12px 18px 14px;
}
</style>
