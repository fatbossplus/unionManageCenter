import { get, post, put, del } from './request'
import type { PageResult } from './user'

export interface OrgListParams {
  keyword?: string
  type?: string
  status?: string
  region?: string
  startDate?: string
  endDate?: string
  page: number
  pageSize: number
}

export interface OrgItem {
  id: string
  name: string
  type: string
  status: 'active' | 'pending' | 'frozen'
  region: string
  memberCount: number
  leader: string
  createdAt: string
}

const statusMap: Record<number, OrgItem['status']> = { 1: 'active', 2: 'pending', 3: 'frozen' }
const typeMap: Record<string, string> = { ec: '电商联盟', service: '服务联盟', content: '内容联盟', other: '其他' }

function normalizeOrg(raw: any): OrgItem {
  return {
    id: String(raw.id),
    name: raw.name,
    type: typeMap[raw.type] || raw.type,
    status: statusMap[raw.status] || 'active',
    region: raw.region || '',
    memberCount: raw.member_count ?? 0,
    leader: raw.leader?.real_name || raw.leader?.username || '',
    createdAt: raw.created_at?.slice(0, 10) || '',
  }
}

export const getOrgList = async (p: OrgListParams): Promise<PageResult<OrgItem>> => {
  const raw: any = await get('/orgs', p as any)
  return { list: (raw.list || []).map(normalizeOrg), total: raw.total ?? 0 }
}
export const getOrgDetail = (id: string) => get<OrgItem>(`/orgs/${id}`)
export const createOrg    = (data: Partial<OrgItem>) => post<OrgItem>('/orgs', data as any)
export const updateOrg    = (id: string, data: Partial<OrgItem>) => put<OrgItem>(`/orgs/${id}`, data as any)
export const deleteOrg    = (id: string) => del<void>(`/orgs/${id}`)
