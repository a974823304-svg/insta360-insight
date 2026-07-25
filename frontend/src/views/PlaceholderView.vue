<template>
  <div class="placeholder">
    <div class="card">
      <div class="title">{{ title }}</div>
      <div class="desc">该模块在 Demo 原型中暂未实现,可作为后续迭代入口。</div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
const route = useRoute()
const title = computed(() => route.meta?.title || '敬请期待')
</script>

<style lang="scss" scoped>
.placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
}
.card {
  background: var(--bg-elev);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 40px 56px;
  text-align: center;
  .title {
    font-size: 22px;
    font-weight: 700;
    background: linear-gradient(90deg, var(--brand) 0%, var(--brand-3) 100%);
    -webkit-background-clip: text;
    background-clip: text;
    color: transparent;
  }
  .desc {
    margin-top: 12px;
    color: var(--text-secondary);
  }
}
</style>
