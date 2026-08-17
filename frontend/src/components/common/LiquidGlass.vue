<template>
  <div
    ref="rootRef"
    class="liquid-glass"
    :class="{ 'liquid-glass--clickable': clickable }"
    :style="rootStyle"
  >
    <svg v-if="displacementEnabled" class="liquid-glass__svg" aria-hidden="true" focusable="false">
      <defs>
        <filter
          :id="filterId"
          x="-35%"
          y="-35%"
          width="170%"
          height="170%"
          color-interpolation-filters="sRGB"
        >
          <feImage
            x="0"
            y="0"
            width="100%"
            height="100%"
            :href="displacementMap"
            preserveAspectRatio="xMidYMid slice"
            result="DISPLACEMENT_MAP"
          />
          <!-- Edge mask derived from the displacement map itself -->
          <feColorMatrix
            in="DISPLACEMENT_MAP"
            type="matrix"
            values="0.3 0.3 0.3 0 0 0.3 0.3 0.3 0 0 0.3 0.3 0.3 0 0 0 0 0 1 0"
            result="EDGE_INTENSITY"
          />
          <feComponentTransfer in="EDGE_INTENSITY" result="EDGE_MASK">
            <feFuncA type="discrete" :tableValues="edgeMaskTable" />
          </feComponentTransfer>
          <!-- Undisplaced source kept for the center area -->
          <feOffset in="SourceGraphic" dx="0" dy="0" result="CENTER_ORIGINAL" />
          <!-- Per-channel displacement for chromatic aberration -->
          <feDisplacementMap
            in="SourceGraphic"
            in2="DISPLACEMENT_MAP"
            :scale="redScale"
            xChannelSelector="R"
            yChannelSelector="B"
            result="RED_DISPLACED"
          />
          <feColorMatrix
            in="RED_DISPLACED"
            type="matrix"
            values="1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0"
            result="RED_CHANNEL"
          />
          <feDisplacementMap
            in="SourceGraphic"
            in2="DISPLACEMENT_MAP"
            :scale="greenScale"
            xChannelSelector="R"
            yChannelSelector="B"
            result="GREEN_DISPLACED"
          />
          <feColorMatrix
            in="GREEN_DISPLACED"
            type="matrix"
            values="0 0 0 0 0 0 1 0 0 0 0 0 0 0 0 0 0 0 1 0"
            result="GREEN_CHANNEL"
          />
          <feDisplacementMap
            in="SourceGraphic"
            in2="DISPLACEMENT_MAP"
            :scale="blueScale"
            xChannelSelector="R"
            yChannelSelector="B"
            result="BLUE_DISPLACED"
          />
          <feColorMatrix
            in="BLUE_DISPLACED"
            type="matrix"
            values="0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0 1 0"
            result="BLUE_CHANNEL"
          />
          <feBlend in="GREEN_CHANNEL" in2="BLUE_CHANNEL" mode="screen" result="GB_COMBINED" />
          <feBlend in="RED_CHANNEL" in2="GB_COMBINED" mode="screen" result="RGB_COMBINED" />
          <feGaussianBlur :stdDeviation="aberrationBlur" in="RGB_COMBINED" result="ABERRATED_BLURRED" />
          <feComposite in="ABERRATED_BLURRED" in2="EDGE_MASK" operator="in" result="EDGE_ABERRATION" />
          <feComponentTransfer in="EDGE_MASK" result="INVERTED_MASK">
            <feFuncA type="table" tableValues="1 0" />
          </feComponentTransfer>
          <feComposite in="CENTER_ORIGINAL" in2="INVERTED_MASK" operator="in" result="CENTER_CLEAN" />
          <feComposite in="EDGE_ABERRATION" in2="CENTER_CLEAN" operator="over" />
        </filter>
      </defs>
    </svg>

    <!-- Backdrop layer: refracts and frosts whatever sits behind the glass -->
    <span class="liquid-glass__backdrop" :style="backdropStyle"></span>
    <!-- Tint layer keeps content legible over busy backgrounds -->
    <span class="liquid-glass__tint" :style="tintStyle"></span>
    <!-- Specular border layers pick up the light around the cursor -->
    <span class="liquid-glass__border liquid-glass__border--screen" :style="borderScreenStyle"></span>
    <span class="liquid-glass__border liquid-glass__border--overlay" :style="borderOverlayStyle"></span>
    <!-- Hover / press glow for clickable glass -->
    <span v-if="clickable" class="liquid-glass__glow" :style="glowStyle"></span>

    <div class="liquid-glass__content">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import {
  liquidGlassDisplacementMap,
  nextLiquidGlassId,
  supportsLiquidGlassDisplacement
} from '@/utils/liquidGlass'

