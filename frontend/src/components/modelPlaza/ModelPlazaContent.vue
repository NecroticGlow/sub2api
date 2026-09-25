<template>
  <div class="plaza" :class="{ 'plaza-embedded': embedded }">
    <section
      v-if="embedded"
      class="console-summary"
      :aria-label="c.consoleTitle"
    >
      <div class="console-intro">
        <span class="console-icon"><Icon name="grid" size="lg" /></span>
        <div>
          <h2>{{ c.consoleTitle }}</h2>
          <p>{{ c.consoleIntro }}</p>
        </div>
      </div>
      <div class="console-actions">
        <button
          class="refresh-action"
          :disabled="loading"
          :aria-label="c.refresh"
          :title="c.refresh"
          @click="emit('retry')"
        >
          <Icon name="refresh" size="sm" />
        </button>
        <RouterLink :to="accessTarget" class="primary-action"
          >{{ c.getKey }}<Icon name="arrowRight" size="sm"
        /></RouterLink>
      </div>
      <div class="console-stats" aria-live="polite">
        <span
          ><strong>{{ loading ? '—' : catalog.length }}</strong
          >{{ c.models }}</span
        >
        <span
          ><strong>{{ loading ? '—' : providers.length }}</strong
          >{{ c.providers }}</span
        >
        <span
          ><strong>{{ loading ? '—' : groups.length }}</strong
          >{{ c.routes }}</span
        >
        <span class="console-sync"
          ><span class="status-dot"></span>{{ c.statsNote }}</span
        >
      </div>
    </section>
    <section v-else class="hero">
      <div class="hero-copy">
        <span class="eyebrow"
          ><span class="status-dot"></span> WANWU · MODEL LIBRARY</span
        >
        <h1>
          {{ c.headline }}<br /><span>{{ c.headlineAccent }}</span>
        </h1>
        <p>{{ c.intro }}</p>
        <div class="actions">
          <a href="#model-catalog" class="primary-action"
            >{{ c.explore }} <Icon name="arrowDown" size="sm" /></a
          ><RouterLink :to="accessTarget" class="text-action"
            >{{ c.getKey }} <Icon name="arrowRight" size="sm"
          /></RouterLink>
        </div>
      </div>
      <div class="hero-art" aria-hidden="true">
        <div class="orbit outer"></div>
        <div class="orbit inner"></div>
        <div class="orbit-center">
          <Icon name="grid" size="xl" /><span>WANWU</span>
        </div>
        <span
          v-for="(name, i) in ['gpt', 'claude', 'gemini', 'deepseek']"
          :key="name"
          class="orbit-node"
          :class="`node-${i}`"
          ><ModelIcon :model="name" size="28px"
        /></span>
        <span class="orbit-caption">{{ c.artCaption }}</span>
      </div>
      <div class="hero-stats" aria-live="polite">
        <div>
          <strong>{{ loading ? '—' : catalog.length }}</strong
          ><span>{{ c.models }}</span>
        </div>
        <div>
          <strong>{{ loading ? '—' : providers.length }}</strong
          ><span>{{ c.providers }}</span>
        </div>
        <div>
          <strong>{{ loading ? '—' : groups.length }}</strong
          ><span>{{ c.routes }}</span>
        </div>
        <p><Icon name="infoCircle" size="sm" />{{ c.statsNote }}</p>
      </div>
    </section>
    <div
      v-if="descriptionHtml"
      class="plaza-description"
      v-html="descriptionHtml"
    ></div>
    <section
      id="model-catalog"
      class="catalog-section"
      aria-labelledby="catalog-heading"
      :aria-busy="loading"
    >
      <div class="section-heading">
        <div>
          <span v-if="!embedded" class="eyebrow">THE COLLECTION</span>
          <h2 id="catalog-heading">
            {{ embedded ? c.catalogTitle : c.browse }}
          </h2>
        </div>
        <span class="unit-hint">{{ c.priceUnit }}</span>
      </div>
      <div class="catalog-layout">
        <aside :aria-label="c.providerFilter">
          <span class="filter-heading">{{ c.providerFilter }}</span>
          <div class="provider-options">
            <button
              :class="{ active: selectedProvider === 'all' }"
              :aria-pressed="selectedProvider === 'all'"
              @click="selectedProvider = 'all'"
            >
              <Icon name="grid" size="sm" /><span>{{ c.all }}</span
              ><small>{{ catalog.length }}</small>
            </button>
            <button
              v-for="provider in providers"
              :key="provider.id"
              :class="{ active: selectedProvider === provider.id }"
              :aria-pressed="selectedProvider === provider.id"
              @click="selectedProvider = provider.id"
            >
              <ModelIcon :model="provider.icon" size="19px" /><span>{{
                provider.label
              }}</span
              ><small>{{ provider.count }}</small>
            </button>
          </div>
          <div class="sidebar-note">
            <Icon name="lightbulb" size="md" /><strong>{{
              c.chooseTitle
            }}</strong>
            <p>{{ c.chooseNote }}</p>
          </div>
        </aside>
        <div class="catalog-main">
          <div class="toolbar">
            <label class="search-field"
              ><Icon name="search" size="sm" /><input
                v-model="searchQuery"
                type="search"
                :aria-label="c.search"
                :placeholder="c.search"
            /></label>
            <select v-model="selectedGroup" :aria-label="c.groupFilter" class="group-filter">
              <option value="all">{{ c.allGroups }} ({{ groups.length }})</option>
              <option v-for="group in groups" :key="group.id" :value="String(group.id)">{{ group.name }} ({{ group.models.length }})</option>
            </select>
            <select v-model="sortOrder" :aria-label="c.sort">
              <option value="popular">{{ c.sortPopular }}</option>
              <option value="name">{{ c.sortName }}</option>
              <option value="price">{{ c.sortPrice }}</option>
            </select>
            <select v-model="selectedPlatform" :aria-label="t('modelPlaza.filters.platformLabel')" data-testid="platform-filter">
              <option value="all">{{ t('modelPlaza.filters.platformLabel') }} · {{ t('modelPlaza.filters.all') }}</option>
              <option v-for="platform in platforms" :key="platform" :value="platform">{{ platform }}</option>
            </select>
            <select v-model="selectedRate" :aria-label="t('modelPlaza.filters.rateLabel')" data-testid="rate-filter">
              <option value="all">{{ t('modelPlaza.filters.rateLabel') }} · {{ t('modelPlaza.filters.all') }}</option>
              <option v-for="rate in rates" :key="rate" :value="String(rate)">{{ rate }}x</option>
            </select>
            <div class="view-switch" :aria-label="c.view">
              <button
                :aria-label="c.cardView"
                :title="c.cardView"
                :aria-pressed="view === 'cards'"
                :class="{ active: view === 'cards' }"
                @click="view = 'cards'"
              >
                <Icon name="grid" size="sm" /></button
              ><button
                :aria-label="c.tableView"
                :title="c.tableView"
                :aria-pressed="view === 'table'"
                :class="{ active: view === 'table' }"
                @click="view = 'table'"
              >
                <Icon name="menu" size="sm" />
              </button>
            </div>
          </div>
          <div class="results-meta">
            <span aria-live="polite"
              >{{ visibleModels.length }} {{ c.results }}</span
            ><button v-if="filtersActive" @click="resetFilters">
              {{ c.reset }} <Icon name="x" size="xs" /></button
            ><span v-else>{{ c.standardPrice }}</span>
          </div>
          <div
            v-if="loading"
            class="model-grid"
            role="status"
            :aria-label="c.loading"
          >
            <div
              v-for="n in 6"
              :key="n"
              class="model-card skeleton"
              aria-hidden="true"
            >
              <span></span><i></i><i></i><i></i>
            </div>
          </div>
          <div v-else-if="error" class="empty-state" role="alert">
            <Icon name="refresh" size="xl" />
            <h3>{{ c.loadFailed }}</h3>
            <p>{{ c.retryNote }}</p>
            <button class="primary-action" @click="emit('retry')">
              {{ c.retry }}
            </button>
          </div>
          <div v-else-if="!visibleModels.length" class="empty-state">
            <Icon :name="filtersActive ? 'search' : 'grid'" size="xl" />
            <h3>{{ filtersActive ? c.noSearchResult : c.emptyTitle }}</h3>
            <p>{{ filtersActive ? c.searchNote : c.emptyNote }}</p>
            <button
              v-if="filtersActive"
              class="primary-action"
              @click="resetFilters"
            >
              {{ c.reset }}</button
            ><RouterLink
              v-else
              :to="authStore.isAdmin ? '/admin/groups' : accessTarget"
              class="primary-action"
              >{{ authStore.isAdmin ? c.configure : c.getKey }}
              <Icon name="arrowRight" size="sm"
            /></RouterLink>
          </div>
          <div v-else-if="view === 'cards'" class="model-grid">
            <article
              v-for="model in visibleModels"
              :key="model.key"
              class="model-card"
              :class="{ selected: selectedKey === model.key }"
            >
              <div class="card-top">
                <span class="model-logo"
                  ><ModelIcon :model="model.name" size="27px" /></span
                ><span v-if="Number.isFinite(model.popularityRank)" class="rank-badge" :class="{ 'rank-leading': model.popularityRank <= 3 }">TOP {{ model.popularityRank }}</span><span v-else class="category">{{ categoryLabel(model) }}</span>
              </div>
              <span class="provider-label">{{
                providerLabel(modelProvider(model.name, model.platform))
              }}</span>
              <h3>{{ model.name }}</h3>
              <p class="model-description">
                {{
                  model.routes.find((r) => r.group.description.trim())?.group
                    .description || c.modelSummary
                }}
              </p>
              <div class="tags">
                <span>{{ model.platform }}</span
                ><span>{{ model.routes.length }} {{ c.routeCount }}</span>
              </div>
              <div class="prices">
                <div>
                  <span>{{ c.input }}</span
                  ><strong>{{
                    displayPrice(minimumTokenPrice(model, 'input_price'))
                  }}</strong>
                </div>
                <div>
                  <span>{{ c.output }}</span
                  ><strong>{{
                    displayPrice(minimumTokenPrice(model, 'output_price'))
                  }}</strong>
                </div>
                <div class="cache-read-price">
                  <span>{{ c.cacheRead }}</span>
                  <strong>{{ displayPrice(minimumTokenPrice(model, 'cache_read_price')) }}</strong>
                </div>
                <div class="cache-write-price">
                  <span>{{ c.cacheWrite }}</span>
                  <strong>{{ displayPrice(minimumTokenPrice(model, 'cache_write_price')) }}</strong>
                </div>
                <div
                  v-if="minimumTokenPrice(model, 'cache_write_1h_price') !== null"
                  class="cache-write-hour-price"
                >
                  <span>{{ c.cacheWrite1h }}</span>
                  <strong>{{ displayPrice(minimumTokenPrice(model, 'cache_write_1h_price')) }}</strong>
                </div>
              </div>
              <button
                class="card-details"
                :aria-expanded="selectedKey === model.key"
                aria-controls="model-details"
                @click="selectModel(model)"
              >
                <span>{{ c.details }}</span
                ><Icon name="arrowRight" size="sm" />
              </button>
            </article>
          </div>
          <div v-else class="table-groups">
            <PlazaGroupSection
              v-for="group in filteredGroups"
              :key="group.id"
              :group="group"
            />
          </div>
          <section
            v-if="selectedModel && view === 'cards'"
            id="model-details"
            ref="detailsElement"
            tabindex="-1"
            class="model-details"
            :aria-label="c.details"
          >
            <div class="details-heading">
              <div>
                <span class="eyebrow">MODEL DETAILS</span>
                <h3>{{ selectedModel.name }}</h3>
              </div>
              <button :aria-label="c.close" @click="selectedKey = null">
                <Icon name="x" size="md" />
              </button>
            </div>
            <p class="details-note">{{ c.detailsNote }}</p>
            <div class="actions">
              <button
                class="secondary-action"
                @click="copyModel(selectedModel.name)"
              >
                <Icon :name="copied ? 'check' : 'copy'" size="sm" />{{
                  copied ? c.copied : c.copy
                }}</button
              ><RouterLink :to="accessTarget" class="primary-action"
                >{{ c.getKey }} <Icon name="arrowRight" size="sm"
              /></RouterLink>
            </div>
            <p v-if="copyError" role="status" class="details-note">
              {{ c.copyError }}
            </p>
            <div class="table-groups">
              <PlazaGroupSection
                v-for="route in selectedModel.routes"
                :key="route.group.id"
                :group="{ ...route.group, models: [route.model] }"
              />
            </div>
          </section>
        </div>
      </div>
    </section>
    <section v-if="!embedded" class="getting-started">
      <div>
        <span class="eyebrow">BUILD SOMETHING GREAT</span>
        <h2>{{ c.startTitle }}</h2>
        <p>{{ c.startNote }}</p>
        <RouterLink :to="accessTarget" class="text-action"
          >{{ c.getKey }} <Icon name="arrowRight" size="sm"
        /></RouterLink>
      </div>
      <ol>
        <li
          v-for="(step, i) in [c.stepOne, c.stepTwo, c.stepThree]"
          :key="step"
        >
          <span>0{{ i + 1 }}</span
          >{{ step }}
        </li>
      </ol>
    </section>
    <p v-if="!authStore.isAuthenticated" class="anonymous-note">
      <Icon name="infoCircle" size="xs" />{{ c.anonymousHint }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import Icon from '@/components/icons/Icon.vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import PlazaGroupSection from './PlazaGroupSection.vue'
import {
  buildCatalog,
  filterCatalog,
  minimumTokenPrice,
  modelProvider,
  type CatalogModel
} from './catalog'
import { catalogCopy } from './catalogCopy'
import type { ModelPlazaResponse } from '@/api/modelPlaza'
import { useAuthStore } from '@/stores/auth'

const props = defineProps<{
  response: ModelPlazaResponse | null
  loading: boolean
  error?: boolean
  embedded?: boolean
}>()
const emit = defineEmits<{ retry: [] }>()
const { locale, t } = useI18n()
const c = computed(
  () => catalogCopy[locale.value.startsWith('zh') ? 'zh' : 'en']
)
const authStore = useAuthStore()
const selectedGroup = ref('all')
const selectedPlatform = ref('all')
const selectedRate = ref('all')
const selectedProvider = ref('all'),
  searchQuery = ref(''),
  sortOrder = ref('popular')
const view = ref<'cards' | 'table'>('cards')
const selectedKey = ref<string | null>(null)
const copied = ref(false),
  copyError = ref(false)
const detailsElement = ref<HTMLElement | null>(null)
const accessTarget = computed(() =>
  authStore.isAuthenticated ? '/keys' : '/login?redirect=/keys'
)
const groups = computed(() =>
  (props.response?.groups ?? []).filter((g) => g.models.length > 0)
)
const catalog = computed(() => buildCatalog(groups.value))
const platforms = computed(() => [...new Set(groups.value.map(group => group.platform).filter(Boolean))].sort())
const rates = computed(() => [...new Set(groups.value.map(group => group.user_rate_multiplier ?? group.rate_multiplier))].sort((a, b) => a - b))
const selectedGroups = computed(() => groups.value.filter(group =>
  (selectedGroup.value === 'all' || String(group.id) === selectedGroup.value) &&
  (selectedPlatform.value === 'all' || group.platform === selectedPlatform.value) &&
  (selectedRate.value === 'all' || String(group.user_rate_multiplier ?? group.rate_multiplier) === selectedRate.value)
))
const groupCatalog = computed(() => buildCatalog(selectedGroups.value))
const filtersActive = computed(
  () => selectedGroup.value !== 'all' || selectedPlatform.value !== 'all' || selectedRate.value !== 'all' || selectedProvider.value !== 'all' || searchQuery.value.trim() !== ''
)
const descriptionHtml = computed(() =>
  DOMPurify.sanitize(
    marked.parse(props.response?.description?.trim() || '') as string
  )
)
const labels: Record<string, string> = {
  openai: 'OpenAI',
  anthropic: 'Anthropic',
  gemini: 'Google',
  deepseek: 'DeepSeek',
  grok: 'xAI',
  zhipu: '智谱 Z.ai'
}
const icons: Record<string, string> = {
  openai: 'gpt',
  anthropic: 'claude',
  gemini: 'gemini',
  deepseek: 'deepseek',
  grok: 'grok',
  zhipu: 'glm'
}
function providerLabel(id: string) {
  return labels[id] || id
}
const providers = computed(() =>
  [...new Set(catalog.value.map((m) => modelProvider(m.name, m.platform)))]
    .sort()
    .map((id) => ({
      id,
      label: providerLabel(id),
      icon: icons[id] || id,
      count: catalog.value.filter(
        (m) => modelProvider(m.name, m.platform) === id
      ).length
    }))
)
const visibleModels = computed(() =>
  filterCatalog(groupCatalog.value, selectedProvider.value, searchQuery.value).sort(
    (a, b) => {
      if (sortOrder.value === 'popular' && a.popularityRank !== b.popularityRank) return a.popularityRank - b.popularityRank
      if (sortOrder.value === 'price') {
        const pa = minimumTokenPrice(a, 'input_price') ?? Infinity,
          pb = minimumTokenPrice(b, 'input_price') ?? Infinity
        if (pa !== pb) return pa - pb
      }
      return a.name.localeCompare(b.name, undefined, { numeric: true })
    }
  )
)
const filteredGroups = computed(() => {
  const keys = new Set(visibleModels.value.map((m) => m.key))
  return selectedGroups.value
    .map((g) => ({
      ...g,
      models: g.models.filter((m) =>
        keys.has(`${m.platform}:${m.name.toLowerCase()}`)
      )
    }))
    .filter((g) => g.models.length)
})
const selectedModel = computed(() =>
  visibleModels.value.find((m) => m.key === selectedKey.value)
)
watch([selectedGroup, selectedPlatform, selectedRate, selectedProvider, searchQuery, view], () => {
  selectedKey.value = null
})
watch(providers, (value) => {
  if (
    selectedProvider.value !== 'all' &&
    !value.some((p) => p.id === selectedProvider.value)
  )
    selectedProvider.value = 'all'
})
watch(groups, (value) => {
  if (selectedGroup.value !== 'all' && !value.some(g => String(g.id) === selectedGroup.value)) selectedGroup.value = 'all'
  if (selectedPlatform.value !== 'all' && !value.some(g => g.platform === selectedPlatform.value)) selectedPlatform.value = 'all'
  if (selectedRate.value !== 'all' && !value.some(g => String(g.user_rate_multiplier ?? g.rate_multiplier) === selectedRate.value)) selectedRate.value = 'all'
})
function resetFilters() {
  selectedGroup.value = 'all'
  selectedPlatform.value = 'all'
  selectedRate.value = 'all'
  selectedProvider.value = 'all'
  searchQuery.value = ''
}
function displayPrice(value: number | null) {
  return value === null
    ? '—'
    : `¥${value.toLocaleString('en-US', { maximumFractionDigits: 6 })}`
}
function categoryLabel(model: CatalogModel) {
  const mode = model.routes[0]?.model.pricing?.billing_mode
  return mode === 'image'
    ? c.value.imageModel
    : mode === 'video'
      ? c.value.videoModel
      : mode === 'per_request'
        ? c.value.requestModel
        : c.value.languageModel
}
async function selectModel(model: CatalogModel) {
  if (selectedKey.value === model.key) {
    selectedKey.value = null
    return
  }
  selectedKey.value = model.key
  copied.value = false
  copyError.value = false
  await nextTick()
  detailsElement.value?.focus({ preventScroll: true })
  detailsElement.value?.scrollIntoView({
    behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches
      ? 'instant'
      : 'smooth',
    block: 'start'
  })
}
async function copyModel(name: string) {
  try {
    await navigator.clipboard.writeText(name)
    copied.value = true
    copyError.value = false
  } catch {
    copyError.value = true
  }
}
</script>

<style scoped>
.plaza {
  --ink: #1b302d;
  --muted: #677975;
  --line: #e0e8e4;
  --paper: #fff;
  --soft: #f5f8f6;
  --accent: #147d64;
  color: var(--ink);
}
.dark .plaza {
  --ink: #e1ede7;
  --muted: #99b0a5;
  --line: #30423a;
  --paper: #192820;
  --soft: #142219;
  --accent: #76d5ae;
}
.rank-badge { font-size: 11px; font-weight: 700; color: var(--muted); border-radius: 6px; padding: 5px 8px; background: var(--soft); letter-spacing: .06em; }
.rank-leading { color: var(--accent); background: color-mix(in srgb, var(--accent) 12%, transparent); }
.group-filter { max-width: 220px; min-width: 0; }
.plaza :is(button, a, input, select):focus-visible {
  outline: 3px solid #55b99b;
  outline-offset: 4px;
}
.hero {
  display: grid;
  grid-template-columns: 1.4fr 1fr;
  overflow: hidden;
  border: 1px solid var(--line);
  border-radius: 22px;
  background: var(--soft);
}
.hero-copy {
  padding: 44px 44px 36px;
  z-index: 1;
}
.eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.17em;
  color: var(--muted);
}
.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--accent);
  box-shadow: 0 0 0 4px #15957013;
}
.hero-copy h1 {
  margin: 22px 0 18px;
  font-size: clamp(28px, 3vw, 43px);
  line-height: 1.4;
  font-weight: 650;
  letter-spacing: -0.055em;
}
.hero-copy h1 span {
  color: var(--accent);
}
.hero-copy > p {
  max-width: 440px;
  color: var(--muted);
  font-size: 13px;
  line-height: 1.9;
}
.actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 24px;
  margin-top: 26px;
}
.primary-action,
.secondary-action {
  display: inline-flex;
  justify-content: center;
  align-items: center;
  gap: 12px;
  border-radius: 9px;
  padding: 11px 17px;
  font-size: 12px;
  font-weight: 600;
}
.primary-action {
  background: #176b54;
  color: #fff;
  border: 1px solid #176b54;
}
.primary-action:hover {
  background: #10543f;
}
.secondary-action {
  border: 1px solid var(--line);
  background: var(--paper);
  color: var(--ink);
}
.text-action {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
  font-weight: 600;
  color: var(--accent);
}
.hero-art {
  position: relative;
  min-height: 310px;
  align-self: center;
}
.orbit {
  position: absolute;
  left: 50%;
  top: 46%;
  border: 1px solid #5e9d7929;
  border-radius: 50%;
  transform: translate(-50%, -50%);
}
.outer {
  width: 282px;
  height: 282px;
  border-style: dashed;
}
.inner {
  width: 187px;
  height: 187px;
}
.orbit-center {
  position: absolute;
  left: 50%;
  top: 46%;
  transform: translate(-50%, -50%) rotate(-8deg);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 9px;
  width: 94px;
  height: 94px;
  border: 1px solid var(--line);
  border-radius: 24px;
  background: var(--paper);
  box-shadow: 0 13px 40px #184c3010;
  color: var(--accent);
}
.orbit-center span {
  font-size: 9px;
  letter-spacing: 0.14em;
}
.orbit-node {
  position: absolute;
  display: grid;
  place-items: center;
  width: 55px;
  height: 55px;
  background: white;
  border: 1px solid #e4eae5;
  border-radius: 16px;
  box-shadow: 0 5px 18px #163c2909;
}
.node-0 {
  left: calc(50% - 115px);
  top: 29px;
  transform: rotate(-10deg);
}
.node-1 {
  left: calc(50% + 80px);
  top: 64px;
  transform: rotate(10deg);
}
.node-2 {
  left: calc(50% + 31px);
  top: 218px;
  transform: rotate(8deg);
}
.node-3 {
  left: calc(50% - 156px);
  top: 181px;
  transform: rotate(-7deg);
}
.orbit-caption {
  position: absolute;
  left: 50%;
  top: 295px;
  transform: translateX(-50%);
  white-space: nowrap;
  font-size: 10px;
  color: var(--muted);
  letter-spacing: 0.08em;
}
.hero-stats {
  grid-column: 1/-1;
  display: flex;
  align-items: center;
  gap: 36px;
  padding: 20px 44px;
  border-top: 1px solid var(--line);
}
.hero-stats div {
  display: flex;
  align-items: baseline;
  gap: 10px;
}
.hero-stats strong {
  font-size: 22px;
  font-weight: 650;
  font-variant-numeric: tabular-nums;
}
.hero-stats span,
.hero-stats p {
  font-size: 11px;
  color: var(--muted);
}
.hero-stats p {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 7px;
}
.catalog-section {
  margin-top: 44px;
  scroll-margin-top: 90px;
}
.section-heading {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 24px;
  gap: 12px;
}
.section-heading h2,
.getting-started h2 {
  font-size: 22px;
  font-weight: 650;
  letter-spacing: -0.04em;
  margin-top: 7px;
}
.unit-hint {
  color: var(--muted);
  font-size: 11px;
}
.catalog-layout > * {
  min-width: 0;
}
.catalog-layout {
  display: grid;
  grid-template-columns: 176px minmax(0, 1fr);
  gap: 28px;
}
.filter-heading {
  display: block;
  margin: 8px 0 15px 10px;
  font-size: 11px;
  color: var(--muted);
}
.provider-options {
  display: flex;
  flex-direction: column;
  gap: 5px;
}
.provider-options button {
  display: flex;
  align-items: center;
  gap: 11px;
  width: 100%;
  padding: 12px 10px;
  border-radius: 8px;
  font-size: 13px;
  color: var(--muted);
  text-align: left;
}
.provider-options small {
  margin-left: auto;
  font-size: 10px;
  opacity: 0.8;
}
.provider-options button:hover {
  background: var(--soft);
}
.provider-options button.active {
  background: #19896312;
  color: var(--accent);
  font-weight: 650;
}
.sidebar-note {
  margin-top: 30px;
  padding: 18px 13px;
  border-top: 1px solid var(--line);
  color: var(--muted);
}
.sidebar-note strong {
  display: block;
  font-size: 12px;
  color: var(--ink);
  margin: 12px 0 8px;
}
.sidebar-note p {
  font-size: 11px;
  line-height: 1.9;
}
.toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
}
.search-field {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 11px 13px;
  border: 1px solid var(--line);
  border-radius: 9px;
  color: var(--muted);
  background: var(--paper);
  min-width: 0;
}
.search-field input {
  width: 100%;
  border: 0;
  outline: none;
  background: transparent;
  color: var(--ink);
  font-size: 13px;
}
.toolbar select {
  max-width: 150px;
  padding: 11px 28px 11px 12px;
  font-size: 11px;
  border-radius: 9px;
  border: 1px solid var(--line);
  background-color: var(--paper);
  color: var(--ink);
}
.view-switch {
  display: flex;
  border: 1px solid var(--line);
  padding: 4px;
  border-radius: 9px;
  background: var(--paper);
}
.view-switch button {
  padding: 7px;
  color: var(--muted);
  border-radius: 5px;
}
.view-switch button.active {
  background: var(--soft);
  color: var(--accent);
}
.results-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin: 15px 0 18px;
  font-size: 10px;
  color: var(--muted);
}
.results-meta button {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  color: var(--accent);
}
.model-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}
.model-card {
  display: flex;
  flex-direction: column;
  min-width: 0;
  padding: 19px 18px 0;
  border: 1px solid var(--line);
  border-radius: 13px;
  background: var(--paper);
  transition:
    border-color 0.18s,
    box-shadow 0.18s,
    transform 0.18s;
}
.model-card:hover,
.model-card.selected {
  border-color: #7cb39b;
  box-shadow: 0 7px 22px #1d482508;
}
.model-card:hover {
  transform: translateY(-2px);
}
.card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 7px;
  margin-bottom: 20px;
}
.model-logo {
  display: grid;
  place-items: center;
  width: 43px;
  height: 43px;
  background: #fff;
  border: 1px solid #e9eee9;
  border-radius: 12px;
}
.category {
  font-size: 10px;
  color: var(--muted);
  background: var(--soft);
  border-radius: 5px;
  padding: 4px 7px;
}
.provider-label {
  font-size: 11px;
  color: var(--muted);
  letter-spacing: 0.025em;
}
.model-card h3 {
  font-size: 17px;
  font-weight: 650;
  line-height: 1.45;
  letter-spacing: -0.035em;
  overflow-wrap: anywhere;
  margin: 5px 0 9px;
}
.model-description {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 37px;
  font-size: 12px;
  line-height: 1.8;
  color: var(--muted);
}
.tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  margin: 14px 0 18px;
}
.tags span {
  border: 1px solid var(--line);
  color: var(--muted);
  font-size: 10px;
  border-radius: 4px;
  padding: 2px 5px;
}
.prices {
  margin-top: auto;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  padding-top: 14px;
  border-top: 1px solid var(--line);
}
.prices div {
  display: flex;
  flex-direction: column;
  gap: 5px;
}
.prices span {
  font-size: 11px;
  color: var(--muted);
}
.prices strong {
  font-size: 17px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}
