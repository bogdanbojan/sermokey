<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { api, ApiError } from '../api/client'
import type { ReadingEntry, ReadingSection } from '../types/api'
import { navigate } from '../router/router'
import { useLanguage, t, format } from '../i18n/language'
import { useWordStatus, type WordStatus } from '../reading/status'
import { useSavedWords, type WordContext } from '../reading/saved'
import { readingPrefs, savePosition, useLookupCounts } from '../reading/prefs'
import { read as readStored, write as writeStored } from '../storage'
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
import ReaderDisplayMenu from '../components/ReaderDisplayMenu.vue'

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

// Arrowing quickly through sections starts several fetches. Whichever
// resolved last used to win, so section 2's body could render while the
// URL said section 5 — and the position saved afterwards was read from the
// *current* props, recording a section the learner was not looking at.
let loadToken = 0

async function load() {
  const token = ++loadToken
  // Captured before the await, so a late response cannot attribute itself
  // to whatever the props happen to say by the time it lands.
  const requested = { lang: lang.value, textId: props.textId, section: props.section }

  phase.value = 'loading'
  error.value = ''
  closePeek()
  closePinned()
  try {
    const fetched = await api.getTextSection(requested.lang, requested.textId, requested.section)
    if (token !== loadToken) return
    section.value = fetched
    phase.value = 'ready'
    savePosition(requested.lang, requested.textId, {
      section: requested.section,
      paragraph: 0,
    })
    window.scrollTo({ top: 0 })
  } catch (e) {
    if (token !== loadToken) return
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

/**
 * Keyboard activation of a focused token.
 *
 * Bound on the article rather than per token so the handler count does not
 * scale with the length of the chapter.
 */
function onTokenKeydown(e: KeyboardEvent) {
  if (e.key !== 'Enter' && e.key !== ' ') return
  const el = e.target instanceof Element ? e.target : null
  if (!el?.classList.contains('tok')) return
  const at = locate(el)
  if (!at) return
  const entry = entryAt(at.token, at.paragraph)
  if (!entry) return
  e.preventDefault()
  pin(entry, at.token, at.paragraph)
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

// `selectionchange` fires on every character of a drag, and the handler
// re-tokenises the paragraph and forces a layout flush. Coalescing to one
// run per frame keeps a drag across a long paragraph smooth.
let selectionFrame = 0
function onSelectionChange() {
  if (selectionFrame) return
  selectionFrame = window.requestAnimationFrame(() => {
    selectionFrame = 0
    measureSelection()
  })
}

function measureSelection() {
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
    window.getSelection()?.removeAllRanges()
    selection.value = null
  } catch (e) {
    // The cleanup above used to sit *outside* the catch, so a network
    // failure wiped the selection and the toolbar with it and showed
    // nothing — leaving the learner unable to tell a missing dictionary
    // entry from a dead backend.
    lookupError.value = e instanceof ApiError ? e.message : t.value.reading.error
  }
}

const lookupError = ref('')

// --- Navigation ----------------------------------------------------------
const hasPrev = computed(() => props.section > 0)
const hasNext = computed(() => !!section.value && props.section < section.value.sectionCount - 1)
const isLastSection = computed(() => !!section.value && !hasNext.value)

function goSection(delta: number) {
  navigate({ name: 'read', textId: props.textId, section: props.section + delta })
}

function goToSection(index: number) {
  if (index === props.section) return
  navigate({ name: 'read', textId: props.textId, section: index })
}

const savedFromText = computed(() => saved.value.fromText(props.textId).length)

// --- Keyboard ------------------------------------------------------------
// The flashcards view already establishes single-key shortcuts, so the
// reader matches. It also makes the feature operable without a pointer,
// rather than a mouse-only toy.
function onKeydown(e: KeyboardEvent) {
  if (e.metaKey || e.ctrlKey || e.altKey) return

  // Escape is handled before the typing guard. The pinned panel's sense
  // selector is made of radio inputs, so focusing one used to make Escape
  // unreachable — a dialog you could not dismiss from the keyboard.
  if (e.key === 'Escape') {
    closePinned()
    closePeek()
    selection.value = null
    lookupError.value = ''
    return
  }

  const el = e.target as HTMLElement | null
  if (el && (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA' || el.isContentEditable)) return

  // Shift+Arrow is how a keyboard user extends a native selection, which
  // is the whole basis of phrase capture. Swallowing it for page
  // navigation removed the one keyboard route into the feature.
  if (e.shiftKey) return

  switch (e.key) {
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
  if (selectionFrame) window.cancelAnimationFrame(selectionFrame)
})

// --- Display controls ----------------------------------------------------
const bodyStyle = computed(() => ({
  fontSize: `${readingPrefs.value.fontSize}rem`,
  // Leading was hardcoded at 1.75. After font size it is the single most
  // effective accommodation for dyslexic and low-vision readers, so it is
  // a preference rather than a constant.
  lineHeight: String(readingPrefs.value.lineHeight),
  fontFamily: readingPrefs.value.serif
    ? 'Georgia, "Iowan Old Style", "Times New Roman", serif'
    : 'var(--font-sans)',
}))

/**
 * Whether the device can actually hover.
 *
 * The coachmark's first instruction is "hover a word", which no touch user
 * can follow; it was shown unconditionally.
 */
const canHover = typeof window !== 'undefined' && window.matchMedia('(hover: hover)').matches

// --- First-run coachmark -------------------------------------------------
// Hover and drag-select are invisible affordances; a reader who does not
// discover them sees a plain wall of French.
const COACH_KEY = 'sermokey:reading-coach-seen'
const showCoach = ref(readStored(COACH_KEY, (u) => (u === true ? true : null), false) !== true)
function dismissCoach() {
  showCoach.value = false
  writeStored(COACH_KEY, true)
}
</script>

<template>
  <section class="reader">
    <div class="toolbar-row">
      <button class="btn back" @click="navigate({ name: 'home' })">{{ t.reading.back }}</button>

      <ReaderDisplayMenu v-if="phase === 'ready'" />
    </div>

    <p v-if="phase === 'loading'" class="muted" role="status">{{ t.reading.loading }}</p>
    <div v-else-if="phase === 'error'" class="error-block" role="alert">
      <p class="error-text">{{ error }}</p>
      <button class="btn" @click="load()">{{ t.common.retry }}</button>
    </div>

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
          <!-- "Hover a word" is not an instruction a touch user can
               follow; it was shown to every phone on first run. -->
          <li v-if="canHover">{{ t.reading.coachHover }}</li>
          <li>{{ t.reading.coachClick }}</li>
          <li>{{ t.reading.coachSelect }}</li>
        </ul>
        <button class="btn" @click="dismissCoach">{{ t.reading.coachDismiss }}</button>
      </div>

      <!-- The three underline styles meant nothing without a key. The
           strings for this have existed, translated, since the feature
           shipped; nothing rendered them. -->
      <div v-if="readingPrefs.highlighting" class="legend measure">
        <span class="legend-label">{{ t.reading.statusLegendTitle }}</span>
        <span class="legend-item"><span class="tok-sample unknown" aria-hidden="true" />{{ t.reading.statusUnknown }}</span>
        <span class="legend-item"><span class="tok-sample learning" aria-hidden="true" />{{ t.reading.statusLearning }}</span>
        <span class="legend-item"><span class="tok-sample known" aria-hidden="true" />{{ t.reading.statusKnown }}</span>
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
        @keydown="onTokenKeydown"
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

        <!-- Prev/next alone meant 46 clicks and 46 round trips to reach
             the end of an imported novel. -->
        <label v-if="section.sectionCount > 2" class="section-jump">
          <span class="visually-hidden">{{ t.reading.sectionJump }}</span>
          <select
            :value="section.index"
            @change="goToSection(Number(($event.target as HTMLSelectElement).value))"
          >
            <option v-for="n in section.sectionCount" :key="n - 1" :value="n - 1">
              {{ format(t.reading.sectionProgress, { index: n, total: section.sectionCount }) }}
            </option>
          </select>
        </label>

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

    <p v-if="lookupError" class="lookup-error" role="alert">
      {{ lookupError }}
      <button class="link" @click="lookupError = ''">{{ t.common.dismiss }}</button>
    </p>

    <!-- Keyed on the position, not on the surface. "est" at token 7 of
         paragraph 3 and "est" at token 7 of paragraph 9 produced the same
         key *and* the same surface, so the panel was neither remounted nor
         reset by its own surface watcher — leaving a stale sense index
         that made the Save button a silent no-op. -->
    <WordPanel
      v-if="pinned"
      :key="`${pinned.paragraph}:${pinned.token}:${pinned.entry.wordId}`"
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

.legend {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
  font-size: 0.78rem;
  color: var(--color-text-muted);
}
.legend-label {
  font-weight: 600;
}
.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}
.tok-sample {
  display: inline-block;
  width: 1.4rem;
  height: 0.85rem;
  border-bottom: 2px solid transparent;
}
.tok-sample.unknown {
  border-bottom-color: var(--color-accent-400);
  background: var(--color-accent-50);
}
.tok-sample.learning {
  border-bottom-style: dotted;
  border-bottom-color: var(--color-ink-400);
}
.tok-sample.known {
  border-bottom-color: transparent;
  background: var(--color-bg);
}

.section-jump select {
  font: inherit;
  font-size: 0.85rem;
  padding: 0.4rem 0.5rem;
  min-height: 2.5rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-surface);
  color: var(--color-text);
}

.error-block {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: var(--space-3);
}
.lookup-error {
  position: fixed;
  left: 50%;
  bottom: var(--space-5);
  transform: translateX(-50%);
  z-index: 66;
  background: var(--color-danger);
  color: #fff;
  border-radius: var(--radius-sm);
  padding: var(--space-2) var(--space-3);
  font-size: 0.85rem;
  margin: 0;
}
.lookup-error .link {
  background: none;
  border: none;
  font: inherit;
  font-weight: 600;
  color: inherit;
  text-decoration: underline;
  cursor: pointer;
  margin-left: var(--space-2);
}

.body {
  /* ~62 characters per line at the default size. Bringhurst puts the
     comfortable range at 45-75; L2 reading sits at the lower end because
     the eye returns to the start of the line more often. */
  max-width: 34em;
  /* line-height comes from readingPrefs via :style, so it is adjustable. */
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
