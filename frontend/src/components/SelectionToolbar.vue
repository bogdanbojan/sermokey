<script setup lang="ts">
import { computed } from 'vue'
// The floating toolbar that appears over a multi-word selection.
//
// Phrase capture is not a nice-to-have: formulaic sequences are learned and
// stored as units (Wray 2002), and the curated vocabulary itself already
// contains multi-word items like "salle de sport". A reader who can only
// save single tokens cannot capture "tout à coup" or "avoir besoin de" at
// all.
//
// The pattern — select text, a small toolbar appears at the selection — is
// the one Medium and Notion established, so it needs no explanation, and it
// rides on native selection, which means it works with touch long-press and
// with the keyboard for free.
import { t, format } from '../i18n/language'
import { MAX_PHRASE_TOKENS } from '../reading/tokens'

const props = defineProps<{
  text: string
  x: number
  y: number
  tooLong: boolean
}>()
const emit = defineEmits<{ lookup: []; dismiss: [] }>()

/**
 * Keeps the toolbar on screen.
 *
 * The anchor is the top of the selection minus a few pixels, and the
 * toolbar is translated fully above it — so selecting text on the first
 * visible line put it off the top of the viewport, where it could not be
 * reached at all. WordPopover already clamps and flips; this did not.
 */
const ESTIMATED_HEIGHT = 40
const EDGE = 8

const placement = computed(() => {
  const flipBelow = props.y < ESTIMATED_HEIGHT + EDGE
  const halfWidth = 176 // half of the 22rem max-width
  const maxX = Math.max(halfWidth + EDGE, window.innerWidth - halfWidth - EDGE)
  return {
    left: `${Math.min(Math.max(props.x, halfWidth + EDGE), maxX)}px`,
    top: flipBelow ? `${props.y + 24}px` : `${props.y}px`,
    transform: flipBelow ? 'translate(-50%, 0)' : 'translate(-50%, -100%)',
  }
})
</script>

<template>
  <div
    class="toolbar"
    :style="placement"
    role="toolbar"
    :aria-label="t.reading.selectionLookup"
    @mousedown.prevent
    @touchstart.prevent
  >
    <template v-if="tooLong">
      <span class="too-long">{{ format(t.reading.selectionTooLong, { max: MAX_PHRASE_TOKENS }) }}</span>
    </template>
    <template v-else>
      <span class="phrase">{{ text }}</span>
      <button class="toolbar-btn" @click="emit('lookup')">{{ t.reading.selectionLookup }}</button>
    </template>
  </div>
</template>

<style scoped>
.toolbar {
  position: fixed;
  z-index: 65;
  display: flex;
  align-items: center;
  gap: var(--space-2);
  background: var(--color-ink-950);
  color: #fff;
  border-radius: var(--radius-sm);
  padding: 0.35rem 0.5rem;
  box-shadow: var(--shadow-card);
  font-size: 0.82rem;
  max-width: min(22rem, 90vw);
}
.phrase {
  font-style: italic;
  opacity: 0.8;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.toolbar-btn {
  border: none;
  background: rgba(255, 255, 255, 0.15);
  color: #fff;
  border-radius: calc(var(--radius-sm) - 3px);
  padding: 0.45rem 0.7rem;
  min-height: 2.25rem;
  font-weight: 600;
  cursor: pointer;
  font-size: 0.8rem;
  white-space: nowrap;
}
.toolbar-btn:hover {
  background: rgba(255, 255, 255, 0.28);
}
.too-long {
  opacity: 0.85;
}
</style>
