import { get, post } from './request'
import type { PageResult } from './user'

export interface MessageItem {
  id: string
  title: string
  content: string
  type: 'system' | 'order' | 'finance' | 'security'
  read: boolean
  createdAt: string
}

export const getMessageList = (p: { page: number; pageSize: number; read?: string }) =>
  get<PageResult<MessageItem>>('/messages', p as any)
export const markRead       = (ids: string[]) => post<void>('/messages/read', { ids })
export const markAllRead    = () => post<void>('/messages/read-all')