/**
 * Vue port of https://github.com/rdev/liquid-glass-react (Apple liquid glass).
 * Renders in normal document flow; size adapts to slot content.
 * On non-Chromium browsers it degrades to a plain frosted glass panel.
 */
const props = withDefaults(
  defineProps<{
    /** Intensity of the edge refraction (0 disables the SVG filter). */
    displacementScale?: number
    /** Backdrop blur in px. */
    blur?: number
    /** Backdrop saturation in %. */
    saturation?: number
    /** Chromatic aberration strength. */
    aberrationIntensity?: number
    /** 0 = rigid, higher = more "liquid" follow/stretch toward the cursor. */
    elasticity?: number
    /** Corner radius in px. */
    cornerRadius?: number
    /** Extra tint color painted inside the glass (defaults to theme-aware). */
    tint?: string
    /** Renders hover/press glow and press scale. */
    clickable?: boolean
    /** Track the cursor for elastic transform + specular light. */
    interactive?: boolean
  }>(),
  {
    displacementScale: 48,
    blur: 8,
    saturation: 150,
    aberrationIntensity: 2,
    elasticity: 0.12,
    cornerRadius: 16,
    tint: undefined,
    clickable: false,
    interactive: true
  }
)

const displacementMap = liquidGlassDisplacementMap
const filterId = nextLiquidGlassId()

const rootRef = ref<HTMLElement | null>(null)
const hovered = ref(false)
const pressed = ref(false)
const reducedMotion = ref(false)
const displacementSupported = ref(false)

// Cursor state (viewport coords + offset from element center, in % of size)
const mouseX = ref(0)
const mouseY = ref(0)
const offsetX = ref(0)
const offsetY = ref(0)
const hasMouse = ref(false)

const displacementEnabled = computed(
  () => displacementSupported.value && props.displacementScale > 0
)

// --- SVG filter params (ported 1:1 from liquid-glass-react, standard mode) ---
const edgeMaskTable = computed(() => `0 ${props.aberrationIntensity * 0.05} 1`)
const redScale = computed(() => props.displacementScale * -1)
const greenScale = computed(() => props.displacementScale * (-1 - props.aberrationIntensity * 0.05))
const blueScale = computed(() => props.displacementScale * (-1 - props.aberrationIntensity * 0.1))
const aberrationBlur = computed(() =>
  Math.max(0.1, 0.5 - props.aberrationIntensity * 0.1).toFixed(2)
)

// --- Elastic transform (activation zone fades the effect in near the glass) ---
const ACTIVATION_ZONE = 200

interface ElementMetrics {
  centerX: number
  centerY: number
  width: number
  height: number
}

function measure(): ElementMetrics | null {
  const el = rootRef.value
  if (!el) return null
  const rect = el.getBoundingClientRect()
  return {
    centerX: rect.left + rect.width / 2,
    centerY: rect.top + rect.height / 2,
    width: el.offsetWidth || rect.width,
    height: el.offsetHeight || rect.height
  }
}

