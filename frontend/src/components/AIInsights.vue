<template>
  <div class="ai">
    <div class="head">
      <div class="title-row">
        <span class="badge">✨ AI 洞察</span>
        <span class="hint">基于当前筛选自动生成</span>
      </div>
    </div>
    <ul class="list">
      <li v-for="(it, i) in items" :key="i" class="item" :class="`sev-${it.severity}`">
        <div class="dot" />
        <div class="body">
          <div class="t">{{ it.title }}</div>
          <div class="d">{{ it.body }}</div>
        </div>
      </li>
    </ul>
    <el-link type="primary" :underline="false" class="more">
      查看全部洞察 →
    </el-link>
  </div>
</template>

<script setup>
defineProps({ items: { type: Array, default: () => [] } })
</script>

<style lang="scss" scoped>
.ai {
  background: var(--bg-elev);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 10px 12px;
  height: 100%;
  display: flex;
  flex-direction: column;
}
.head {
  margin-bottom: 8px;
  .title-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .badge {
    font-size: 11px;
    font-weight: 600;
    padding: 2px 7px;
    border-radius: 6px;
    background: linear-gradient(135deg, rgba(61, 217, 235, 0.18), rgba(160, 123, 255, 0.18));
    color: var(--brand);
    border: 1px solid rgba(61, 217, 235, 0.25);
  }
  .hint {
    font-size: 10px;
    color: var(--text-muted);
  }
}
.list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex: 1;
}
.item {
  display: flex;
  gap: 8px;
  padding: 7px 10px;
  background: var(--bg-elev-2);
  border-radius: 8px;
  border-left: 3px solid;
  .dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    margin-top: 5px;
    flex-shrink: 0;
  }
  .t {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-primary);
  }
  .d {
    font-size: 11px;
    color: var(--text-secondary);
    margin-top: 2px;
    line-height: 1.45;
  }
  &.sev-success {
    border-color: var(--success);
    .dot { background: var(--success); }
  }
  &.sev-info {
    border-color: var(--brand);
    .dot { background: var(--brand); }
  }
  &.sev-warning {
    border-color: var(--warning);
    .dot { background: var(--warning); }
  }
}
.more {
  margin-top: 8px;
  font-size: 11px;
  text-align: right;
}
</style>
