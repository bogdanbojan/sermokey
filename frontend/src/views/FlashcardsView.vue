<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { api, ApiError } from '../api/client'
import type { VocabWord } from '../types/api'
import { navigate } from '../router/router'
import { useLanguage, t, format } from '../i18n/language'
import { useSrsDeck, cardID, vocabDirection, type DeckStats, type Direction } from '../srs/store'
import type { Grade } from '../srs/sm2'

const props = defineProps<{ category: string }>()
const { lang } = useLanguage()
const deck = computed(() => useSrsDeck(lang.value))

// One flashcard = one word reviewed in one direction. Recognition ("see
// the word, recall its meaning") and production ("see the meaning, recall
// the word") get independent SRS ids/scheduling — see srs/store.ts's doc
// comment for why that split matters.
interface FlashCard {
  id: string
  word: VocabWord
  direction: 'recognition' | 'production'
}

const direction = vocabDirection

function buildCards(words: VocabWord[], dir: Direction | 'mixed'): FlashCard[] {
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
const allWords = ref<VocabWord[]>([])
const queue = ref<FlashCard[]>([])
const index = ref(0)
const revealed = ref(false)
const reviewedCount = ref(0)
const categoryLabel = ref('')

const current = computed<FlashCard | undefined>(() => queue.value[index.value])

async function load(includeNotDue = false) {
  phase.value = 'loading'
  error.value = ''
  revealed.value = false
  reviewedCount.value = 0
  index.value = 0

  try {
    allWords.value = await api.getVocabulary(lang.value, props.category || undefined)
    if (props.category) {
      const cats = await api.getVocabCategories(lang.value)
      categoryLabel.value = cats.find((c) => c.id === props.category)?.nameNative ?? props.category
    } else {
      categoryLabel.value = t.value.vocab.mixedTitle
    }

    if (allWords.value.length === 0) {
      error.value = t.value.vocab.error
      phase.value = 'error'
      return
    }

    const allCards = buildCards(allWords.value, direction.value)
    const newCardLimit = includeNotDue ? allCards.length : 15
    queue.value = deck.value.sessionQueue(allCards, newCardLimit)

    if (queue.value.length === 0 && !includeNotDue) {
      phase.value = 'empty'
    } else {
      phase.value = 'studying'
    }
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : t.value.vocab.error
    phase.value = 'error'
  }
}

onMounted(() => load())
watch([() => props.category, lang, direction], () => load())

function reveal() {
  revealed.value = true
}

function grade(g: Grade) {
  if (!current.value) return
  deck.value.recordReview(current.value.id, g)
  reviewedCount.value++
  if (index.value + 1 >= queue.value.length) {
    phase.value = 'done'
  } else {
    index.value++
    revealed.value = false
  }
}

function onKeydown(e: KeyboardEvent) {
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

const stats = computed<DeckStats | null>(() => {
  if (!allWords.value.length) return null
  return deck.value.stats(buildCards(allWords.value, direction.value))
})

// Gender is shown alongside whichever side of the card currently displays
// the target-language word: immediately for recognition (it's part of the
// prompt), only after reveal for production (it would otherwise give the
// answer away).
const genderBadge = computed(() => {
  if (!current.value || current.value.word.pos !== 'noun' || !current.value.word.gender) return ''
  if (current.value.direction === 'production' && !revealed.value) return ''
  return current.value.word.gender === 'm' ? '♂' : '♀'
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
  <section tabindex="0" @keydown="onKeydown">
    <button class="btn back" @click="navigate({ name: 'home' })">{{ t.vocab.back }}</button>

    <div v-if="phase !== 'loading' && phase !== 'error'" class="direction-switch" role="group" :aria-label="t.vocab.directionLabel">
      <button
        type="button"
        class="direction-btn"
        :class="{ active: direction === 'recognition' }"
        :title="t.vocab.directionRecognitionHint"
        @click="direction = 'recognition'"
      >
        {{ t.vocab.directionRecognition }}
      </button>
      <button
        type="button"
        class="direction-btn"
        :class="{ active: direction === 'production' }"
        :title="t.vocab.directionProductionHint"
        @click="direction = 'production'"
      >
        {{ t.vocab.directionProduction }}
      </button>
      <button
        type="button"
        class="direction-btn"
        :class="{ active: direction === 'mixed' }"
        :title="t.vocab.directionMixedHint"
        @click="direction = 'mixed'"
      >
        {{ t.vocab.directionMixed }}
      </button>
    </div>

    <p v-if="phase === 'loading'" class="muted">{{ t.vocab.loading }}</p>
    <p v-else-if="phase === 'error'" class="error-text">{{ error }}</p>

    <template v-else-if="phase === 'empty'">
      <div class="card empty-card">
        <h1>{{ t.vocab.nothingDue }}</h1>
        <p class="muted">{{ t.vocab.nothingDueDetail }}</p>
        <button class="btn btn-primary" @click="load(true)">{{ t.vocab.studyAnyway }}</button>
      </div>
    </template>

    <template v-else-if="phase === 'studying' && current">
      <div class="progress-row">
        <span class="muted">{{ categoryLabel }}</span>
        <span class="muted">{{ format(t.vocab.cardProgress, { index: index + 1, total: queue.length }) }}</span>
      </div>
      <div class="progress-bar">
        <div class="progress-fill" :style="{ width: `${(index / queue.length) * 100}%` }" />
      </div>

      <div class="card flashcard">
        <div class="flashcard-front">
          <span v-if="genderBadge" class="gender-badge" :class="current.word.gender">{{ genderBadge }}</span>
          <span class="lemma">{{ promptText }}</span>
          <span class="pos-pill">{{ current.direction === 'recognition' ? current.word.pos : 'en' }}</span>
        </div>

        <div v-if="!revealed" class="reveal-area">
          <button class="btn btn-primary btn-block" @click="reveal">{{ t.vocab.reveal }}</button>
        </div>

        <div v-else class="answer-area">
          <p class="translation">{{ answerText }}</p>
          <div class="grade-buttons">
            <button class="btn grade-again" @click="grade('again')"><kbd>1</kbd> {{ t.vocab.again }}</button>
            <button class="btn grade-hard" @click="grade('hard')"><kbd>2</kbd> {{ t.vocab.hard }}</button>
            <button class="btn grade-good" @click="grade('good')"><kbd>3</kbd> {{ t.vocab.good }}</button>
            <button class="btn grade-easy" @click="grade('easy')"><kbd>4</kbd> {{ t.vocab.easy }}</button>
          </div>
        </div>
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

        <div class="done-actions">
          <button class="btn" @click="navigate({ name: 'home' })">{{ t.quiz.backHome }}</button>
          <button class="btn btn-primary" @click="load()">{{ t.quiz.tryAgain }}</button>
        </div>
      </div>
    </template>
  </section>
</template>

<style scoped>
.back {
  margin-bottom: var(--space-4);
}
.error-text {
  color: var(--color-danger);
}
section {
  outline: none;
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
  padding: 0.4rem 0.8rem;
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
.grade-buttons {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-2);
  width: 100%;
}
.grade-buttons .btn {
  flex-direction: column;
  gap: 0.2rem;
  font-size: 0.85rem;
}
.grade-buttons kbd {
  font-size: 0.7rem;
  opacity: 0.6;
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
  grid-template-columns: repeat(4, 1fr);
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
