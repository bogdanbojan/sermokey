// Reader preferences, reading position, and lookup counts. All three are
// small, local, and exist to make the reader feel like a place you return
// to rather than a page you load.
import { ref, watch, type Ref } from 'vue'
import type { LanguageCode } from '../types/api'
import { isRecord, read, remove as storageRemove, write } from '../storage'

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
  /** Leading, as a multiple of the font size. Exposed rather than fixed
   *  because after font size it is the single most effective accommodation
   *  for dyslexic and low-vision readers. */
  lineHeight: number
}

// Frozen, because it was previously handed out *by reference* as the
// fallback from `loadPrefs`. `readingPrefs.value` was then literally this
// object, so the first `readingPrefs.value.fontSize = …` silently rewrote
// the shipped defaults for the rest of the session. Nothing compared
// against them, so the bug was invisible until the reset control needed to
// know what "default" meant. Freezing makes a recurrence throw — these are
// ES modules, so assignment to a frozen property is a TypeError — rather
// than corrupting state quietly.
const DEFAULT_PREFS: ReadingPrefs = Object.freeze({
  fontSize: 1.125,
  serif: true,
  highlighting: true,
  lineHeight: 1.75,
})

/** Shared so the display menu can disable a stepper at its limit and
 *  compare the current state against the shipped defaults. */
export const DEFAULT_READING_PREFS: ReadingPrefs = DEFAULT_PREFS

export function isDefaultReadingPrefs(p: ReadingPrefs): boolean {
  return (
    p.fontSize === DEFAULT_PREFS.fontSize &&
    p.serif === DEFAULT_PREFS.serif &&
    p.highlighting === DEFAULT_PREFS.highlighting &&
    p.lineHeight === DEFAULT_PREFS.lineHeight
  )
}

/**
 * Restores the shipped defaults.
 *
 * Steppers can be walked back by hand, but only if you remember where
 * they started — and after a few taps nobody does. One line in a panel is
 * a cheaper answer than asking the reader to count clicks backwards.
 */
export function resetReadingPrefs() {
  readingPrefs.value = { ...DEFAULT_PREFS }
}

export const MIN_FONT_SIZE = 0.875
export const MAX_FONT_SIZE = 1.625
export const MIN_LINE_HEIGHT = 1.3
export const MAX_LINE_HEIGHT = 2.4

function parsePrefs(u: unknown): ReadingPrefs | null {
  if (!isRecord(u)) return null
  const fontSize = typeof u['fontSize'] === 'number' ? u['fontSize'] : DEFAULT_PREFS.fontSize
  return {
    // Clamped on read as well as on write: a value from a future build
    // with a wider range must not render the body at an unusable size.
    fontSize: Math.min(MAX_FONT_SIZE, Math.max(MIN_FONT_SIZE, fontSize)),
    serif: typeof u['serif'] === 'boolean' ? u['serif'] : DEFAULT_PREFS.serif,
    highlighting:
      typeof u['highlighting'] === 'boolean' ? u['highlighting'] : DEFAULT_PREFS.highlighting,
    lineHeight:
      typeof u['lineHeight'] === 'number'
        ? Math.min(MAX_LINE_HEIGHT, Math.max(MIN_LINE_HEIGHT, u['lineHeight']))
        : DEFAULT_PREFS.lineHeight,
  }
}

function loadPrefs(): ReadingPrefs {
  // Spread, never the constant itself: the result is mutable state.
  return { ...read(PREFS_KEY, parsePrefs, DEFAULT_PREFS) }
}

/** Module-level singleton, like the language and direction preferences. */
export const readingPrefs = ref<ReadingPrefs>(loadPrefs())

watch(readingPrefs, (v) => write(PREFS_KEY, v), { deep: true })

// --- Reading position ----------------------------------------------------

const POSITION_PREFIX = 'sermokey:reading-position:'

export interface ReadingPosition {
  section: number
  /** Index of the paragraph scrolled to, so a reload lands where you were
   *  rather than at the top of the chapter. */
  paragraph: number
}

function parsePosition(u: unknown): ReadingPosition | null {
  if (!isRecord(u)) return null
  if (typeof u['section'] !== 'number') return null
  return {
    section: u['section'],
    paragraph: typeof u['paragraph'] === 'number' ? u['paragraph'] : 0,
  }
}

function positionKey(lang: LanguageCode, textId: string): string {
  return POSITION_PREFIX + lang + ':' + textId
}

export function loadPosition(lang: LanguageCode, textId: string): ReadingPosition | undefined {
  return read(positionKey(lang, textId), parsePosition, undefined as ReadingPosition | undefined)
}

export function savePosition(lang: LanguageCode, textId: string, pos: ReadingPosition) {
  return write(positionKey(lang, textId), pos)
}

/** Forgets a text's position. Called when the text itself is deleted, so
 *  removing a book does not leave a dangling "resume" marker behind. */
export function forgetPosition(lang: LanguageCode, textId: string) {
  storageRemove(positionKey(lang, textId))
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

function parseCounts(u: unknown): LookupCounts | null {
  if (!isRecord(u)) return null
  const out: LookupCounts = {}
  for (const [k, v] of Object.entries(u)) if (typeof v === 'number') out[k] = v
  return out
}

// `ReturnType<typeof ref<T>>` resolves against the no-argument overload and
// yields `Ref<T | undefined>`, which is why this used to be read through
// `?? {}` and `?.` everywhere despite never actually being undefined.
const counts = new Map<LanguageCode, Ref<LookupCounts>>()

function countsRef(lang: LanguageCode): Ref<LookupCounts> {
  let existing = counts.get(lang)
  if (!existing) {
    existing = ref<LookupCounts>(read(LOOKUPS_PREFIX + lang, parseCounts, {}))
    counts.set(lang, existing)
  }
  return existing
}

export function useLookupCounts(lang: LanguageCode) {
  const store = countsRef(lang)

  function record(wordId: string): number {
    const next = { ...store.value }
    next[wordId] = (next[wordId] ?? 0) + 1
    store.value = next
    write(LOOKUPS_PREFIX + lang, next)
    return next[wordId]!
  }

  function countFor(wordId: string): number {
    return store.value[wordId] ?? 0
  }

  /** Whether to prompt the reader to save this word. */
  function shouldNudge(wordId: string): boolean {
    return countFor(wordId) >= NUDGE_AFTER_LOOKUPS
  }

  return { record, countFor, shouldNudge }
}
