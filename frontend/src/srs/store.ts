// localStorage-backed persistence for spaced-repetition card state. This is
// a deliberate architecture choice: the app has no accounts/backend
// persistence anywhere else (it's a local, single-user tool), so review
// state lives entirely in the browser — the same approach Anki itself takes
// as a local-first app. Decks are keyed per language so French and Spanish
// progress never mix.
//
// Because that makes the browser the only copy, every read goes through
// `storage.read` with a validating parser rather than a cast, every write
// reports whether it actually landed, and an external write from another
// tab reloads the deck instead of being overwritten by this tab's stale
// snapshot. See storage.ts for what each of those three used to do wrong.
//
// Recognition ("see the word, recall its meaning") and production ("see
// the meaning, recall the word") are tracked as independent cards with
// independent scheduling, not two sides of one card — the literature
// (Laufer & Goldstein 2004; Nation 2001) treats them as genuinely different
// skills that mature at different rates, so a word can be "mastered" for
// recognition while still being "new" for production. See Direction below.
import { ref, type Ref } from 'vue'
import type { LanguageCode } from '../types/api'
import { isRecord, read, subscribe, write, type WriteResult } from '../storage'
import {
  bucketOf,
  isDue,
  isLeech,
  newCardState,
  normalizeState,
  previewIntervals,
  review,
  type CardBucket,
  type CardState,
  type Grade,
} from './sm2'

const STORAGE_PREFIX = 'sermokey:srs:'
const DIRECTION_STORAGE_KEY = 'sermokey:vocab-direction'

export type Direction = 'recognition' | 'production'

/** Builds the composite SRS card ID for one word in one direction. */
export function cardID(wordId: string, direction: Direction): string {
  return `${wordId}::${direction}`
}

// Module-level singleton (same pattern as i18n/language.ts's `lang`): one
// review-direction preference shared across Home's progress hints and the
// flashcards screen itself, persisted so it survives reloads.
function parseDirection(u: unknown): Direction | 'mixed' | null {
  return u === 'production' || u === 'mixed' || u === 'recognition' ? u : null
}

export const vocabDirection = ref<Direction | 'mixed'>(
  read(DIRECTION_STORAGE_KEY, parseDirection, 'recognition'),
)

export function setVocabDirection(value: Direction | 'mixed') {
  vocabDirection.value = value
  write(DIRECTION_STORAGE_KEY, value)
}

/**
 * Maps a list of words to the SRS card IDs that matter for the given
 * review direction — used wherever only the stats (due/new counts) are
 * needed, not the full card rendering (see FlashcardsView.vue for the
 * richer version that also keeps the word/direction for display).
 */
export function cardIdsFor(words: { id: string }[], direction: Direction | 'mixed'): Cardlike[] {
  if (direction === 'mixed') {
    return words.flatMap((w) => [{ id: cardID(w.id, 'recognition') }, { id: cardID(w.id, 'production') }])
  }
  return words.map((w) => ({ id: cardID(w.id, direction) }))
}

type Deck = Record<string, CardState> // keyed by cardID(word.id, direction)

function storageKey(lang: LanguageCode): string {
  return STORAGE_PREFIX + lang
}

/**
 * Validates a stored deck.
 *
 * Entries that do not look like card state are dropped individually rather
 * than failing the whole deck: one corrupt card should cost one card, not
 * a learner's entire review history.
 */
function parseDeck(u: unknown): Deck | null {
  if (!isRecord(u)) return null
  const deck: Deck = {}
  for (const [id, raw] of Object.entries(u)) {
    if (!isRecord(raw)) continue
    if (typeof raw['easeFactor'] !== 'number' || typeof raw['intervalDays'] !== 'number') continue
    deck[id] = normalizeState(raw as unknown as CardState)
  }
  return deck
}

function loadDeck(lang: LanguageCode): Deck {
  return read(storageKey(lang), parseDeck, {})
}

// One reactive deck per language, lazily loaded on first access.
const decks = new Map<LanguageCode, Ref<Deck>>()

function deckRef(lang: LanguageCode): Ref<Deck> {
  let existing = decks.get(lang)
  if (!existing) {
    existing = ref(loadDeck(lang))
    decks.set(lang, existing)
    // Another tab just wrote this deck. Without this the two tabs each
    // hold a snapshot and whole-object writes mean the last one to save
    // silently discards everything the other did.
    subscribe(storageKey(lang), () => {
      existing!.value = loadDeck(lang)
    })
  }
  return existing
}

export interface DeckStats {
  total: number
  new: number
  learning: number
  review: number
  mastered: number
  dueNow: number
  /** Cards failed often enough to be worth rewriting rather than redrilling. */
  leeches: number
}

/** Anything that can sit in the SRS queue: a unique card ID is all that's required. */
export interface Cardlike {
  id: string
}

export interface SessionQueueOptions {
  /** How many never-seen cards to introduce. */
  newCardLimit?: number
  /** Ceiling on due cards, so two weeks away does not produce a 400-card wall. */
  reviewLimit?: number
  /**
   * Admit cards that are scheduled but not yet due. This is what the
   * "study anyway" button means; without it that button produced an empty
   * queue on any deck where everything had been seen.
   */
  includeNotDue?: boolean
}

