<script setup lang="ts">
import { onErrorCaptured, ref } from 'vue'
import { useRoute } from './router/router'
import { t } from './i18n/language'
import { clearAll, exportAll } from './storage'
import AppHeader from './components/AppHeader.vue'
import StorageNotice from './components/StorageNotice.vue'
import HomeView from './views/HomeView.vue'
import TheoryView from './views/TheoryView.vue'
import QuizSetupView from './views/QuizSetupView.vue'
import QuizView from './views/QuizView.vue'
import FlashcardsView from './views/FlashcardsView.vue'
import ReaderView from './views/ReaderView.vue'

const route = useRoute()

// An error boundary. Without one, a single thrown render — which corrupt
// localStorage used to cause on every single load — left a white page with
// no way back, because the bad value was re-read on the next boot too.
//
// The recovery screen leads with the backup, not the reset: whatever is
// still readable in storage is worth more than a working app.
const crashed = ref(false)
onErrorCaptured((err) => {
  console.error('sermokey: unhandled render error', err)
  crashed.value = true
  return false
})

function downloadBackup() {
  const blob = new Blob([JSON.stringify(exportAll(), null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `sermokey-backup-${new Date().toISOString().slice(0, 10)}.json`
  a.click()
  URL.revokeObjectURL(url)
}

function reload() {
  window.location.reload()
}

function reset() {
  if (!window.confirm(t.value.storage.resetConfirm)) return
  clearAll()
  window.location.reload()
}
</script>

<template>
  <AppHeader />

  <main v-if="crashed" class="page">
    <div class="card recovery">
      <h1>{{ t.storage.crashTitle }}</h1>
      <p class="muted">{{ t.storage.crashDetail }}</p>
      <div class="recovery-actions">
        <button class="btn btn-primary" @click="downloadBackup">{{ t.storage.backup }}</button>
        <button class="btn" @click="reload">{{ t.storage.reload }}</button>
        <button class="btn danger" @click="reset">{{ t.storage.reset }}</button>
      </div>
    </div>
  </main>

  <main v-else class="page">
    <HomeView v-if="route.name === 'home'" />
    <TheoryView v-else-if="route.name === 'theory'" :tense="route.tense" />
    <QuizSetupView v-else-if="route.name === 'quiz-setup'" :tense="route.tense" />
    <QuizView v-else-if="route.name === 'quiz'" :tense="route.tense" :type="route.type" :count="route.count" />
    <FlashcardsView v-else-if="route.name === 'flashcards'" :deck="route.deck" />
    <ReaderView v-else-if="route.name === 'read'" :text-id="route.textId" :section="route.section" />
    <StorageNotice />
  </main>
</template>

<style scoped>
.recovery {
  text-align: center;
  padding: var(--space-8) var(--space-5);
}
.recovery-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  justify-content: center;
  margin-top: var(--space-5);
}
.danger {
  border-color: var(--color-danger);
  color: var(--color-danger);
}
</style>