function fadeInFactor(m: ElementMetrics): number {
  const edgeX = Math.max(0, Math.abs(mouseX.value - m.centerX) - m.width / 2)
  const edgeY = Math.max(0, Math.abs(mouseY.value - m.centerY) - m.height / 2)
  const edgeDistance = Math.sqrt(edgeX * edgeX + edgeY * edgeY)
  return edgeDistance > ACTIVATION_ZONE ? 0 : 1 - edgeDistance / ACTIVATION_ZONE
}

const elasticTransform = ref('')

function updateElasticTransform() {
  if (!trackingEnabled.value || props.elasticity <= 0) {
    elasticTransform.value = ''
    return
  }
  const m = measure()
  if (!m || !hasMouse.value) {
    elasticTransform.value = ''
    return
  }
  const fade = fadeInFactor(m)
  if (fade <= 0) {
    elasticTransform.value = ''
    return
  }
  const deltaX = mouseX.value - m.centerX
  const deltaY = mouseY.value - m.centerY
  const tx = deltaX * props.elasticity * 0.1 * fade
  const ty = deltaY * props.elasticity * 0.1 * fade

  const centerDistance = Math.sqrt(deltaX * deltaX + deltaY * deltaY)
  let scalePart = ''
  if (centerDistance > 0) {
    const nx = Math.abs(deltaX / centerDistance)
    const ny = Math.abs(deltaY / centerDistance)
    const stretch = Math.min(centerDistance / 300, 1) * props.elasticity * fade
    const scaleX = Math.max(0.8, 1 + nx * stretch * 0.3 - ny * stretch * 0.15)
    const scaleY = Math.max(0.8, 1 + ny * stretch * 0.3 - nx * stretch * 0.15)
    scalePart = ` scaleX(${scaleX.toFixed(4)}) scaleY(${scaleY.toFixed(4)})`
  }
  elasticTransform.value = `translate(${tx.toFixed(2)}px, ${ty.toFixed(2)}px)${scalePart}`
}

const trackingEnabled = computed(() => props.interactive && !reducedMotion.value)

let rafId = 0
function onMouseMove(e: MouseEvent) {
  if (!trackingEnabled.value) return
  mouseX.value = e.clientX
  mouseY.value = e.clientY
  hasMouse.value = true
  if (rafId) return
  rafId = requestAnimationFrame(() => {
    rafId = 0
    const m = measure()
    if (m && m.width > 0 && m.height > 0) {
      offsetX.value = ((mouseX.value - m.centerX) / m.width) * 100
      offsetY.value = ((mouseY.value - m.centerY) / m.height) * 100
    }
    updateElasticTransform()
  })
}

// --- Styles ---
const rootStyle = computed(() => {
  const transform = pressed.value && props.clickable ? 'scale(0.97)' : elasticTransform.value
  return {
    borderRadius: `${props.cornerRadius}px`,
    transform: transform || undefined,
    transition: pressed.value ? 'transform 0.1s ease-out' : 'transform 0.2s ease-out'
  }
})

const backdropStyle = computed(() => {
  const style: Record<string, string> = {
    backdropFilter: `blur(${props.blur}px) saturate(${props.saturation}%)`,
    WebkitBackdropFilter: `blur(${props.blur}px) saturate(${props.saturation}%)`
  }
  if (displacementEnabled.value) {
    style.filter = `url(#${filterId})`
  }
  return style
})

const tintStyle = computed(() => (props.tint ? { background: props.tint } : {}))

function borderGradient(baseA: number, baseB: number): string {
  const angle = 135 + offsetX.value * 1.2
  const stopA = Math.max(10, 33 + offsetY.value * 0.3)
  const stopB = Math.min(90, 66 + offsetY.value * 0.4)
  const alphaA = baseA + Math.abs(offsetX.value) * 0.008
  const alphaB = baseB + Math.abs(offsetX.value) * 0.012
  return `linear-gradient(${angle}deg,
    rgba(255, 255, 255, 0) 0%,
    rgba(255, 255, 255, ${alphaA.toFixed(3)}) ${stopA}%,
    rgba(255, 255, 255, ${alphaB.toFixed(3)}) ${stopB}%,
    rgba(255, 255, 255, 0) 100%)`
}