export const DEFAULT_NEW_CARD_LIMIT = 15
export const DEFAULT_REVIEW_LIMIT = 120

export function useSrsDeck(lang: LanguageCode) {
  const deck = deckRef(lang)

  function stateFor(cardId: string): CardState | undefined {
    return deck.value[cardId]
  }

  /**
   * Returns the review queue for a session: due cards first (oldest due
   * first), then new cards, interleaved so a long session does not front-load
   * every review before showing anything fresh. This mirrors Anki's
   * well-established default UX — reviews take priority, and new cards are
   * introduced in small batches rather than all at once, so a 700-word deck
   * doesn't dump 700 "new" cards on day one.
   */
  function sessionQueue<T extends Cardlike>(cards: T[], options: SessionQueueOptions = {}): T[] {
    const {
      newCardLimit = DEFAULT_NEW_CARD_LIMIT,
      reviewLimit = DEFAULT_REVIEW_LIMIT,
      includeNotDue = false,
    } = options

    const today = new Date()
    const due: T[] = []
    const upcoming: T[] = []
    const fresh: T[] = []

    for (const c of cards) {
      const state = deck.value[c.id]
      if (!state) {
        fresh.push(c)
      } else if (state.suspended) {
        // A leech is withheld until the learner deals with it.
        continue
      } else if (isDue(state, today)) {
        due.push(c)
      } else if (includeNotDue) {
        upcoming.push(c)
      }
    }

    // A comparator that never returns 0 is not a valid ordering: equal due
    // dates (the common case, since they are day-granular strings) came out
    // in a different order on every call, so no session was reproducible.
    const byDueDate = (a: T, b: T) => {
      const x = deck.value[a.id]?.dueDate ?? ''
      const y = deck.value[b.id]?.dueDate ?? ''
      if (x < y) return -1
      if (x > y) return 1
      return a.id < b.id ? -1 : a.id > b.id ? 1 : 0
    }
    due.sort(byDueDate)
    upcoming.sort(byDueDate)

    const reviews = [...due, ...upcoming].slice(0, reviewLimit)
    return interleave(reviews, fresh.slice(0, newCardLimit))
  }

  /**
   * Applies a grade and reports whether the result was persisted.
   *
   * The caller must not advance past a card on anything other than 'ok' —
   * the review did not survive the reload and telling the learner otherwise
   * is how an hour of work disappears.
   */
  function recordReview(cardId: string, grade: Grade): { previous?: CardState; saved: WriteResult } {
    // One clock for the whole operation. Two `new Date()` calls either side
    // of midnight disagree about what day it is.
    const today = new Date()
    const previous = deck.value[cardId]
    const current = previous ?? newCardState(today)
    const next = review(current, grade, today)
    deck.value = { ...deck.value, [cardId]: next }
    return { previous, saved: write(storageKey(lang), deck.value) }
  }

  /** Restores a card to a prior state. Backs the undo button. */
  function restore(cardId: string, previous: CardState | undefined): WriteResult {
    const next = { ...deck.value }
    if (previous) next[cardId] = previous
    else delete next[cardId]
    deck.value = next
    return write(storageKey(lang), deck.value)
  }

  /** Clears the leech flag so a reworked card re-enters the rotation. */
  function unsuspend(cardId: string): WriteResult {
    const state = deck.value[cardId]
    if (!state) return 'ok'
    deck.value = { ...deck.value, [cardId]: { ...state, suspended: undefined, lapses: 0 } }
    return write(storageKey(lang), deck.value)
  }

  /** What each grade would do to this card, for the grade-button labels. */
  function preview(cardId: string) {
    return previewIntervals(deck.value[cardId] ?? newCardState())
  }

  function stats<T extends Cardlike>(cards: T[]): DeckStats {
    const today = new Date()
    const counts: Record<CardBucket, number> = { new: 0, learning: 0, review: 0, mastered: 0 }
    let dueNow = 0
    let leeches = 0
    for (const c of cards) {
      const state = deck.value[c.id]
      counts[bucketOf(state)]++
      if (state && isDue(state, today)) dueNow++
      if (state && isLeech(state)) leeches++
    }
    return { total: cards.length, dueNow, leeches, ...counts }
  }

  function resetDeck(): WriteResult {
    deck.value = {}
    return write(storageKey(lang), {})
  }

  return { deck, stateFor, sessionQueue, recordReview, restore, unsuspend, preview, stats, resetDeck }
}

/**
 * Merges reviews and new cards so fresh material is spread through the
 * session rather than all queued behind the reviews. Reviews still lead,
 * because forgetting something already learned costs more than delaying
 * something not yet learned.
 */
function interleave<T>(reviews: T[], fresh: T[]): T[] {
  if (fresh.length === 0) return reviews
  if (reviews.length === 0) return fresh
  const out: T[] = []
  const gap = Math.max(1, Math.floor(reviews.length / fresh.length))
  let f = 0
  for (let i = 0; i < reviews.length; i++) {
    out.push(reviews[i]!)
    if ((i + 1) % gap === 0 && f < fresh.length) out.push(fresh[f++]!)
  }
  while (f < fresh.length) out.push(fresh[f++]!)
  return out
}
