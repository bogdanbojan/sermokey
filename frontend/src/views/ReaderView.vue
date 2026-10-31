<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { api, ApiError } from '../api/client'
import type { ReadingEntry, ReadingSection } from '../types/api'
import { navigate } from '../router/router'
import { useLanguage, t, format } from '../i18n/language'
import { useWordStatus, type WordStatus } from '../reading/status'
import { useSavedWords, type WordContext } from '../reading/saved'
import {
  readingPrefs,
  savePosition,
  useLookupCounts,
  MIN_FONT_SIZE,
  MAX_FONT_SIZE,
} from '../reading/prefs'
import {
  renderTokens,
  selectionToRange,
  rangeText,
  sentenceAround,
  tokenIndexOf,
  MAX_PHRASE_TOKENS,
  type TokenRange,
} from '../reading/tokens'
import ReaderParagraph from '../components/ReaderParagraph.vue'
import WordPopover from '../components/WordPopover.vue'
import WordPanel from '../components/WordPanel.vue'
import SelectionToolbar from '../components/SelectionToolbar.vue'

const props = defineProps<{ textId: string; section: number }>()

const { lang } = useLanguage()
const { statusOf } = useWordStatus(lang)

// Status for every entry in the current section's lexicon.
//
// Computed, not a function handed to each paragraph: doing the lookups
// here puts the reads on the deck, the saved-word store and the
// known/ignored sets inside a tracked computation, so saving a word
// repaints it immediately. It also costs one pass per change instead of a
// call per token per render.
//
// Function words and proper nouns are resolved to 'no-entry' at this
// point: neither is vocabulary, and decorating every "le", "de" and
// "Claire" would bury the words actually worth noticing.
const entryStatus = computed<Record<number, WordStatus | 'no-entry'>>(() => {
  const out: Record<number, WordStatus | 'no-entry'> = {}
  const lexicon = section.value?.lexicon ?? []
  for (let i = 0; i < lexicon.length; i++) {
    const e = lexicon[i]!
    out[i] = e.function || e.proper ? 'no-entry' : statusOf(e.wordId)
  }
  return out
})
const saved = computed(() => useSavedWords(lang.value))
const lookups = computed(() => useLookupCounts(lang.value))

type Phase = 'loading' | 'error' | 'ready'
const phase = ref<Phase>('loading')
const error = ref('')
const section = ref<ReadingSection | null>(null)
const bodyEl = ref<HTMLElement | null>(null)

async function load() {
  phase.value = 'loading'
  error.value = ''
  closePeek()
  closePinned()
  try {
    section.value = await api.getTextSection(lang.value, props.textId, props.section)
    phase.value = 'ready'
    savePosition(lang.value, props.textId, { section: props.section, paragraph: 0 })
    window.scrollTo({ top: 0 })
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : t.value.reading.error
    phase.value = 'error'
  }
}

onMounted(load)
watch([() => props.textId, () => props.section, lang], load)

// --- Hover peek ----------------------------------------------------------
// A short intent delay. Without it the popup fires on every word the
// pointer crosses on its way somewhere else, which is both visually noisy
// and exactly the kind of constant interruption that breaks reading fluency.
const HOVER_INTENT_MS = 180

const peekEntry = ref<ReadingEntry | null>(null)
const peekX = ref(0)
const peekY = ref(0)
let hoverTimer: number | undefined

function entryAt(tokenIndex: number, paragraphIndex: number): ReadingEntry | undefined {
  const para = section.value?.paragraphs[paragraphIndex]
  if (!para) return undefined
  const ref_ = para.spans[tokenIndex * 3 + 2]
  return ref_ !== undefined && ref_ >= 0 ? section.value?.lexicon[ref_] : undefined
}

function locate(target: EventTarget | null): { token: number; paragraph: number } | undefined {
  const el = target instanceof Element ? target : null
  if (!el) return undefined
  const token = tokenIndexOf(el)
  const paraEl = el.closest<HTMLElement>('[data-paragraph]')
  const paragraph = paraEl ? Number(paraEl.dataset.paragraph) : NaN
  if (token === undefined || !Number.isInteger(paragraph)) return undefined
  return { token, paragraph }
}

