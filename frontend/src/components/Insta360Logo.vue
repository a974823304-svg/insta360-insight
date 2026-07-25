<!--
  Insta360 品牌标识
  -----------------
  复刻 Insta360 官方品牌视觉:
    - 字母 "i" 上的小圆点被替换为一颗彩色渐变球(品牌签名)
    - "Insta360" 字标紧随其右
  颜色取自品牌官方色卡: 红 -> 橙 -> 青 的对角渐变。
  纯 SVG, 不依赖任何外部资源, 直接 inline 使用。
-->
<template>
  <svg
    :width="size"
    :height="size * (28 / 140)"
    viewBox="0 0 140 28"
    class="insta360-logo"
    aria-label="Insta360"
    role="img"
  >
    <defs>
      <linearGradient :id="gradId" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" stop-color="#FF3D5A" />
        <stop offset="50%" stop-color="#FFB547" />
        <stop offset="100%" stop-color="#3DD9EB" />
      </linearGradient>
    </defs>

    <!-- "I" 的主干 + 顶部彩色球(代替 i 的小圆点) -->
    <circle :cx="orbX" :cy="orbY" :r="orbR" :fill="`url(#${gradId})`" />
    <rect :x="orbX - 1" :y="orbY + orbR + 1" width="2" :height="stemH" fill="#ffffff" />

    <!-- "Insta360" 字标 -->
    <text
      x="14"
      y="22"
      font-family="-apple-system, BlinkMacSystemFont, 'PingFang SC', 'Microsoft YaHei', 'Segoe UI', Roboto, sans-serif"
      font-size="18"
      font-weight="700"
      letter-spacing="0.4"
      fill="#ffffff"
    >Insta360</text>
  </svg>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  size: { type: Number, default: 120 }
})

// 让 svg 内部坐标与外部像素尺寸解耦, 渲染始终清晰
const orbX = 5
const orbY = 8
const orbR = 5
const stemH = 10

// 同一页可能渲染多次, 给每个实例一个唯一的 gradient id, 避免互相覆盖
const gradId = computed(() => `insta360-grad-${Math.random().toString(36).slice(2, 9)}`)
</script>

<style scoped>
.insta360-logo {
  display: block;
  user-select: none;
}
</style>