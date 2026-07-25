// 数据洞察相关 API
// 全部走 /api 代理(前端 dev server 5173 -> 后端 8080)
import request from './request'

// querystring 序列化时, axios 会自动把数组参数重复展开为同名 key
// 例如: { date_range: ['2024-01-01', '2024-01-31'] }
//        -> ?date_range=2024-01-01&date_range=2024-01-31
//        后端用 c.QueryArray 接收, 完美匹配

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
