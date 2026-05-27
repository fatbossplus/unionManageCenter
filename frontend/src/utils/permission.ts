/**
 * 权限工具函数
 * 直接从 Pinia store 读取，避免循环依赖
 */
import { useUserStore } from '@/stores/user'

/**
 * 判断当前登录管理员是否拥有指定权限码
 * - superadmin 角色始终返回 true
 * - 其他角色从 store.permissions[] 中查找
 */
export function hasPermission(code: string): boolean {
  const store = useUserStore()
  return store.hasPermission(code)
}

/**
 * 判断是否为超级管理员
 */
export function isSuperAdmin(): boolean {
  const store = useUserStore()
  return store.isSuperAdmin
}
