<script setup lang="ts">
// One paragraph, rendered as a run of token spans.
//
// One <span> per word over the verbatim text is what makes everything else
// work: native selection and Range map onto token indices via data-token,
// hit testing is the browser's problem, and per-word status decoration is a
// CSS class rather than a layout engine.
import { computed } from 'vue'
import type { ReadingParagraph } from '../types/api'
import { renderTokens } from '../reading/tokens'
import type { WordStatus } from '../reading/status'

const props = defineProps<{
  paragraph: ReadingParagraph
  index: number
  /**
   * Status per lexicon index, precomputed by the view.
   *
   * A plain map rather than a lookup function, on purpose. A function prop
   * hides its reactive reads inside a call made during render, which left
   * saving a word updating the deck but not repainting the word on the
   * page until something else forced a re-render. A computed map puts the
   * dependencies where Vue can see them — and it turns hundreds of
   * per-token calls per render into one pass per change.
   */
  entryStatus: Record<number, WordStatus | 'no-entry'>
  highlighting: boolean
  /** Token index of the pinned word, if it is in this paragraph. */
  pinned?: number
}>()

const tokens = computed(() => renderTokens(props.paragraph))

function classesFor(entryIndex: number, tokenIndex: number): string[] {
  const out: string[] = []
  if (props.highlighting) {
    out.push(props.entryStatus[entryIndex] ?? 'no-entry')
  }
  if (tokenIndex === props.pinned) out.push('pinned')
  return out
}
</script>

<template>
  <p class="para" :data-paragraph="index">
    <template v-for="tk in tokens" :key="tk.index">
      <span
        v-if="tk.word"
        class="tok"
        :class="classesFor(tk.entry, tk.index)"
        :data-token="tk.index"
        >{{ tk.text }}</span
      >
      <!-- Inter-word material carries data-token too, so a selection that
           starts or ends on a space still resolves to a token index
           instead of being silently discarded. -->
      <span v-else class="gap" :data-token="tk.index">{{ tk.text }}</span>
    </template>
  </p>
</template>

<style scoped>
.para {
  margin: 0 0 1.1em;
}

.tok {
  /* A 2px transparent underline is reserved on every word so that
     decorating one does not shift the line. */
  border-bottom: 2px solid transparent;
  border-radius: 2px;
  cursor: pointer;
  transition: background-color 0.1s ease;
}
.tok:hover {
  background: var(--color-accent-50);
}

/* Status decoration.
   Underlines rather than background fills: blocks of colour behind running
   text wreck the cream-paper/navy-ink palette and measurably hurt legibility
   over a long passage, whereas a rule under the word reads as annotation.
   Line style varies with status as well as hue, so the states remain
   distinguishable without colour (WCAG 1.4.1). */
.tok.unknown {
  border-bottom-color: var(--color-accent-400);
  background: var(--color-accent-50);
}
.tok.learning {
  border-bottom-style: dotted;
  border-bottom-color: var(--color-ink-400);
}
/* known, ignored and no-entry are deliberately undecorated: the page
   should get quieter as the reader's deck grows. That fading is the
   reward. */
.tok.known,
.tok.ignored,
.tok.no-entry {
  border-bottom-color: transparent;
}

.tok.pinned {
  background: var(--color-ink-100);
  border-bottom-color: var(--color-primary);
}

.gap {
  white-space: pre-wrap;
}
</style>
