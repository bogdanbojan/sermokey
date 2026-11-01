<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { api, ApiError } from '../api/client'
import { navigate, type DeckRef } from '../router/router'
import { useLanguage, t, format } from '../i18n/language'
import {
  useSrsDeck,
  cardID,
  setVocabDirection,
  vocabDirection,
  type DeckStats,
  type Direction,
} from '../srs/store'
import { isLeech, type CardState, type Grade } from '../srs/sm2'
import { useSavedWords, type SavedWord } from '../reading/saved'
import { useDecks } from '../reading/decks'

const props = defineProps<{ deck: DeckRef }>()
const { lang } = useLanguage()
const deck = computed(() => useSrsDeck(lang.value))
const savedWords = computed(() => useSavedWords(lang.value))
const decks = computed(() => useDecks(lang.value))

// Words saved while reading are stored locally and carry the sentence they
// were met in; curated words come from the API and do not. Both reduce to
// the same shape here so the session logic below does not care where a card
// came from — which is what let reading reuse the flashcard screen rather
// than grow a parallel one.
interface StudyWord {
  id: string
  lemma: string
  pos: string
  gender?: string
  translationEn: string
  context?: SavedWord['context']
}

function fromSaved(w: SavedWord): StudyWord {
  return { id: w.id, lemma: w.lemma, pos: w.pos, gender: w.gender, translationEn: w.gloss, context: w.context }
}

/** Resolves a DeckRef to its words and its display label. */
async function resolveDeck(ref: DeckRef): Promise<{ words: StudyWord[]; label: string }> {
  switch (ref.kind) {
    case 'saved':
      return { words: savedWords.value.all().map(fromSaved), label: t.value.reading.savedDeckTitle }
    case 'text': {
      const words = savedWords.value.fromText(ref.id).map(fromSaved)
      const title = savedWords.value.fromText(ref.id)[0]?.context?.textTitle ?? ref.id
      return { words, label: format(t.value.reading.textDeckTitle, { title }) }
    }
    case 'deck': {
      const words = savedWords.value.inDeck(ref.id).map(fromSaved)
      // A deck whose entry has been forgotten still lists its words, so a
      // stale link is an empty-ish deck rather than an error.
      return { words, label: decks.value.get(ref.id)?.name ?? ref.id }
    }
    case 'category': {
      const vocab = await api.getVocabulary(lang.value, ref.id || undefined)
      const words: StudyWord[] = vocab.map((w) => ({
        id: w.id,
        lemma: w.lemma,
        pos: w.pos,
        gender: w.gender,
        translationEn: w.translationEn,
      }))
      if (!ref.id) return { words, label: t.value.vocab.mixedTitle }
      const cats = await api.getVocabCategories(lang.value)
      return { words, label: cats.find((c) => c.id === ref.id)?.nameNative ?? ref.id }
    }
  }
}

// One flashcard = one word reviewed in one direction. Recognition ("see
// the word, recall its meaning") and production ("see the meaning, recall
// the word") get independent SRS ids/scheduling — see srs/store.ts's doc
// comment for why that split matters.
interface FlashCard {
  id: string
  word: StudyWord
  direction: 'recognition' | 'production'
}

const direction = vocabDirection

function buildCards(words: StudyWord[], dir: Direction | 'mixed'): FlashCard[] {
  if (dir === 'mixed') {
    return words.flatMap((w) => [
      { id: cardID(w.id, 'recognition'), word: w, direction: 'recognition' as const },
      { id: cardID(w.id, 'production'), word: w, direction: 'production' as const },
    ])
  }
  return words.map((w) => ({ id: cardID(w.id, dir), word: w, direction: dir }))
}

type Phase = 'loading' | 'error' | 'empty' | 'studying' | 'done'

const phase = ref<Phase>('loading')
const error = ref('')
const allWords = ref<StudyWord[]>([])
const queue = ref<FlashCard[]>([])
const index = ref(0)
const revealed = ref(false)
const reviewedCount = ref(0)
const categoryLabel = ref('')

const current = computed<FlashCard | undefined>(() => queue.value[index.value])

// Switching deck or language mid-flight must not let the older response
// land on top of the newer one. Every load takes a ticket and drops its
// result if a newer load has started since.
let loadToken = 0

