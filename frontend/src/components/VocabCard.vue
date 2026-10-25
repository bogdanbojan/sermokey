<script setup lang="ts">
import { navigate } from '../router/router'
import { t, useLanguage } from '../i18n/language'
import { useRecents } from '../recent'

const props = defineProps<{
  id: string
  nameNative: string
  nameEn: string
  wordCount: number
  isMixed?: boolean
  dueCount?: number
  newCount?: number
}>()

const { lang } = useLanguage()

function start() {
  useRecents(lang.value).record('vocab', props.id, props.nameNative)
  navigate({ name: 'flashcards', category: props.id })
}
</script>

<template>
  <div class="card vocab-card" :class="{ mixed: isMixed }">
    <p class="pill">{{ isMixed ? t.vocab.mixedTitle : nameEn }}</p>
    <h3>{{ nameNative }}</h3>
    <p class="muted usage">{{ props.wordCount }} {{ t.vocab.wordsLabel }}</p>
    <p v-if="dueCount || newCount" class="progress-hint">
      <span v-if="dueCount" class="hint-due">{{ dueCount }} {{ t.vocab.dueLabel }}</span>
      <span v-if="newCount" class="hint-new muted">{{ newCount }} {{ t.vocab.newLabel.toLowerCase() }}</span>
    </p>
    <button class="btn btn-primary btn-block" @click="start">
      {{ t.vocab.startSession }}
    </button>
  </div>
</template>

<style scoped>
.vocab-card {
  display: flex;
  flex-direction: column;
}
.vocab-card.mixed {
  border-color: var(--color-accent);
  border-width: 2px;
}
.usage {
  flex: 1;
  font-size: 0.92rem;
}
.progress-hint {
  display: flex;
  gap: var(--space-2);
  font-size: 0.8rem;
  margin-top: var(--space-1);
}
.hint-due {
  color: var(--color-accent-text);
  font-weight: 600;
}
.vocab-card .btn {
  margin-top: var(--space-3);
}
</style>
