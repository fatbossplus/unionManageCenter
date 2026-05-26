import { get } from './request'

export interface DashboardStats {
  totalUsers: number
  activeOrgs: number
  monthlyRevenue: number
  pendingOrders: number
  onlineUsers: number
  todayRevenue: number
  todayNewUsers: number
}

export interface TrendPoint {
  date: string
  users: number
  revenue: number
}

export interface OrgTypeItem {
  name: string
  value: number
  color: string
}

export interface OrgRankItem {
  name: string
  revenue: number
}

export interface EventItem {
  id: string
  color: string
  text: string
  ts: number
}

export const getDashboardStats = async (): Promise<DashboardStats> => {
  const raw: any = await get('/dashboard/stats')
  return {
    totalUsers:     raw.total_users     ?? 0,
    activeOrgs:     raw.active_orgs     ?? 0,
    monthlyRevenue: raw.monthly_revenue ?? 0,
    pendingOrders:  raw.pending_orders  ?? 0,
    onlineUsers:    raw.online_users    ?? 0,
    todayRevenue:   raw.today_revenue   ?? 0,
    todayNewUsers:  raw.today_new_users ?? 0,
  }
}

export const getTrendData = (period: 'month' | 'quarter' | 'year') =>
  get<TrendPoint[]>('/dashboard/trend', { period })

export const getOrgTypeDistrib = () => get<OrgTypeItem[]>('/dashboard/org-types')
export const getOrgRank        = () => get<OrgRankItem[]>('/dashboard/org-rank')
export const getRealtimeEvents = async (): Promise<EventItem[]> => {
  const raw: any[] = await get('/dashboard/events') as any[] ?? []
  return raw.map(e => ({ id: String(e.id), color: e.color, text: e.text, ts: e.ts }))
}