async function load(includeNotDue = false) {
  const token = ++loadToken
  phase.value = 'loading'
  error.value = ''
  saveWarning.value = ''
  lastGrade.value = null
  revealed.value = false
  reviewedCount.value = 0
  index.value = 0

  try {
    const resolved = await resolveDeck(props.deck)
    if (token !== loadToken) return
    allWords.value = resolved.words
    categoryLabel.value = resolved.label

    // An empty deck is not a failure. Reporting "could not load the
    // vocabulary" is what a brand-new learner saw on their very first
    // visit to the saved-words deck, which is the one deck that is
    // *supposed* to start empty.
    if (allWords.value.length === 0) {
      phase.value = 'empty'
      return
    }

    const allCards = buildCards(allWords.value, direction.value)
    queue.value = deck.value.sessionQueue(allCards, {
      newCardLimit: includeNotDue ? allCards.length : undefined,
      includeNotDue,
    })

    // Guard on the queue itself rather than on the flag. "Study anyway"
    // used to force the studying phase even when the queue came back
    // empty, and the template has no branch for a studying phase with no
    // current card — so the screen simply went blank.
    phase.value = queue.value.length === 0 ? 'empty' : 'studying'
  } catch (e) {
    if (token !== loadToken) return
    error.value = e instanceof ApiError ? e.message : t.value.vocab.error
    phase.value = 'error'
  }
}

onMounted(() => {
  load()
  // Bound to the window, not to the section. The handler used to sit on a
  // `tabindex="0"` container that nothing ever focused and whose focus ring
  // was removed in CSS, so the advertised 1-4 shortcuts did nothing until
  // the learner happened to click the background — and stopped working
  // again after the first grade, because grading unmounts the focused
  // button.
  window.addEventListener('keydown', onKeydown)
})
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))

watch([() => props.deck, lang, direction], () => load())

function reveal() {
  revealed.value = true
}

/** The state a card was in before the last grade, so it can be put back. */
const lastGrade = ref<{ cardId: string; previous?: CardState; index: number } | null>(null)
const saveWarning = ref('')

function grade(g: Grade) {
  if (!current.value) return
  const cardId = current.value.id
  const { previous, saved } = deck.value.recordReview(cardId, g)

  // A review that did not persist has not happened. Advancing anyway is
  // how an hour of study disappears on the next reload with no warning.
  if (saved !== 'ok') {
    saveWarning.value = t.value.vocab.saveFailed
    return
  }

  saveWarning.value = ''
  lastGrade.value = { cardId, previous, index: index.value }
  reviewedCount.value++

  // A card that was just failed is re-queued for this session. SM-2 as
  // published repeats an item until it is recalled; deferring it to
  // tomorrow removes the most effective part of the loop.
  if (g === 'again') queue.value = [...queue.value, current.value]

  if (index.value + 1 >= queue.value.length) {
    phase.value = 'done'
  } else {
    index.value++
    revealed.value = false
  }
}

function undoGrade() {
  const last = lastGrade.value
  if (!last) return
  deck.value.restore(last.cardId, last.previous)
  // Drop the re-queued copy an "again" appended, so undo really is a
  // return to the previous state rather than a partial one.
  if (queue.value.length > 0 && queue.value[queue.value.length - 1]?.id === last.cardId) {
    queue.value = queue.value.slice(0, -1)
  }
  index.value = last.index
  revealed.value = true
  reviewedCount.value = Math.max(0, reviewedCount.value - 1)
  phase.value = 'studying'
  lastGrade.value = null
}

