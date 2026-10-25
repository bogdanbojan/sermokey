<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { api, ApiError } from '../api/client'
import type { Exercise } from '../types/api'
import { gradeTextAnswer } from '../utils/text'
import { navigate } from '../router/router'
import { useLanguage, t, format } from '../i18n/language'

const props = defineProps<{ tense: string; type: string; count: number }>()
const { lang } = useLanguage()

type Phase = 'loading' | 'error' | 'question' | 'finished'
type Feedback = { correct: boolean; accentWarning: boolean }
type MissedItem = { exercise: Exercise; userAnswer: string }

const phase = ref<Phase>('loading')
const error = ref('')
const exercises = ref<Exercise[]>([])
const index = ref(0)
const score = ref(0)
const missed = ref<MissedItem[]>([])

const answered = ref(false)
const selectedChoice = ref<string | null>(null)
const textInput = ref('')
const feedback = ref<Feedback | null>(null)

const textInputEl = ref<HTMLInputElement | null>(null)

const current = computed<Exercise | undefined>(() => exercises.value[index.value])
const isLast = computed(() => index.value === exercises.value.length - 1)
const blankDisplay = computed(() => (answered.value && current.value ? current.value.answer : '______'))
const blankClass = computed(() => {
  if (!answered.value) return 'pending'
  return feedback.value?.correct ? 'correct' : 'incorrect'
})

async function load() {
  phase.value = 'loading'
  error.value = ''
  index.value = 0
  score.value = 0
  missed.value = []
  resetQuestionState()
  try {
    exercises.value = await api.getExercises(lang.value, props.tense, props.type, props.count)
    if (exercises.value.length === 0) {
      error.value = t.value.quiz.noExercises
      phase.value = 'error'
    } else {
      phase.value = 'question'
      focusInputIfNeeded()
    }
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : t.value.quiz.error
    phase.value = 'error'
  }
}

function resetQuestionState() {
  answered.value = false
  selectedChoice.value = null
  textInput.value = ''
  feedback.value = null
}

function focusInputIfNeeded() {
  nextTick(() => textInputEl.value?.focus())
}

function submitChoice(choice: string) {
  if (answered.value || !current.value) return
  selectedChoice.value = choice
  answered.value = true
  const correct = choice === current.value.answer
  feedback.value = { correct, accentWarning: false }
  if (correct) {
    score.value++
  } else {
    missed.value.push({ exercise: current.value, userAnswer: choice })
  }
}

function submitText() {
  if (answered.value || !current.value || !textInput.value.trim()) return
  const result = gradeTextAnswer(textInput.value, current.value.answer)
  answered.value = true
  if (result.correct) {
    score.value++
    feedback.value = { correct: true, accentWarning: result.accentWarning }
  } else {
    feedback.value = { correct: false, accentWarning: false }
    missed.value.push({ exercise: current.value, userAnswer: textInput.value })
  }
}

function next() {
  if (!answered.value) return
  if (isLast.value) {
    phase.value = 'finished'
    return
  }
  index.value++
  resetQuestionState()
  focusInputIfNeeded()
}

function onInputEnter() {
  if (!answered.value) {
    submitText()
  } else {
    next()
  }
}

onMounted(load)
watch([() => props.tense, () => props.type, () => props.count, lang], load)
</script>

