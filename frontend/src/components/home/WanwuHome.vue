<template>
  <div
    id="home-top"
    ref="root"
    class="wanwu-home"
    :class="{ 'motion-off': motionOff, 'locale-en': language === 'en' }"
  >
    <a class="skip-link" href="#home-main">{{ c.skip }}</a>
    <div class="reading-progress" aria-hidden="true"></div>
    <header class="home-header">
      <a class="wordmark" href="#home-top" :aria-label="siteName">
        <img :src="siteLogo || '/logo.svg'" alt="" width="34" height="34" />
        <span>{{ siteName }}<small>WANWU · CONNECTED INTELLIGENCE</small></span>
      </a>
      <nav
        class="desktop-nav"
        :aria-label="language === 'zh' ? '主导航' : 'Main navigation'"
      >
        <a href="#possibilities">{{ c.scenesNav }}</a>
        <RouterLink v-if="showModelPlazaEntry" to="/model-plaza">{{
          c.modelsNav
        }}</RouterLink>
        <a href="#connect">{{ c.connectNav }}</a>
      </nav>
      <div class="header-actions">
        <LocaleSwitcher class="home-locale" />
        <RouterLink class="nav-login" :to="entryTarget"
          >{{ isAuthenticated ? c.dashboard : c.login }}
          <span aria-hidden="true">↗</span></RouterLink
        >
        <button
          class="menu-toggle"
          :aria-label="menuOpen ? c.closeMenu : c.menu"
          :aria-expanded="menuOpen"
          aria-controls="home-mobile-nav"
          @click="menuOpen = !menuOpen"
        >
          <span>{{ menuOpen ? '×' : '☰' }}</span>
        </button>
      </div>
      <nav
        v-if="menuOpen"
        id="home-mobile-nav"
        class="mobile-nav"
        :aria-label="c.menu"
      >
        <a href="#possibilities" @click="menuOpen = false"
          >{{ c.scenesNav }} <span>01</span></a
        >
        <RouterLink
          v-if="showModelPlazaEntry"
          to="/model-plaza"
          @click="menuOpen = false"
          >{{ c.modelsNav }} <span>02</span></RouterLink
        >
        <a href="#connect" @click="menuOpen = false"
          >{{ c.connectNav }} <span>03</span></a
        >
      </nav>
    </header>

    <main id="home-main">
      <section class="hero" aria-labelledby="home-title">
        <div class="hero-art">
          <picture>
            <source
              media="(max-width: 760px)"
              srcset="/artwork/wanwu-home-v1/hero-960.webp"
            />
            <img
              src="/artwork/wanwu-home-v1/hero-1672.webp"
              :alt="c.heroAlt"
              width="1672"
              height="941"
              fetchpriority="high"
              decoding="async"
            />
          </picture>
        </div>
        <div class="hero-shade" aria-hidden="true"></div>
        <div class="orbital-field" aria-hidden="true">
          <div class="orbit orbit-one"></div>
          <div class="orbit orbit-two"></div>
          <span
            v-for="n in 9"
            :key="n"
            class="light-particle"
            :style="{ '--i': n }"
          ></span>
        </div>
        <div class="hero-content page-width">
          <p class="eyebrow hero-enter">
            <span class="status-dot"></span>{{ c.kicker }}
          </p>
          <h1 id="home-title" class="hero-enter">
            <span>{{ c.title }}</span
            ><em>{{ c.titleAccent }}</em>
          </h1>
          <p class="hero-intro hero-enter">{{ c.intro }}</p>
          <div class="hero-actions hero-enter">
            <RouterLink class="button button-light" :to="entryTarget"
              >{{ c.start }} <span aria-hidden="true">↗</span></RouterLink
            >
            <RouterLink
              v-if="showModelPlazaEntry"
              class="button button-outline"
              to="/model-plaza"
              >{{ c.explore }} <span aria-hidden="true">→</span></RouterLink
            >
          </div>
        </div>
        <div class="hero-caption" aria-hidden="true">
          <span>THE ORIGIN / 001</span>
          <p>{{ c.heroCaption }}</p>
        </div>
        <div class="hero-bottom page-width">
          <a href="#possibilities" class="scroll-cue"
            ><span aria-hidden="true">↓</span>{{ c.scroll }}</a
          >
          <button
            class="motion-control"
            :disabled="reduced"
            :aria-pressed="paused || reduced"
            @click="paused = !paused"
          >
            <span aria-hidden="true">{{ motionOff ? '▷' : 'Ⅱ' }}</span
            >{{ reduced ? c.reduced : paused ? c.resume : c.pause }}
          </button>
        </div>
      </section>

      <div class="ecosystem page-width">
        <p>{{ c.ecosystem }}</p>
        <div class="provider-wordmarks" aria-label="GPT, DeepSeek, Grok">
          <span>GPT<span class="brand-star">✳</span></span
          ><span>deepseek<span class="brand-wave">≈</span></span
          ><span class="grok-wordmark">grok<span>↗</span></span>
        </div>
        <small>{{ c.ecosystemNote }}</small>
      </div>

      <section
        id="possibilities"
        class="possibilities page-width section-space"
        aria-labelledby="possibilities-title"
      >
        <div class="section-heading reveal">
          <div>
            <p class="eyebrow">01 — A WORLD OF POSSIBILITIES</p>
            <h2 id="possibilities-title">
              {{ c.sceneTitle }}<br /><em>{{ c.sceneAccent }}</em>
            </h2>
          </div>
          <p class="section-intro">{{ c.sceneIntro }}</p>
        </div>
        <div class="world-grid">
          <article
            v-for="(scene, index) in c.scenes"
            :key="scene.image"
            class="world-card reveal"
            :style="{ '--reveal-delay': `${index * 90}ms` }"
          >
            <img
              :src="`/artwork/wanwu-home-v1/${scene.image}-800.webp`"
              :srcset="`/artwork/wanwu-home-v1/${scene.image}-480.webp 480w, /artwork/wanwu-home-v1/${scene.image}-800.webp 800w`"
              sizes="(max-width: 640px) 90vw, 33vw"
              :alt="scene.alt"
              width="800"
              height="1000"
              loading="lazy"
              decoding="async"
            />
            <div class="world-card-shade" aria-hidden="true"></div>
            <span class="scene-index">{{ scene.tag }}</span>
            <div class="world-card-copy">
              <h3>{{ scene.title }}</h3>
              <p>{{ scene.text }}</p>
            </div>
          </article>
        </div>
      </section>

      <section
        class="union-section section-space"
        aria-labelledby="union-title"
      >
        <div class="union-layout page-width">
          <div class="union-copy reveal">
            <p class="eyebrow">02 — EVERYTHING, CONNECTED</p>
            <h2 id="union-title">
              {{ c.unionTitle }}<br /><em>{{ c.unionAccent }}</em>
            </h2>
            <p class="section-intro">{{ c.unionIntro }}</p>
            <div class="feature-list">
              <div
                v-for="(feature, index) in c.features"
                :key="feature.title"
                class="feature-row"
              >
                <span class="feature-index">0{{ index + 1 }}</span>
                <div>
                  <h3>{{ feature.title }}</h3>
                  <p>{{ feature.text }}</p>
                </div>
              </div>
            </div>
          </div>
          <div
            class="connection-art reveal"
            role="img"
            :aria-label="
              language === 'zh'
                ? '应用通过万物连接不同模型'
                : 'Applications connect to different models through Wanwu'
            "
          >
            <div class="connection-ring ring-outer"></div>
            <div class="connection-ring ring-inner"></div>
            <div class="connection-axis axis-one"></div>
            <div class="connection-axis axis-two"></div>
            <span class="connection-node node-gpt">GPT</span
            ><span class="connection-node node-deepseek">DeepSeek</span
            ><span class="connection-node node-grok">Grok</span
            ><span class="connection-node node-you">YOUR IDEAS</span>
            <div class="connection-core">
              <span class="core-symbol">W</span><strong>{{ siteName }}</strong
              ><span>ONE CONNECTION</span>
            </div>
            <span class="signal signal-one"></span
            ><span class="signal signal-two"></span>
            <p class="connection-caption">MANY FORMS. ONE FLOW.</p>
          </div>
        </div>
      </section>

      <section
        v-if="showModelPlazaEntry"
        class="popular-section page-width section-space"
        aria-labelledby="popular-title"
      >
        <div class="section-heading reveal">
          <div>
            <p class="eyebrow">03 — EXPLORE INTELLIGENCE</p>
            <h2 id="popular-title">{{ c.modelsTitle }}</h2>
            <p class="section-intro">{{ c.modelsIntro }}</p>
          </div>
          <RouterLink class="text-link" to="/model-plaza"
            >{{ c.allModels }} <span aria-hidden="true">↗</span></RouterLink
          >
        </div>
        <div class="popular-content reveal" :aria-busy="modelsLoading">
          <div v-if="modelsLoading" class="model-message" role="status">
            {{ c.loading }}
          </div>
          <div v-else-if="!popularModels.length" class="model-message">
            <p>{{ modelsFailed ? c.modelsFailed : c.modelsEmpty }}</p>
            <RouterLink class="text-link" to="/model-plaza"
              >{{ c.explore }} ↗</RouterLink
            >
          </div>
          <div v-else class="popular-grid">
            <RouterLink
              v-for="model in popularModels"
              :key="model.key"
              to="/model-plaza"
              class="popular-card"
            >
              <div class="model-card-top">
                <span class="model-monogram">{{
                  model.name.slice(0, 1).toUpperCase()
                }}</span
                ><span>TOP {{ model.popularityRank }}</span>
              </div>
              <h3>{{ model.name }}</h3>
              <p>
                {{ model.routes.length }} {{ c.groups }}
                <span aria-hidden="true">↗</span>
              </p>
            </RouterLink>
          </div>
          <p class="model-note">{{ c.modelNote }}</p>
        </div>
      </section>

      <section
        id="connect"
        class="connect-section page-width section-space"
        aria-labelledby="connect-title"
      >
        <div class="connect-copy reveal">
          <p class="eyebrow">04 — FROM IDEA TO API</p>
          <h2 id="connect-title">
            {{ c.connectTitle }}<br /><em>{{ c.connectAccent }}</em>
          </h2>
          <p class="section-intro">{{ c.connectIntro }}</p>
          <ol class="connect-steps">
            <li v-for="(step, index) in c.steps" :key="step.title">
              <span>0{{ index + 1 }}</span>
              <div>
                <h3>{{ step.title }}</h3>
                <p>{{ step.text }}</p>
              </div>
            </li>
          </ol>
          <div class="connect-actions">
            <RouterLink
              class="button button-light"
              :to="isAuthenticated ? '/keys' : '/login'"
              >{{ c.getKey }} <span aria-hidden="true">↗</span></RouterLink
            ><a
              v-if="docUrl"
              class="text-link"
              :href="docUrl"
              target="_blank"
              rel="noopener noreferrer"
              >{{ c.docs }} ↗</a
            >
          </div>
        </div>
        <div class="code-panel reveal">
          <div class="code-panel-top">
            <span class="code-dots" aria-hidden="true"
              ><i></i><i></i><i></i></span
            ><span>hello_world</span><span class="code-protocol">API</span>
          </div>
          <div class="code-toolbar">
            <div
              class="code-tabs"
              role="group"
              :aria-label="language === 'zh' ? '代码语言' : 'Code language'"
            >
              <button
                :aria-pressed="codeLanguage === 'curl'"
                @click="setCodeLanguage('curl')"
              >
                cURL</button
              ><button
                :aria-pressed="codeLanguage === 'python'"
                @click="setCodeLanguage('python')"
              >
                Python
              </button>
            </div>
            <button class="copy-code" @click="copyCode">
              {{ copied ? c.copied : c.copy }}
              <span aria-hidden="true">{{ copied ? '✓' : '⧉' }}</span>
            </button>
          </div>
          <pre tabindex="0"><code>{{ sampleCode }}</code></pre>
          <p class="code-hint">
            <span class="status-dot"></span>{{ c.codeHint }}
          </p>
          <p v-if="copyFailed" class="copy-error" role="status">
            {{ c.copyFailed }}
          </p>
          <span v-else class="sr-only" role="status">{{
            copied ? c.copied : ''
          }}</span>
        </div>
      </section>

      <section
        class="faq-section page-width section-space reveal"
        aria-labelledby="faq-title"
      >
        <div>
          <p class="eyebrow">A LITTLE MORE CLARITY</p>
          <h2 id="faq-title">{{ c.faqTitle }}</h2>
        </div>
        <div class="faq-list">
          <details v-for="faq in c.faqs" :key="faq.question">
            <summary>
              {{ faq.question }}<span aria-hidden="true">+</span>
            </summary>
            <p>{{ faq.answer }}</p>
          </details>
        </div>
      </section>

      <section class="closing-section" aria-labelledby="closing-title">
        <div class="closing-orbit" aria-hidden="true"></div>
        <div class="page-width reveal">
          <p class="eyebrow">THE NEXT POSSIBILITY IS YOU</p>
          <h2 id="closing-title">
            {{ c.closingTitle }}<br /><em>{{ c.closingAccent }}</em>
          </h2>
          <p>{{ c.closingText }}</p>
          <RouterLink class="button button-light" :to="entryTarget"
            >{{ c.start }} <span aria-hidden="true">↗</span></RouterLink
          >
        </div>
      </section>
    </main>
    <footer class="home-footer page-width">
      <div class="footer-brand">
        <strong>{{ siteName }}</strong
        ><span>{{ c.footerLine }}</span>
      </div>
      <div class="footer-links">
        <RouterLink v-if="showModelPlazaEntry" to="/model-plaza">{{
          c.modelsNav
        }}</RouterLink
        ><RouterLink :to="entryTarget">{{
          isAuthenticated ? c.dashboard : c.login
        }}</RouterLink
        ><a href="#home-top">{{ c.backTop }} ↑</a>
      </div>
      <div class="footer-bottom">
        <span>© {{ new Date().getFullYear() }} {{ siteName }}</span
        ><span>{{ c.artwork }}</span>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { getModelPlaza } from '@/api/modelPlaza'