function onKeydown(e: KeyboardEvent) {
  if (e.metaKey || e.ctrlKey || e.altKey) return
  const el = e.target as HTMLElement | null
  // Never steal a key from a field the learner is typing into.
  if (el && (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA' || el.isContentEditable)) return

  if (e.key === 'z' && lastGrade.value) {
    e.preventDefault()
    undoGrade()
    return
  }
  if (phase.value !== 'studying') return
  if (!revealed.value) {
    if (e.code === 'Space' || e.key === 'Enter') {
      e.preventDefault()
      reveal()
    }
    return
  }
  switch (e.key) {
    case '1':
      grade('again')
      break
    case '2':
      grade('hard')
      break
    case '3':
      grade('good')
      break
    case '4':
      grade('easy')
      break
  }
}

/**
 * What each button will do to the current card, as a short label.
 *
 * Grading blind is the main reason learners mis-calibrate "hard" against
 * "good"; Anki has shown the resulting interval on the buttons forever.
 */
const intervalLabels = computed<Record<Grade, string>>(() => {
  const empty = { again: '', hard: '', good: '', easy: '' }
  if (!current.value) return empty
  const preview = deck.value.preview(current.value.id)
  const label = (days: number) => {
    if (days <= 0) return t.value.vocab.intervalToday
    if (days < 30) return format(t.value.vocab.intervalDays, { count: days })
    return format(t.value.vocab.intervalMonths, { count: Math.round(days / 30) })
  }
  return {
    again: label(preview.again),
    hard: label(preview.hard),
    good: label(preview.good),
    easy: label(preview.easy),
  }
})

/** A card failed often enough that redrilling it is not working. */
const currentLeech = computed(() => {
  if (!current.value) return null
  const state = deck.value.stateFor(current.value.id)
  return state && isLeech(state) ? state : null
})

const stats = computed<DeckStats | null>(() => {
  if (!allWords.value.length) return null
  return deck.value.stats(buildCards(allWords.value, direction.value))
})

// Gender is shown alongside whichever side of the card currently displays
// the target-language word: immediately for recognition (it's part of the
// prompt), only after reveal for production (it would otherwise give the
// answer away). German adds a third, neuter gender ('n') alongside
// French/Spanish's masculine/feminine.
const genderBadge = computed(() => {
  if (!current.value || current.value.word.pos !== 'noun' || !current.value.word.gender) return ''
  if (current.value.direction === 'production' && !revealed.value) return ''
  if (current.value.word.gender === 'm') return '♂'
  if (current.value.word.gender === 'f') return '♀'
  return 'N'
})

const promptText = computed(() => {
  if (!current.value) return ''
  return current.value.direction === 'recognition' ? current.value.word.lemma : current.value.word.translationEn
})
const answerText = computed(() => {
  if (!current.value) return ''
  return current.value.direction === 'recognition' ? current.value.word.translationEn : current.value.word.lemma
})
</script>

<template>
  <section>
    <button class="btn back" @click="navigate({ name: 'home' })">{{ t.vocab.back }}</button>

    <div v-if="phase !== 'loading' && phase !== 'error'" class="direction-switch" role="group" :aria-label="t.vocab.directionLabel">
      <button
        type="button"
        class="direction-btn"
        :class="{ active: direction === 'recognition' }"
        :aria-pressed="direction === 'recognition'"
        :title="t.vocab.directionRecognitionHint"
        @click="setVocabDirection('recognition')"
      >
        {{ t.vocab.directionRecognition }}
      </button>
      <button
        type="button"
        class="direction-btn"
        :class="{ active: direction === 'production' }"
        :aria-pressed="direction === 'production'"
        :title="t.vocab.directionProductionHint"
        @click="setVocabDirection('production')"
      >
        {{ t.vocab.directionProduction }}
      </button>
      <button
        type="button"
        class="direction-btn"
        :class="{ active: direction === 'mixed' }"
        :aria-pressed="direction === 'mixed'"
        :title="t.vocab.directionMixedHint"
        @click="setVocabDirection('mixed')"
      >
        {{ t.vocab.directionMixed }}
      </button>
    </div>

    <p v-if="phase === 'loading'" class="muted" role="status">{{ t.vocab.loading }}</p>
    <div v-else-if="phase === 'error'" class="error-block" role="alert">
      <p class="error-text">{{ error }}</p>
      <button class="btn" @click="load()">{{ t.common.retry }}</button>
    </div>

    <template v-else-if="phase === 'empty'">
      <div class="card empty-card">
        <h1>{{ allWords.length === 0 ? t.reading.noneSaved : t.vocab.nothingDue }}</h1>
        <p class="muted">{{ allWords.length === 0 ? t.reading.sectionIntro : t.vocab.nothingDueDetail }}</p>
        <button v-if="allWords.length > 0" class="btn btn-primary" @click="load(true)">
          {{ t.vocab.studyAnyway }}
        </button>
        <button v-else class="btn btn-primary" @click="navigate({ name: 'home' })">
          {{ t.quiz.backHome }}
        </button>
      </div>
    </template>

    <template v-else-if="phase === 'studying' && current">
      <div class="progress-row">
        <span class="muted">{{ categoryLabel }}</span>
        <span class="muted">{{ format(t.vocab.cardProgress, { index: index + 1, total: queue.length }) }}</span>
      </div>
      <div class="progress-bar">
        <div class="progress-fill" :style="{ width: `${((index + 1) / queue.length) * 100}%` }" />
      </div>

      <div class="card flashcard">
        <div class="flashcard-front">
          <span v-if="genderBadge" class="gender-badge" :class="current.word.gender">{{ genderBadge }}</span>
          <span class="lemma">{{ promptText }}</span>
          <!-- On a production card the prompt is the English gloss, so the
               part-of-speech pill would be describing the wrong language. -->
          <span v-if="current.direction === 'recognition'" class="pos-pill">{{ current.word.pos }}</span>
        </div>

        <div v-if="!revealed" class="reveal-area">
          <button class="btn btn-primary btn-block" @click="reveal">
            {{ t.vocab.reveal }} <kbd class="reveal-key">space</kbd>
          </button>
        </div>

        <div v-else class="answer-area">
          <p class="translation" role="status">{{ answerText }}</p>

          <p v-if="currentLeech" class="leech-note">
            <strong>{{ t.vocab.leechLabel }}</strong>
            {{ format(t.vocab.leechHint, { count: currentLeech.lapses }) }}
          </p>

          <!-- The sentence the word was met in, for words captured while
               reading. A gloss recalled in context is retained better than
               a bare pair (Webb 2008), and it is the cheapest possible way
               to give the card a meaningful back side. -->
          <p v-if="current.word.context" class="context">
            <span>{{ current.word.context.sentence }}</span>
            <button
              class="context-link"
              @click="navigate({ name: 'read', textId: current.word.context.textId, section: current.word.context.section })"
            >
              {{ current.word.context.textTitle }}
            </button>
          </p>
          <div class="grade-buttons">
            <button class="btn grade-again" @click="grade('again')">
              <span class="grade-label"><kbd>1</kbd> {{ t.vocab.again }}</span>
              <span class="grade-interval">{{ intervalLabels.again }}</span>
            </button>
            <button class="btn grade-hard" @click="grade('hard')">
              <span class="grade-label"><kbd>2</kbd> {{ t.vocab.hard }}</span>
              <span class="grade-interval">{{ intervalLabels.hard }}</span>
            </button>
            <button class="btn grade-good" @click="grade('good')">
              <span class="grade-label"><kbd>3</kbd> {{ t.vocab.good }}</span>
              <span class="grade-interval">{{ intervalLabels.good }}</span>
            </button>
            <button class="btn grade-easy" @click="grade('easy')">
              <span class="grade-label"><kbd>4</kbd> {{ t.vocab.easy }}</span>
              <span class="grade-interval">{{ intervalLabels.easy }}</span>
            </button>
          </div>

          <p v-if="saveWarning" class="error-text" role="alert">{{ saveWarning }}</p>
        </div>
      </div>

      <!-- A mis-tap on a four-up grid of buttons permanently alters the
           card's schedule, so the action has to be reversible. -->
      <div v-if="lastGrade" class="undo-row">
        <button class="btn btn-small" @click="undoGrade">{{ t.vocab.undo }} <kbd>Z</kbd></button>
      </div>
    </template>

    <template v-else-if="phase === 'done'">
      <div class="card done-card">
        <h1>{{ t.vocab.sessionDone }}</h1>
        <p class="muted">{{ format(t.vocab.sessionDoneDetail, { count: reviewedCount }) }}</p>

        <div v-if="stats" class="stats-grid">
          <div class="stat"><strong>{{ stats.new }}</strong><span class="muted">{{ t.vocab.newLabel }}</span></div>
          <div class="stat"><strong>{{ stats.learning }}</strong><span class="muted">{{ t.vocab.learningLabel }}</span></div>
          <div class="stat"><strong>{{ stats.review }}</strong><span class="muted">{{ t.vocab.reviewLabel }}</span></div>
          <div class="stat"><strong>{{ stats.mastered }}</strong><span class="muted">{{ t.vocab.masteredLabel }}</span></div>
        </div>

        <p v-if="stats && stats.leeches > 0" class="leech-note">
          {{ format(t.vocab.leechHint, { count: stats.leeches }) }}
        </p>

        <div class="done-actions">
          <button class="btn" @click="navigate({ name: 'home' })">{{ t.quiz.backHome }}</button>
          <button class="btn btn-primary" @click="load()">{{ t.quiz.tryAgain }}</button>
        </div>
      </div>
    </template>

    <!-- Defence in depth: every phase above is guarded by a condition that
         could in principle all be false at once, and a blank screen with no
         explanation is the worst possible outcome of that. -->
    <div v-else class="card empty-card">
      <p class="muted">{{ t.vocab.nothingDue }}</p>
      <button class="btn btn-primary" @click="load(true)">{{ t.vocab.studyAnyway }}</button>
    </div>
  </section>
</template>

<style scoped>
.back {
  margin-bottom: var(--space-4);
}
.error-text {
  color: var(--color-danger);
}

.direction-switch {
  display: inline-flex;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: 2px;
  background: var(--color-bg);
  margin-bottom: var(--space-4);
}
.direction-btn {
  border: none;
  background: none;
  padding: 0.5rem 0.8rem;
  min-height: 2.5rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-text-muted);
  border-radius: calc(var(--radius-sm) - 2px);
  cursor: pointer;
}
.direction-btn:hover {
  color: var(--color-text);
}
.direction-btn.active {
  background: var(--color-primary);
  color: var(--color-primary-contrast);
}

.progress-row {
  display: flex;
  justify-content: space-between;
  font-size: 0.9rem;
  margin-bottom: var(--space-2);
}
.progress-bar {
  height: 6px;
  background: var(--color-border);
  border-radius: 999px;
  overflow: hidden;
  margin-bottom: var(--space-4);
}
.progress-fill {
  height: 100%;
  background: var(--color-accent);
  transition: width 0.2s ease;
}

.flashcard {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: var(--space-8) var(--space-5);
  min-height: 16rem;
  justify-content: center;
  gap: var(--space-5);
}
.flashcard-front {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
}
.lemma {
  font-size: 2.2rem;
  font-weight: 700;
}
.gender-badge {
  font-size: 1.2rem;
  font-weight: 700;
  border-radius: 999px;
  width: 2rem;
  height: 2rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.gender-badge.m {
  background: var(--color-bg);
  color: var(--color-primary);
  border: 1px solid var(--color-primary);
}
.gender-badge.f {
  background: var(--color-warning-bg);
  color: var(--color-warning-text);
  border: 1px solid var(--color-warning-text);
}
.gender-badge.n {
  background: var(--color-success-bg);
  color: var(--color-success);
  border: 1px solid var(--color-success);
}
.pos-pill {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
}
.reveal-area {
  width: 100%;
  max-width: 20rem;
}
.answer-area {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-4);
}
.translation {
  font-size: 1.4rem;
  color: var(--color-text-muted);
}
.context {
  font-size: 0.85rem;
  font-style: italic;
  color: var(--color-text-muted);
  max-width: 32em;
  border-left: 2px solid var(--color-border);
  padding-left: var(--space-3);
  text-align: left;
  margin: 0;
}
.context-link {
  display: block;
  border: none;
  background: none;
  padding: 0;
  margin-top: var(--space-1);
  font: inherit;
  font-style: normal;
  font-weight: 600;
  color: var(--color-primary);
  cursor: pointer;
}
.grade-buttons {
  display: grid;
  /* Four fixed columns left roughly 35px of label at 375px, which is not
     enough for "À revoir" or "Nochmal" in any of the three languages.
     auto-fit collapses to two rows on a phone instead of truncating. */
  grid-template-columns: repeat(auto-fit, minmax(7.5rem, 1fr));
  gap: var(--space-2);
  width: 100%;
}
.grade-buttons .btn {
  flex-direction: column;
  gap: 0.15rem;
  font-size: 0.85rem;
  /* WCAG 2.5.5: these are the most-pressed controls in the app. */
  min-height: 2.75rem;
}
.grade-label {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  white-space: nowrap;
}
.grade-buttons kbd {
  font-size: 0.7rem;
  /* 0.6 on an already-muted colour put this under the contrast floor. */
  opacity: 0.8;
}
.grade-interval {
  font-size: 0.7rem;
  font-weight: 500;
  opacity: 0.75;
  font-variant-numeric: tabular-nums;
}
.reveal-key {
  font-size: 0.7rem;
  opacity: 0.75;
  margin-left: 0.35rem;
}

.undo-row {
  display: flex;
  justify-content: center;
  margin-top: var(--space-3);
}
.btn-small {
  padding: 0.3rem 0.7rem;
  font-size: 0.8rem;
}

.leech-note {
  font-size: 0.85rem;
  color: var(--color-warning-text);
  background: var(--color-warning-bg);
  border-radius: var(--radius-sm);
  padding: var(--space-2) var(--space-3);
  margin: 0;
  max-width: 32em;
}

.error-block {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: var(--space-3);
}
.grade-again {
  border-color: var(--color-danger);
  color: var(--color-danger);
}
.grade-hard {
  border-color: var(--color-warning-text);
  color: var(--color-warning-text);
}
.grade-good {
  border-color: var(--color-primary);
  color: var(--color-primary);
}
.grade-easy {
  border-color: var(--color-success);
  color: var(--color-success);
}

.empty-card,
.done-card {
  text-align: center;
  padding: var(--space-8) var(--space-5);
}
.empty-card .btn,
.done-card h1 {
  margin-top: var(--space-4);
}
.stats-grid {
  display: grid;
  /* Same reason as .grade-buttons: "In Bearbeitung" does not fit a quarter
     of a 375px viewport. */
  grid-template-columns: repeat(auto-fit, minmax(5.5rem, 1fr));
  gap: var(--space-3);
  margin: var(--space-5) 0;
}
.stat {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}
.stat strong {
  font-size: 1.4rem;
}
.done-actions {
  display: flex;
  gap: var(--space-2);
  justify-content: center;
}
</style>
