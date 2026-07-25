<template>
  <div class="kpi" :class="`kpi-${kpi.key}`">
    <div class="head">
      <span class="label">{{ kpi.label }}</span>
      <span class="icon" :style="{ background: iconBg }">{{ icon }}</span>
    </div>
    <div class="value-row">
      <span class="value num">{{ kpi.value }}</span>
    </div>
    <div class="delta">
      <span class="arrow" :class="kpi.delta_up ? 'up' : 'down'">
        {{ kpi.delta_up ? '▲' : '▼' }}
      </span>
      <span class="pct num" :class="kpi.delta_up ? 'up' : 'down'">
        {{ kpi.delta_pct }}%
      </span>
      <span class="desc">{{ kpi.description }}</span>
    </div>
    <div class="bar">
      <div class="bar-fill" :style="barStyle" />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  kpi: { type: Object, required: true }
})

// 每个 KPI 一套配色 + 图标,提升识别度
const META = {
  creators:   { icon: '👥', from: '#5EA1FF', to: '#3DD9EB' },
  followers:  { icon: '🫀', from: '#A07BFF', to: '#5EA1FF' },
  views:      { icon: '▶', from: '#3DD9EB', to: '#7DD96E' },
  engagement: { icon: '❤', from: '#FF6B6B', to: '#FFB547' },
  collabs:    { icon: '📄', from: '#FFB547', to: '#FF6B6B' },
  new:       { icon: '✨', from: '#7DD96E', to: '#3DD9EB' }
}

const icon = computed(() => META[props.kpi.key]?.icon || '📈')
const iconBg = computed(() => {
  const m = META[props.kpi.key]
  return m ? `linear-gradient(135deg, ${m.from}, ${m.to})` : '#888'
})
const barStyle = computed(() => {
  const m = META[props.kpi.key]
  return m ? { background: `linear-gradient(90deg, ${m.from}, ${m.to})`, width: '72%' } : {}
})
</script>

<style lang="scss" scoped>
.kpi {
  background: var(--bg-elev);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 8px 10px;
  position: relative;
  overflow: hidden;
  transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.3s;
  &:hover {
    transform: translateY(-2px);
    box-shadow: var(--shadow-2);
  }
  &::after {
    content: '';
    position: absolute;
    inset: 0;
    background: radial-gradient(400px 200px at 100% 0%, rgba(61, 217, 235, 0.05), transparent 60%);
    pointer-events: none;
  }
}
.head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  .label {
    font-size: 11px;
    color: var(--text-secondary);
  }
  .icon {
    width: 22px;
    height: 22px;
    border-radius: 6px;
    display: grid;
    place-items: center;
    color: #0B1020;
    font-weight: 700;
    font-size: 12px;
  }
}
.value-row {
  margin-top: 2px;
  .value {
    font-size: 20px;
    font-weight: 700;
    color: var(--text-primary);
    letter-spacing: -0.5px;
  }
}
.delta {
  margin-top: 2px;
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 10px;
  .arrow.up { color: var(--success); }
  .arrow.down { color: var(--danger); }
  .pct.up { color: var(--success); font-weight: 600; }
  .pct.down { color: var(--danger); font-weight: 600; }
  .desc { color: var(--text-muted); }
}
.bar {
  margin-top: 6px;
  height: 2px;
  background: var(--bg-elev-2);
  border-radius: 999px;
  overflow: hidden;
  .bar-fill {
    height: 100%;
    border-radius: 999px;
    animation: fill 1.2s ease-out;
  }
}
@keyframes fill {
  from { width: 0; }
}
</style>
