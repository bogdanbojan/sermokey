<script setup lang="ts">
// The persistence warning and the backup controls.
//
// This exists because the browser is the only copy of a learner's review
// history. A failed write used to be a `console.warn`, so the UI carried on
// as though the review had landed; and there was no export path at all, so
// a cleared profile or a full disk was simply unrecoverable.
import { computed, ref } from 'vue'
import { t } from '../i18n/language'
import { exportAll, importAll, isBackup, storageHealth } from '../storage'

const restoring = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)
const restoreError = ref('')

const problem = computed(() => {
  if (storageHealth.value === 'quota') {
    return { title: t.value.storage.quotaTitle, detail: t.value.storage.quotaDetail }
  }
  if (storageHealth.value === 'unavailable') {
    return { title: t.value.storage.unavailableTitle, detail: t.value.storage.unavailableDetail }
  }
  return null
})

function download() {
  const blob = new Blob([JSON.stringify(exportAll(), null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `sermokey-backup-${new Date().toISOString().slice(0, 10)}.json`
  a.click()
  URL.revokeObjectURL(url)
}

async function onFile(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  restoring.value = true
  restoreError.value = ''
  try {
    const parsed: unknown = JSON.parse(await file.text())
    if (!isBackup(parsed)) {
      restoreError.value = t.value.storage.restoreConfirm
      return
    }
    if (!window.confirm(t.value.storage.restoreConfirm)) return
    importAll(parsed)
    // Every store ref was populated at import time and has no way to learn
    // that its backing key just changed underneath it.
    window.location.reload()
  } catch {
    restoreError.value = t.value.storage.restoreConfirm
  } finally {
    restoring.value = false
    if (fileInput.value) fileInput.value.value = ''
  }
}
</script>

<template>
  <div>
    <div v-if="problem" class="storage-banner" role="alert">
      <div>
        <strong>{{ problem.title }}</strong>
        <p>{{ problem.detail }}</p>
      </div>
      <button class="btn" @click="download">{{ t.storage.backup }}</button>
    </div>

    <div class="backup-row">
      <button class="link-btn" @click="download">{{ t.storage.backup }}</button>
      <button class="link-btn" :disabled="restoring" @click="fileInput?.click()">
        {{ t.storage.restore }}
      </button>
      <input
        ref="fileInput"
        type="file"
        accept="application/json,.json"
        class="visually-hidden"
        @change="onFile"
      />
    </div>
    <p v-if="restoreError" class="restore-error" role="alert">{{ restoreError }}</p>
  </div>
</template>

<style scoped>
.storage-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: var(--space-3);
  background: var(--color-warning-bg);
  color: var(--color-warning-text);
  border-bottom: 1px solid var(--color-warning-text);
  padding: var(--space-3) var(--space-5);
}
.storage-banner p {
  margin: 0.2rem 0 0;
  font-size: 0.85rem;
}

.backup-row {
  display: flex;
  gap: var(--space-3);
  justify-content: center;
  padding: var(--space-5) var(--space-3) var(--space-3);
}
.link-btn {
  background: none;
  border: none;
  padding: 0.35rem;
  font: inherit;
  font-size: 0.78rem;
  color: var(--color-text-muted);
  text-decoration: underline;
  cursor: pointer;
}
.link-btn:hover {
  color: var(--color-text);
}
.restore-error {
  text-align: center;
  color: var(--color-danger);
  font-size: 0.8rem;
}
</style>
