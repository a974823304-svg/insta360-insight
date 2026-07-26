// 数据洞察相关 API
// 全部走 /api 代理(前端 dev server 5173 -> 后端 8080)
import request from './request'

// querystring 序列化由 request.js 的自定义 paramsSerializer 处理:
// 数组参数展开为同名 key 重复(platforms=抖音&platforms=B站), 不带 [] 后缀,
// 与后端 gin c.QueryArray("platforms") 完美匹配。

const insight = {
  kpi: (q) => request.get('/kpi', { params: q }).then(r => r.data),
  viewsTrend: (q) => request.get('/views-trend', { params: q }).then(r => r.data),
  platformShare: (q) => request.get('/platform-distribution', { params: q }).then(r => r.data),
  tracks: (q) => request.get('/track-performance', { params: q }).then(r => r.data),
  radar: (q) => request.get('/explosive-radar', { params: q }).then(r => r.data),
  age: (q) => request.get('/audience-age', { params: q }).then(r => r.data),
  insights: (q) => request.get('/insights', { params: q }).then(r => r.data),
  topCreators: (q) => request.get('/top-creators', { params: q }).then(r => r.data),
  options: () => request.get('/filters/options').then(r => r.data)
}

export default insight
