<script setup lang="ts">
import { onMounted } from 'vue'
import { onLaunch } from '@dcloudio/uni-app'
import { useThemeStore } from '@/stores/theme'
import { injectIcons } from '@/assets/iconfont'

onLaunch(() => {
  useThemeStore().init()
  injectIcons()
})

onMounted(() => {
  injectIcons()
})
</script>

<style lang="scss">
@import '@/styles/global.scss';
</style>
