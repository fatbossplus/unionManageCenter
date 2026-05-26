import { get, post } from './request'
import type { PageResult } from './user'

export interface OrderListParams {
  keyword?: string
  type?: string
  status?: string
  payMethod?: string
  minAmount?: string
  maxAmount?: string
  startDate?: string
  endDate?: string
  page: number
  pageSize: number
}

export interface OrderItem {
  id: string
  orderNo: string
  type: string
  status: 'pending' | 'paid' | 'refunded' | 'cancelled'
  payMethod: string
  amount: number
  userName: string
  orgName: string
  createdAt: string
}

export const getOrderList   = (p: OrderListParams) => get<PageResult<OrderItem>>('/orders', p as any)
export const getOrderDetail = (id: string) => get<OrderItem>(`/orders/${id}`)
export const refundOrder    = (id: string, reason: string) => post<void>(`/orders/${id}/refund`, { reason })
