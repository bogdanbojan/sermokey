<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, ApiError } from '../api/client'
import { navigate } from '../router/router'
import { useLanguage, t } from '../i18n/language'

const props = defineProps<{ tense: string }>()
const { lang } = useLanguage()

const tenseLabel = ref('')
const loading = ref(true)
const error = ref('')

const exerciseType = ref<'mixed' | 'mcq' | 'text'>('mixed')
const count = ref(10)

async function load() {
  loading.value = true
  error.value = ''
  if (props.tense === 'mixed') {
    tenseLabel.value = t.value.home.mixedTitle
    loading.value = false
    return
  }
  try {
    const theory = await api.getTheory(lang.value, props.tense)
    tenseLabel.value = theory.nameNative
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : t.value.quizSetup.error
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch([() => props.tense, lang], load)

function start() {
  navigate({ name: 'quiz', tense: props.tense, type: exerciseType.value, count: count.value })
}
</script>

<template>
  <section>
    <button class="btn back" @click="navigate({ name: 'home' })">{{ t.quizSetup.back }}</button>

    <p v-if="loading" class="muted">{{ t.quizSetup.loading }}</p>
    <p v-else-if="error" class="error-text">{{ error }}</p>

    <template v-else>
      <h1>{{ t.quizSetup.practiceTitlePrefix }} {{ tenseLabel }}</h1>

      <div class="card setup-card">
        <div class="field">
          <h3>{{ t.quizSetup.exerciseType }}</h3>
          <label class="radio">
            <input type="radio" value="mixed" v-model="exerciseType" />
            {{ t.quizSetup.mixed }}
          </label>
          <label class="radio">
            <input type="radio" value="mcq" v-model="exerciseType" />
            {{ t.quizSetup.mcqOnly }}
          </label>
          <label class="radio">
            <input type="radio" value="text" v-model="exerciseType" />
            {{ t.quizSetup.textOnly }}
          </label>
        </div>

        <div class="field">
          <h3>{{ t.quizSetup.questionCount }}</h3>
          <div class="count-options">
            <button
              v-for="n in [5, 10, 15, 20]"
              :key="n"
              class="btn"
              :class="{ 'btn-primary': count === n }"
              @click="count = n"
            >
              {{ n }}
            </button>
          </div>
        </div>

        <button class="btn btn-primary btn-block start-btn" @click="start">{{ t.quizSetup.start }}</button>
      </div>
    </template>
  </section>
</template>

<style scoped>
.back {
  margin-bottom: var(--space-4);
}
.setup-card {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}
.field h3 {
  margin-bottom: var(--space-2);
}
.radio {
  display: block;
  margin-bottom: var(--space-2);
  cursor: pointer;
}
.radio input {
  margin-right: var(--space-2);
}
.count-options {
  display: flex;
  gap: var(--space-2);
}
.count-options .btn {
  flex: 1;
}
.start-btn {
  margin-top: var(--space-2);
}
.error-text {
  color: var(--color-danger);
}
</style>