import {
  buildCatalog,
  type CatalogModel
} from '@/components/modelPlaza/catalog'
import { homeCopy } from './homeCopy'
import { useHomeMotion } from './useHomeMotion'

const props = defineProps<{
  siteName: string
  siteLogo: string
  docUrl: string
  apiBaseUrl: string
  isAuthenticated: boolean
  dashboardPath: string
  showModelPlazaEntry: boolean
}>()
const { locale } = useI18n()
const language = computed(() => (locale?.value?.startsWith('en') ? 'en' : 'zh'))
const c = computed(() => homeCopy[language.value])
const root = ref<HTMLElement | null>(null)
const { paused, reduced, motionOff } = useHomeMotion(root)
const menuOpen = ref(false)
const entryTarget = computed(() =>
  props.isAuthenticated ? props.dashboardPath : '/login'
)
const popularModels = ref<CatalogModel[]>([])
const modelsLoading = ref(false)
const modelsFailed = ref(false)
const codeLanguage = ref<'curl' | 'python'>('curl')
const copied = ref(false)
const copyFailed = ref(false)
let copyTimer: ReturnType<typeof setTimeout> | undefined
let request: AbortController | undefined
let requestVersion = 0

watch(
  () => props.showModelPlazaEntry,
  async (show) => {
    const version = ++requestVersion
    request?.abort()
    popularModels.value = []
    modelsFailed.value = false
    modelsLoading.value = show
    if (!show) return
    request = new AbortController()
    try {
      const catalog = await getModelPlaza({ signal: request.signal })
      if (version !== requestVersion) return
      const availableModels = buildCatalog(
        catalog.groups.filter((group) => !group.is_exclusive)
      )
      const rankedModels = availableModels.filter((model) => Number.isFinite(model.popularityRank))
      popularModels.value = (rankedModels.length ? rankedModels : availableModels)
        .sort((a, b) => a.popularityRank - b.popularityRank)
        .slice(0, 4)
    } catch {
      if (version === requestVersion) modelsFailed.value = true
    } finally {
      if (version === requestVersion) modelsLoading.value = false
    }
  },
  { immediate: true }
)

