<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { api, ApiError } from '../api/client'
import type { ReadingText, TenseSummary, VocabCategory, VocabWord } from '../types/api'
import TenseCard from '../components/TenseCard.vue'
import VocabCard from '../components/VocabCard.vue'
import DeckCard from '../components/DeckCard.vue'
import TextCard from '../components/TextCard.vue'
import ImportText from '../components/ImportText.vue'
import Accordion from '../components/Accordion.vue'
import { useSavedWords } from '../reading/saved'
import { useDecks, textDeckID } from '../reading/decks'
import { useLanguage, t, format, pluralize } from '../i18n/language'
import { useSrsDeck, cardIdsFor, vocabDirection } from '../srs/store'
import { useRecents } from '../recent'
import { forgetPosition } from '../reading/prefs'
import { isRecord, read, write } from '../storage'
import { CLUSTER_ORDER, clusterFor, assertEveryCategoryMapped, type ClusterKey } from '../vocabClusters'
import { navigate } from '../router/router'

const { lang } = useLanguage()
const tenses = ref<TenseSummary[]>([])
const vocabCategories = ref<VocabCategory[]>([])
const vocabWords = ref<VocabWord[]>([])
const texts = ref<ReadingText[]>([])
// Reading ships for a subset of languages, so its absence is a normal
// state the tab reflects, not an error that blocks the page.
const readingAvailable = ref(true)
const loading = ref(true)
const error = ref('')
// Errors from an action (deleting a text) are deliberately *not* the same
// ref as errors from loading the page. They used to be, and because the
// page error is rendered as a `v-else-if` ahead of all the content, a
// failed DELETE replaced the tabs, the search box and the entire library
// with one line of red text and no way back.
const actionError = ref('')

let loadToken = 0

async function load() {
  const token = ++loadToken
  loading.value = true
  error.value = ''
  try {
    const [tensesResult, categoriesResult, wordsResult] = await Promise.all([
      api.getTenses(lang.value),
      api.getVocabCategories(lang.value),
      api.getVocabulary(lang.value),
    ])
    if (token !== loadToken) return
    tenses.value = tensesResult
    vocabCategories.value = categoriesResult
    vocabWords.value = wordsResult
    assertEveryCategoryMapped(lang.value, categoriesResult.map((c) => c.id))

    // Fetched separately and tolerantly: a language with no corpus 404s
    // here, and that must not take down grammar and vocabulary with it.
    try {
      const fetched = await api.getTexts(lang.value)
      if (token !== loadToken) return
      texts.value = fetched
      readingAvailable.value = true
    } catch {
      if (token !== loadToken) return
      texts.value = []
      readingAvailable.value = false
    }
  } catch (e) {
    if (token !== loadToken) return
    error.value = e instanceof ApiError ? e.message : t.value.home.error
  } finally {
    if (token === loadToken) loading.value = false
  }
}

onMounted(load)
watch(lang, load)

// --- Tabs: Grammar / Vocabulary, persisted like the language toggle --------
const TAB_STORAGE_KEY = 'sermokey:home-tab'
type Tab = 'grammar' | 'vocab' | 'texts'
function parseTab(u: unknown): Tab | null {
  return u === 'vocab' || u === 'texts' || u === 'grammar' ? u : null
}
const activeTab = ref<Tab>(read(TAB_STORAGE_KEY, parseTab, 'grammar'))
watch(activeTab, (v) => write(TAB_STORAGE_KEY, v))

const visibleTexts = computed(() => texts.value.filter((x) => matches(x.title, x.blurb)))

function onImported(text: ReadingText) {
  // Replace rather than append: re-importing the same file returns the
  // same id, and the library should show one copy.
  texts.value = [text, ...texts.value.filter((x) => x.id !== text.id)]
}

