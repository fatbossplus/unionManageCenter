import { get, post, put, del } from './request'

export interface UserListParams {
  keyword?: string
  orgName?: string
  role?: string
  status?: string
  certStatus?: string
  source?: string
  startDate?: string
  endDate?: string
  inactive?: string
  activeMonth?: string
  page: number
  pageSize: number
}

export interface UserItem {
  id: string
  username: string
  email: string
  avatar?: string
  orgName: string
  role: string
  status: 'active' | 'pending' | 'disabled'
  certStatus: 'certified' | 'pending' | 'none'
  createdAt: string
  lastLoginAt: string
}

export interface PageResult<T> {
  list: T[]
  total: number
}

export const getUserList   = (params: UserListParams) =>
  get<PageResult<UserItem>>('/users', params as any)
export const getUserDetail = (id: string) => get<UserItem>(`/users/${id}`)
export const createUser    = (data: Partial<UserItem>) => post<UserItem>('/users', data as any)
export const updateUser    = (id: string, data: Partial<UserItem>) =>
  put<UserItem>(`/users/${id}`, data as any)
export const deleteUser    = (id: string) => del<void>(`/users/${id}`)
export const batchEnable   = (ids: string[]) => post<void>('/users/batch-enable', { ids })
export const batchDisable  = (ids: string[]) => post<void>('/users/batch-disable', { ids })
