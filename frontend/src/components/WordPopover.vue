<script setup lang="ts">
// The hover peek.
//
// Deliberately read-only and button-free. A hover popup containing controls
// forces the reader to chase it with the pointer, and the gap between word
// and popup makes it flicker — the classic hover-menu failure that needs
// safe-triangle hacks to paper over. Splitting the interaction solves it
// properly: hover answers "what does this mean?" and nothing else, and a
// click pins a panel where actions can live without moving.
//
// It also shows at most two glosses. Eight senses in a tooltip is noise
// while reading, and looking up every word in depth is precisely what
// destroys reading fluency (Grabe 2009).
import { computed } from 'vue'
import type { ReadingEntry } from '../types/api'
import { t } from '../i18n/language'

const props = defineProps<{
  entry: ReadingEntry | null
  x: number
  y: number
}>()

const MAX_PEEK_SENSES = 2

const senses = computed(() => props.entry?.senses.slice(0, MAX_PEEK_SENSES) ?? [])
const extraSenses = computed(() =>
  Math.max(0, (props.entry?.senses.length ?? 0) - MAX_PEEK_SENSES),
)

// Clamped so the popup never leaves the viewport on a word near the right
// edge or the last line.
const style = computed(() => {
  const margin = 12
  const width = 320
  const left = Math.min(Math.max(margin, props.x - width / 2), window.innerWidth - width - margin)
  const flipAbove = props.y > window.innerHeight - 160
  return {
    left: `${left}px`,
    top: flipAbove ? 'auto' : `${props.y + 18}px`,
    bottom: flipAbove ? `${window.innerHeight - props.y + 14}px` : 'auto',
    width: `${width}px`,
  }
})
</script>

<template>
  <!-- aria-hidden and pointer-events:none: this is a visual peek only. The
       same information is reachable by keyboard and screen reader through
       the pinned panel, which is a real focusable dialog. -->
  <div v-if="entry" class="peek" :style="style" aria-hidden="true">
    <div class="head">
      <span class="lemma">{{ entry.lemma }}</span>
      <span v-if="entry.gender" class="gender" :class="entry.gender">{{ entry.gender }}</span>
      <span class="pos">{{ entry.pos }}</span>
    </div>

    <!-- The deinflection chain. Showing it is what connects "parlons" back
         to the conjugation table the learner studied, instead of letting it
         look like an unrelated new word. -->
    <p v-if="entry.formNote" class="form-note">{{ entry.formNote }}</p>

    <ul class="senses">
      <li v-for="(s, i) in senses" :key="i">{{ s.gloss }}</li>
    </ul>

    <p class="hint">
      <span v-if="extraSenses">+{{ extraSenses }}</span>
      <span>{{ t.reading.lookupHint }}</span>
    </p>
  </div>
</template>

<style scoped>
.peek {
  position: fixed;
  z-index: 60;
  pointer-events: none;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-card);
  padding: var(--space-3);
  font-size: 0.9rem;
  line-height: 1.45;
}
@media (prefers-reduced-motion: no-preference) {
  .peek {
    animation: peek-in 0.09s ease-out;
  }
  @keyframes peek-in {
    from {
      opacity: 0;
      transform: translateY(-2px);
    }
  }
}
.head {
  display: flex;
  align-items: baseline;
  gap: var(--space-2);
}
.lemma {
  font-weight: 700;
  color: var(--color-heading);
}
.gender {
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  border-radius: 999px;
  padding: 0 0.35rem;
  border: 1px solid currentColor;
}
.gender.m {
  color: var(--color-primary);
}
.gender.f {
  color: var(--color-warning-text);
}
.gender.n {
  color: var(--color-success);
}
.pos {
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
}
.form-note {
  font-size: 0.78rem;
  color: var(--color-accent-text);
  margin: var(--space-1) 0 0;
}
.senses {
  margin: var(--space-2) 0 0;
  padding-left: 1.1rem;
}
.senses li {
  margin-bottom: 0.1rem;
}
.hint {
  display: flex;
  justify-content: space-between;
  gap: var(--space-2);
  font-size: 0.72rem;
  color: var(--color-text-muted);
  margin: var(--space-2) 0 0;
}
</style>
