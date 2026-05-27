import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { get } from '@/api/request'

interface UserInfo {
  id: string
  username: string
  role: string
  email: string
  real_name?: string
  avatar?: string
}

export const useUserStore = defineStore('user', () => {
  const info       = ref<UserInfo | null>(null)
  const token      = ref<string>(uni.getStorageSync('token') || '')
  const permissions = ref<string[]>(
    JSON.parse(uni.getStorageSync('permissions') || '[]')
  )

  const isSuperAdmin = computed(() => info.value?.role === 'superadmin')

  function setToken(t: string) {
    token.value = t
    uni.setStorageSync('token', t)
  }

  function setInfo(u: UserInfo) {
    info.value = u
  }

  function setPermissions(codes: string[]) {
    permissions.value = codes
    uni.setStorageSync('permissions', JSON.stringify(codes))
  }

  /** 登录后调用：加载权限码列表 */
  async function loadPermissions() {
    try {
      const codes = await get('/auth/permissions') as string[]
      setPermissions(codes)
    } catch { /* ignore */ }
  }

  /** 判断是否有某权限（superadmin 始终返回 true） */
  function hasPermission(code: string): boolean {
    if (isSuperAdmin.value) return true
    return permissions.value.includes(code)
  }

  function logout() {
    token.value = ''
    info.value = null
    permissions.value = []
    uni.removeStorageSync('token')
    uni.removeStorageSync('permissions')
    uni.reLaunch({ url: '/pages/login/index' })
  }

  return { info, token, permissions, isSuperAdmin,
           setToken, setInfo, setPermissions, loadPermissions,
           hasPermission, logout }
})
