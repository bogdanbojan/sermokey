<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { api, ApiError } from '../api/client'
import type { TenseSummary, VocabCategory, VocabWord } from '../types/api'
import TenseCard from '../components/TenseCard.vue'
import VocabCard from '../components/VocabCard.vue'
import Accordion from '../components/Accordion.vue'
import { useLanguage, t } from '../i18n/language'
import { useSrsDeck, cardIdsFor, vocabDirection } from '../srs/store'
import { useRecents } from '../recent'
import { CLUSTER_ORDER, clusterFor, type ClusterKey } from '../vocabClusters'
import { navigate } from '../router/router'

const { lang } = useLanguage()
const tenses = ref<TenseSummary[]>([])
const vocabCategories = ref<VocabCategory[]>([])
const vocabWords = ref<VocabWord[]>([])
const loading = ref(true)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [tensesResult, categoriesResult, wordsResult] = await Promise.all([
      api.getTenses(lang.value),
      api.getVocabCategories(lang.value),
      api.getVocabulary(lang.value),
    ])
    tenses.value = tensesResult
    vocabCategories.value = categoriesResult
    vocabWords.value = wordsResult
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : t.value.home.error
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(lang, load)

// --- Tabs: Grammar / Vocabulary, persisted like the language toggle --------
const TAB_STORAGE_KEY = 'sermokey:home-tab'
const activeTab = ref<'grammar' | 'vocab'>(
  localStorage.getItem(TAB_STORAGE_KEY) === 'vocab' ? 'vocab' : 'grammar',
)
watch(activeTab, (v) => localStorage.setItem(TAB_STORAGE_KEY, v))

// --- Search --------------------------------------------------------------
const searchQuery = ref('')
function matches(nameNative: string, nameEn: string): boolean {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return true
  return nameNative.toLowerCase().includes(q) || nameEn.toLowerCase().includes(q)
}

// --- Grammar: group tenses by their existing category, in an accordion per group ---
const grammarGroups = computed(() => {
  const order: string[] = []
  const byCategory = new Map<string, TenseSummary[]>()
  for (const tense of tenses.value) {
    if (!byCategory.has(tense.category)) {
      byCategory.set(tense.category, [])
      order.push(tense.category)
    }
    byCategory.get(tense.category)!.push(tense)
  }
  return order.map((category, i) => ({
    key: category,
    title: category,
    items: byCategory.get(category)!.filter((tense) => matches(tense.nameNative, tense.nameEn)),
    defaultOpen: i === 0,
  }))
})
const visibleGrammarGroups = computed(() => grammarGroups.value.filter((g) => g.items.length > 0))
const grammarHasNoResults = computed(
  () => searchQuery.value.trim() !== '' && visibleGrammarGroups.value.length === 0,
)

// --- Vocabulary: group categories into 6 semantic clusters, in an accordion per cluster ---
const vocabClusters = computed(() => {
  const byCluster = new Map<ClusterKey, VocabCategory[]>()
  for (const cat of vocabCategories.value) {
    const key = clusterFor(lang.value, cat.id)
    if (!byCluster.has(key)) byCluster.set(key, [])
    byCluster.get(key)!.push(cat)
  }
  return CLUSTER_ORDER.map((key, i) => ({
    key,
    title: t.value.home.clusters[key],
    items: (byCluster.get(key) ?? []).filter((cat) => matches(cat.nameNative, cat.nameEn)),
    defaultOpen: i === 0,
  }))
})
const visibleVocabClusters = computed(() => vocabClusters.value.filter((c) => c.items.length > 0))
const vocabHasNoResults = computed(
  () => searchQuery.value.trim() !== '' && visibleVocabClusters.value.length === 0,
)

// Force-expand any group/cluster containing a search match, otherwise fall
// back to the persisted/default state — "search reveals results" is a
// well-established pattern for collapsed content.
const EXPANDED_KEY = 'sermokey:expanded'
function loadExpandedMap(): Record<string, boolean> {
  try {
    const raw = localStorage.getItem(EXPANDED_KEY)
    return raw ? (JSON.parse(raw) as Record<string, boolean>) : {}
  } catch {
    return {}
  }
}
const expandedMap = ref<Record<string, boolean>>(loadExpandedMap())
function isExpanded(id: string, defaultOpen: boolean): boolean {
  if (searchQuery.value.trim() !== '') return true
  return expandedMap.value[id] ?? defaultOpen
}
function setExpanded(id: string, value: boolean) {
  expandedMap.value = { ...expandedMap.value, [id]: value }
  localStorage.setItem(EXPANDED_KEY, JSON.stringify(expandedMap.value))
}

// --- "Continue where you left off" ----------------------------------------
const recents = computed(() => useRecents(lang.value).items.value)
function goToRecent(kind: 'theory' | 'quiz' | 'vocab', id: string) {
  if (kind === 'theory') navigate({ name: 'theory', tense: id })
  else if (kind === 'quiz') navigate({ name: 'quiz-setup', tense: id })
  else navigate({ name: 'flashcards', category: id })
}

const deck = computed(() => useSrsDeck(lang.value))
const wordsByCategory = computed(() => {
  const map = new Map<string, VocabWord[]>()
  for (const w of vocabWords.value) {
    if (!map.has(w.category)) map.set(w.category, [])
    map.get(w.category)!.push(w)
  }
  return map
})
function statsFor(words: VocabWord[]) {
  if (!words.length) return { dueNow: 0, new: 0 }
  const s = deck.value.stats(cardIdsFor(words, vocabDirection.value))
  return { dueNow: s.dueNow, new: s.new }
}
const totalVocabWords = computed(() => vocabCategories.value.reduce((sum, c) => sum + c.wordCount, 0))
const mixedStats = computed(() => statsFor(vocabWords.value))
</script>

