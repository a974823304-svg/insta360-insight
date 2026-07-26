// 前端内嵌的兜底数据集
// ------------------------------------------------------------
// 当 Go 后端不可用(未启动/网络不通)时, store 会用这里的数据填充 UI,
// 保证 Demo 在任何环境下都能展示完整效果, 不会看到空白。
//
// 真实部署: 后端一定会启动, 这份数据仅作兜底。
// 与 backend/internal/mock/insight_data.go 保持字段一致(中文 key)。
export default {
  kpi: [
    { key: 'creators',  label: '达人数', value: '12,856', raw: 12856, delta_pct: 18.6, delta_up: true, unit: '', description: '较上期' },
    { key: 'followers', label: '总粉丝', value: '287.6M', raw: 287600000, delta_pct: 21.3, delta_up: true, unit: '', description: '较上期' },
    { key: 'views',     label: '总播放量', value: '2.38B', raw: 2380000000, delta_pct: 24.7, delta_up: true, unit: '', description: '较上期' },
    { key: 'engagement',label: '互动量', value: '46.7M', raw: 46700000, delta_pct: 19.8, delta_up: true, unit: '', description: '较上期' },
    { key: 'collabs',   label: '合作内容', value: '3,682', raw: 3682, delta_pct: 17.2, delta_up: true, unit: '', description: '较上期' }
  ],
  viewsTrend: buildViewsTrend(),
  platformShare: [
    { platform: '抖音',   share: 52.4, views: 1247000000, color: '#000000' },
    { platform: 'B站',    share: 26.8, views: 638000000,  color: '#00A1D6' },
    { platform: '小红书', share: 20.8, views: 495000000,  color: '#FF2442' }
  ],
  tracks: [
    { track: '滑雪', views: 878000000, color: '#5EA1FF' },
    { track: '冲浪', views: 642000000, color: '#3DD9EB' },
    { track: '骑行', views: 456000000, color: '#A07BFF' },
    { track: '潜水', views: 228000000, color: '#4DD0E1' },
    { track: '攀岩', views: 176000000, color: '#7DD96E' }
  ],
  radar: [
    { dimension: '内容质量',   value: 92, avg: 78 },
    { dimension: '粉丝互动力', value: 78, avg: 70 },
    { dimension: '商业配合力', value: 85, avg: 72 },
    { dimension: '成活性',     value: 81, avg: 68 },
    { dimension: '成长力',     value: 88, avg: 73 }
  ],
  age: [
    { bucket: '18-24 岁',  share: 31.2, color: '#3DD9EB' },
    { bucket: '25-34 岁',  share: 42.7, color: '#FFB547' },
    { bucket: '35-44 岁',  share: 17.3, color: '#A07BFF' },
    { bucket: '45 岁以上', share: 8.8,  color: '#FF6B6B' }
  ],
  insights: [
    { icon: 'surge', title: '冲浪赛道增长迅猛', body: '冲浪相关内容在近 30 天内播放量增长 38.7%,显著高于整体水平(24.7%),建议加大优选达人合作力度。', severity: 'success' },
    { icon: 'star',  title: '抖音表现突出', body: '抖音平台播放量占比提升至 52.4%, 同比增长 24.6%, 建议加大抖音渠道投放及头部达人合作。', severity: 'info' },
    { icon: 'alert', title: '黑马达人浮现', body: '粉丝数 < 5 万但播放量 > 30 万的内容创作者 +62%,具备较高投资价值,建议关注 #极限运动 标签下的新锐创作者。', severity: 'warning' }
  ],
  topCreators: [
    { rank: 1,  avatar: '🤿', name: 'Chris Burkard',  platform: '抖音',   followers: 1020000,  total_views: 186300000, engagement: 6.72,  growth_30d: 12.4, blacklist: false, explosive: 89.7, tags: ['#骑行','#极限','#摄影'] },
    { rank: 2,  avatar: '🚴', name: 'Sophie Laurent', platform: '小红书', followers: 326800,   total_views: 64200000,  engagement: 8.91,  growth_30d: 18.7, blacklist: false, explosive: 88.6, tags: ['#冲浪','#滑翔伞','#生活方式'] },
    { rank: 3,  avatar: '⛷️', name: 'Jake Wetter',    platform: '抖音',   followers: 48700,    total_views: 18600000,  engagement: 12.38, growth_30d: 42.3, blacklist: true,  explosive: 91.2, tags: ['#滑雪','#极限运动','#户外'] },
    { rank: 4,  avatar: '🏄', name: 'Marina Costa',   platform: 'B站',    followers: 890400,   total_views: 124500000, engagement: 9.65,  growth_30d: 15.1, blacklist: false, explosive: 86.4, tags: ['#冲浪','#海洋','#环保'] },
    { rank: 5,  avatar: '🧗', name: 'Liam Hoffmann',  platform: '抖音',   followers: 215600,   total_views: 38900000,  engagement: 7.43,  growth_30d: 9.8,  blacklist: false, explosive: 82.1, tags: ['#攀岩','#登山','#探险'] },
    { rank: 6,  avatar: '🏂', name: 'Yuki Tanaka',    platform: 'B站',    followers: 1540000,  total_views: 248700000, engagement: 8.12,  growth_30d: 21.6, blacklist: false, explosive: 87.9, tags: ['#滑雪','#单板','#冬季'] },
    { rank: 7,  avatar: '🚵', name: 'Felix Becker',   platform: '小红书', followers: 432900,   total_views: 71200000,  engagement: 6.85,  growth_30d: 6.4,  blacklist: false, explosive: 79.5, tags: ['#骑行','#公路车','#训练'] },
    { rank: 8,  avatar: '🤽', name: 'Aria Chen',      platform: '抖音',   followers: 678200,   total_views: 96500000,  engagement: 10.27, growth_30d: 28.4, blacklist: false, explosive: 88.2, tags: ['#潜水','#自由潜','#旅行'] },
    { rank: 9,  avatar: '🏃', name: 'Marco Rivera',   platform: '抖音',   followers: 312400,   total_views: 52300000,  engagement: 5.96,  growth_30d: 4.7,  blacklist: false, explosive: 76.8, tags: ['#越野跑','#训练','#营养'] },
    { rank: 10, avatar: '🎿', name: 'Elena Petrova',  platform: '小红书', followers: 198500,   total_views: 28700000,  engagement: 9.74,  growth_30d: 11.9, blacklist: false, explosive: 84.3, tags: ['#滑雪','#阿尔卑斯','#度假'] }
  ],
  creatorKpi: [
    { key: 'creators',  label: '达人总数', value: '20', raw: 20, delta_pct: 8.5, delta_up: true, description: '较上期' },
    { key: 'new',       label: '本月新增', value: '3', raw: 3, delta_pct: 50, delta_up: true, description: '较上期' },
    { key: 'followers', label: '平均粉丝', value: '1.82M', raw: 1820000, delta_pct: 4.2, delta_up: true, description: '较上期' },
    { key: 'engagement',label: '平均互动率', value: '6.8%', raw: 6.8, delta_pct: 0.6, delta_up: true, description: '较上期' },
    { key: 'collabs',   label: '合作占比', value: '45%', raw: 45, delta_pct: 12, delta_up: true, description: '较上期' }
  ],
  creatorTrend: buildCreatorTrend(),
  creatorPlatforms: [
    { platform: '抖音',   share: 48.2, views: 880000000, color: '#000000' },
    { platform: 'B站',    share: 30.5, views: 556000000, color: '#00A1D6' },
    { platform: '小红书', share: 21.3, views: 388000000, color: '#FF2442' }
  ],
  creatorTracks: [
    { track: '滑雪', views: 620000000, color: '#5EA1FF' },
    { track: '冲浪', views: 510000000, color: '#3DD9EB' },
    { track: '骑行', views: 430000000, color: '#A07BFF' },
    { track: '潜水', views: 360000000, color: '#4DD0E1' },
    { track: '攀岩', views: 308000000, color: '#7DD96E' }
  ],
  creatorAudience: {
    age: [
      { bucket: '18-24 岁',  share: 33.5, color: '#3DD9EB' },
      { bucket: '25-34 岁',  share: 41.2, color: '#FFB547' },
      { bucket: '35-44 岁',  share: 17.8, color: '#A07BFF' },
      { bucket: '45 岁以上', share: 7.5,  color: '#FF6B6B' }
    ],
    gender: [
      { gender: '男', share: 58.4, color: '#5EA1FF' },
      { gender: '女', share: 41.6, color: '#FF6B6B' }
    ]
  },
  creatorList: buildCreatorList(),
  // ===== 内容分析 =====
  contentKpi: [
    { key: 'total', label: '内容总数', value: '1,284', raw: 1284, delta_pct: 12.3, delta_up: true, description: '较上期' },
    { key: 'avg_views', label: '平均播放', value: '86.4K', raw: 86400, delta_pct: 5.1, delta_up: true, description: '较上期' },
    { key: 'hit_rate', label: '爆款率', value: '8.2%', raw: 8.2, delta_pct: 1.4, delta_up: true, description: '较上期' },
    { key: 'engage', label: '平均互动率', value: '7.3%', raw: 7.3, delta_pct: 0.3, delta_up: false, description: '较上期' },
    { key: 'freq', label: '周更新频次', value: '42', raw: 42, delta_pct: 6.0, delta_up: true, description: '较上期' }
  ],
  contentTrend: buildContentTrend(),
  contentForms: [
    { platform: '教程', share: 38.0, views: 920000, color: '#3DD9EB' },
    { platform: '测评', share: 27.0, views: 650000, color: '#5EA1FF' },
    { platform: '创意短片', share: 21.0, views: 510000, color: '#A07BFF' },
    { platform: '挑战赛', share: 14.0, views: 340000, color: '#7DD96E' }
  ],
  contentTopics: [
    { track: '滑雪', views: 720000, color: '#5EA1FF' },
    { track: '潜水', views: 610000, color: '#3DD9EB' },
    { track: '骑行', views: 530000, color: '#A07BFF' },
    { track: '旅行', views: 470000, color: '#7DD96E' },
    { track: 'Vlog', views: 390000, color: '#FFB547' }
  ],
  contentDurations: [
    { bucket: '≤30s', share: 34.0, color: '#3DD9EB' },
    { bucket: '30-60s', share: 41.0, color: '#5EA1FF' },
    { bucket: '1-3min', share: 18.0, color: '#A07BFF' },
    { bucket: '3min+', share: 7.0, color: '#7DD96E' }
  ],
  contentList: buildContentList(),
  // ===== 市场洞察 =====
  marketKpi: [
    { key: 'size', label: '品类规模', value: '¥3.2B', raw: 3200000000, delta_pct: 9.8, delta_up: true, description: '较上期' },
    { key: 'growth', label: '品类增速', value: '14.6%', raw: 14.6, delta_pct: 2.1, delta_up: true, description: '较上期' },
    { key: 'share', label: 'Insta360 市占', value: '31.2%', raw: 31.2, delta_pct: 3.4, delta_up: true, description: '较上期' },
    { key: 'comp', label: '在榜竞品数', value: '6', raw: 6, delta_pct: 0, delta_up: true, description: '较上期' },
    { key: 'buzz', label: '行业声量', value: '4.7M', raw: 4700000, delta_pct: 7.5, delta_up: true, description: '较上期' }
  ],
  marketTrend: buildMarketTrend(),
  marketCompetitors: [
    { platform: 'Insta360', share: 31.2, views: 1460000, color: '#3DD9EB' },
    { platform: 'GoPro', share: 28.4, views: 1330000, color: '#FF6B6B' },
    { platform: 'DJI', share: 24.7, views: 1150000, color: '#5EA1FF' },
    { platform: '其他', share: 15.7, views: 730000, color: '#A07BFF' }
  ],
  marketRegions: [
    { track: '华东', views: 1520000, color: '#3DD9EB' },
    { track: '华南', views: 1180000, color: '#5EA1FF' },
    { track: '华北', views: 960000, color: '#A07BFF' },
    { track: '西南', views: 640000, color: '#7DD96E' },
    { track: '海外', views: 1300000, color: '#FFB547' }
  ],
  marketPrices: [
    { bucket: '<1000', share: 22.0, color: '#3DD9EB' },
    { bucket: '1000-3000', share: 41.0, color: '#5EA1FF' },
    { bucket: '3000-5000', share: 26.0, color: '#A07BFF' },
    { bucket: '5000+', share: 11.0, color: '#7DD96E' }
  ],
  marketList: [
    { name: 'Insta360', category: '全景/运动相机', buzz: 1460000, growth: 17.2, sentiment: 81.0 },
    { name: 'GoPro', category: '运动相机', buzz: 1330000, growth: 4.1, sentiment: 73.0 },
    { name: 'DJI', category: '无人机/相机', buzz: 1150000, growth: 9.8, sentiment: 76.0 },
    { name: 'Sony', category: '微单', buzz: 880000, growth: 6.3, sentiment: 79.0 },
    { name: '大疆Action', category: '运动相机', buzz: 620000, growth: 12.5, sentiment: 74.0 },
    { name: 'AKASO', category: '入门运动相机', buzz: 340000, growth: 21.0, sentiment: 68.0 }
  ],
  // ===== 品牌分析 =====
  brandKpi: [
    { key: 'buzz', label: '品牌声量', value: '2.9M', raw: 2900000, delta_pct: 11.2, delta_up: true, description: '较上期' },
    { key: 'sent', label: '好感度', value: '81%', raw: 81, delta_pct: 1.8, delta_up: true, description: '较上期' },
    { key: 'partners', label: '合作品牌数', value: '8', raw: 8, delta_pct: 14.3, delta_up: true, description: '较上期' },
    { key: 'roi', label: '内容互动 ROI', value: '4.2', raw: 4.2, delta_pct: 0.5, delta_up: true, description: '较上期' },
    { key: 'search', label: '搜索指数', value: '68.5', raw: 68.5, delta_pct: 3.2, delta_up: false, description: '较上期' }
  ],
  brandTrend: buildBrandTrend(),
  brandPlatforms: [
    { platform: '抖音', share: 44.0, views: 1280000, color: '#000000' },
    { platform: 'B站', share: 26.0, views: 754000, color: '#00A1D6' },
    { platform: '小红书', share: 22.0, views: 638000, color: '#FF2442' },
    { platform: '微博', share: 8.0, views: 232000, color: '#E6162D' }
  ],
  brandSentiment: [
    { bucket: '正面', share: 67.0, color: '#7DD96E' },
    { bucket: '中性', share: 24.0, color: '#FFB547' },
    { bucket: '负面', share: 9.0, color: '#FF6B6B' }
  ],
  brandKeywords: [
    { word: '画质', weight: 92.0 }, { word: '防抖', weight: 88.0 },
    { word: '全景', weight: 81.0 }, { word: '运动相机', weight: 76.0 },
    { word: 'Vlog', weight: 70.0 }, { word: '旅行', weight: 64.0 },
    { word: '性价比', weight: 58.0 }, { word: '续航', weight: 49.0 }
  ],
  brandList: [
    { name: '红牛', industry: '饮料', contents: 8, exposure: 5400000, engagement: 320000, roi: 4.2 },
    { name: '始祖鸟', industry: '户外', contents: 6, exposure: 4100000, engagement: 260000, roi: 3.9 },
    { name: '携程', industry: '旅行', contents: 5, exposure: 3300000, engagement: 210000, roi: 3.5 },
    { name: '迪卡侬', industry: '运动', contents: 7, exposure: 2900000, engagement: 190000, roi: 3.1 },
    { name: '大疆', industry: '无人机', contents: 4, exposure: 2600000, engagement: 175000, roi: 3.8 },
    { name: '索尼', industry: '影像', contents: 3, exposure: 1800000, engagement: 120000, roi: 3.3 },
    { name: 'Keep', industry: '健身', contents: 5, exposure: 1500000, engagement: 98000, roi: 2.9 },
    { name: '小米', industry: '科技', contents: 2, exposure: 980000, engagement: 64000, roi: 2.6 }
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
      { label: '抖音', value: '抖音' },
      { label: 'B站', value: 'B站' },
      { label: '小红书', value: '小红书' }
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

// 生成 30 天的确定性趋势(与 Go 后端 TrendGenerator 等价)
function buildViewsTrend() {
  const days = 30
  const out = []
  const now = new Date()
  now.setHours(0, 0, 0, 0)
  const base = new Date(now.getTime() - (days - 1) * 86400000)
  const anomalyTag = {
    8:  '抖音热门话题 #citysurf 推升当日播放量 +28%',
    17: '雪季开板,滑雪赛道自然爆发 +18%',
    23: 'B站推荐算法调整, 长视频流量回升 +12%'
  }
  for (let i = 0; i < days; i++) {
    const d = new Date(base.getTime() + i * 86400000)
    const dow = d.getDay()
    const trend = 72_000_000 + i * 1_200_000
    const weekend = (dow === 0 || dow === 6) ? 18_000_000 : 0
    const noise = ((i * 7 + 13) % 9) * 1_500_000
    const views = trend + weekend + noise
    const prev = Math.round(views * 0.78)
    const mm = String(d.getMonth() + 1).padStart(2, '0')
    const dd = String(d.getDate()).padStart(2, '0')
    const point = {
      date: `${mm}-${dd}`,
      views,
      prev_views: prev,
      has_anomaly: !!anomalyTag[i],
      anomaly_tag: anomalyTag[i] || '',
      ratio: prev > 0 ? +(((views - prev) / prev) * 100).toFixed(2) : 0
    }
    out.push(point)
  }
  return out
}

// 达人分析: 近 30 天累计粉丝趋势(M 量级, 与 Go CreatorTrend 等价)
function buildCreatorTrend() {
  const days = 30
  const out = []
  const now = new Date()
  now.setHours(0, 0, 0, 0)
  const base = new Date(now.getTime() - (days - 1) * 86400000)
  for (let i = 0; i < days; i++) {
    const d = new Date(base.getTime() + i * 86400000)
    const views = 180_000_000 + i * 180_000 + ((i * 5) % 9) * 40_000
    const prev = 180_000_000 + i * 150_000
    const mm = String(d.getMonth() + 1).padStart(2, '0')
    const dd = String(d.getDate()).padStart(2, '0')
    out.push({ date: `${mm}-${dd}`, views, prev_views: prev })
  }
  return out
}

// 达人分析: 20 个达人(与 Go CreatorList 等价)
function buildCreatorList() {
  const seeds = [
    ['Chris Burkard', '🤿'], ['Sophie Laurent', '🚴'], ['Jake Wetter', '⛷️'],
    ['Marina Costa', '🏄'], ['Liam Hoffmann', '🧗'], ['Yuki Tanaka', '🏂'],
    ['Felix Becker', '🚵'], ['Aria Chen', '🤽'], ['Marco Rivera', '🏃'],
    ['Elena Petrova', '🎿'], ['Noah Kim', '🏄'], ['Mia Wong', '🚴'],
    ['Leo Schmidt', '⛷️'], ['Zoe Martin', '🧗'], ['Kenji Sato', '🏂'],
    ['Lara Lopez', '🤽'], ['Owen Brooks', '🏃'], ['Nina Roth', '🎿'],
    ['Pablo Cruz', '🚵'], ['Sara Lind', '🤿']
  ]
  const platforms = ['抖音', 'B站', '小红书']
  const tracks = ['滑雪', '冲浪', '骑行', '潜水', '攀岩']
  return seeds.map((s, i) => {
    const followers = 80_000 + ((i * 37_000) % 1_500_000)
    return {
      rank: i + 1,
      avatar: s[1],
      name: s[0],
      platform: platforms[i % 3],
      followers,
      total_views: followers * (3 + (i % 5)),
      engagement: +(5.0 + (i % 6) + (i % 3) * 0.4).toFixed(2),
      growth_30d: +(((i * 7) % 45) - 5.0).toFixed(1),
      blacklist: i === 2,
      explosive: +(70.0 + (i % 25) + (i % 4) * 1.5).toFixed(1),
      tags: ['#' + tracks[i % tracks.length], '#极限']
    }
  })
}

// 内容分析: 近 30 天内容播放量趋势
function buildContentTrend() {
  const days = 30, out = [], now = new Date()
  now.setHours(0, 0, 0, 0)
  const base = new Date(now.getTime() - (days - 1) * 86400000)
  for (let i = 0; i < days; i++) {
    const d = new Date(base.getTime() + i * 86400000)
    const views = 2400000 + i * 120000 + ((i * 7) % 11) * 30000
    const prev = 2400000 + i * 95000
    const mm = String(d.getMonth() + 1).padStart(2, '0')
    const dd = String(d.getDate()).padStart(2, '0')
    out.push({ date: `${mm}-${dd}`, views, prev_views: prev })
  }
  return out
}

// 内容分析: 15 条爆款内容（注意 isHit 驼峰, 与 Go ContentItem json 标签 / 视图 formatter 一致）
function buildContentList() {
  const forms = ['教程', '测评', '创意短片', '挑战赛']
  const topics = ['滑雪', '潜水', '骑行', '旅行', 'Vlog']
  return Array.from({ length: 15 }, (_, i) => ({
    id: i + 1,
    title: topics[i % topics.length] + '第' + (i + 1) + '期',
    form: forms[i % forms.length],
    topic: topics[i % topics.length],
    views: 50000 + ((i * 53000) % 1200000),
    engagement: +(5.0 + (i % 9) + (i % 4) * 0.6).toFixed(2),
    isHit: i % 3 === 0
  }))
}

// 市场洞察: 近 30 天品类声量趋势
function buildMarketTrend() {
  const days = 30, out = [], now = new Date()
  now.setHours(0, 0, 0, 0)
  const base = new Date(now.getTime() - (days - 1) * 86400000)
  for (let i = 0; i < days; i++) {
    const d = new Date(base.getTime() + i * 86400000)
    const views = 140000 + i * 9000 + ((i * 5) % 9) * 2000
    const prev = 140000 + i * 7000
    const mm = String(d.getMonth() + 1).padStart(2, '0')
    const dd = String(d.getDate()).padStart(2, '0')
    out.push({ date: `${mm}-${dd}`, views, prev_views: prev })
  }
  return out
}

// 品牌分析: 近 30 天品牌声量趋势
function buildBrandTrend() {
  const days = 30, out = [], now = new Date()
  now.setHours(0, 0, 0, 0)
  const base = new Date(now.getTime() - (days - 1) * 86400000)
  for (let i = 0; i < days; i++) {
    const d = new Date(base.getTime() + i * 86400000)
    const views = 90000 + i * 6000 + ((i * 4) % 7) * 1500
    const prev = 90000 + i * 5000
    const mm = String(d.getMonth() + 1).padStart(2, '0')
    const dd = String(d.getDate()).padStart(2, '0')
    out.push({ date: `${mm}-${dd}`, views, prev_views: prev })
  }
  return out
}
