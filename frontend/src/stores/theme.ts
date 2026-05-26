import { defineStore } from 'pinia'
import { ref } from 'vue'

export type ThemeKey = 'a' | 'b' | 'c'

export const useThemeStore = defineStore('theme', () => {
  const current = ref<ThemeKey>(
    (uni.getStorageSync('theme') as ThemeKey) || 'b'
  )

  function apply(theme: ThemeKey) {
    current.value = theme
    uni.setStorageSync('theme', theme)
    // #ifdef H5
    document.documentElement.setAttribute('data-theme', theme)
    // #endif
  }

  function init() {
    apply(current.value)
  }

  return { current, apply, init }
})