const borderScreenStyle = computed(() => ({
  background: borderGradient(0.12, 0.4)
}))

const borderOverlayStyle = computed(() => ({
  background: borderGradient(0.32, 0.6)
}))

const glowStyle = computed(() => ({
  opacity: pressed.value ? '0.55' : hovered.value ? '0.4' : '0'
}))

// --- Lifecycle ---
function onMouseEnter() {
  hovered.value = true
}
function onMouseLeave() {
  hovered.value = false
  pressed.value = false
}
function onMouseDown() {
  pressed.value = true
}
function onMouseUp() {
  pressed.value = false
}

onMounted(() => {
  displacementSupported.value = supportsLiquidGlassDisplacement()
  reducedMotion.value =
    typeof window !== 'undefined' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches

  window.addEventListener('mousemove', onMouseMove, { passive: true })
  const el = rootRef.value
  if (el) {
    el.addEventListener('mouseenter', onMouseEnter)
    el.addEventListener('mouseleave', onMouseLeave)
    el.addEventListener('mousedown', onMouseDown)
    el.addEventListener('mouseup', onMouseUp)
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('mousemove', onMouseMove)
  const el = rootRef.value
  if (el) {
    el.removeEventListener('mouseenter', onMouseEnter)
    el.removeEventListener('mouseleave', onMouseLeave)
    el.removeEventListener('mousedown', onMouseDown)
    el.removeEventListener('mouseup', onMouseUp)
  }
  if (rafId) cancelAnimationFrame(rafId)
})
</script>

<style scoped>
.liquid-glass {
  position: relative;
  isolation: isolate;
  overflow: hidden;
  box-shadow:
    0 8px 32px rgba(15, 23, 42, 0.1),
    0 2px 8px rgba(15, 23, 42, 0.06);
  will-change: transform;
}

:global(.dark) .liquid-glass {
  box-shadow:
    0 12px 40px rgba(0, 0, 0, 0.35),
    0 2px 10px rgba(0, 0, 0, 0.25);
}

.liquid-glass--clickable {
  cursor: pointer;
}

.liquid-glass__svg {
  position: absolute;
  width: 0;
  height: 0;
  pointer-events: none;
}

.liquid-glass__backdrop,
.liquid-glass__tint,
.liquid-glass__border,
.liquid-glass__glow {
  position: absolute;
  inset: 0;
  border-radius: inherit;
  pointer-events: none;
}

.liquid-glass__tint {
  background: rgba(255, 255, 255, 0.38);
}

:global(.dark) .liquid-glass__tint {
  background: rgba(15, 23, 42, 0.3);
}

.liquid-glass__border {
  padding: 1.5px;
  -webkit-mask:
    linear-gradient(#000 0 0) content-box,
    linear-gradient(#000 0 0);
  -webkit-mask-composite: xor;
  mask:
    linear-gradient(#000 0 0) content-box,
    linear-gradient(#000 0 0);
  mask-composite: exclude;
  box-shadow:
    0 0 0 0.5px rgba(255, 255, 255, 0.5) inset,
    0 1px 3px rgba(255, 255, 255, 0.25) inset,
    0 1px 4px rgba(0, 0, 0, 0.35);
}

.liquid-glass__border--screen {
  mix-blend-mode: screen;
  opacity: 0.25;
}

.liquid-glass__border--overlay {
  mix-blend-mode: overlay;
}

.liquid-glass__glow {
  background-image: radial-gradient(
    circle at 50% 0%,
    rgba(255, 255, 255, 0.9) 0%,
    rgba(255, 255, 255, 0) 60%
  );
  mix-blend-mode: overlay;
  transition: opacity 0.2s ease-out;
}

.liquid-glass__content {
  position: relative;
  z-index: 1;
}
</style>