const endpoint = computed(() => {
  const base = (props.apiBaseUrl || window.location.origin).replace(/\/+$/, '')
  return base.endsWith('/v1') ? base : `${base}/v1`
})
const sampleModel = computed(
  () =>
    popularModels.value.find((model) => /^gpt-/i.test(model.name))?.name ??
    'YOUR_MODEL_ID'
)
const sampleCode = computed(() => {
  if (codeLanguage.value === 'python') {
    return `import os\nfrom openai import OpenAI\n\nclient = OpenAI(\n    api_key=os.environ["WANWU_API_KEY"],\n    base_url=${JSON.stringify(endpoint.value)},\n)\n\nresponse = client.chat.completions.create(\n    model=${JSON.stringify(sampleModel.value)},\n    messages=[{\n        "role": "user",\n        "content": ${JSON.stringify(c.value.prompt)}\n    }]\n)\nprint(response.choices[0].message.content)`
  }
  // Encode the JSON body independently from shell quoting; no real secret is embedded.
  const body = JSON.stringify(
    {
      model: sampleModel.value,
      messages: [{ role: 'user', content: c.value.prompt }]
    },
    null,
    2
  ).replace(/'/g, "'\\''")
  const url = `${endpoint.value}/chat/completions`.replace(/'/g, "'\\''")
  return `curl '${url}' \\\n  -H "Authorization: Bearer $WANWU_API_KEY" \\\n  -H "Content-Type: application/json" \\\n  -d '${body}'`
})
function setCodeLanguage(value: 'curl' | 'python') {
  codeLanguage.value = value
  copied.value = false
  copyFailed.value = false
}
async function copyCode() {
  try {
    await navigator.clipboard.writeText(sampleCode.value)
    copied.value = true
    copyFailed.value = false
    clearTimeout(copyTimer)
    copyTimer = setTimeout(() => {
      copied.value = false
    }, 2200)
  } catch {
    copied.value = false
    copyFailed.value = true
  }
}
onBeforeUnmount(() => {
  requestVersion++
  request?.abort()
  clearTimeout(copyTimer)
})
</script>

<style scoped>
.wanwu-home {
  --ink: #eef0e7;
  --muted: #a3afa6;
  --jade: #b9d7b6;
  --line: rgba(211, 232, 217, 0.13);
  --page-progress: 0;
  --drift-x: 0px;
  --drift-y: 0px;
  --scroll-drift: 0px;
  color: var(--ink);
  background: #080f0d;
  overflow: clip;
  isolation: isolate;
  font-family: 'Inter', 'PingFang SC', 'Microsoft YaHei', sans-serif;
  -webkit-font-smoothing: antialiased;
}
.wanwu-home *,
.wanwu-home *::before,
.wanwu-home *::after {
  box-sizing: border-box;
}
.wanwu-home a {
  text-decoration: none;
}
.wanwu-home button,
.wanwu-home a,
.wanwu-home summary {
  -webkit-tap-highlight-color: transparent;
}
.wanwu-home :is(a, button, summary, pre):focus-visible {
  outline: 2px solid #d8efc8;
  outline-offset: 5px;
}
.wanwu-home ::selection {
  background: #a9c8ab;
  color: #102417;
}
.page-width {
  width: min(1280px, calc(100% - 112px));
  margin-inline: auto;
}
.section-space {
  padding-block: 110px;
}
.skip-link {
  position: fixed;
  z-index: 100;
  top: 10px;
  left: 20px;
  transform: translateY(-150%);
  background: #e5eddd;
  color: #132317;
  padding: 12px 20px;
  border-radius: 8px;
}
.skip-link:focus {
  transform: translateY(0);
}
.reading-progress {
  position: fixed;
  z-index: 90;
  top: 0;
  left: 0;
  width: 100%;
  height: 2px;
  background: #c5dab5;
  transform: scaleX(var(--page-progress));
  transform-origin: left;
}
.home-header {
  height: 92px;
  padding: 0 4%;
  position: absolute;
  inset: 0 0 auto;
  z-index: 30;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  border-bottom: 1px solid rgba(227, 243, 234, 0.08);
}
.wordmark {
  display: inline-flex;
  gap: 11px;
  align-items: center;
  color: var(--ink);
  min-width: 160px;
}
.wordmark img {
  width: 34px;
  height: 34px;
  border-radius: 9px;
  object-fit: contain;
}
.wordmark > span {
  font-weight: 650;
  font-size: 22px;
  letter-spacing: 0.06em;
}
.wordmark small {
  display: block;
  font-size: 7px;
  letter-spacing: 0.15em;
  color: #a1afa5;
  font-weight: 400;
  margin-top: 2px;
}
.desktop-nav,
.header-actions {
  display: flex;
  align-items: center;
  gap: 30px;
}
.desktop-nav a {
  font-size: 13px;
  color: #c7d1c9;
  transition: color 0.2s;
}
.desktop-nav a:hover,
.footer-links a:hover {
  color: white;
}
.header-actions {
  gap: 15px;
}
.nav-login {
  display: inline-flex;
  align-items: center;
  gap: 22px;
  border: 1px solid #7b8f7f;
  border-radius: 100px;
  padding: 10px 20px;
  font-size: 13px;
  color: var(--ink);
  transition: background 0.2s;
}
.nav-login:hover {
  background: #21382b;
}
.home-header :deep(.home-locale > button) {
  color: #d1dbd3;
}
.home-header :deep(.home-locale > button:hover) {
  background: #21352c;
}
.menu-toggle {
  display: none;
  width: 40px;
  height: 40px;
  color: var(--ink);
  font-size: 24px;
}
.mobile-nav {
  position: absolute;
  top: 100%;
  left: 0;
  width: 100%;
  padding: 18px 6%;
  background: #0d1a14;
  border-bottom: 1px solid var(--line);
  box-shadow: 0 20px 50px #0007;
}
.mobile-nav a {
  display: flex;
  justify-content: space-between;
  padding: 18px 0;
  color: var(--ink);
  border-bottom: 1px solid var(--line);
}
.mobile-nav a:last-child {
  border: 0;
}
.mobile-nav span {
  color: #819a89;
  font-size: 11px;
}
.hero {
  min-height: 780px;
  height: min(950px, 100svh);
  position: relative;
  display: flex;
  align-items: center;
}
.hero-art,
.hero-art picture,
.hero-art img,
.hero-shade,
.orbital-field {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}
.hero-art {
  overflow: hidden;
}
.hero-art picture {
  transform: translate3d(
      var(--drift-x),
      calc(var(--drift-y) + var(--scroll-drift)),
      0
    )
    scale(1.05);
  transition: transform 1s cubic-bezier(0.16, 1, 0.3, 1);
}
.hero-art img {
  object-fit: cover;
  object-position: 63% center;
}
.hero-shade {
  background:
    linear-gradient(90deg, #06100d99 0%, #06100d22 44%, transparent 70%),
    linear-gradient(
      0deg,
      #080f0d 0%,
      transparent 24%,
      #07100b22 85%,
      #07100b88 100%
    );
}
.hero-content {
  position: relative;
  z-index: 3;
  padding-top: 85px;
  pointer-events: none;
}
.hero-content a {
  pointer-events: auto;
}
.eyebrow {
  display: flex;
  gap: 10px;
  align-items: center;
  font-size: 10px;
  font-weight: 500;
  letter-spacing: 0.2em;
  color: #b5cab8;
  line-height: 1.8;
  margin-bottom: 24px;
}
.status-dot {
  display: inline-block;
  flex-shrink: 0;
  width: 5px;
  height: 5px;
  background: #c5dfb3;
  border-radius: 50%;
  box-shadow: 0 0 12px #b8e69b66;
}
.hero h1 {
  font-size: clamp(58px, 6.4vw, 94px);
  line-height: 1.23;
  font-weight: 500;
  letter-spacing: -0.055em;
  margin: 0;
}
.hero h1 span,
.hero h1 em {
  display: block;
}
.wanwu-home em {
  font-style: normal;
  color: #c2d2b9;
}
.hero h1 em {
  color: #e2e5cf;
}
.hero-intro {
  max-width: 390px;
  font-size: 16px;
  line-height: 1.95;
  color: #c4cec6;
  margin: 29px 0 34px;
}
.hero-actions,
.connect-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 14px;
}
.button {
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  gap: 34px;
  min-height: 50px;
  padding: 0 25px;
  border-radius: 100px;
  font-size: 13px;
  font-weight: 600;
  transition:
    transform 0.3s,
    background 0.3s,
    box-shadow 0.3s;
}
.button > span {
  font-size: 19px;
  font-weight: 400;
}
.button:hover {
  transform: translateY(-3px);
}
.button-light {
  color: #12201a;
  background: #e8eddb;
  border: 1px solid #e8eddb;
}
.button-light:hover {
  background: #fbfff0;
  box-shadow: 0 6px 30px #d9ebc41a;
}
.button-outline {
  color: #e0e8e0;
  background: #07110c22;
  border: 1px solid #b6c9b752;
  backdrop-filter: blur(8px);
}
.button-outline:hover {
  background: #1e352980;
}
.hero-caption {
  position: absolute;
  z-index: 3;
  right: 7%;
  bottom: 17%;
  text-align: right;
  color: #bccabd;
}
.hero-caption span {
  font:
    9px ui-monospace,
    monospace;
  letter-spacing: 0.15em;
  color: #d5decc88;
}
.hero-caption p {
  font-size: 12px;
  letter-spacing: 0.12em;
  margin-top: 9px;
}
.hero-bottom {
  position: absolute;
  bottom: 30px;
  left: 0;
  right: 0;
  z-index: 5;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.scroll-cue {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 10px;
  letter-spacing: 0.12em;
  color: #a4b5a9;
}
.scroll-cue > span {
  width: 26px;
  height: 34px;
  border: 1px solid #a4b5a949;
  border-radius: 100px;
  display: grid;
  place-items: center;
  animation: scroll-breathe 3s ease-in-out infinite;
}
.motion-control {
  color: #b1c1b5;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 9px 12px;
  border: 1px solid #a4b5a930;
  border-radius: 20px;
  font-size: 10px;
  background: #07110d55;
}
.motion-control:disabled {
  cursor: default;
}
.orbital-field {
  pointer-events: none;
  z-index: 2;
  overflow: hidden;
  opacity: 0.5;
}
.orbit {
  position: absolute;
  width: 45vw;
  height: 45vw;
  right: 4%;
  top: 16%;
  border-radius: 50%;
  border: 1px solid #e7ebca12;
  animation: orbit-turn 70s linear infinite;
}
.orbit::after {
  content: '';
  position: absolute;
  top: 15%;
  left: 14%;
  height: 4px;
  width: 4px;
  border-radius: 50%;
  background: #e3eccd;
  box-shadow: 0 0 17px 5px #d2e3b659;
}
.orbit-two {
  width: 52vw;
  height: 52vw;
  right: 0.5%;
  top: 10%;
  animation-direction: reverse;
  animation-duration: 95s;
  opacity: 0.6;
}
.light-particle {
  position: absolute;
  width: 2px;
  height: 2px;
  background: #d7edc7;
  border-radius: 50%;
  left: calc(44% + var(--i) * 5%);
  top: calc(17% + var(--i) * 7%);
  opacity: 0.4;
  animation: particle-drift calc(8s + var(--i) * 1s) ease-in-out infinite
    alternate;
  animation-delay: calc(var(--i) * -1.2s);
}
.hero-enter {
  animation: hero-arrive 1s cubic-bezier(0.16, 1, 0.3, 1) both;
}
.hero h1 {
  animation-delay: 0.12s;
}
.hero-intro {
  animation-delay: 0.22s;
}
.hero-actions {
  animation-delay: 0.32s;
}
.ecosystem {
  min-height: 130px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 35px;
  border-bottom: 1px solid var(--line);
}
.ecosystem p {
  font-size: 12px;
  color: #8fa194;
}
.ecosystem small {
  font-size: 9px;
  color: #809485;
  max-width: 155px;
  line-height: 1.7;
  text-align: right;
}
.provider-wordmarks {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 75px;
  color: #d6dfd4;
}
.provider-wordmarks > span {
  display: flex;
  align-items: center;
  gap: 9px;
  font-size: 25px;
  font-weight: 600;
  letter-spacing: -0.06em;
}
.brand-star {
  font-size: 28px;
  color: #9bad9c;
}
.brand-wave {
  font-size: 37px;
  line-height: 1;
  color: #9bad9c;
}
.grok-wordmark {
  font-family: Georgia, serif;
}
.grok-wordmark > span {
  font: 28px sans-serif;
  color: #9bad9c;
}
.section-heading {
  display: flex;
  justify-content: space-between;
  align-items: end;
  gap: 50px;
  margin-bottom: 42px;
}
.wanwu-home h2 {
  font-size: clamp(29px, 3.2vw, 48px);
  line-height: 1.5;
  letter-spacing: -0.045em;
  font-weight: 500;
  margin: 0;
}
.wanwu-home h2 em {
  color: #98ac9a;
}
.section-intro {
  font-size: 14px;
  line-height: 1.9;
  color: var(--muted);
  max-width: 380px;
}
.section-heading > .section-intro {
  padding-bottom: 5px;
  max-width: 290px;
}
.section-heading .eyebrow {
  margin-bottom: 19px;
}
.world-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 22px;
}
.world-card {
  min-height: 470px;
  aspect-ratio: 0.79;
  position: relative;
  overflow: hidden;
  border-radius: 9px;
  border: 1px solid #cde5d31c;
  background: #0c1711;
  transform-style: preserve-3d;
}
.world-card img {
  width: 100%;
  height: 100%;
  position: absolute;
  object-fit: cover;
  transition: transform 1.3s cubic-bezier(0.2, 0.7, 0.2, 1);
}
.world-card:hover img {
  transform: scale(1.065);
}
.world-card:hover {
  border-color: #d2e5c741;
}
.world-card-shade {
  position: absolute;
  inset: 0;
  background: linear-gradient(
    180deg,
    #07130c55,
    transparent 20%,
    transparent 40%,
    #050e09dd 80%,
    #050e09 100%
  );
}
.scene-index {
  position: absolute;
  left: 27px;
  top: 25px;
  font:
    9px ui-monospace,
    monospace;
  letter-spacing: 0.12em;
  color: #d2e3cd;
}
.world-card-copy {
  position: absolute;
  bottom: 26px;
  left: 27px;
  right: 27px;
}
.world-card-copy h3 {
  font-size: 25px;
  font-weight: 500;
  letter-spacing: -0.035em;
  margin-bottom: 12px;
}
.world-card-copy p {
  color: #b0bdb1;
  font-size: 12px;
  line-height: 1.85;
  max-width: 30em;
}
.union-section {
  background:
    radial-gradient(ellipse at 78% 45%, #18352555, transparent 57%),
    linear-gradient(150deg, #10201866, #0a1510 70%);
  border-block: 1px solid var(--line);
}
.union-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 50px;
  align-items: center;
}
.union-copy > .section-intro {
  margin: 20px 0 30px;
  max-width: 395px;
}
.feature-list {
  display: grid;
  gap: 22px;
  max-width: 390px;
}
.feature-row {
  display: flex;
  gap: 20px;
}
.feature-index {
  color: #708b76;
  font:
    10px ui-monospace,
    monospace;
  padding-top: 5px;
}
.feature-row h3 {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 7px;
}
.feature-row p {
  font-size: 12px;
  line-height: 1.8;
  color: #8fa495;
}
.connection-art {
  aspect-ratio: 1;
  width: 100%;
  max-width: 530px;
  position: relative;
  justify-self: end;
}
.connection-ring {
  position: absolute;
  border-radius: 50%;
  border: 1px solid #b6d5b71f;
}
.ring-outer {
  inset: 6%;
}
.ring-inner {
  inset: 23%;
  border-style: dashed;
  opacity: 0.7;
  animation: orbit-turn 120s linear infinite;
}
.connection-axis {
  position: absolute;
  left: 50%;
  top: 6%;
  height: 88%;
  width: 1px;
  background: linear-gradient(transparent, #abc8ae40, transparent);
}
.axis-one {
  transform: rotate(45deg);
}
.axis-two {
  transform: rotate(-45deg);
}
.connection-core {
  position: absolute;
  inset: 31%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  gap: 7px;
  border-radius: 50%;
  border: 1px solid #b7d3b642;
  background: radial-gradient(circle at 35% 25%, #54845c33, #0d1f13 75%);
  box-shadow:
    0 0 90px #bde4a910,
    inset 0 0 25px #bde4a909;
  animation: core-breathe 7s ease-in-out infinite;
}
.core-symbol {
  font:
    italic 42px Georgia,
    serif;
  color: #d9e7c6;
}
.connection-core strong {
  font-size: 16px;
  font-weight: 500;
}
.connection-core > span:last-child {
  font:
    6px ui-monospace,
    monospace;
  letter-spacing: 0.18em;
  color: #809787;
}
.connection-node {
  position: absolute;
  z-index: 2;
  border: 1px solid #91aa9545;
  background: #0e2118;
  border-radius: 40px;
  padding: 13px 20px;
  font-size: 12px;
  color: #bfcfc0;
}
.node-gpt {
  left: 12%;
  top: 14%;
}
.node-deepseek {
  right: 5%;
  top: 27%;
}
.node-grok {
  bottom: 13%;
  right: 16%;
}
.node-you {
  bottom: 30%;
  left: 0;
  font-size: 9px;
  letter-spacing: 0.1em;
}
.signal {
  position: absolute;
  inset: 6%;
  border-radius: 50%;
  animation: orbit-turn 18s linear infinite;
}
.signal::before {
  content: '';
  position: absolute;
  top: 15%;
  left: 14%;
  width: 4px;
  height: 4px;
  background: #d1e2b6;
  border-radius: 50%;
  box-shadow: 0 0 12px 4px #c8e9a13b;
}
.signal-two {
  animation-duration: 31s;
  animation-direction: reverse;
  opacity: 0.6;
}
.connection-caption {
  position: absolute;
  bottom: -5%;
  text-align: center;
  width: 100%;
  color: #56755e;
  letter-spacing: 0.3em;
  font-size: 8px;
}
.popular-section .section-intro {
  margin-top: 18px;
  max-width: 540px;
}
.text-link {
  display: inline-flex;
  align-items: center;
  gap: 16px;
  color: #cbd8c7;
  font-size: 12px;
  line-height: 1.8;
  white-space: nowrap;
}
.text-link:hover {
  color: white;
}
.text-link span {
  font-size: 18px;
}
.popular-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}
.popular-card {
  border: 1px solid var(--line);
  border-radius: 8px;
  background: #10201755;
  padding: 24px 21px;
  color: var(--ink);
  transition:
    transform 0.3s,
    background 0.3s,
    border-color 0.3s;
}
.popular-card:hover {
  transform: translateY(-5px);
  border-color: #b6d0b84a;
  background: #1a302155;
}
.model-card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.model-card-top > span:last-child {
  font:
    9px ui-monospace,
    monospace;
  letter-spacing: 0.1em;
  color: #718d77;
}
.model-monogram {
  width: 33px;
  height: 33px;
  display: grid;
  place-items: center;
  background: #bed8b81a;
  color: #d0e1c4;
  border-radius: 50%;
  font:
    17px Georgia,
    serif;
}
.popular-card h3 {
  font-size: 16px;
  letter-spacing: -0.02em;
  margin: 27px 0 12px;
  overflow-wrap: anywhere;
}
.popular-card p {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #859a8c;
}
.popular-card p span {
  color: #c1d1bd;
  font-size: 17px;
}
.model-note {
  color: #7b9383;
  font-size: 11px;
  margin-top: 21px;
}
.model-message {
  border: 1px solid var(--line);
  padding: 35px;
  font-size: 13px;
  color: #a4b3a8;
}
.model-message .text-link {
  margin-top: 15px;
}
.connect-section {
  display: grid;
  grid-template-columns: 1fr 1.05fr;
  gap: 75px;
  border-top: 1px solid var(--line);
}
.connect-copy > .section-intro {
  margin: 20px 0 34px;
}
.connect-steps {
  list-style: none;
  display: grid;
  gap: 27px;
  padding: 0;
  margin: 0 0 35px;
}
.connect-steps li {
  display: flex;
  gap: 18px;
}
.connect-steps li > span {
  width: 29px;
  height: 29px;
  border: 1px solid #a8c1ad32;
  border-radius: 50%;
  display: grid;
  place-items: center;
  font:
    9px ui-monospace,
    monospace;
  color: #b9d3b6;
  flex-shrink: 0;
}
.connect-steps h3 {
  font-size: 13px;
  font-weight: 500;
  margin: 3px 0 8px;
}
.connect-steps p {
  font-size: 12px;
  line-height: 1.8;
  color: #879d8f;
  max-width: 320px;
}
.connect-actions {
  gap: 25px;
}
.code-panel {
  min-width: 0;
  align-self: center;
  background: linear-gradient(145deg, #14231b, #0a130f 70%);
  border: 1px solid #819c8a35;
  border-radius: 10px;
  box-shadow: 0 24px 85px #0003;
}
.code-panel-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--line);
  padding: 18px 22px;
  font:
    10px ui-monospace,
    monospace;
  color: #82968a;
}
.code-dots {
  display: flex;
  gap: 5px;
}
.code-dots i {
  width: 6px;
  height: 6px;
  background: #6c8374;
  border-radius: 50%;
}
.code-dots i:nth-child(2) {
  opacity: 0.65;
}
.code-dots i:nth-child(3) {
  opacity: 0.35;
}
.code-protocol {
  letter-spacing: 0.2em;
  font-size: 8px;
}
.code-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 15px 22px 8px;
}
.code-tabs {
  display: flex;
  gap: 3px;
  background: #030a0766;
  padding: 3px;
  border-radius: 5px;
}
.code-tabs button {
  font:
    10px ui-monospace,
    monospace;
  color: #9caea0;
  padding: 6px 11px;
  border-radius: 3px;
}
.code-tabs button[aria-pressed='true'] {
  background: #304936;
  color: #e1eed9;
}
.copy-code {
  color: #a5b8a9;
  font-size: 10px;
  display: inline-flex;
  gap: 8px;
  padding: 8px 0;
}
.copy-code:hover {
  color: white;
}
.code-panel pre {
  padding: 22px;
  margin: 0;
  min-height: 316px;
  overflow-x: auto;
  color: #bad6bb;
  font:
    11px/1.95 ui-monospace,
    'SFMono-Regular',
    Consolas,
    monospace;
  scrollbar-color: #375240 #0c1711;
  scrollbar-width: thin;
}
.code-hint {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 17px 22px;
  border-top: 1px solid var(--line);
  font-size: 9px;
  line-height: 1.7;
  color: #7e9988;
}
.copy-error {
  color: #e6c5a1;
  font-size: 11px;
  padding: 0 22px 16px;
}
.faq-section {
  display: grid;
  grid-template-columns: 0.8fr 1.2fr;
  gap: 75px;
  border-top: 1px solid var(--line);
}
.faq-section h2 {
  font-size: clamp(27px, 2.6vw, 37px);
}
.faq-section .eyebrow {
  font-size: 8px;
}
.faq-list details {
  border-bottom: 1px solid var(--line);
}
.faq-list details:first-child {
  border-top: 1px solid var(--line);
}
.faq-list summary {
  cursor: pointer;
  list-style: none;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
  padding: 23px 0;
  font-size: 13px;
  color: #d1dacf;
}
.faq-list summary::-webkit-details-marker {
  display: none;
}
.faq-list summary span {
  font-size: 19px;
  color: #8ea994;
  transition: transform 0.25s;
}
.faq-list details[open] summary span {
  transform: rotate(45deg);
}
.faq-list details > p {
  color: #91a496;
  line-height: 1.9;
  font-size: 12px;
  padding: 0 30px 25px 0;
}
.closing-section {
  overflow: hidden;
  position: relative;
  border-block: 1px solid var(--line);
  padding: 105px 0;
  text-align: center;
  background: radial-gradient(
    ellipse at 50% 100%,
    #30583855 0,
    #142c1c44 35%,
    transparent 70%
  );
}
.closing-section .eyebrow {
  justify-content: center;
  font-size: 9px;
}
.closing-section h2 {
  font-size: clamp(37px, 4.7vw, 66px);
  line-height: 1.4;
  position: relative;
}
.closing-section h2 em {
  color: #c4d3b7;
}
.closing-section p:not(.eyebrow) {
  margin: 25px 0 29px;
  font-size: 13px;
  color: #8da28f;
}
.closing-section .page-width {
  position: relative;
  z-index: 1;
}
.closing-orbit {
  position: absolute;
  border: 1px solid #c3daba15;
  border-radius: 50%;
  width: 1000px;
  height: 1000px;
  top: 35%;
  left: calc(50% - 500px);
  box-shadow:
    0 0 0 100px #bfe2c503,
    0 0 0 220px #bfe2c503;
  animation: core-breathe 12s ease-in-out infinite;
}
.home-footer {
  padding-top: 48px;
  padding-bottom: 25px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 40px;
}
.footer-brand {
  display: flex;
  flex-direction: column;
  gap: 9px;
}
.footer-brand strong {
  font-size: 21px;
  font-weight: 500;
  letter-spacing: 0.08em;
}
.footer-brand span {
  font-size: 10px;
  color: #6e8776;
  letter-spacing: 0.08em;
}
.footer-links {
  display: flex;
  gap: 30px;
  align-items: start;
  justify-content: end;
  font-size: 11px;
  color: #9daf9f;
  padding-top: 6px;
}
.footer-bottom {
  grid-column: 1/-1;
  display: flex;
  justify-content: space-between;
  padding-top: 23px;
  border-top: 1px solid var(--line);
  color: #617d6a;
  font-size: 9px;
}
.has-observer .reveal {
  opacity: 0;
  transform: translateY(28px);
  transition:
    opacity 0.85s var(--reveal-delay, 0ms),
    transform 0.85s cubic-bezier(0.2, 0.7, 0.2, 1) var(--reveal-delay, 0ms);
}
.has-observer .reveal.is-visible {
  opacity: 1;
  transform: translateY(0);
}
.motion-off *,
.motion-off *::before,
.motion-off *::after {
  animation-play-state: paused !important;
  transition: none !important;
}
.motion-off .hero-enter {
  animation: none;
}
.motion-off .reveal {
  opacity: 1;
  transform: none;
}
.motion-off .world-card:hover img,
.motion-off .popular-card:hover,
.motion-off .button:hover {
  transform: none;
}
.locale-en .hero h1 {
  font-size: clamp(40px, 4.6vw, 66px);
  max-width: 650px;
  line-height: 1.15;
}
.locale-en .hero-intro {
  max-width: 400px;
}
.locale-en .section-heading h2,
.locale-en .union-copy h2,
.locale-en .connect-copy h2 {
  font-size: clamp(29px, 3vw, 43px);
}
.locale-en .world-card-copy h3 {
  font-size: 27px;
}
.locale-en .closing-section h2 {
  font-size: clamp(33px, 4.5vw, 60px);
}
@keyframes hero-arrive {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
@keyframes orbit-turn {
  to {
    transform: rotate(360deg);
  }
}
@keyframes core-breathe {
  0%,
  100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.035);
  }
}
@keyframes particle-drift {
  from {
    transform: translate3d(0, 0, 0);
    opacity: 0.15;
  }
  to {
    transform: translate3d(-12px, -32px, 0);
    opacity: 0.65;
  }
}
@keyframes scroll-breathe {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(4px);
  }
}
@media (min-width: 1600px) {
  .hero-art img {
    object-position: center 48%;
  }
  .hero-content {
    width: min(1380px, calc(100% - 180px));
  }
}
@media (max-width: 1100px) {
  .page-width {
    width: calc(100% - 64px);
  }
  .home-header {
    padding-inline: 32px;
  }
  .desktop-nav {
    gap: 22px;
  }
  .wordmark small {
    display: none;
  }
  .wordmark > span {
    font-size: 20px;
  }
  .wordmark {
    min-width: 130px;
  }
  .hero {
    min-height: 720px;
  }
  .hero-art img {
    object-position: 60% center;
  }
  .hero-caption {
    right: 5%;
  }
  .hero-intro {
    max-width: 340px;
  }
  .provider-wordmarks {
    gap: 35px;
  }
  .ecosystem small {
    display: none;
  }
  .section-space {
    padding-block: 85px;
  }
  .world-grid {
    gap: 15px;
  }
  .world-card {
    min-height: 430px;
  }
  .world-card-copy {
    left: 22px;
    right: 22px;
  }
  .world-card-copy h3 {
    font-size: 23px;
  }
  .world-card-copy p {
    font-size: 11px;
  }
  .scene-index {
    left: 22px;
  }
  .union-layout {
    gap: 25px;
  }
  .connection-node {
    padding: 10px 14px;
    font-size: 10px;
  }
  .node-you {
    font-size: 8px;
  }
  .connection-core strong {
    font-size: 13px;
  }
  .core-symbol {
    font-size: 32px;
  }
  .connect-section {
    gap: 40px;
  }
  .popular-card {
    padding: 21px 15px;
  }
  .popular-card h3 {
    font-size: 14px;
  }
}
@media (max-width: 760px) {
  .home-header {
    height: 76px;
    padding-inline: 23px;
  }
  .desktop-nav {
    display: none;
  }
  .header-actions {
    gap: 7px;
  }
  .wordmark {
    min-width: 0;
  }
  .wordmark > span {
    font-size: 18px;
  }
  .wordmark img {
    width: 29px;
    height: 29px;
  }
  .nav-login {
    padding: 8px 14px;
    gap: 14px;
    font-size: 11px;
  }
  .menu-toggle {
    display: block;
  }
  .page-width {
    width: calc(100% - 46px);
  }
  .section-space {
    padding-block: 70px;
  }
  .hero {
    height: 850px;
    min-height: min(850px, 100svh);
    align-items: start;
  }
  .hero-content {
    padding-top: 133px;
  }
  .hero h1 {
    font-size: clamp(49px, 9.5vw, 68px);
    line-height: 1.23;
  }
  .hero .eyebrow {
    font-size: 9px;
    margin-bottom: 23px;
  }
  .hero-intro {
    max-width: 320px;
    font-size: 13px;
    line-height: 1.85;
    margin-top: 21px;
    margin-bottom: 25px;
  }
  .button {
    min-height: 45px;
    padding-inline: 20px;
    gap: 27px;
    font-size: 12px;
  }
  .hero-art {
    top: 180px;
    height: calc(100% - 180px);
  }
  .hero-art picture {
    transform: translate3d(var(--drift-x), var(--drift-y), 0) scale(1.03);
  }
  .hero-art img {
    object-fit: cover;
    object-position: 76% center;
    opacity: 0.85;
  }
  .hero-shade {
    background: linear-gradient(
      180deg,
      #080f0d 0%,
      #080f0d 22%,
      #080f0dc9 36%,
      #080f0d22 57%,
      #080f0d00 77%,
      #080f0d 100%
    );
  }
  .hero-caption {
    bottom: 14%;
    right: 7%;
    font-size: 10px;
  }
  .hero-caption p {
    font-size: 10px;
  }
  .hero-caption span {
    font-size: 7px;
  }
  .hero-bottom {
    bottom: 26px;
  }
  .scroll-cue {
    font-size: 9px;
  }
  .motion-control {
    font-size: 9px;
  }
  .orbit {
    top: 44%;
    width: 83vw;
    height: 83vw;
    right: -4%;
  }
  .orbit-two {
    top: 41%;
    right: -13%;
    width: 102vw;
    height: 102vw;
  }
  .light-particle {
    top: calc(45% + var(--i) * 4%);
  }
  .ecosystem {
    flex-direction: column;
    justify-content: center;
    gap: 24px;
    min-height: 140px;
    padding-block: 20px 30px;
  }
  .ecosystem p {
    font-size: 10px;
    letter-spacing: 0.08em;
  }
  .provider-wordmarks {
    gap: clamp(28px, 8vw, 65px);
  }
  .provider-wordmarks > span {
    font-size: 22px;
  }
  .section-heading {
    display: block;
    margin-bottom: 30px;
  }
  .section-heading > .section-intro {
    margin-top: 20px;
    max-width: 360px;
  }
  .section-intro {
    font-size: 13px;
  }
  .eyebrow {
    font-size: 8px;
    margin-bottom: 18px;
  }
  .wanwu-home h2 {
    font-size: 33px;
    line-height: 1.45;
  }
  .world-grid {
    grid-template-columns: 1fr;
    gap: 22px;
  }
  .world-card {
    min-height: 420px;
    aspect-ratio: 0.98;
  }
  .world-card img {
    object-position: center 40%;
  }
  .world-card-copy {
    left: 26px;
    right: 26px;
    bottom: 27px;
  }
  .world-card-copy h3 {
    font-size: 28px;
  }
  .world-card-copy p {
    font-size: 13px;
    max-width: 28em;
  }
  .world-card-shade {
    background: linear-gradient(
      180deg,
      #06110d44,
      transparent 35%,
      #06110d99 65%,
      #06110df0 92%
    );
  }
  .scene-index {
    left: 26px;
    font-size: 9px;
  }
  .union-layout,
  .connect-section,
  .faq-section {
    grid-template-columns: 1fr;
    gap: 45px;
  }
  .union-copy > .section-intro {
    max-width: 370px;
  }
  .connection-art {
    max-width: 420px;
    justify-self: center;
  }
  .connection-core strong {
    font-size: 15px;
  }
  .feature-list {
    gap: 24px;
  }
  .feature-row h3 {
    font-size: 14px;
  }
  .popular-section .section-heading > .text-link {
    margin-top: 25px;
  }
  .popular-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
  }
  .popular-card {
    padding: 22px 16px;
  }
  .popular-card h3 {
    font-size: 14px;
    margin-top: 24px;
  }
  .model-card-top > span:last-child {
    font-size: 8px;
  }
  .model-note {
    font-size: 10px;
    line-height: 1.8;
  }
  .connect-section {
    gap: 40px;
  }
  .code-panel pre {
    font-size: 10px;
    padding: 20px 17px;
  }
  .code-panel-top,
  .code-toolbar {
    padding-inline: 17px;
  }
  .code-hint {
    padding-inline: 17px;
    font-size: 8px;
  }
  .code-panel pre {
    min-height: 305px;
  }
  .faq-section {
    gap: 30px;
  }
  .faq-list summary {
    font-size: 13px;
    padding-block: 21px;
  }
  .closing-section {
    padding-block: 80px;
  }
  .closing-section h2 {
    font-size: 38px;
  }
  .home-footer {
    gap: 30px;
  }
  .footer-links {
    gap: 19px;
    flex-wrap: wrap;
    align-content: start;
    font-size: 10px;
  }
  .footer-brand strong {
    font-size: 19px;
  }
  .footer-brand span {
    font-size: 8px;
  }
  .footer-bottom {
    font-size: 8px;
  }
  .locale-en .hero h1 {
    font-size: clamp(33px, 7.9vw, 54px);
    max-width: 100%;
  }
  .locale-en .hero-intro {
    max-width: 320px;
    font-size: 13px;
  }
  .locale-en .hero-actions .button {
    gap: 17px;
  }
  .locale-en .hero-caption {
    display: none;
  }
  .locale-en .closing-section h2 {
    font-size: 33px;
  }
}
@media (max-width: 380px) {
  .home-header {
    padding-inline: 16px;
  }
  .page-width {
    width: calc(100% - 36px);
  }
  .header-actions {
    gap: 2px;
  }
  .nav-login {
    padding-inline: 11px;
  }
  .nav-login span {
    display: none;
  }
  .hero h1 {
    font-size: 48px;
  }
  .hero-actions {
    gap: 10px;
  }
  .button {
    padding-inline: 18px;
    gap: 19px;
  }
  .popular-card {
    padding-inline: 12px;
  }
  .world-card-copy {
    left: 22px;
    right: 22px;
  }
  .closing-section h2 {
    font-size: 32px;
  }
  .footer-links {
    gap: 12px;
  }
}
@media (prefers-reduced-motion: reduce) {
  .wanwu-home *,
  .wanwu-home *::before,
  .wanwu-home *::after {
    animation: none !important;
    transition: none !important;
    scroll-behavior: auto !important;
  }
  .wanwu-home .reveal {
    opacity: 1;
    transform: none;
  }
  .hero-art picture {
    transform: none;
  }
}
</style>
