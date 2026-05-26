import { get, post } from './request'
import type { PageResult } from './user'

export interface FinanceListParams {
  keyword?: string
  status?: string
  accountType?: string
  period?: string
  minAmount?: string
  maxAmount?: string
  startDate?: string
  endDate?: string
  page: number
  pageSize: number
}

export interface FinanceItem {
  id: string
  orgName: string
  amount: number
  status: 'pending' | 'processing' | 'done'
  accountType: string
  accountNo: string
  period: string
  settledAt?: string
  createdAt: string
}

export const getFinanceList  = (p: FinanceListParams) => get<PageResult<FinanceItem>>('/finance', p as any)
export const settleFinance   = (id: string) => post<void>(`/finance/${id}/settle`)