<template>
  <section>
    <button class="btn back" @click="navigate({ name: 'home' })">{{ t.quiz.back }}</button>

    <p v-if="phase === 'loading'" class="muted">{{ t.quiz.loadingExercises }}</p>
    <p v-else-if="phase === 'error'" class="error-text">{{ error }}</p>

    <template v-else-if="phase === 'question' && current">
      <div class="progress-row">
        <span class="muted">{{ t.quiz.questionLabel }} {{ index + 1 }} / {{ exercises.length }}</span>
        <span class="muted">{{ t.quiz.scoreLabel }} {{ score }}</span>
      </div>
      <div class="progress-bar">
        <div class="progress-fill" :style="{ width: `${(index / exercises.length) * 100}%` }" />
      </div>

      <div class="card question-card">
        <div class="meta-row">
          <span class="pill">{{ current.tenseName }}</span>
          <span class="pill">{{ t.quiz.subjectLabel }} {{ current.personLabel }}</span>
        </div>

        <p class="infinitive-hint">
          <em>{{ current.infinitive }}</em> <span class="muted">— {{ current.translationEn }}</span>
        </p>

        <p class="sentence">
          <span>{{ current.sentencePrefix }}</span>
          <template v-if="current.type === 'text' && !answered">
            <input
              ref="textInputEl"
              v-model="textInput"
              class="blank-input"
              type="text"
              autocomplete="off"
              autocapitalize="off"
              spellcheck="false"
              @keyup.enter="onInputEnter"
            />
          </template>
          <span v-else class="blank" :class="blankClass">{{ blankDisplay }}</span>
          <span>{{ current.sentenceSuffix }}</span>
        </p>

        <div v-if="current.type === 'mcq' && !answered" class="choices">
          <button v-for="choice in current.choices" :key="choice" class="btn choice-btn" @click="submitChoice(choice)">
            {{ choice }}
          </button>
        </div>

        <button
          v-if="current.type === 'text' && !answered"
          class="btn btn-primary submit-btn"
          :disabled="!textInput.trim()"
          @click="submitText"
        >
          {{ t.quiz.check }}
        </button>

        <div v-if="answered" class="feedback-block">
          <p v-if="feedback?.correct && !feedback.accentWarning" class="feedback correct">{{ t.quiz.correct }}</p>
          <p v-else-if="feedback?.correct && feedback.accentWarning" class="feedback warning">
            {{ format(t.quiz.correctWithAccentWarning, { answer: current.answer }) }}
          </p>
          <p v-else class="feedback incorrect">
            {{ format(t.quiz.incorrect, { answer: current.answer }) }}
          </p>
          <p v-if="!feedback?.correct && current.type === 'text'" class="muted">
            {{ format(t.quiz.youWrote, { text: missed[missed.length - 1]?.userAnswer ?? '' }) }}
          </p>

          <button class="btn btn-primary next-btn" @click="next">
            {{ isLast ? t.quiz.seeResults : t.quiz.next }}
          </button>
        </div>
      </div>
    </template>

    <template v-else-if="phase === 'finished'">
      <div class="card results-card">
        <h1>{{ t.quiz.results }}</h1>
        <p class="score-line">
          <strong>{{ score }}</strong> / {{ exercises.length }}
          <span class="muted">({{ Math.round((score / exercises.length) * 100) }}%)</span>
        </p>

        <div v-if="missed.length > 0" class="missed-section">
          <h3>{{ t.quiz.toReview }}</h3>
          <ul class="missed-list">
            <li v-for="(m, i) in missed" :key="i">
              <strong>{{ m.exercise.infinitive }}</strong>
              <span class="muted"> ({{ m.exercise.tenseName }}, {{ m.exercise.personLabel }})</span>
              — {{ t.quiz.correctLabel }}: <strong>{{ m.exercise.answer }}</strong>
              <span class="muted"> · {{ t.quiz.wroteLabel }}: {{ m.userAnswer }}</span>
            </li>
          </ul>
        </div>
        <p v-else class="muted">{{ t.quiz.noMistakes }}</p>

        <div class="results-actions">
          <button class="btn" @click="navigate({ name: 'home' })">{{ t.quiz.backHome }}</button>
          <button class="btn btn-primary" @click="load">{{ t.quiz.tryAgain }}</button>
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

.meta-row {
  display: flex;
  gap: var(--space-2);
  margin-bottom: var(--space-4);
}

.infinitive-hint {
  margin-bottom: var(--space-2);
}

.sentence {
  font-size: 1.25rem;
  line-height: 1.6;
}

.blank-input {
  border: none;
  border-bottom: 2px solid var(--color-primary);
  background: transparent;
  font-size: inherit;
  font-family: inherit;
  padding: 0 var(--space-1);
  width: 10ch;
  text-align: center;
}
.blank-input:focus {
  outline: none;
  border-bottom-color: var(--color-accent);
}

.blank {
  display: inline-block;
  padding: 0 var(--space-2);
  border-radius: var(--radius-sm);
  font-weight: 700;
}
.blank.pending {
  color: var(--color-text-muted);
  letter-spacing: 0.1em;
}
.blank.correct {
  background: var(--color-success-bg);
  color: var(--color-success);
}
.blank.incorrect {
  background: var(--color-danger-bg);
  color: var(--color-danger);
}

.choices {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: var(--space-2);
  margin-top: var(--space-4);
}
.choice-btn {
  text-align: center;
}

.submit-btn {
  margin-top: var(--space-4);
}

.feedback-block {
  margin-top: var(--space-4);
  padding-top: var(--space-4);
  border-top: 1px solid var(--color-border);
}
.feedback {
  font-weight: 600;
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  margin-bottom: var(--space-2);
}
.feedback.correct {
  background: var(--color-success-bg);
  color: var(--color-success);
}
.feedback.warning {
  background: var(--color-warning-bg);
  color: var(--color-warning-text);
}
.feedback.incorrect {
  background: var(--color-danger-bg);
  color: var(--color-danger);
}
.next-btn {
  margin-top: var(--space-3);
}

.results-card h1 {
  margin-bottom: var(--space-2);
}
.score-line {
  font-size: 1.4rem;
  margin-bottom: var(--space-4);
}
.missed-section {
  margin-bottom: var(--space-4);
}
.missed-list {
  padding-left: 1.2rem;
}
.missed-list li {
  margin-bottom: var(--space-2);
}
.results-actions {
  display: flex;
  gap: var(--space-2);
}
.results-actions .btn {
  flex: 1;
}
</style>
