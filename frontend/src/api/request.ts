const BASE_URL = (import.meta as any).env?.VITE_API_BASE || '/api/v1'

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: Record<string, unknown>
}

export async function request<T>(url: string, options: RequestOptions = {}): Promise<T> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: BASE_URL + url,
      method: options.method || 'GET',
      data: options.data,
      header: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${uni.getStorageSync('token')}`,
      },
      success: (res: any) => {
        if (res.statusCode === 401) {
          uni.navigateTo({ url: '/pages/login/index' })
          reject(new Error('Unauthorized'))
          return
        }
        if (res.data?.code !== 0) {
          uni.showToast({ title: res.data?.message || '请求失败', icon: 'none' })
          reject(new Error(res.data?.message))
          return
        }
        resolve(res.data.data as T)
      },
      fail: (err: any) => {
        uni.showToast({ title: '网络异常', icon: 'none' })
        reject(err)
      },
    })
  })
}

export const get  = <T>(url: string, params?: Record<string, unknown>) =>
  request<T>(url, { method: 'GET', data: params })
export const post = <T>(url: string, data?: Record<string, unknown>) =>
  request<T>(url, { method: 'POST', data })
export const put  = <T>(url: string, data?: Record<string, unknown>) =>
  request<T>(url, { method: 'PUT', data })
export const del  = <T>(url: string) =>
  request<T>(url, { method: 'DELETE' })
