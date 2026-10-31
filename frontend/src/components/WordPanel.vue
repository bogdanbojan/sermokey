<script setup lang="ts">
// The pinned word panel: everything the hover peek deliberately leaves out.
//
// On a narrow screen this is a bottom sheet rather than a popover. A
// popover anchored to a word on a phone covers the word and the line it
// sits in, which is exactly the context the reader needs to judge the
// sense. A sheet docks out of the way and leaves the sentence visible.
import { computed, ref, watch } from 'vue'
import type { ReadingEntry, ReadingAlt } from '../types/api'
import { t, useLanguage, format } from '../i18n/language'
import { useSavedWords, toSavedWord, type WordContext } from '../reading/saved'
import { useWordStatuses } from '../reading/known'
import { useLookupCounts } from '../reading/prefs'

const props = defineProps<{
  entry: ReadingEntry
  context?: WordContext
}>()
const emit = defineEmits<{ close: [] }>()

const { lang } = useLanguage()
const saved = computed(() => useSavedWords(lang.value))
const statuses = computed(() => useWordStatuses(lang.value))
const lookups = computed(() => useLookupCounts(lang.value))

// A reading is the primary entry or one of its alternatives; the panel
// treats them uniformly so correcting a mis-ranked homograph ("est" read as
// the compass point rather than as être) is one click, not a dead end.
type Reading = Pick<ReadingEntry, 'lemma' | 'pos' | 'gender' | 'formNote' | 'senses' | 'wordId' | 'curated'>

const readings = computed<Reading[]>(() => [props.entry, ...(props.entry.alternatives ?? [])])
const activeReading = ref(0)
const current = computed<Reading>(() => readings.value[activeReading.value] ?? props.entry)

// Which gloss becomes the flashcard's answer.
//
// Making the reader pick is the point, not friction for its own sake: the
// Involvement Load Hypothesis (Laufer & Hulstijn 2001) identifies
// evaluation — weighing which meaning fits here — as a driver of retention,
// and a card carrying eight comma-separated senses teaches none of them.
const chosenSense = ref(0)

watch(
  () => props.entry.surface,
  () => {
    activeReading.value = 0
    chosenSense.value = 0
  },
)
watch(activeReading, () => (chosenSense.value = 0))

const isSaved = computed(() => saved.value.has(current.value.wordId))
const isKnown = computed(() => statuses.value.isKnown(current.value.wordId))
const isIgnored = computed(() => statuses.value.isIgnored(current.value.wordId))

// The nudge. Repeatedly looking a word up is the strongest signal that it
// matters and is not sticking — the "need" component of Involvement Load.
const nudgeCount = computed(() => lookups.value.countFor(current.value.wordId))
const showNudge = computed(
  () => !isSaved.value && !isKnown.value && lookups.value.shouldNudge(current.value.wordId),
)

function save() {
  const gloss = current.value.senses[chosenSense.value]?.gloss ?? ''
  if (!gloss) return
  saved.value.add(
    toSavedWord(
      { ...current.value, surface: props.entry.surface },
      gloss,
      props.context,
    ),
  )
  statuses.value.clear(current.value.wordId)
}

function unsave() {
  saved.value.remove(current.value.wordId)
}

function markKnown() {
  saved.value.remove(current.value.wordId)
  statuses.value.markKnown(current.value.wordId)
}

function ignore() {
  saved.value.remove(current.value.wordId)
  statuses.value.markIgnored(current.value.wordId)
}

function undo() {
  statuses.value.clear(current.value.wordId)
}

function altLabel(a: Reading | ReadingAlt): string {
  return `${a.lemma} · ${a.senses[0]?.gloss ?? ''}`
}
</script>

<template>
  <!-- A real dialog, unlike the peek: focusable, labelled, dismissible, so
       the whole feature is usable without a pointer. -->
  <aside class="panel" role="dialog" :aria-label="entry.lemma" tabindex="-1">
    <header class="panel-head">
      <div class="title">
        <span class="lemma">{{ current.lemma }}</span>
        <span v-if="current.gender" class="gender" :class="current.gender">{{ current.gender }}</span>
        <span class="pos">{{ current.pos }}</span>
        <span v-if="current.curated" class="pill curated">★</span>
      </div>
      <button class="icon-btn" :aria-label="t.reading.coachDismiss" @click="emit('close')">×</button>
    </header>

    <p v-if="current.formNote" class="form-note">
      <span class="surface">{{ entry.surface }}</span> — {{ current.formNote }}
    </p>
    <p v-if="entry.uncertain" class="uncertain">{{ t.reading.uncertainNote }}</p>

    <p v-if="showNudge" class="nudge">{{ format(t.reading.nudge, { count: nudgeCount }) }}</p>

    <template v-if="current.senses.length">
      <p class="section-label">{{ t.reading.chooseSense }}</p>
      <ul class="senses">
        <li v-for="(s, i) in current.senses" :key="i">
          <label class="sense">
            <input v-model="chosenSense" type="radio" :value="i" name="sense" />
            <span>{{ s.gloss }}</span>
          </label>
        </li>
      </ul>
    </template>
    <p v-else class="muted">{{ t.reading.noDefinition }}</p>

    <!-- Alternatives, so a wrong primary reading is recoverable in place. -->
    <template v-if="readings.length > 1">
      <p class="section-label">{{ t.reading.otherReadings }}</p>
      <div class="alts">
        <button
          v-for="(a, i) in readings"
          :key="i"
          class="alt-chip"
          :class="{ active: i === activeReading }"
          @click="activeReading = i"
        >
          {{ altLabel(a) }}
        </button>
      </div>
    </template>

    <!-- The sentence the word was met in, carried onto the card. -->
    <p v-if="context" class="context">
      <span class="section-label">{{ t.reading.contextLabel }}</span>
      <span class="context-text">{{ context.sentence }}</span>
    </p>

    <p v-if="isKnown" class="state-note">
      {{ t.reading.knownNote }} · <button class="link" @click="undo">{{ t.reading.undo }}</button>
    </p>
    <p v-else-if="isIgnored" class="state-note">
      {{ t.reading.ignoredNote }} · <button class="link" @click="undo">{{ t.reading.undo }}</button>
    </p>

    <div class="actions">
      <button v-if="!isSaved" class="btn btn-primary" :disabled="!current.senses.length" @click="save">
        {{ t.reading.save }}
      </button>
      <button v-else class="btn saved-btn" @click="unsave">
        {{ t.reading.saved }} · {{ t.reading.unsave }}
      </button>
      <button v-if="!isKnown" class="btn" @click="markKnown">{{ t.reading.markKnown }}</button>
      <button v-if="!isIgnored" class="btn subtle" @click="ignore">{{ t.reading.ignore }}</button>
    </div>

    <!-- CC BY-SA requires attribution; the lexicon is Wiktionary-derived. -->
    <p class="attribution muted">{{ t.reading.attribution }}</p>
  </aside>
