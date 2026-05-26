<script setup lang="ts">
import { onLaunch } from '@dcloudio/uni-app'
import { useThemeStore } from '@/stores/theme'

onLaunch(() => {
  useThemeStore().init()
})
</script>

<style lang="scss">
@import '@/styles/global.scss';
</style>
