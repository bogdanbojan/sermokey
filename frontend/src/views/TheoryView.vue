<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, ApiError } from '../api/client'
import type { Theory } from '../types/api'
import { navigate } from '../router/router'
import { useLanguage, t } from '../i18n/language'

const props = defineProps<{ tense: string }>()
const { lang } = useLanguage()

const theory = ref<Theory | null>(null)
const loading = ref(true)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  theory.value = null
  try {
    theory.value = await api.getTheory(lang.value, props.tense)
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : t.value.theory.error
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch([() => props.tense, lang], load)
</script>

<template>
  <section>
    <button class="btn back" @click="navigate({ name: 'home' })">{{ t.theory.back }}</button>

    <p v-if="loading" class="muted">{{ t.theory.loading }}</p>
    <p v-else-if="error" class="error-text">{{ error }}</p>

    <template v-else-if="theory">
      <h1>{{ theory.nameNative }} <span class="muted name-en">({{ theory.nameEn }})</span></h1>
      <p>{{ theory.usage }}</p>

      <div class="card section-card">
        <h3>{{ t.theory.formation }}</h3>
        <p>{{ theory.formation }}</p>
      </div>

      <div v-for="group in theory.groups" :key="group.groupName + group.exampleVerb" class="card section-card">
        <h3>{{ group.groupName }}</h3>
        <p class="muted">{{ group.explanation }}</p>
        <table>
          <thead>
            <tr>
              <th v-for="p in t.theory.personLabels" :key="p">{{ p }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="group.endings.some((e) => e)">
              <td v-for="(e, i) in group.endings" :key="'e' + i">{{ e || '—' }}</td>
            </tr>
            <tr>
              <td v-for="(f, i) in group.exampleForms" :key="'f' + i"><strong>{{ f || '—' }}</strong></td>
            </tr>
          </tbody>
        </table>
        <p class="muted example-label">{{ t.theory.example }} : {{ group.exampleVerb }}</p>
      </div>

      <div v-if="theory.irregularNotes" class="card section-card notes">
        <h3>{{ t.theory.toRemember }}</h3>
        <p>{{ theory.irregularNotes }}</p>
      </div>

      <button class="btn btn-primary btn-block practice-btn" @click="navigate({ name: 'quiz-setup', tense: theory.id })">
        {{ t.theory.practiceThisTense }}
      </button>
    </template>
  </section>
</template>

<style scoped>
.back {
  margin-bottom: var(--space-4);
}
.name-en {
  font-weight: 400;
  font-size: 0.65em;
}
.section-card {
  margin-bottom: var(--space-4);
}
.example-label {
  margin: var(--space-2) 0 0;
}
.practice-btn {
  margin-top: var(--space-5);
}
.error-text {
  color: var(--color-danger);
}
</style>
