<script setup lang="ts">
// Typography and display settings for the reading surface.
//
// One "Aa" button opening a panel, rather than six micro-buttons sitting
// in the toolbar. That is the pattern every long-form reader converges on
// — Kindle, Apple Books, Instapaper, Pocket — for two reasons worth
// stating: these are set-once preferences that do not deserve permanent
// chrome on the screen they are meant to get out of the way of, and a row
// of 32px icon buttons is both unreadable and untappable on a phone.
//
// Collapsing them also makes Reset affordable. As a seventh toolbar button
// it was not worth the clutter; as one line in a panel it answers a real
// question a stepper cannot — "where was the default?" — because after a
// few taps there is no way to tell.
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { t } from '../i18n/language'
import {
  readingPrefs,
  resetReadingPrefs,
  isDefaultReadingPrefs,
  MIN_FONT_SIZE,
  MAX_FONT_SIZE,
  MIN_LINE_HEIGHT,
  MAX_LINE_HEIGHT,
} from '../reading/prefs'

const open = ref(false)
const rootEl = ref<HTMLElement | null>(null)
const triggerEl = ref<HTMLButtonElement | null>(null)

function nudgeFontSize(delta: number) {
  const next = Math.round((readingPrefs.value.fontSize + delta) * 1000) / 1000
  readingPrefs.value.fontSize = Math.min(MAX_FONT_SIZE, Math.max(MIN_FONT_SIZE, next))
}

function nudgeLineHeight(delta: number) {
  const next = Math.round((readingPrefs.value.lineHeight + delta) * 100) / 100
  readingPrefs.value.lineHeight = Math.min(MAX_LINE_HEIGHT, Math.max(MIN_LINE_HEIGHT, next))
}

// Steppers disable at the ends rather than silently clamping, so a reader
// can tell the difference between "this does nothing" and "this is broken".
const atMinSize = computed(() => readingPrefs.value.fontSize <= MIN_FONT_SIZE)
const atMaxSize = computed(() => readingPrefs.value.fontSize >= MAX_FONT_SIZE)
const atMinLeading = computed(() => readingPrefs.value.lineHeight <= MIN_LINE_HEIGHT)
const atMaxLeading = computed(() => readingPrefs.value.lineHeight >= MAX_LINE_HEIGHT)

const isDefault = computed(() => isDefaultReadingPrefs(readingPrefs.value))

function toggle() {
  open.value = !open.value
}

function close() {
  if (!open.value) return
  open.value = false
  triggerEl.value?.focus()
}

function onDocumentPointerDown(e: PointerEvent) {
  if (!open.value) return
  const target = e.target as Node | null
  if (target && rootEl.value?.contains(target)) return
  open.value = false
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && open.value) {
    e.stopPropagation()
    close()
  }
}

onMounted(() => document.addEventListener('pointerdown', onDocumentPointerDown))
onBeforeUnmount(() => document.removeEventListener('pointerdown', onDocumentPointerDown))
</script>

