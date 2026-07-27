// 前端演示用「合成真数据」
// ------------------------------------------------------------
// 形状等价于 backend/internal/service/source/douyin_adapter.go 的真实映射输出:
//   - kpi        来自 /data/external/user/overview/ (fans_total/play_total/digg+comment+share/collab_total)
//   - viewsTrend 来自 /data/external/user/play/      (list[date, views] -> 含 prev_views/ratio)
//
// 用途: GitHub Pages 静态托管、无后端时的演示数据。有凭证时, 同一套前端代码
// 通过后端适配器拿到的就是实时真数据(见 README「真实数据接入」+ 后端 httptest 映射测试)。
//
// 单账号视角(诚实边界): 抖音开放平台个人号授权 access_token 能拉到的就是
// 「你自己账号」的真实数据, 因此这里是一个真实量级账号的合理值, 不做多平台混排。
// 数值与 CP1 后端测试用的真实形状响应(fans_total=1234567 等)保持一致, 便于对照。

function buildDemoViewsTrend() {
  // 30 天播放序列, 与真实 play 接口 list[date,views] 等价, 含 prev_views / ratio。
  const days = 30
  const out = []
  const now = new Date()
  now.setHours(0, 0, 0, 0)
  const base = new Date(now.getTime() - (days - 1) * 86400000)
  const anomalyTag = { 16: '雪季开板, 滑雪内容自然爆发 +22%' }
  let views = 820000
  for (let i = 0; i < days; i++) {
    const d = new Date(base.getTime() + i * 86400000)
    const dow = d.getDay()
    const weekend = (dow === 0 || dow === 6) ? 260000 : 0
    const noise = ((i * 7 + 5) % 9) * 22000
    const v = views + weekend + noise
    const prev = i === 0 ? Math.round(v * 0.82) : Math.round(out[i - 1].views * 1.015)
    const mm = String(d.getMonth() + 1).padStart(2, '0')
    const dd = String(d.getDate()).padStart(2, '0')
    out.push({
      date: `${mm}-${dd}`,
      views: v,
      prev_views: prev,
      has_anomaly: !!anomalyTag[i],
      anomaly_tag: anomalyTag[i] || '',
      ratio: prev > 0 ? +(((v - prev) / prev) * 100).toFixed(2) : 0
    })
    views += 18000
  }
  return out
}

export default {
  // 与 douyin_adapter.mapOverviewToKpi 输出同构(creators/followers/views/engagement/collabs)
  kpi: [
    { key: 'creators',  label: '授权账号数', value: '1',        raw: 1,        delta_pct: 0,   delta_up: true,  unit: '',   description: '当前授权抖音账号' },
    { key: 'followers', label: '粉丝总量',   value: '123.5万',  raw: 1234567,  delta_pct: 4.8, delta_up: true,  unit: '万', description: '抖音平台' },
    { key: 'views',     label: '播放总量',   value: '4567.9万', raw: 45678901, delta_pct: 6.2, delta_up: true,  unit: '万', description: '累计播放' },
    { key: 'engagement',label: '互动总量',   value: '95.9万',   raw: 959257,   delta_pct: 5.1, delta_up: true,  unit: '万', description: '赞+评+转' },
    { key: 'collabs',   label: '商业合作数', value: '42',       raw: 42,       delta_pct: 12.0,delta_up: true,  unit: '',   description: '星图合作' }
  ],
  viewsTrend: buildDemoViewsTrend(),
  // 单账号视角: 平台分布即抖音 100%
  platformShare: [
    { platform: '抖音', share: 100, views: 45678901, color: '#FE2C55' }
  ],
  tracks: [
    { track: '滑雪', views: 15200000, color: '#5EA1FF' },
    { track: '冲浪', views: 11300000, color: '#3DD9EB' },
    { track: '骑行', views: 9600000,  color: '#A07BFF' },
    { track: '潜水', views: 7100000,  color: '#4DD0E1' },
    { track: '攀岩', views: 5300000,  color: '#7DD96E' }
  ],
  radar: [
    { dimension: '内容质量',   value: 89, avg: 78 },
    { dimension: '粉丝互动力', value: 82, avg: 70 },
    { dimension: '商业配合力', value: 76, avg: 72 },
    { dimension: '成活性',     value: 84, avg: 68 },
    { dimension: '成长力',     value: 88, avg: 73 }
  ],
  age: [
    { bucket: '18-24 岁',  share: 34.5, color: '#3DD9EB' },
    { bucket: '25-34 岁',  share: 41.8, color: '#FFB547' },
    { bucket: '35-44 岁',  share: 16.4, color: '#A07BFF' },
    { bucket: '45 岁以上', share: 7.3,  color: '#FF6B6B' }
  ],
  insights: [
    { icon: 'surge', title: '滑雪内容增长迅猛', body: '近 30 天滑雪相关播放增长 22%, 显著高于账号整体, 建议加大冬季内容投放。', severity: 'success' },
    { icon: 'star',  title: '粉丝互动活跃', body: '粉丝互动力雷达值 82, 高于同类均值 70, 商业配合潜力高。', severity: 'info' },
    { icon: 'alert', title: '关注内容节奏', body: '部分非周末时段播放偏低, 建议错峰发布并复用爆款选题。', severity: 'warning' }
  ],
  topCreators: [
    { rank: 1, avatar: '🎬', name: '我的抖音账号', platform: '抖音', followers: 1234567, total_views: 45678901, engagement: 7.8, growth_30d: 4.8, blacklist: false, explosive: 86.4, tags: ['#滑雪', '#极限', '#旅行'] }
  ],
  options: {
    regions: [
      { label: '北美', value: '北美' }, { label: '欧洲', value: '欧洲' },
      { label: '亚太', value: '亚太' }, { label: '全球', value: '全球' }
    ],
    tracks: [
      { label: '滑雪', value: '滑雪' }, { label: '冲浪', value: '冲浪' },
      { label: '骑行', value: '骑行' }, { label: '潜水', value: '潜水' },
      { label: '攀岩', value: '攀岩' }
    ],
    platforms: [
      { label: '抖音', value: '抖音' }
    ],
    ageBands: [
      { label: '18-24 岁', value: '18-24 岁' },
      { label: '25-34 岁', value: '25-34 岁' },
      { label: '35-44 岁', value: '35-44 岁' },
      { label: '45 岁以上', value: '45 岁以上' }
    ],
    presets: [
      { label: '近 7 天',  value: '7d' },
      { label: '近 30 天', value: '30d' },
      { label: '近 90 天', value: '90d' },
      { label: '本季度',   value: 'this_quarter' }
    ]
  }
}
