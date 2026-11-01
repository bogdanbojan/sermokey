<script setup lang="ts">
// Chooses where saved words go.
//
// Deliberately *not* a dialog that interrupts the save. Saving a word is
// the highest-frequency action in the reader, and a picker in front of it
// would tax every single one. So the Save button stays a single tap and
// this is disclosure: the destination is shown next to it, and opens only
// when the reader wants to redirect it.
//
// The chosen destination then sticks, because a reader working through a
// book wants that book's deck for every word in it, not once per word.
import { computed, nextTick, ref } from 'vue'
import { t, format } from '../i18n/language'
import { useDecks, textDeckID, type Deck } from '../reading/decks'
import type { LanguageCode } from '../types/api'

const props = defineProps<{
  lang: LanguageCode
  /** The text being read, if any. Supplies the 'auto' destination. */
  currentText?: { id: string; title: string }
}>()

const decks = computed(() => useDecks(props.lang))
const open = ref(false)
const creating = ref(false)
const newName = ref('')
const newNameInput = ref<HTMLInputElement | null>(null)

/** The deck a save would go to right now, or undefined for My words only. */
const destination = computed<Deck | undefined>(() => decks.value.resolveTarget(props.currentText))

const destinationLabel = computed(() => destination.value?.name ?? t.value.reading.deckMyWords)

/** Decks offered in the list: the current text's, then every custom deck. */
const options = computed<Deck[]>(() => {
  const out: Deck[] = []
  if (props.currentText) {
    const id = textDeckID(props.currentText.id)
    out.push(
      decks.value.get(id) ?? {
        id,
        name: props.currentText.title,
        textId: props.currentText.id,
        createdAt: '',
      },
    )
  }
  for (const d of decks.value.customDecks.value) out.push(d)
  return out
})

/** True when the row represents the destination currently in effect. */
function isActive(deckId: string | undefined): boolean {
  return (destination.value?.id ?? undefined) === deckId
}

function choose(deckId: string | undefined) {
  // Picking the current text's deck returns to 'auto' rather than pinning
  // that id, so moving on to the next book follows along instead of
  // silently still filing words under the previous one.
  if (deckId && props.currentText && deckId === textDeckID(props.currentText.id)) {
    decks.value.setTarget('auto')
  } else {
    decks.value.setTarget(deckId ?? '')
  }
  open.value = false
}

async function startCreating() {
  creating.value = true
  await nextTick()
  newNameInput.value?.focus()
}

function confirmCreate() {
  const name = newName.value.trim()
  if (!name) return
  const deck = decks.value.create(name)
  decks.value.setTarget(deck.id)
  newName.value = ''
  creating.value = false
  open.value = false
}

function cancelCreate() {
  creating.value = false
  newName.value = ''
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && open.value) {
    e.stopPropagation()
    if (creating.value) cancelCreate()
    else open.value = false
  }
}

defineExpose({ destination })
</script>

<template>
  <div class="deck-picker" @keydown="onKeydown">
    <!-- The destination is always visible, so a word is never filed
         somewhere the reader cannot see. -->
    <p class="destination">
      <span class="dest-label">{{ t.reading.deckDestination }}</span>
      <button
        type="button"
        class="dest-btn"
        :aria-expanded="open"
        aria-haspopup="listbox"
        @click="open = !open"
      >
        <span class="dest-name">{{ destinationLabel }}</span>
        <span class="chevron" :class="{ up: open }" aria-hidden="true">▾</span>
      </button>
    </p>

    <div v-if="open" class="deck-list" role="listbox" :aria-label="t.reading.deckDestination">
      <button
        type="button"
        class="deck-option"
        role="option"
        :aria-selected="isActive(undefined)"
        :class="{ active: isActive(undefined) }"
        @click="choose(undefined)"
      >
        <span class="check" aria-hidden="true">{{ isActive(undefined) ? '✓' : '' }}</span>
        {{ t.reading.deckMyWords }}
      </button>

      <button
        v-for="d in options"
        :key="d.id"
        type="button"
        class="deck-option"
        role="option"
        :aria-selected="isActive(d.id)"
        :class="{ active: isActive(d.id) }"
        @click="choose(d.id)"
      >
        <span class="check" aria-hidden="true">{{ isActive(d.id) ? '✓' : '' }}</span>
        <span class="deck-name">{{ d.name }}</span>
        <span v-if="d.textId" class="deck-kind">{{ t.reading.tab }}</span>
      </button>

      <template v-if="creating">
        <form class="new-deck" @submit.prevent="confirmCreate">
          <input
            ref="newNameInput"
            v-model="newName"
            type="text"
            :placeholder="t.reading.deckNewPlaceholder"
            :aria-label="t.reading.deckNewPlaceholder"
            maxlength="60"
            @keydown.esc.prevent="cancelCreate"
          />
          <button type="submit" class="btn btn-primary btn-tiny" :disabled="!newName.trim()">
            {{ t.reading.deckCreate }}
          </button>
        </form>
      </template>
      <button v-else type="button" class="deck-option new" @click="startCreating">
        <span class="check" aria-hidden="true">+</span>
        {{ t.reading.deckNew }}
      </button>
    </div>

    <p v-if="destination && !open" class="dest-hint">
      {{ format(t.reading.deckAddedTo, { name: destinationLabel }) }}
    </p>
  </div>
</template>

<style scoped>
.deck-picker {
  margin-top: var(--space-2);
}

.destination {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin: 0;
  font-size: 0.78rem;
  color: var(--color-text-muted);
}
.dest-label {
  flex-shrink: 0;
}
.dest-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  min-width: 0;
  background: none;
  border: none;
  padding: 0.35rem 0.25rem;
  font: inherit;
  font-weight: 600;
  color: var(--color-primary);
  cursor: pointer;
}
.dest-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.chevron {
  /* 0.7em rendered the glyph as an ambiguous dot. It has to read as a
     disclosure affordance, since it is the only cue that the destination
     is changeable. */
  font-size: 0.9em;
  line-height: 1;
  opacity: 0.8;
  transition: transform 0.15s ease;
}
.chevron.up {
  transform: rotate(180deg);
}

.deck-list {
  margin-top: var(--space-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg);
  max-height: 11rem;
  overflow-y: auto;
}
.deck-option {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  width: 100%;
  background: none;
  border: none;
  /* A list row is a touch target before it is a label. */
  min-height: 2.5rem;
  padding: 0.4rem var(--space-2);
  font: inherit;
  font-size: 0.82rem;
  text-align: left;
  color: var(--color-text);
  cursor: pointer;
}
.deck-option:hover {
  background: var(--color-surface);
}
.deck-option.active {
  font-weight: 600;
}
.deck-option.new {
  color: var(--color-primary);
  font-weight: 600;
}
.check {
  width: 1rem;
  flex-shrink: 0;
  color: var(--color-accent-text);
}
.deck-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.deck-kind {
  margin-left: auto;
  flex-shrink: 0;
  font-size: 0.68rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
}

.new-deck {
  display: flex;
  gap: var(--space-2);
  padding: var(--space-2);
  border-top: 1px solid var(--color-border);
}
.new-deck input {
  flex: 1;
  min-width: 0;
  font: inherit;
  font-size: 0.82rem;
  padding: 0.4rem 0.5rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-surface);
  color: var(--color-text);
}
.btn-tiny {
  padding: 0.35rem 0.6rem;
  font-size: 0.78rem;
}

.dest-hint {
  margin: var(--space-1) 0 0;
  font-size: 0.72rem;
  color: var(--color-text-muted);
}

@media (prefers-reduced-motion: reduce) {
  .chevron {
    transition: none;
  }
}
</style>