<template>
  <div ref="rootEl" class="display-menu" @keydown="onKeydown">
    <button
      ref="triggerEl"
      type="button"
      class="trigger"
      :class="{ open }"
      :aria-expanded="open"
      aria-haspopup="dialog"
      :aria-label="t.reading.displayMenu"
      :title="t.reading.displayMenu"
      @click="toggle"
    >
      <span class="aa" aria-hidden="true">Aa</span>
    </button>

    <div
      v-if="open"
      class="panel"
      role="dialog"
      :aria-label="t.reading.displayMenu"
    >
      <div class="row">
        <span class="row-label" id="size-label">{{ t.reading.textSizeLabel }}</span>
        <div class="stepper" role="group" aria-labelledby="size-label">
          <button type="button" :disabled="atMinSize" :title="t.reading.smaller" :aria-label="t.reading.smaller" @click="nudgeFontSize(-0.125)">
            <span aria-hidden="true">A−</span>
          </button>
          <button type="button" :disabled="atMaxSize" :title="t.reading.larger" :aria-label="t.reading.larger" @click="nudgeFontSize(0.125)">
            <span class="big" aria-hidden="true">A+</span>
          </button>
        </div>
      </div>

      <div class="row">
        <span class="row-label" id="leading-label">{{ t.reading.lineHeightLabel }}</span>
        <div class="stepper" role="group" aria-labelledby="leading-label">
          <button type="button" :disabled="atMinLeading" :title="t.reading.tighter" :aria-label="t.reading.tighter" @click="nudgeLineHeight(-0.15)">
            <span aria-hidden="true">↕−</span>
          </button>
          <button type="button" :disabled="atMaxLeading" :title="t.reading.looser" :aria-label="t.reading.looser" @click="nudgeLineHeight(0.15)">
            <span aria-hidden="true">↕+</span>
          </button>
        </div>
      </div>

      <div class="row">
        <span class="row-label" id="face-label">{{ t.reading.typefaceLabel }}</span>
        <div class="segmented" role="group" aria-labelledby="face-label">
          <button
            type="button"
            class="serif-sample"
            :class="{ active: readingPrefs.serif }"
            :aria-pressed="readingPrefs.serif"
            @click="readingPrefs.serif = true"
          >
            {{ t.reading.serif }}
          </button>
          <button
            type="button"
            :class="{ active: !readingPrefs.serif }"
            :aria-pressed="!readingPrefs.serif"
            @click="readingPrefs.serif = false"
          >
            {{ t.reading.sans }}
          </button>
        </div>
      </div>

      <!-- A labelled switch, not a bare "H". The letter was the keyboard
           shortcut masquerading as a label and meant nothing on sight. -->
      <label class="row switch-row">
        <span class="row-label">
          {{ t.reading.highlightLabel }}
          <kbd>H</kbd>
        </span>
        <input v-model="readingPrefs.highlighting" type="checkbox" class="switch" />
      </label>

      <button type="button" class="reset" :disabled="isDefault" @click="resetReadingPrefs">
        {{ t.reading.resetDisplay }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.display-menu {
  position: relative;
}

.trigger {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  /* One target instead of six, and a real one. */
  min-width: 2.75rem;
  min-height: 2.5rem;
  padding: 0.3rem 0.6rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-surface);
  color: var(--color-text);
  cursor: pointer;
}
.trigger:hover {
  border-color: var(--color-text-muted);
}
.trigger.open {
  background: var(--color-primary);
  border-color: var(--color-primary);
  color: var(--color-primary-contrast);
}
.aa {
  font-family: Georgia, 'Iowan Old Style', 'Times New Roman', serif;
  font-size: 1rem;
  font-weight: 600;
  line-height: 1;
}

.panel {
  position: absolute;
  right: 0;
  top: calc(100% + var(--space-2));
  z-index: 68;
  width: 16rem;
  padding: var(--space-3);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}
.row-label {
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--color-text);
}
.row-label kbd {
  margin-left: 0.3rem;
  font-size: 0.68rem;
  opacity: 0.8;
}

.stepper,
.segmented {
  display: inline-flex;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  overflow: hidden;
  background: var(--color-bg);
}
.stepper button,
.segmented button {
  border: none;
  background: none;
  min-width: 2.5rem;
  min-height: 2.25rem;
  padding: 0 0.5rem;
  font: inherit;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--color-text-muted);
  cursor: pointer;
}
.stepper button + button,
.segmented button + button {
  border-left: 1px solid var(--color-border);
}
.stepper button:hover:not(:disabled),
.segmented button:hover {
  color: var(--color-text);
}
.stepper button:disabled {
  opacity: 0.4;
  cursor: default;
}
.segmented button.active {
  background: var(--color-primary);
  color: var(--color-primary-contrast);
}
.serif-sample {
  font-family: Georgia, 'Iowan Old Style', 'Times New Roman', serif;
}
.big {
  font-size: 1rem;
}

.switch-row {
  cursor: pointer;
}
.switch {
  width: 1.15rem;
  height: 1.15rem;
  accent-color: var(--color-primary);
  cursor: pointer;
}

.reset {
  margin-top: var(--space-1);
  padding: 0.5rem;
  min-height: 2.25rem;
  border: none;
  border-top: 1px solid var(--color-border);
  background: none;
  font: inherit;
  font-size: 0.8rem;
  color: var(--color-primary);
  cursor: pointer;
}
.reset:disabled {
  color: var(--color-text-muted);
  opacity: 0.6;
  cursor: default;
}

/* On a phone the panel docks to the screen edges rather than hanging off
   a 44px trigger, where a 16rem popover would overflow the viewport. */
@media (max-width: 640px) {
  .panel {
    position: fixed;
    right: var(--space-3);
    left: var(--space-3);
    top: auto;
    bottom: calc(var(--space-3) + env(safe-area-inset-bottom));
    width: auto;
  }
}
</style>
