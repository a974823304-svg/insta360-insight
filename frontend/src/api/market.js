// 市场洞察相关 API, 全部走 /api 代理(5173 -> 8080)
// 注意: 响应拦截器返回的是完整 { code, data } 信封, 这里统一拆出 data,
// 使 store 拿到的就是业务数组/对象(与 store 内部的数组校验一致)。
import request from './request'

const market = {
  kpi:         (q) => request.get('/market/kpi', { params: q }).then(r => r.data),
  trend:       (q) => request.get('/market/trend', { params: q }).then(r => r.data),
  competitors: (q) => request.get('/market/competitors', { params: q }).then(r => r.data),
  regions:     (q) => request.get('/market/regions', { params: q }).then(r => r.data),
  prices:      (q) => request.get('/market/prices', { params: q }).then(r => r.data),
  list:        (q) => request.get('/market/list', { params: q }).then(r => r.data)
}

export default market