async function onRemoveText(id: string) {
  const text = texts.value.find((x) => x.id === id)
  if (!text) return
  // Say what happens to the words. Deleting a book must not look like it
  // might delete the vocabulary built from it — that is the part the
  // reader actually worked for, and it is kept.
  const savedFromText = useSavedWords(lang.value).fromText(id).length
  const message =
    format(t.value.reading.removeConfirm, { title: text.title }) +
    (savedFromText > 0 ? '\n\n' + pluralize(t.value.reading.removeKeepsWords, savedFromText) : '')
  if (!window.confirm(message)) return
  actionError.value = ''
  try {
    await api.deleteText(lang.value, id)

    // Materialise the text's deck before it stops being derivable.
    //
    // A text deck is normally derived from its words' context, and that
    // derivation is now scoped to texts still in the library — otherwise a
    // forgotten deck is rebuilt on the next render and can never be
    // removed. The consequence, if nothing is done here, is that a text
    // deck that was only ever derived (because its words were filed into a
    // custom deck instead) would silently disappear along with the book.
    //
    // Deleting a book must not delete a grouping the reader can still see.
    // Storing it turns it into an orphan they keep for as long as they
    // like, and remove when they choose.
    const saved = useSavedWords(lang.value)
    if (saved.fromText(id).length > 0) {
      decks.value.ensureTextDeck(id, text.title)
    }

    texts.value = texts.value.filter((x) => x.id !== id)
    // Deleting the text used to leave its reading position and its
    // "continue" chip behind. The chip then matched no text, fell through
    // to the vocabulary branch of goToRecent, and navigated to a category
    // that does not exist — a dead end whose only exit was the back button.
    forgetPosition(lang.value, id)
    useRecents(lang.value).forget(id)
  } catch (e) {
    actionError.value = e instanceof ApiError ? e.message : t.value.home.error
  }
}
const textsHaveNoResults = computed(
  () => searchQuery.value.trim() !== '' && visibleTexts.value.length === 0,
)

// Words captured while reading surface in the vocabulary tab alongside the
// curated categories, so the reading loop closes where a learner would
// already look for their cards rather than in a separate place.
const savedWords = computed(() => useSavedWords(lang.value).all())
const savedStats = computed(() => statsFor(savedWords.value.map((w) => ({ id: w.id }))))

// --- Named decks ----------------------------------------------------------
const decks = computed(() => useDecks(lang.value))

/**
 * Decks with their word counts and due counts.
 *
 * A deck with no words is still listed: it was created deliberately, and
 * hiding it would make creating one look like it had failed.
 */
/** Ids of texts currently in the library, for spotting orphaned decks. */
const libraryTextIDs = computed(() => new Set(texts.value.map((x) => x.id)))

const deckCards = computed(() => {
  const saved = useSavedWords(lang.value)
  const inLibrary = libraryTextIDs.value
  const byId = new Map(decks.value.decks.value.map((d) => [d.id, d]))

  // A text you have saved words from gets a deck here even if those words
  // were filed into a custom deck instead. "Words I met in La Parure" is a
  // fact about the words, so the deck is derivable — and a reader who sent
  // a word to "Maupassant" while reading it should still find it under the
  // text, rather than having the text deck silently not exist.
  //
  // Derived *only for texts still in the library*. Deriving unconditionally
  // is what made a text deck impossible to remove: forgetting one deleted
  // the stored entry, and the very next render rebuilt it from the words'
  // context. An explicitly stored deck still shows after its text is gone,
  // so an orphan stays visible until the reader deals with it.
  for (const w of saved.all()) {
    const textId = w.context?.textId
    if (!textId || !inLibrary.has(textId)) continue
    const id = textDeckID(textId)
    if (byId.has(id)) continue
    byId.set(id, {
      id,
      name: w.context?.textTitle || textId,
      textId,
      createdAt: w.savedAt,
    })
  }

  return [...byId.values()]
    .map((d) => {
      const words = saved.inDeck(d.id)
      // A text deck whose text has been deleted. The grouping outlived the
      // thing it grouped, so it is the reader's to remove.
      const orphaned = !!d.textId && !inLibrary.has(d.textId)
      return {
        deck: d,
        orphaned,
        count: words.length,
        stats: statsFor(words.map((w) => ({ id: w.id }))),
      }
    })
    // Decks the reader made come first: they were a deliberate act, where a
    // text's deck appears merely because a word was saved there. Within a
    // kind, the fullest deck leads.
    .sort((a, b) => {
      const kind = Number(!!a.deck.textId) - Number(!!b.deck.textId)
      if (kind !== 0) return kind
      return b.count - a.count || a.deck.name.localeCompare(b.deck.name)
    })
})