</template>

<style scoped>
.panel {
  position: fixed;
  right: var(--space-5);
  bottom: var(--space-5);
  width: min(26rem, calc(100vw - 2 * var(--space-5)));
  max-height: 70vh;
  overflow-y: auto;
  z-index: 70;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
  padding: var(--space-4);
  outline: none;
}
/* On a phone, dock to the bottom edge so the panel never covers the
   sentence the reader is judging the sense against. */
@media (max-width: 640px) {
  .panel {
    right: 0;
    left: 0;
    bottom: 0;
    width: 100%;
    max-height: 60vh;
    border-radius: var(--radius-lg) var(--radius-lg) 0 0;
    border-bottom: none;
  }
}
.panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-2);
}
.title {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: var(--space-2);
}
.lemma {
  font-size: 1.3rem;
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
.pill.curated {
  border-color: var(--color-accent);
  color: var(--color-accent-text);
  padding: 0 0.4rem;
}
.icon-btn {
  border: none;
  background: none;
  font-size: 1.4rem;
  line-height: 1;
  cursor: pointer;
  color: var(--color-text-muted);
  padding: 0 0.2rem;
}
.icon-btn:hover {
  color: var(--color-text);
}
.form-note {
  font-size: 0.82rem;
  color: var(--color-accent-text);
  margin: var(--space-1) 0 0;
}
.surface {
  font-weight: 600;
}
.uncertain {
  font-size: 0.78rem;
  color: var(--color-warning-text);
  margin: var(--space-1) 0 0;
}
.nudge {
  font-size: 0.82rem;
  background: var(--color-warning-bg);
  color: var(--color-warning-text);
  border-radius: var(--radius-sm);
  padding: var(--space-2);
  margin: var(--space-3) 0 0;
}
.section-label {
  font-size: 0.72rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
  margin: var(--space-3) 0 var(--space-1);
}
.senses {
  list-style: none;
  margin: 0;
  padding: 0;
}
.sense {
  display: flex;
  gap: var(--space-2);
  align-items: flex-start;
  padding: 0.25rem 0;
  cursor: pointer;
  font-size: 0.92rem;
}
.alts {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  /* Without this a long gloss makes a nowrap chip wider than the sheet and
     pushes it off the right edge of a phone screen: flex items refuse to
     shrink below their content width unless told they may. */
  min-width: 0;
}
.alt-chip {
  border: 1px solid var(--color-border);
  background: var(--color-bg);
  border-radius: 999px;
  padding: 0.25rem 0.7rem;
  font-size: 0.78rem;
  cursor: pointer;
  color: var(--color-text-muted);
  flex: 0 1 auto;
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.alt-chip.active {
  border-color: var(--color-primary);
  color: var(--color-primary);
  font-weight: 600;
}
.context {
  margin: var(--space-3) 0 0;
}
.context-text {
  display: block;
  font-size: 0.85rem;
  font-style: italic;
  color: var(--color-text-muted);
  border-left: 2px solid var(--color-border);
  padding-left: var(--space-3);
}
.state-note {
  font-size: 0.8rem;
  color: var(--color-text-muted);
  margin: var(--space-3) 0 0;
}
.link {
  border: none;
  background: none;
  padding: 0;
  color: var(--color-primary);
  cursor: pointer;
  font-weight: 600;
  font-size: inherit;
}
.actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin-top: var(--space-4);
}
.actions .btn {
  font-size: 0.85rem;
  padding: 0.45rem 0.8rem;
}
.saved-btn {
  border-color: var(--color-success);
  color: var(--color-success);
}
.subtle {
  color: var(--color-text-muted);
}
.attribution {
  font-size: 0.68rem;
  margin: var(--space-3) 0 0;
}
</style>