function onPointerOver(e: PointerEvent) {
  // Touch has no hover; tapping goes straight to the pinned panel, so
  // firing a peek here would just flash a popup before the real one.
  if (e.pointerType === 'touch') return
  const el = e.target instanceof Element ? e.target : null
  if (!el?.classList.contains('tok')) return
  const at = locate(el)
  if (!at) return
  const entry = entryAt(at.token, at.paragraph)
  if (!entry) return closePeek()

  const rect = el.getBoundingClientRect()
  window.clearTimeout(hoverTimer)
  hoverTimer = window.setTimeout(() => {
    peekX.value = rect.left + rect.width / 2
    peekY.value = rect.bottom
    peekEntry.value = entry
  }, HOVER_INTENT_MS)
}

function onPointerOut(e: PointerEvent) {
  const el = e.target instanceof Element ? e.target : null
  if (el?.classList.contains('tok')) closePeek()
}

function closePeek() {
  window.clearTimeout(hoverTimer)
  peekEntry.value = null
}

// --- Pinned panel --------------------------------------------------------
const pinned = ref<{ entry: ReadingEntry; token: number; paragraph: number } | null>(null)
const pinnedContext = ref<WordContext | undefined>()

function pin(entry: ReadingEntry, token: number, paragraph: number) {
  closePeek()
  pinned.value = { entry, token, paragraph }
  pinnedContext.value = contextFor(token, paragraph)
  // Every pin is a lookup, and repeated lookups are what trigger the
  // save nudge.
  lookups.value.record(entry.wordId)
}

function closePinned() {
  pinned.value = null
  pinnedContext.value = undefined
}

/** The sentence a word sits in, carried onto its flashcard. */
function contextFor(token: number, paragraph: number): WordContext | undefined {
  const para = section.value?.paragraphs[paragraph]
  if (!para || !section.value) return undefined
  const { sentence, start, end } = sentenceAround(renderTokens(para), token)
  return {
    textId: section.value.textId,
    textTitle: section.value.title,
    section: section.value.index,
    sentence,
    start,
    end,
  }
}

function onClick(e: MouseEvent) {
  const el = e.target instanceof Element ? e.target : null
  if (!el?.classList.contains('tok')) return
  // A click that ends a drag is a selection, not a word tap; the selection
  // toolbar handles it.
  if (!window.getSelection()?.isCollapsed) return
  const at = locate(el)
  if (!at) return
  const entry = entryAt(at.token, at.paragraph)
  if (entry) pin(entry, at.token, at.paragraph)
}

// --- Phrase selection ----------------------------------------------------
const selection = ref<{ text: string; x: number; y: number; tooLong: boolean } | null>(null)
let selectionRange: TokenRange | null = null

function onSelectionChange() {
  const range = selectionToRange(bodyEl.value)
  if (!range) {
    selection.value = null
    selectionRange = null
    return
  }
  const para = section.value?.paragraphs[range.paragraph]
  if (!para) return
  const tokens = renderTokens(para)
  const text = rangeText(tokens, range)
  if (!text) {
    selection.value = null
    return
  }
  const wordCount = tokens.slice(range.from, range.to + 1).filter((tk) => tk.word).length
  const rect = window.getSelection()?.getRangeAt(0).getBoundingClientRect()
  if (!rect) return

  selectionRange = range
  selection.value = {
    text,
    x: rect.left + rect.width / 2,
    y: rect.top - 6,
    tooLong: wordCount > MAX_PHRASE_TOKENS,
  }
}

async function lookupSelection() {
  if (!selectionRange || !selection.value) return
  const q = selection.value.text
  try {
    const entries = await api.lookup(lang.value, q)
    const range = selectionRange
    if (entries.length > 0) {
      pin(entries[0]!, range.from, range.paragraph)
    } else {
      // Nothing in the dictionary is a normal outcome for an arbitrary
      // span, and the reader should still be able to save it — a phrase
      // they noticed is worth keeping even if Wiktionary has no entry.
      pin(
        {
          surface: q,
          lemma: q,
          pos: 'other',
          senses: [],
          wordId: `${q}|lecture`,
          curated: false,
        },
        range.from,
        range.paragraph,
      )
    }
  } catch {
    /* lookup failure leaves the selection in place; nothing is lost */
  }
  window.getSelection()?.removeAllRanges()
  selection.value = null
}

// --- Navigation ----------------------------------------------------------
const hasPrev = computed(() => props.section > 0)
const hasNext = computed(() => !!section.value && props.section < section.value.sectionCount - 1)
const isLastSection = computed(() => !!section.value && !hasNext.value)

function goSection(delta: number) {
  navigate({ name: 'read', textId: props.textId, section: props.section + delta })
}

const savedFromText = computed(() => saved.value.fromText(props.textId).length)