.card-details {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 15px 0;
  margin-top: 4px;
  font-size: 12px;
  font-weight: 600;
  color: var(--accent);
}
.table-groups {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
}
.table-groups :deep(section) {
  min-width: 0;
}
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 64px 24px;
  border: 1px dashed var(--line);
  border-radius: 14px;
  background: var(--paper);
  color: var(--muted);
}
.empty-state h3 {
  font-size: 17px;
  margin: 18px 0 10px;
  font-weight: 600;
  color: var(--ink);
}
.empty-state p {
  font-size: 12px;
  line-height: 1.8;
  max-width: 380px;
  margin-bottom: 22px;
}
.model-details {
  padding: 22px;
  margin-top: 24px;
  background: var(--soft);
  border: 1px solid var(--line);
  border-radius: 14px;
  scroll-margin-top: 90px;
}
.details-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}
.details-heading h3 {
  font-size: 21px;
  font-weight: 650;
  margin-top: 8px;
  overflow-wrap: anywhere;
}
.details-heading button {
  padding: 6px;
}
.details-note {
  font-size: 11px;
  color: var(--muted);
  line-height: 1.8;
  margin-top: 12px;
}
.model-details .actions {
  margin: 16px 0 22px;
  gap: 12px;
}
.getting-started {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 35px;
  padding: 32px 38px;
  border: 1px solid var(--line);
  background: var(--soft);
  border-radius: 15px;
  margin-top: 42px;
}
.getting-started p {
  color: var(--muted);
  font-size: 12px;
  margin: 12px 0 18px;
  line-height: 1.8;
}
.getting-started ol {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 19px;
}
.getting-started li {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 12px;
}
.getting-started li span {
  color: var(--accent);
  font-size: 11px;
  font-family: monospace;
  border: 1px solid var(--line);
  background: var(--paper);
  padding: 5px 7px;
  border-radius: 6px;
}
.anonymous-note {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 6px;
  margin-top: 20px;
  color: var(--muted);
  font-size: 10px;
}
.plaza-description {
  margin-top: 20px;
  padding: 20px;
  border: 1px solid var(--line);
  border-radius: 12px;
  background: var(--paper);
  overflow-wrap: anywhere;
  font-size: 13px;
  line-height: 1.8;
}
.plaza-description :deep(a) {
  text-decoration: underline;
  color: var(--accent);
}
.plaza-description :deep(ul) {
  padding-left: 20px;
  list-style: disc;
}
.plaza-description :deep(ol) {
  padding-left: 20px;
  list-style: decimal;
}
.plaza-description :deep(h1),
.plaza-description :deep(h2),
.plaza-description :deep(h3) {
  margin: 12px 0;
  font-weight: 600;
}
.skeleton {
  height: 280px;
  padding-bottom: 20px;
}
.skeleton span,
.skeleton i {
  display: block;
  height: 15px;
  border-radius: 5px;
  background: var(--soft);
  animation: plaza-pulse 1.8s ease-in-out infinite;
}
.skeleton span {
  width: 43px;
  height: 43px;
  margin-bottom: 28px;
}
.skeleton i {
  margin-bottom: 17px;
}
.skeleton i:last-child {
  width: 65%;
  margin-top: auto;
}
@keyframes plaza-pulse {
  50% {
    opacity: 0.4;
  }
}
@media (max-width: 1150px) {
  .model-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .hero-stats {
    gap: 24px;
  }
  .hero-stats p {
    display: none;
  }
}
@media (max-width: 760px) {
  .hero-copy {
    padding: 30px 24px;
  }
  .hero-art {
    transform: scale(0.8);
    margin-left: -30px;
  }
  .catalog-layout > * {
    min-width: 0;
  }
  .catalog-layout {
    grid-template-columns: 1fr;
    gap: 18px;
  }
  .provider-options {
    flex-direction: row;
    overflow-x: auto;
    padding-bottom: 5px;
  }
  .provider-options button {
    width: auto;
    white-space: nowrap;
    flex-shrink: 0;
    border: 1px solid var(--line);
    padding: 9px 12px;
  }
  .filter-heading,
  .sidebar-note {
    display: none;
  }
  .section-heading {
    margin-bottom: 16px;
  }
  .hero-stats {
    padding: 17px 24px;
  }
  .getting-started {
    padding: 26px;
    gap: 24px;
  }
  .catalog-section {
    margin-top: 30px;
  }
}
@media (max-width: 520px) {
  .hero {
    display: block;
  }
  .hero-art {
    display: none;
  }
  .hero-copy h1 {
    font-size: 32px;
  }
  .hero-stats {
    justify-content: space-between;
    gap: 12px;
  }
  .hero-stats div {
    flex-direction: column;
    gap: 3px;
  }
  .hero-stats strong {
    font-size: 23px;
  }
  .hero-stats span {
    font-size: 10px;
  }
  .toolbar {
    flex-wrap: wrap;
  }
  .search-field {
    flex-basis: 100%;
  }
  .toolbar select {
    flex: 1;
    max-width: none;
  }
  .unit-hint {
    max-width: 130px;
    text-align: right;
    font-size: 9px;
    line-height: 1.6;
  }
  .model-grid {
    grid-template-columns: 1fr;
  }
  .model-card {
    padding: 21px 21px 0;
  }
  .model-card h3 {
    font-size: 19px;
  }
  .model-description {
    font-size: 12px;
    min-height: auto;
  }
  .card-top {
    margin-bottom: 15px;
  }
  .prices strong {
    font-size: 22px;
  }
  .card-details {
    font-size: 12px;
    padding: 17px 0;
  }
  .getting-started {
    grid-template-columns: 1fr;
  }
  .model-details {
    padding: 15px;
  }
  .results-meta {
    font-size: 9px;
  }
  .hero-copy .actions {
    gap: 18px;
  }
}
.console-summary {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 22px;
  padding: 24px 26px 0;
  border: 1px solid var(--line);
  border-radius: 16px;
  background: linear-gradient(110deg, var(--paper), var(--soft));
}
.console-intro {
  display: flex;
  flex: 1;
  gap: 15px;
  align-items: center;
}
.console-icon {
  display: grid;
  place-items: center;
  flex-shrink: 0;
  width: 44px;
  height: 44px;
  border-radius: 13px;
  color: var(--accent);
  background: #19896312;
}
.console-intro h2 {
  font-size: 19px;
  font-weight: 650;
  letter-spacing: -0.03em;
}
.console-intro p {
  margin-top: 6px;
  color: var(--muted);
  font-size: 12px;
  line-height: 1.7;
}
.console-actions {
  display: flex;
  align-items: center;
  gap: 9px;
}
.refresh-action {
  display: grid;
  place-items: center;
  width: 40px;
  height: 40px;
  border: 1px solid var(--line);
  border-radius: 9px;
  color: var(--muted);
  background: var(--paper);
}
.refresh-action:hover {
  color: var(--accent);
}
.refresh-action:disabled {
  opacity: 0.5;
  cursor: wait;
}
.console-stats {
  display: flex;
  flex-basis: 100%;
  align-items: center;
  flex-wrap: wrap;
  gap: 14px 28px;
  padding: 16px 0;
  border-top: 1px solid var(--line);
  color: var(--muted);
  font-size: 11px;
}
.console-stats > span {
  display: inline-flex;
  gap: 8px;
  align-items: center;
}
.console-stats strong {
  font-size: 18px;
  color: var(--ink);
  font-weight: 650;
}
.console-stats .console-sync {
  margin-left: auto;
}
.plaza-embedded .catalog-section {
  margin-top: 26px;
}
.plaza-embedded .section-heading {
  margin-bottom: 18px;
  align-items: center;
}
.plaza-embedded .section-heading h2 {
  font-size: 16px;
  margin-top: 0;
  letter-spacing: 0;
}
.plaza-embedded .catalog-layout {
  grid-template-columns: minmax(0, 1fr);
  gap: 18px;
}
.plaza-embedded .filter-heading,
.plaza-embedded .sidebar-note {
  display: none;
}
.plaza-embedded .provider-options {
  flex-direction: row;
  overflow-x: auto;
  gap: 8px;
  padding-bottom: 5px;
}
.plaza-embedded .provider-options button {
  width: auto;
  white-space: nowrap;
  flex-shrink: 0;
  padding: 9px 13px;
  border: 1px solid var(--line);
  background: var(--paper);
}
.plaza-embedded .provider-options button.active {
  border-color: var(--accent);
  background: #19896312;
}
.plaza-embedded .model-grid {
  grid-template-columns: repeat(auto-fill, minmax(min(100%, 260px), 1fr));
}
.plaza-embedded .model-details {
  scroll-margin-top: 90px;
}
@media (max-width: 760px) {
  .console-summary {
    padding: 18px 18px 0;
    gap: 16px;
  }
  .console-intro {
    flex-basis: 100%;
  }
  .console-stats {
    gap: 12px 20px;
  }
  .console-stats .console-sync {
    display: none;
  }
}
@media (prefers-reduced-motion: reduce) {
  .model-card {
    transition: none;
  }
  .model-card:hover {
    transform: none;
  }
  .skeleton span,
  .skeleton i {
    animation: none;
  }
}
</style>
