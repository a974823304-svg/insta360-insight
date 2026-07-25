// 达人分析相关 API, 全部走 /api 代理(5173 -> 8080)
// 注意: 响应拦截器返回的是完整 { code, data } 信封, 这里统一拆出 data,
// 使 store 拿到的就是业务数组/对象(与 store 内部的数组校验一致)。
import request from './request'

const creator = {
  kpi:       (q) => request.get('/creator/kpi', { params: q }).then(r => r.data),
  trend:     (q) => request.get('/creator/trend', { params: q }).then(r => r.data),
  platforms: (q) => request.get('/creator/platforms', { params: q }).then(r => r.data),
  tracks:    (q) => request.get('/creator/tracks', { params: q }).then(r => r.data),
  audience:  (q) => request.get('/creator/audience', { params: q }).then(r => r.data),
  list:      (q) => request.get('/creator/list', { params: q }).then(r => r.data)
}

export default creator