// --- Keyboard ------------------------------------------------------------
// The flashcards view already establishes single-key shortcuts, so the
// reader matches. It also makes the feature operable without a pointer,
// rather than a mouse-only toy.
function onKeydown(e: KeyboardEvent) {
  if (e.metaKey || e.ctrlKey || e.altKey) return
  const typing = (e.target as HTMLElement)?.tagName === 'INPUT'
  if (typing) return

  switch (e.key) {
    case 'Escape':
      closePinned()
      closePeek()
      break
    case 'h':
      readingPrefs.value.highlighting = !readingPrefs.value.highlighting
      break
    case 'ArrowRight':
      if (hasNext.value) goSection(1)
      break
    case 'ArrowLeft':
      if (hasPrev.value) goSection(-1)
      break
  }
}

onMounted(() => {
  document.addEventListener('selectionchange', onSelectionChange)
  window.addEventListener('keydown', onKeydown)
})
onBeforeUnmount(() => {
  document.removeEventListener('selectionchange', onSelectionChange)
  window.removeEventListener('keydown', onKeydown)
  window.clearTimeout(hoverTimer)
})

// --- Display controls ----------------------------------------------------
function nudgeFontSize(delta: number) {
  const next = Math.round((readingPrefs.value.fontSize + delta) * 1000) / 1000
  readingPrefs.value.fontSize = Math.min(MAX_FONT_SIZE, Math.max(MIN_FONT_SIZE, next))
}

const bodyStyle = computed(() => ({
  fontSize: `${readingPrefs.value.fontSize}rem`,
  fontFamily: readingPrefs.value.serif
    ? 'Georgia, "Iowan Old Style", "Times New Roman", serif'
    : 'var(--font-sans)',
}))

// --- First-run coachmark -------------------------------------------------
// Hover and drag-select are invisible affordances; a reader who does not
// discover them sees a plain wall of French.
const COACH_KEY = 'sermokey:reading-coach-seen'
const showCoach = ref(localStorage.getItem(COACH_KEY) !== '1')
function dismissCoach() {
  showCoach.value = false
  localStorage.setItem(COACH_KEY, '1')
}
</script>

<template>
  <section class="reader">
    <div class="toolbar-row">
      <button class="btn back" @click="navigate({ name: 'home' })">{{ t.reading.back }}</button>

      <div v-if="phase === 'ready'" class="display-controls" :aria-label="t.reading.displayLabel" role="group">
        <button class="ctl" :title="t.reading.smaller" @click="nudgeFontSize(-0.125)">A−</button>
        <button class="ctl" :title="t.reading.larger" @click="nudgeFontSize(0.125)">A+</button>
        <button class="ctl" @click="readingPrefs.serif = !readingPrefs.serif">
          {{ readingPrefs.serif ? t.reading.serif : t.reading.sans }}
        </button>
        <button
          class="ctl"
          :class="{ on: readingPrefs.highlighting }"
          :title="readingPrefs.highlighting ? t.reading.highlightOn : t.reading.highlightOff"
          @click="readingPrefs.highlighting = !readingPrefs.highlighting"
        >
          <kbd>H</kbd>
        </button>
      </div>
    </div>

    <p v-if="phase === 'loading'" class="muted">{{ t.reading.loading }}</p>
    <p v-else-if="phase === 'error'" class="error-text">{{ error }}</p>

    <template v-else-if="section">
      <header class="head measure">
        <h1>{{ section.title }}</h1>
        <p class="muted progress">
          {{ format(t.reading.sectionProgress, { index: section.index + 1, total: section.sectionCount }) }}
          <span v-if="savedFromText"> · {{ format(t.reading.savedCount, { count: savedFromText }) }}</span>
        </p>
      </header>

      <div v-if="showCoach" class="coach measure">
        <strong>{{ t.reading.coachTitle }}</strong>
        <ul>
          <li>{{ t.reading.coachHover }}</li>
          <li>{{ t.reading.coachClick }}</li>
          <li>{{ t.reading.coachSelect }}</li>
        </ul>
        <button class="btn" @click="dismissCoach">{{ t.reading.coachDismiss }}</button>
      </div>

      <h2 v-if="section.sectionTitle" class="section-title measure">{{ section.sectionTitle }}</h2>

      <!-- The reading surface. Measure, leading and face are tuned for
           long-form L2 prose, not for UI. -->
      <article
        ref="bodyEl"
        class="body"
        :style="bodyStyle"
        @pointerover="onPointerOver"
        @pointerout="onPointerOut"
        @click="onClick"
      >
        <ReaderParagraph
          v-for="(p, i) in section.paragraphs"
          :key="i"
          :paragraph="p"
          :index="i"
          :entry-status="entryStatus"
          :highlighting="readingPrefs.highlighting"
          :pinned="pinned?.paragraph === i ? pinned.token : undefined"
        />
      </article>

      <nav class="section-nav">
        <button class="btn" :disabled="!hasPrev" @click="goSection(-1)">{{ t.reading.prevSection }}</button>
        <button class="btn" :disabled="!hasNext" @click="goSection(1)">{{ t.reading.nextSection }}</button>
      </nav>

      <!-- Closing the loop. Extensive reading on its own yields very slow
           uptake; pairing it with deliberate review is what makes the
           words stick (Laufer 2005), so the handoff is offered the moment
           the text ends rather than left for the reader to remember. -->
      <div v-if="isLastSection" class="card finish-card">
        <h2>{{ t.reading.finishTitle }}</h2>
        <p v-if="savedFromText" class="muted">
          {{ format(t.reading.finishDetail, { count: savedFromText }) }}
        </p>
        <p v-else class="muted">{{ t.reading.noneSaved }}</p>
        <button
          v-if="savedFromText"
          class="btn btn-primary"
          @click="navigate({ name: 'flashcards', deck: { kind: 'text', id: props.textId } })"
        >
          {{ t.reading.reviewSaved }}
        </button>
      </div>

      <p class="muted licence">{{ section.source }} · {{ section.license }}</p>
    </template>

    <WordPopover :entry="peekEntry" :x="peekX" :y="peekY" />

    <SelectionToolbar
      v-if="selection"
      :text="selection.text"
      :x="selection.x"
      :y="selection.y"
      :too-long="selection.tooLong"
      @lookup="lookupSelection"
      @dismiss="selection = null"
    />

    <WordPanel
      v-if="pinned"
      :key="pinned.entry.surface + pinned.token"
      :entry="pinned.entry"
      :context="pinnedContext"
      @close="closePinned"
    />
  </section>
