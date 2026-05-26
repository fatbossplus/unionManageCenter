import { defineStore } from 'pinia'
import { ref } from 'vue'

interface UserInfo {
  id: string
  username: string
  role: string
  email: string
}

export const useUserStore = defineStore('user', () => {
  const info = ref<UserInfo | null>(null)
  const token = ref<string>(uni.getStorageSync('token') || '')

  function setToken(t: string) {
    token.value = t
    uni.setStorageSync('token', t)
  }

  function setInfo(u: UserInfo) {
    info.value = u
  }

  function logout() {
    token.value = ''
    info.value = null
    uni.removeStorageSync('token')
    uni.reLaunch({ url: '/pages/login/index' })
  }

  return { info, token, setToken, setInfo, logout }
})
