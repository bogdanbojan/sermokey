<script setup lang="ts">
import { computed } from 'vue'
import type { ReadingText } from '../types/api'
import { navigate } from '../router/router'
import { t, useLanguage, format } from '../i18n/language'
import { useRecents } from '../recent'
import { useWordStatus, coverageOf } from '../reading/status'
import { loadPosition } from '../reading/prefs'

const props = defineProps<{ text: ReadingText }>()

const { lang } = useLanguage()
const { comprehensionOf } = useWordStatus(lang)

// Coverage is computed against this reader's own deck, so the same text is
// "comfortable" for one person and "hard" for another — both true, and far
// more useful than the static CEFR level, which is kept only as a
// secondary hint.
const coverage = computed(() => coverageOf(props.text, comprehensionOf))

// Until the reader's deck has something in it, coverage is 0 for every
// text and says nothing. The author-assigned CEFR level carries the card
// instead, and the coverage row appears once it has real information.
const bandLabel = computed(() => {
  switch (coverage.value.band) {
    case 'comfortable':
      return t.value.reading.bandComfortable
    case 'challenging':
      return t.value.reading.bandChallenging
    default:
      return t.value.reading.bandHard
  }
})

// Floored, not rounded. The bands are threshold-based, so rounding 94.6%
// up to "95 %" put the displayed figure on the far side of the boundary
// from the badge beside it — a card reading "95 % known · hard" when 95%
// is exactly where "challenging" begins. Flooring guarantees the number
// never claims a threshold the band has not reached.
const shownPercent = computed(() => Math.floor(coverage.value.known * 100))

const position = computed(() => loadPosition(lang.value, props.text.id))
const hasProgress = computed(() => (position.value?.section ?? 0) > 0)

function open() {
  useRecents(lang.value).record('vocab', props.text.id, props.text.title)
  navigate({ name: 'read', textId: props.text.id, section: position.value?.section ?? 0 })
}
</script>

<template>
  <div class="card text-card">
    <div class="meta-row">
      <p class="pill">{{ t.reading.levelLabel }} {{ text.level }}</p>
      <p
        v-if="!coverage.coldStart"
        class="pill band"
        :class="coverage.band"
        :title="t.reading.bandHint"
      >
        {{ bandLabel }}
      </p>
    </div>

    <h3>{{ text.title }}</h3>
    <p v-if="text.author" class="muted author">{{ text.author }}</p>
    <p class="muted blurb">{{ text.blurb }}</p>

    <p class="muted stats">
      {{ format(t.reading.tokensLabel, { count: text.tokenCount }) }} ·
      {{ format(t.reading.sectionsLabel, { count: text.sectionCount }) }}
    </p>

    <!-- The coverage bar doubles as the difficulty signal and as progress:
         it fills as the reader's deck grows. Hidden until the deck has
         something to say, rather than shown pinned at zero. -->
    <div v-if="!coverage.coldStart" class="coverage" :title="t.reading.bandHint">
      <div class="coverage-bar">
        <div class="coverage-fill" :class="coverage.band" :style="{ width: `${coverage.known * 100}%` }" />
      </div>
      <p class="coverage-text">
        <span class="known">{{
          format(t.reading.coverageLabel, { percent: shownPercent })
        }}</span>
        <span class="muted">{{ format(t.reading.newWordsLabel, { count: coverage.newWords }) }}</span>
      </p>
    </div>

    <button class="btn btn-primary btn-block" @click="open">
      {{ hasProgress ? t.reading.continueReading : t.reading.read }}
    </button>
  </div>
</template>

<style scoped>
.text-card {
  display: flex;
  flex-direction: column;
}
.meta-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-bottom: var(--space-2);
}
.band.comfortable {
  border-color: var(--color-success);
  color: var(--color-success);
}
.band.challenging {
  border-color: var(--color-warning-text);
  color: var(--color-warning-text);
}
.band.hard {
  border-color: var(--color-danger);
  color: var(--color-danger);
}
.author {
  font-size: 0.85rem;
  margin-top: calc(-1 * var(--space-2));
}
.blurb {
  flex: 1;
  font-size: 0.92rem;
}
.stats {
  font-size: 0.8rem;
}
.coverage {
  margin: var(--space-2) 0 var(--space-1);
}
.coverage-bar {
  height: 5px;
  background: var(--color-border);
  border-radius: 999px;
  overflow: hidden;
}
.coverage-fill {
  height: 100%;
  transition: width 0.25s ease;
  background: var(--color-accent);
}
.coverage-fill.comfortable {
  background: var(--color-success);
}
.coverage-fill.challenging {
  background: var(--color-accent);
}
.coverage-fill.hard {
  background: var(--color-danger);
}
.coverage-text {
  display: flex;
  justify-content: space-between;
  font-size: 0.78rem;
  margin: var(--space-1) 0 0;
}
.known {
  font-weight: 600;
  color: var(--color-accent-text);
}
.text-card .btn {
  margin-top: var(--space-3);
}
</style>