</template>

<style scoped>
.toolbar-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
  flex-wrap: wrap;
}
.error-text {
  color: var(--color-danger);
}

.display-controls {
  display: inline-flex;
  gap: 2px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: 2px;
  background: var(--color-surface);
}
.ctl {
  border: none;
  background: none;
  padding: 0.3rem 0.6rem;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--color-text-muted);
  border-radius: calc(var(--radius-sm) - 2px);
  cursor: pointer;
}
.ctl:hover {
  color: var(--color-text);
}
.ctl.on {
  background: var(--color-primary);
  color: var(--color-primary-contrast);
}
.ctl kbd {
  font: inherit;
}

.head h1 {
  margin-bottom: var(--space-1);
}
.progress {
  font-size: 0.85rem;
}
.section-title {
  font-size: 1.05rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-muted);
  margin: var(--space-5) 0 var(--space-3);
}

.coach {
  background: var(--color-ink-50);
  border: 1px solid var(--color-ink-200);
  border-radius: var(--radius-md);
  padding: var(--space-3) var(--space-4);
  margin: var(--space-4) 0;
  font-size: 0.88rem;
}
.coach ul {
  margin: var(--space-2) 0;
  padding-left: 1.1rem;
}
.coach .btn {
  padding: 0.3rem 0.8rem;
  font-size: 0.8rem;
}

/* Everything in the reading column shares one measure, so the title,
   section heading and navigation line up with the prose instead of
   hanging off its left edge. Declared in rem rather than em so the chrome
   does not reflow when the reader changes the body size. */
.measure,
.section-nav,
.finish-card,
.licence {
  max-width: 38rem;
  margin-left: auto;
  margin-right: auto;
}

.body {
  /* ~62 characters per line at the default size. Bringhurst puts the
     comfortable range at 45-75; L2 reading sits at the lower end because
     the eye returns to the start of the line more often. */
  max-width: 34em;
  line-height: 1.75;
  margin: 0 auto;
  /* Lets the browser hyphenate French prose at this narrow measure rather
     than leaving ragged gaps. */
  hyphens: auto;
}

.section-nav {
  display: flex;
  justify-content: space-between;
  gap: var(--space-2);
  margin-top: var(--space-6);
}

.finish-card {
  margin-top: var(--space-6);
  text-align: center;
}
.finish-card .btn {
  margin-top: var(--space-2);
}

.licence {
  margin-top: var(--space-6);
  font-size: 0.72rem;
  text-align: center;
}
</style>
