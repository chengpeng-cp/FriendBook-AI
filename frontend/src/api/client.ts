/**
 * Axios 实例与统一拦截器
 *
 * 后端约定：HTTP 状态恒 200，业务成败看 body.code（0 成功）。
 * 拦截器职责：
 *   请求：自动附加 Bearer Token
 *   响应：解包 { code, message, data }，code !== 0 抛业务错误；
 *         Token 过期（20004/10002）自动登出并跳转登录页
 */
import axios, { AxiosError } from 'axios'
import type { ApiResponse } from '@/types'

export const API_BASE = import.meta.env.VITE_API_BASE_URL ?? ''

/** 业务错误（区别于网络错误，携带后端 code） */
export class ApiError extends Error {
  code: number
  constructor(message: string, code: number) {
    super(message)
    this.name = 'ApiError'
    this.code = code
  }
}

const client = axios.create({
  baseURL: API_BASE,
  timeout: 30_000,
})

client.interceptors.request.use((config) => {
  const token = localStorage.getItem('fb_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

client.interceptors.response.use(
  (res) => {
    // 二进制响应（CSV 下载等）直接透传
    if (res.config.responseType === 'blob') return res

    const body = res.data as ApiResponse
    if (body.code !== 0) {
      // Token 失效 / 未授权：清理本地态并跳登录
      if (body.code === 20004 || body.code === 10002) {
        localStorage.removeItem('fb_token')
        localStorage.removeItem('fb_user')
        if (!location.pathname.startsWith('/login')) {
          location.href = '/login'
        }
      }
      return Promise.reject(new ApiError(body.message || '请求失败', body.code))
    }
    return res
  },
  (err: AxiosError<ApiResponse>) => {
    // HTTP 层错误（后端宕机 / 网络断开）
    const msg = err.response?.data?.message ?? '网络异常，请稍后重试'
    return Promise.reject(new ApiError(msg, err.response?.status ?? -1))
  },
)

/** GET 请求并解包 data */
export async function get<T>(url: string, params?: object): Promise<T> {
  const res = await client.get<ApiResponse<T>>(url, { params })
  return res.data.data
}

/** POST 请求并解包 data */
export async function post<T>(url: string, body?: object): Promise<T> {
  const res = await client.post<ApiResponse<T>>(url, body)
  return res.data.data
}

/** PUT 请求并解包 data */
export async function put<T>(url: string, body?: object): Promise<T> {
  const res = await client.put<ApiResponse<T>>(url, body)
  return res.data.data
}

/** DELETE 请求并解包 data */
export async function del<T>(url: string): Promise<T> {
  const res = await client.delete<ApiResponse<T>>(url)
  return res.data.data
}

/** 只关心成功消息（data 为 null 的接口用） */
export async function postForMsg(url: string, body?: object): Promise<string> {
  const res = await client.post<ApiResponse<null>>(url, body)
  return res.data.message
}

export default client
