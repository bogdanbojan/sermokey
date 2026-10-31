// Reader preferences, reading position, and lookup counts. All three are
// small, local, and exist to make the reader feel like a place you return
// to rather than a page you load.
import { ref, watch } from 'vue'
import type { LanguageCode } from '../types/api'

// --- Typography and display ----------------------------------------------

const PREFS_KEY = 'sermokey:reading-prefs'

/**
 * The rest of the app is a UI surface; this is a reading surface, and the
 * defaults differ accordingly. Long-form L2 reading wants a measure around
 * 60-70 characters (Bringhurst's 45-75), looser leading than UI text, and
 * the option of a serif face — all of which are comfort settings a reader
 * expects to control.
 */
export interface ReadingPrefs {
  /** Body size in rem. */
  fontSize: number
  serif: boolean
  /** Status colouring on the page. On by default: it is what makes reading
   *  feel like progress. Off is offered because decoration measurably
   *  competes with fluency for some readers (Grabe 2009), and because
   *  sometimes you just want to read. */
  highlighting: boolean
}

const DEFAULT_PREFS: ReadingPrefs = { fontSize: 1.125, serif: true, highlighting: true }

export const MIN_FONT_SIZE = 0.875
export const MAX_FONT_SIZE = 1.625

function loadPrefs(): ReadingPrefs {
  try {
    const raw = localStorage.getItem(PREFS_KEY)
    return raw ? { ...DEFAULT_PREFS, ...(JSON.parse(raw) as Partial<ReadingPrefs>) } : DEFAULT_PREFS
  } catch {
    return DEFAULT_PREFS
  }
}

/** Module-level singleton, like the language and direction preferences. */
export const readingPrefs = ref<ReadingPrefs>(loadPrefs())

watch(
  readingPrefs,
  (v) => {
    try {
      localStorage.setItem(PREFS_KEY, JSON.stringify(v))
    } catch {
      /* quota or private mode; preferences are not worth failing over */
    }
  },
  { deep: true },
)

// --- Reading position ----------------------------------------------------

const POSITION_PREFIX = 'sermokey:reading-position:'

export interface ReadingPosition {
  section: number
  /** Index of the paragraph scrolled to, so a reload lands where you were
   *  rather than at the top of the chapter. */
  paragraph: number
}

export function loadPosition(lang: LanguageCode, textId: string): ReadingPosition | undefined {
  try {
    const raw = localStorage.getItem(POSITION_PREFIX + lang + ':' + textId)
    return raw ? (JSON.parse(raw) as ReadingPosition) : undefined
  } catch {
    return undefined
  }
}

export function savePosition(lang: LanguageCode, textId: string, pos: ReadingPosition) {
  try {
    localStorage.setItem(POSITION_PREFIX + lang + ':' + textId, JSON.stringify(pos))
  } catch {
    /* non-critical */
  }
}

// --- Lookup counts -------------------------------------------------------

const LOOKUPS_PREFIX = 'sermokey:lookups:'

/**
 * How many times each word has been looked up.
 *
 * Looking a word up repeatedly is the strongest available signal that it
 * matters and is not sticking — the "need" component of the Involvement
 * Load Hypothesis (Laufer & Hulstijn 2001). Kindle turns every lookup into
 * a saved word, which captures everything and therefore also captures a lot
 * of noise. Counting instead, and prompting only once a word has been
 * looked up repeatedly, keeps the deck something the reader chose while
 * still surfacing the words they keep stumbling on.
 */
export const NUDGE_AFTER_LOOKUPS = 3

type LookupCounts = Record<string, number>

const counts = new Map<LanguageCode, ReturnType<typeof ref<LookupCounts>>>()

function countsRef(lang: LanguageCode) {
  let existing = counts.get(lang)
  if (!existing) {
    let initial: LookupCounts = {}
    try {
      const raw = localStorage.getItem(LOOKUPS_PREFIX + lang)
      if (raw) initial = JSON.parse(raw) as LookupCounts
    } catch {
      /* start fresh */
    }
    existing = ref<LookupCounts>(initial)
    counts.set(lang, existing)
  }
  return existing
}

export function useLookupCounts(lang: LanguageCode) {
  const store = countsRef(lang)

  function record(wordId: string): number {
    const next = { ...(store.value ?? {}) }
    next[wordId] = (next[wordId] ?? 0) + 1
    store.value = next
    try {
      localStorage.setItem(LOOKUPS_PREFIX + lang, JSON.stringify(next))
    } catch {
      /* non-critical */
    }
    return next[wordId]
  }

  function countFor(wordId: string): number {
    return store.value?.[wordId] ?? 0
  }

  /** Whether to prompt the reader to save this word. */
  function shouldNudge(wordId: string): boolean {
    return countFor(wordId) >= NUDGE_AFTER_LOOKUPS
  }

  return { record, countFor, shouldNudge }
}
