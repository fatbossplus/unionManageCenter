export function countUp(
  from: number,
  to: number,
  duration: number,
  onUpdate: (value: number) => void
) {
  // #ifdef H5
  const start = performance.now()
  const step = (now: number) => {
    const elapsed = now - start
    const progress = Math.min(elapsed / duration, 1)
    const eased = 1 - Math.pow(1 - progress, 4)
    onUpdate(Math.floor(from + (to - from) * eased))
    if (progress < 1) requestAnimationFrame(step)
  }
  requestAnimationFrame(step)
  // #endif

  // #ifndef H5
  onUpdate(to)
  // #endif
}
