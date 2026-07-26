// 品牌分析相关 API, 全部走 /api 代理(5173 -> 8080)
// 注意: 响应拦截器返回的是完整 { code, data } 信封, 这里统一拆出 data,
// 使 store 拿到的就是业务数组/对象(与 store 内部的数组校验一致)。
import request from './request'

const brand = {
  kpi:       (q) => request.get('/brand/kpi', { params: q }).then(r => r.data),
  trend:     (q) => request.get('/brand/trend', { params: q }).then(r => r.data),
  platforms: (q) => request.get('/brand/platforms', { params: q }).then(r => r.data),
  sentiment: (q) => request.get('/brand/sentiment', { params: q }).then(r => r.data),
  keywords:  (q) => request.get('/brand/keywords', { params: q }).then(r => r.data),
  list:      (q) => request.get('/brand/list', { params: q }).then(r => r.data)
}

export default brand