/** True once there is anything of the reader's own to show. */
const hasOwnDecks = computed(() => savedWords.value.length > 0 || deckCards.value.length > 0)

function forgetDeck(id: string, name: string) {
  if (!window.confirm(format(t.value.reading.deckForgetConfirm, { name }))) return
  // Membership is cleaned up before the deck itself, so no word is left
  // pointing at an id that no longer resolves.
  useSavedWords(lang.value).removeFromAllDecks(id)
  decks.value.remove(id)
}

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
function parseExpanded(u: unknown): Record<string, boolean> | null {
  if (!isRecord(u)) return null
  const out: Record<string, boolean> = {}
  for (const [k, v] of Object.entries(u)) if (typeof v === 'boolean') out[k] = v
  return out
}
const expandedMap = ref<Record<string, boolean>>(read(EXPANDED_KEY, parseExpanded, {}))
function isExpanded(id: string, defaultOpen: boolean): boolean {
  if (searchQuery.value.trim() !== '') return true
  return expandedMap.value[id] ?? defaultOpen
}
function setExpanded(id: string, value: boolean) {
  expandedMap.value = { ...expandedMap.value, [id]: value }
  write(EXPANDED_KEY, expandedMap.value)
}

// --- "Continue where you left off" ----------------------------------------
const recents = computed(() => useRecents(lang.value).items.value)
function goToRecent(kind: 'theory' | 'quiz' | 'vocab', id: string) {
  if (kind === 'theory') navigate({ name: 'theory', tense: id })
  else if (kind === 'quiz') navigate({ name: 'quiz-setup', tense: id })
  // Texts are recorded under the same 'vocab' recent kind; if the id names
  // a text, resume reading it, otherwise it is a vocabulary category.
  else if (texts.value.some((x) => x.id === id)) navigate({ name: 'read', textId: id, section: 0 })
  else navigate({ name: 'flashcards', deck: { kind: 'category', id } })
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
/**
 * Due/new counts per category, computed once per render rather than per
 * card. The template called `statsFor` twice for each of ~27 categories,
 * and each call walked the entire deck allocating a throwaway object per
 * card — so every keystroke in the search box triggered ~54 full deck
 * scans.
 */
const statsByCategory = computed(() => {
  const out = new Map<string, { dueNow: number; new: number }>()
  for (const [categoryId, words] of wordsByCategory.value) {
    out.set(categoryId, statsFor(words))
  }
  return out
})

function statsForCategory(categoryId: string) {
  return statsByCategory.value.get(categoryId) ?? { dueNow: 0, new: 0 }
}

function statsFor(words: { id: string }[]) {
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

    <p v-if="loading" class="muted" role="status">{{ t.home.loading }}</p>
    <div v-else-if="error" class="error-block" role="alert">
      <p class="error-text">{{ error }}</p>
      <button class="btn" @click="load()">{{ t.common.retry }}</button>
    </div>

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
        <button
          v-if="readingAvailable"
          type="button"
          role="tab"
          class="tab-btn"
          :class="{ active: activeTab === 'texts' }"
          :aria-selected="activeTab === 'texts'"
          @click="activeTab = 'texts'"
        >
          {{ t.reading.tab }}
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

      <template v-else-if="activeTab === 'vocab'">
        <!-- Two labelled groups.
             The reader's own collections and the curated curriculum are
             different kinds of thing — one grows with every page read, the
             other is fixed — and they were sharing one undifferentiated
             grid. The only hint at the difference was an accent border on
             two cards out of three, which reads as arbitrary rather than
             as a rule. Grouping them structurally lets the accent carry a
             single meaning: "the everything card of this group". -->
        <template v-if="hasOwnDecks && !searchQuery">
          <h2 class="group-heading">{{ t.reading.deckManage }}</h2>
          <div class="grid grid-cols-2">
            <DeckCard
              v-if="savedWords.length"
              :name="t.reading.savedDeckTitle"
              :kind="t.reading.deckAll"
              :blurb="t.reading.savedDeckBlurb"
              :count="savedWords.length"
              :count-label="t.reading.deckCountLabel"
              :due-count="savedStats.dueNow"
              aggregate
              @start="navigate({ name: 'flashcards', deck: { kind: 'saved' } })"
            />
            <!-- Forgettable when the reader made it, or when its text is
                 gone and the grouping has outlived what it grouped. A live
                 text deck is withheld because removing it would be a
                 visible no-op: it is rebuilt from the words' own context on
                 the next render. The pill says which case a card is in, so
                 the asymmetry is explained rather than just observed. -->
            <DeckCard
              v-for="d in deckCards"
              :key="d.deck.id"
              :name="d.deck.name"
              :kind="d.orphaned ? t.reading.deckOrphaned : d.deck.textId ? t.reading.tab : t.reading.deckManage"
              :count="d.count"
              :count-label="t.reading.deckCountLabel"
              :due-count="d.stats.dueNow"
              :disabled="d.count === 0"
              :orphaned="d.orphaned"
              :forgettable="!d.deck.textId || d.orphaned"
              @start="navigate({ name: 'flashcards', deck: { kind: 'deck', id: d.deck.id } })"
              @forget="forgetDeck(d.deck.id, d.deck.name)"
            />
          </div>

          <h2 class="group-heading">{{ t.vocab.sectionTitle }}</h2>
        </template>

        <div v-if="!searchQuery" class="grid grid-cols-2 mixed-row">
          <VocabCard
            id=""
            :name-native="t.vocab.mixedTitle"
            :name-en="t.vocab.mixedPill"
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
              :due-count="statsForCategory(cat.id).dueNow"
              :new-count="statsForCategory(cat.id).new"
            />
          </div>
        </Accordion>

        <p v-if="vocabHasNoResults" class="muted no-results">{{ t.home.noResults }}</p>
      </template>

      <template v-else>
        <p class="muted intro">{{ t.reading.sectionIntro }}</p>
        <ImportText @imported="onImported" />

        <!-- Scoped to the action, dismissible, and rendered beside the
             library rather than instead of the whole page. -->
        <div v-if="actionError" class="action-error" role="alert">
          <span>{{ actionError }}</span>
          <button class="link" @click="actionError = ''">{{ t.common.dismiss }}</button>
        </div>
        <div class="grid grid-cols-2">
          <TextCard v-for="x in visibleTexts" :key="x.id" :text="x" @remove="onRemoveText" />
        </div>
        <p v-if="textsHaveNoResults" class="muted no-results">{{ t.home.noResults }}</p>
      </template>
    </template>
  </section>
</template>

<style scoped>
.group-heading {
  font-size: 1.05rem;
  margin: var(--space-6) 0 var(--space-3);
  color: var(--color-heading);
}
.group-heading:first-of-type {
  margin-top: var(--space-4);
}

.deck-card h3 {
  margin-top: 0;
}

.error-block {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: var(--space-3);
}
.action-error {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  background: var(--color-danger-bg, #fdeaea);
  color: var(--color-danger);
  border-radius: var(--radius-sm);
  padding: var(--space-2) var(--space-3);
  margin-top: var(--space-3);
  font-size: 0.88rem;
}
.action-error .link {
  background: none;
  border: none;
  font: inherit;
  font-weight: 600;
  color: inherit;
  text-decoration: underline;
  cursor: pointer;
}

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
  /* Three word-labelled tabs do not fit across a 320px screen. Wrapping
     is preferable to the alternatives: abbreviating the labels makes the
     sections harder to recognise, and letting them overflow drags the
     whole page into a horizontal scroll. */
  flex-wrap: wrap;
  max-width: 100%;
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

@media (max-width: 640px) {
  .tab-btn {
    padding: 0.5rem 0.85rem;
    font-size: 0.88rem;
  }
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
/* A border-colour change alone is a colour-only focus indicator, which is
   invisible to an achromatic user. The ring is kept and the colour change
   becomes the secondary cue. */
.search-input:focus-visible {
  outline: 2px solid var(--color-accent);
  outline-offset: 2px;
  border-color: var(--color-accent);
}

.no-results {
  margin-top: var(--space-4);
}

</style>