<template>
  <section>
    <h1>{{ t.home.title }}</h1>
    <p class="muted intro">{{ t.home.intro }}</p>

    <p v-if="loading" class="muted">{{ t.home.loading }}</p>
    <p v-else-if="error" class="error-text">{{ error }}</p>

    <template v-else>
      <!-- Continue where you left off: optimizes for the returning daily
           user, not just first-time discovery. -->
      <section v-if="recents.length" class="continue-section">
        <h2 class="continue-title">{{ t.home.continueTitle }}</h2>
        <div class="continue-chips">
          <button
            v-for="r in recents"
            :key="r.kind + r.id"
            class="continue-chip"
            @click="goToRecent(r.kind, r.id)"
          >
            {{ r.label }}
          </button>
        </div>
      </section>

      <div class="tabs" role="tablist">
        <button
          type="button"
          role="tab"
          class="tab-btn"
          :class="{ active: activeTab === 'grammar' }"
          :aria-selected="activeTab === 'grammar'"
          @click="activeTab = 'grammar'"
        >
          {{ t.home.grammarTab }}
        </button>
        <button
          type="button"
          role="tab"
          class="tab-btn"
          :class="{ active: activeTab === 'vocab' }"
          :aria-selected="activeTab === 'vocab'"
          @click="activeTab = 'vocab'"
        >
          {{ t.home.vocabTab }}
        </button>
      </div>

      <input
        v-model="searchQuery"
        type="search"
        class="search-input"
        :placeholder="t.home.searchPlaceholder"
      />

      <template v-if="activeTab === 'grammar'">
        <div v-if="!searchQuery" class="grid grid-cols-2 mixed-row">
          <TenseCard
            id="mixed"
            :name-native="t.home.mixedTitle"
            name-en="All tenses"
            :usage="t.home.mixedUsage"
            is-mixed
          />
        </div>

        <Accordion
          v-for="group in visibleGrammarGroups"
          :key="group.key"
          :title="group.title"
          :badge="group.items.length"
          :model-value="isExpanded('g:' + group.key, group.defaultOpen)"
          @update:model-value="(v) => setExpanded('g:' + group.key, v)"
        >
          <div class="grid grid-cols-2">
            <TenseCard
              v-for="tense in group.items"
              :key="tense.id"
              :id="tense.id"
              :name-native="tense.nameNative"
              :name-en="tense.nameEn"
              :usage="tense.usage"
            />
          </div>
        </Accordion>

        <p v-if="grammarHasNoResults" class="muted no-results">{{ t.home.noResults }}</p>
      </template>

      <template v-else>
        <div v-if="!searchQuery" class="grid grid-cols-2 mixed-row">
          <VocabCard
            id=""
            :name-native="t.vocab.mixedTitle"
            name-en="All vocabulary"
            :word-count="totalVocabWords"
            :due-count="mixedStats.dueNow"
            :new-count="mixedStats.new"
            is-mixed
          />
        </div>

        <Accordion
          v-for="cluster in visibleVocabClusters"
          :key="cluster.key"
          :title="cluster.title"
          :badge="cluster.items.length"
          :model-value="isExpanded('v:' + cluster.key, cluster.defaultOpen)"
          @update:model-value="(v) => setExpanded('v:' + cluster.key, v)"
        >
          <div class="grid grid-cols-2">
            <VocabCard
              v-for="cat in cluster.items"
              :key="cat.id"
              :id="cat.id"
              :name-native="cat.nameNative"
              :name-en="cat.nameEn"
              :word-count="cat.wordCount"
              :due-count="statsFor(wordsByCategory.get(cat.id) ?? []).dueNow"
              :new-count="statsFor(wordsByCategory.get(cat.id) ?? []).new"
            />
          </div>
        </Accordion>

        <p v-if="vocabHasNoResults" class="muted no-results">{{ t.home.noResults }}</p>
      </template>
    </template>
  </section>
</template>

<style scoped>
.intro {
  max-width: 60ch;
}
.error-text {
  color: var(--color-danger);
}
.mixed-row {
  margin-bottom: var(--space-5);
}

.continue-section {
  margin-bottom: var(--space-5);
}
.continue-title {
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
  margin-bottom: var(--space-2);
}
.continue-chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}
.continue-chip {
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  border-radius: 999px;
  padding: 0.45rem 1rem;
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--color-primary);
  cursor: pointer;
}
.continue-chip:hover {
  border-color: var(--color-primary);
}

.tabs {
  display: inline-flex;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: 3px;
  background: var(--color-surface);
  margin-bottom: var(--space-4);
}
.tab-btn {
  border: none;
  background: none;
  padding: 0.55rem 1.3rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--color-text-muted);
  border-radius: calc(var(--radius-sm) - 2px);
  cursor: pointer;
}
.tab-btn:hover {
  color: var(--color-text);
}
.tab-btn.active {
  background: var(--color-primary);
  color: var(--color-primary-contrast);
}

.search-input {
  display: block;
  width: 100%;
  max-width: 24rem;
  padding: 0.6rem 0.9rem;
  margin-bottom: var(--space-5);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-surface);
  color: var(--color-text);
}
.search-input:focus {
  outline: none;
  border-color: var(--color-accent);
}

.no-results {
  margin-top: var(--space-4);
}
</style>
