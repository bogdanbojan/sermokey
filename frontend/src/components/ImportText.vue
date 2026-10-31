<script setup lang="ts">
// Importing an EPUB.
//
// The whole control is a drop zone with a button inside it, because the
// two gestures people actually use for "add a file I already have" are
// dragging it and picking it — and a drop zone that is also a button costs
// nothing extra to build while covering both. The <input type="file"> stays
// in the DOM, visually hidden rather than removed, so keyboard and
// assistive-technology users get the native file picker.
import { ref, computed } from 'vue'
import { api, ApiError } from '../api/client'
import type { ReadingText } from '../types/api'
import { useLanguage, t, format } from '../i18n/language'

const emit = defineEmits<{ imported: [text: ReadingText] }>()
const { lang } = useLanguage()

const input = ref<HTMLInputElement | null>(null)
const dragging = ref(false)
const busy = ref(false)
const error = ref('')
const progressName = ref('')

// Parsing and analyzing a full-length book takes a moment, so the control
// has to say that it is working. Without it the reader clicks again and
// uploads twice.
const label = computed(() =>
  busy.value ? format(t.value.reading.importing, { name: progressName.value }) : t.value.reading.import,
)

async function upload(file: File | undefined) {
  if (!file || busy.value) return
  error.value = ''
  if (!file.name.toLowerCase().endsWith('.epub')) {
    error.value = t.value.reading.importWrongType
    return
  }
  busy.value = true
  progressName.value = file.name
  try {
    emit('imported', await api.importText(lang.value, file))
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : t.value.reading.importFailed
  } finally {
    busy.value = false
    progressName.value = ''
    if (input.value) input.value.value = '' // allow re-picking the same file
  }
}

function onDrop(e: DragEvent) {
  dragging.value = false
  upload(e.dataTransfer?.files?.[0])
}
</script>

<template>
  <div
    class="import-zone"
    :class="{ dragging, busy }"
    @dragover.prevent="dragging = true"
    @dragleave.prevent="dragging = false"
    @drop.prevent="onDrop"
  >
    <input
      ref="input"
      type="file"
      accept=".epub,application/epub+zip"
      class="visually-hidden"
      :disabled="busy"
      @change="upload(($event.target as HTMLInputElement).files?.[0])"
    />
    <button class="btn" :disabled="busy" @click="input?.click()">
      {{ label }}
    </button>
    <span class="muted hint">{{ t.reading.importHint }}</span>
    <p v-if="error" class="error-text">{{ error }}</p>
  </div>
</template>

<style scoped>
.import-zone {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-3);
  border: 1px dashed var(--color-border);
  border-radius: var(--radius-md);
  padding: var(--space-3) var(--space-4);
  margin-bottom: var(--space-5);
  transition: border-color 0.15s ease, background-color 0.15s ease;
}
.import-zone.dragging {
  border-color: var(--color-accent);
  background: var(--color-accent-50);
}
.import-zone.busy {
  opacity: 0.75;
}
.hint {
  font-size: 0.82rem;
}
.error-text {
  color: var(--color-danger);
  font-size: 0.85rem;
  flex-basis: 100%;
  margin: 0;
}

/* Kept in the accessibility tree and focusable, just not painted —
   display:none would take the native file picker away from keyboard
   users entirely. */
.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
</style>
