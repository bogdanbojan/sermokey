<script setup lang="ts">
// One card in the "My decks" group: the all-words aggregate, a text's deck,
// or a deck the reader made.
//
// Extracted because the saved-words card and the per-deck card had grown
// into near-identical markup with separately-maintained styles, which is
// exactly how two things that should look alike stop looking alike. One
// component means the three kinds cannot drift apart.
import { t } from '../i18n/language'
import { pluralize } from '../i18n/language'
import type { PluralForms } from '../i18n/strings'

defineProps<{
  name: string
  /** Short classifier shown above the name: the deck's kind. */
  kind: string
  count: number
  countLabel: PluralForms
  dueCount?: number
  blurb?: string
  /**
   * The "everything" card of its group.
   *
   * Exactly one card per group carries this, which is what makes the
   * accent read as a rule rather than as decoration. It previously landed
   * on two cards out of three for no stated reason.
   */
  aggregate?: boolean
  /** Only decks the reader created, or orphaned ones, can be forgotten. */
  forgettable?: boolean
  /** The source text has been deleted. Muted rather than alarming: the
   *  words are intact, only the thing that grouped them is gone. */
  orphaned?: boolean
  disabled?: boolean
}>()

const emit = defineEmits<{ start: []; forget: [] }>()
</script>

<template>
  <div class="card deck-card" :class="{ aggregate, orphaned }">
    <div class="deck-head">
      <p class="pill">{{ kind }}</p>
      <button
        v-if="forgettable"
        class="forget-btn"
        :title="t.reading.deckForget"
        :aria-label="t.reading.deckForget"
        @click="emit('forget')"
      >
        ×
      </button>
    </div>

    <h3>{{ name }}</h3>
    <p v-if="blurb" class="muted blurb">{{ blurb }}</p>

    <p class="progress-hint">
      <span v-if="dueCount" class="hint-due">{{ dueCount }} {{ t.vocab.dueLabel }}</span>
      <span class="muted">{{ pluralize(countLabel, count) }}</span>
    </p>

    <button class="btn btn-primary btn-block" :disabled="disabled" @click="emit('start')">
      {{ t.vocab.startSession }}
    </button>
  </div>
</template>

<style scoped>
.deck-card {
  display: flex;
  flex-direction: column;
}
.deck-card.aggregate {
  border-color: var(--color-accent);
  border-width: 2px;
}
.deck-card.orphaned .pill {
  background: var(--color-warning-bg);
  color: var(--color-warning-text);
}

.deck-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-2);
}
.deck-card h3 {
  margin-top: 0;
}
.blurb {
  font-size: 0.92rem;
}

.forget-btn {
  background: none;
  border: none;
  /* A destructive control needs a real target, and a colour cue that is
     not gated behind hover — there is no hover on a phone. */
  min-width: 2.25rem;
  min-height: 2.25rem;
  padding: 0;
  margin: calc(var(--space-1) * -1) calc(var(--space-1) * -1) 0 0;
  font-size: 1.15rem;
  line-height: 1;
  color: var(--color-text-muted);
  cursor: pointer;
  border-radius: var(--radius-sm);
}
.forget-btn:hover,
.forget-btn:focus-visible {
  color: var(--color-danger);
}

.progress-hint {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  font-size: 0.8rem;
  /* Pushes the button to the bottom so a row of cards aligns regardless of
     whether each has a blurb or a due count. */
  margin-top: auto;
  padding-top: var(--space-2);
}
.hint-due {
  color: var(--color-accent-text);
  font-weight: 600;
}
.deck-card .btn {
  margin-top: var(--space-3);
}
</style>
