// 内容分析相关 API, 全部走 /api 代理(5173 -> 8080)
// 注意: 响应拦截器返回的是完整 { code, data } 信封, 这里统一拆出 data,
// 使 store 拿到的就是业务数组/对象(与 store 内部的数组校验一致)。
import request from './request'

const content = {
  kpi:       (q) => request.get('/content/kpi', { params: q }).then(r => r.data),
  trend:     (q) => request.get('/content/trend', { params: q }).then(r => r.data),
  forms:     (q) => request.get('/content/forms', { params: q }).then(r => r.data),
  topics:    (q) => request.get('/content/topics', { params: q }).then(r => r.data),
  durations: (q) => request.get('/content/durations', { params: q }).then(r => r.data),
  list:      (q) => request.get('/content/list', { params: q }).then(r => r.data)
}

export default content
