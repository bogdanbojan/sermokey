<script setup lang="ts">
import { navigate } from '../router/router'
import { t, useLanguage } from '../i18n/language'
import { useRecents } from '../recent'

const props = defineProps<{
  id: string
  nameNative: string
  nameEn: string
  usage: string
  isMixed?: boolean
}>()

const { lang } = useLanguage()

function goTheory() {
  useRecents(lang.value).record('theory', props.id, props.nameNative)
  navigate({ name: 'theory', tense: props.id })
}
function goPractice() {
  useRecents(lang.value).record('quiz', props.id, props.nameNative)
  navigate({ name: 'quiz-setup', tense: props.id })
}
</script>

<template>
  <div class="card tense-card" :class="{ mixed: isMixed }">
    <p class="pill">{{ isMixed ? t.tenseCard.mixedPill : nameEn }}</p>
    <h3>{{ nameNative }}</h3>
    <p class="muted usage">{{ usage }}</p>
    <div class="actions">
      <button v-if="!isMixed" class="btn" @click="goTheory">
        {{ t.tenseCard.theory }}
      </button>
      <button class="btn btn-primary" @click="goPractice">
        {{ t.tenseCard.practice }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.tense-card {
  display: flex;
  flex-direction: column;
}
.tense-card.mixed {
  border-color: var(--color-accent);
  border-width: 2px;
}
.usage {
  flex: 1;
  font-size: 0.92rem;
}
.actions {
  display: flex;
  gap: var(--space-2);
  margin-top: var(--space-3);
}
.actions .btn {
  flex: 1;
}
</style>
