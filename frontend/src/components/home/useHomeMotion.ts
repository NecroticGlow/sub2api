import { computed, onBeforeUnmount, onMounted, ref, watch, type Ref } from 'vue'

/** Small, event-driven parallax; continuous movement stays in compositor-friendly CSS. */
export function useHomeMotion(root: Ref<HTMLElement | null>) {
  const paused = ref(false)
  const reduced = ref(false)
  const hidden = ref(false)
  const motionOff = computed(
    () => paused.value || reduced.value || hidden.value
  )
  let frame = 0
  let pointerX = 0
  let pointerY = 0
  let finePointer = false
  let observer: IntersectionObserver | undefined
  let preference: MediaQueryList | undefined

  function paint() {
    frame = 0
    const el = root.value
    if (!el) return
    const distance = Math.max(
      1,
      document.documentElement.scrollHeight - window.innerHeight
    )
    el.style.setProperty(
      '--page-progress',
      String(Math.min(1, window.scrollY / distance))
    )
    el.style.setProperty('--drift-x', `${motionOff.value ? 0 : pointerX}px`)
    el.style.setProperty('--drift-y', `${motionOff.value ? 0 : pointerY}px`)
    el.style.setProperty(
      '--scroll-drift',
      `${motionOff.value ? 0 : Math.min(window.scrollY * 0.12, 90)}px`
    )
  }
  function schedule() {
    if (!frame) frame = window.requestAnimationFrame(paint)
  }
  function pointer(event: PointerEvent) {
    if (!finePointer || motionOff.value) return
    pointerX = (event.clientX / window.innerWidth - 0.5) * 18
    pointerY = (event.clientY / window.innerHeight - 0.5) * 12
    schedule()
  }
  function leave() {
    pointerX = pointerY = 0
    schedule()
  }
  function preferenceChanged() {
    reduced.value = preference?.matches ?? false
  }
  function visibilityChanged() {
    hidden.value = document.hidden
  }
  function revealAll() {
    root.value
      ?.querySelectorAll('.reveal')
      .forEach((el) => el.classList.add('is-visible'))
  }
  watch(motionOff, (off) => {
    if (off) revealAll()
    schedule()
  })

  onMounted(() => {
    preference = window.matchMedia('(prefers-reduced-motion: reduce)')
    finePointer = window.matchMedia('(pointer: fine)').matches
    preferenceChanged()
    visibilityChanged()
    preference.addEventListener?.('change', preferenceChanged)
    window.addEventListener('scroll', schedule, { passive: true })
    window.addEventListener('resize', schedule, { passive: true })
    document.addEventListener('visibilitychange', visibilityChanged)
    root.value?.addEventListener('pointermove', pointer, { passive: true })
    root.value?.addEventListener('pointerleave', leave)
    if ('IntersectionObserver' in window && !motionOff.value) {
      observer = new IntersectionObserver(
        (entries) => {
          for (const entry of entries) {
            if (entry.isIntersecting) {
              entry.target.classList.add('is-visible')
              observer?.unobserve(entry.target)
            }
          }
        },
        { threshold: 0.08 }
      )
      root.value?.classList.add('has-observer')
      root.value
        ?.querySelectorAll('.reveal')
        .forEach((el) => observer?.observe(el))
    }
    paint()
  })
  onBeforeUnmount(() => {
    if (frame) window.cancelAnimationFrame(frame)
    observer?.disconnect()
    preference?.removeEventListener?.('change', preferenceChanged)
    window.removeEventListener('scroll', schedule)
    window.removeEventListener('resize', schedule)
    document.removeEventListener('visibilitychange', visibilityChanged)
    root.value?.removeEventListener('pointermove', pointer)
    root.value?.removeEventListener('pointerleave', leave)
  })
  return { paused, reduced, motionOff }
}
