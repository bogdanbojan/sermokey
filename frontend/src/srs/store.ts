// localStorage-backed persistence for spaced-repetition card state. This is
// a deliberate architecture choice: the app has no accounts/backend
// persistence anywhere else (it's a local, single-user tool), so review
// state lives entirely in the browser — the same approach Anki itself takes
// as a local-first app. Decks are keyed per language so French and Spanish
// progress never mix.
//
// Recognition ("see the word, recall its meaning") and production ("see
// the meaning, recall the word") are tracked as independent cards with
// independent scheduling, not two sides of one card — the literature
// (Laufer & Goldstein 2004; Nation 2001) treats them as genuinely different
// skills that mature at different rates, so a word can be "mastered" for
// recognition while still being "new" for production. See Direction below.
import { ref, watch, type Ref } from 'vue'
import type { LanguageCode } from '../types/api'
import { newCardState, review, isDue, bucketOf, type CardState, type Grade, type CardBucket } from './sm2'

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
function loadDirectionPref(): Direction | 'mixed' {
  const stored = localStorage.getItem(DIRECTION_STORAGE_KEY)
  return stored === 'production' || stored === 'mixed' ? stored : 'recognition'
}

export const vocabDirection = ref<Direction | 'mixed'>(loadDirectionPref())

watch(vocabDirection, (value) => localStorage.setItem(DIRECTION_STORAGE_KEY, value))

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

function loadDeck(lang: LanguageCode): Deck {
  try {
    const raw = localStorage.getItem(storageKey(lang))
    return raw ? (JSON.parse(raw) as Deck) : {}
  } catch {
    return {}
  }
}

function saveDeck(lang: LanguageCode, deck: Deck) {
  try {
    localStorage.setItem(storageKey(lang), JSON.stringify(deck))
  } catch (e) {
    console.warn('Could not persist flashcard progress to localStorage:', e)
  }
}

// One reactive deck per language, lazily loaded on first access.
const decks = new Map<LanguageCode, Ref<Deck>>()

function deckRef(lang: LanguageCode): Ref<Deck> {
  let existing = decks.get(lang)
  if (!existing) {
    existing = ref(loadDeck(lang))
    decks.set(lang, existing)
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
}

/** Anything that can sit in the SRS queue: a unique card ID is all that's required. */
export interface Cardlike {
  id: string
}

export function useSrsDeck(lang: LanguageCode) {
  const deck = deckRef(lang)

  function stateFor(cardId: string): CardState | undefined {
    return deck.value[cardId]
  }

  /**
   * Returns the review queue for a session: every due card first (oldest
   * due first), then up to `newCardLimit` never-seen cards appended. This
   * mirrors Anki's well-established default UX — reviews always take
   * priority, and new cards are introduced in small batches rather than
   * all at once, so a 700-word deck doesn't dump 700 "new" cards on day one.
   */
  function sessionQueue<T extends Cardlike>(cards: T[], newCardLimit = 15): T[] {
    const today = new Date()
    const due: T[] = []
    const fresh: T[] = []
    for (const c of cards) {
      const state = deck.value[c.id]
      if (state) {
        if (isDue(state, today)) due.push(c)
      } else {
        fresh.push(c)
      }
    }
    due.sort((a, b) => (deck.value[a.id]!.dueDate < deck.value[b.id]!.dueDate ? -1 : 1))
    return [...due, ...fresh.slice(0, newCardLimit)]
  }

  function recordReview(cardId: string, grade: Grade) {
    const current = deck.value[cardId] ?? newCardState()
    const next = review(current, grade)
    deck.value = { ...deck.value, [cardId]: next }
    saveDeck(lang, deck.value)
  }

  function stats<T extends Cardlike>(cards: T[]): DeckStats {
    const today = new Date()
    const counts: Record<CardBucket, number> = { new: 0, learning: 0, review: 0, mastered: 0 }
    let dueNow = 0
    for (const c of cards) {
      const state = deck.value[c.id]
      counts[bucketOf(state)]++
      if (state && isDue(state, today)) dueNow++
    }
    return { total: cards.length, dueNow, ...counts }
  }

  function resetDeck() {
    deck.value = {}
    saveDeck(lang, {})
  }

  return { stateFor, sessionQueue, recordReview, stats, resetDeck }
}
