// Words captured while reading.
//
// This module stores word *content*; scheduling stays entirely in
// srs/store.ts. That split is deliberate and is why adding reading required
// no change to the SRS code at all: the deck is keyed by card id and knows
// nothing about where a word came from, so a word saved from a text and a
// word studied from a category are the same kind of thing to it.
//
// The sentence a word was met in is stored alongside it. A gloss learned in
// context is retained better than a bare pair (Webb 2008), and it gives the
// flashcard a back side worth reading and a way back to the source.
import { ref, type Ref } from 'vue'
import type { LanguageCode, ReadingEntry } from '../types/api'
import { isRecord, read, subscribe, write } from '../storage'
import { textIDOf } from './decks'

const STORAGE_PREFIX = 'sermokey:saved:'

/** Where a saved word was first met. */
export interface WordContext {
  textId: string
  textTitle: string
  section: number
  /** The sentence as it appeared, used verbatim on the flashcard back. */
  sentence: string
  /** Offsets of the word within `sentence`, so it can be emphasised
   *  without a fragile re-search that would highlight the wrong occurrence
   *  when a word repeats. */
  start: number
  end: number
}

export interface SavedWord {
  /** The SRS card id. Minted server-side; see ReadingEntry.wordId. */
  id: string
  lemma: string
  pos: string
  gender?: string
  /** The single gloss the reader chose. Deliberately one, not all of them:
   *  picking is the "evaluation" step that drives retention (Laufer &
   *  Hulstijn 2001), and a card carrying eight senses teaches none. */
  gloss: string
  /** The form actually encountered, when it differs from the lemma. */
  surface?: string
  context?: WordContext
  savedAt: string
  /**
   * Named decks this word belongs to, beyond "My words".
   *
   * Absent on words saved before decks existed, which is why `inDeck`
   * falls back to the stored context for a text deck: "I met this word in
   * La Parure" is already recorded there, so those words appear in the
   * text's deck without a migration.
   */
  decks?: string[]
}

type SavedMap = Record<string, SavedWord>

function storageKey(lang: LanguageCode): string {
  return STORAGE_PREFIX + lang
}

/**
 * Validates a stored map, dropping entries that are not saved words.
 *
 * A single malformed entry costs that entry, not the whole collection —
 * the previous `as SavedMap` cast turned a stored `null` into a crash on
 * the first `wordId in store.value`.
 */
function parseSaved(u: unknown): SavedMap | null {
  if (!isRecord(u)) return null
  const out: SavedMap = {}
  for (const [id, raw] of Object.entries(u)) {
    if (!isRecord(raw)) continue
    if (typeof raw['id'] !== 'string' || typeof raw['lemma'] !== 'string') continue
    out[id] = raw as unknown as SavedWord
  }
  return out
}

function load(lang: LanguageCode): SavedMap {
  return read(storageKey(lang), parseSaved, {})
}

function save(lang: LanguageCode, value: SavedMap) {
  return write(storageKey(lang), value)
}

// One reactive store per language, lazily loaded — the same pattern as
// srs/store.ts, so French and Spanish never mix.
const stores = new Map<LanguageCode, Ref<SavedMap>>()

function storeRef(lang: LanguageCode): Ref<SavedMap> {
  let existing = stores.get(lang)
  if (!existing) {
    existing = ref(load(lang))
    stores.set(lang, existing)
    // Saved words are written explicitly by add/remove rather than by a
    // watcher, so that a failed write is visible to the caller instead of
    // disappearing into a watcher callback.
    subscribe(storageKey(lang), () => {
      existing!.value = load(lang)
    })
  }
  return existing
}

export function useSavedWords(lang: LanguageCode) {
  const store = storeRef(lang)

  function has(wordId: string): boolean {
    return wordId in store.value
  }

  function get(wordId: string): SavedWord | undefined {
    return store.value[wordId]
  }

  function all(): SavedWord[] {
    return Object.values(store.value)
  }

  /** Words first met in one text, for the end-of-text review session. */
  function fromText(textId: string): SavedWord[] {
    return all().filter((w) => w.context?.textId === textId)
  }

  /**
   * Words in one named deck.
   *
   * A text deck unions explicit membership with the source text recorded
   * in the word's context, so words saved before decks existed are still
   * found by the deck named after the text they came from.
   */
  function inDeck(deckId: string): SavedWord[] {
    const textId = textIDOf(deckId)
    return all().filter(
      (w) => w.decks?.includes(deckId) || (textId !== undefined && w.context?.textId === textId),
    )
  }

  /** Adds an already-saved word to a deck. */
  function addToDeck(wordId: string, deckId: string) {
    const word = store.value[wordId]
    if (!word || word.decks?.includes(deckId)) return 'ok' as const
    const next: SavedWord = { ...word, decks: [...(word.decks ?? []), deckId] }
    store.value = { ...store.value, [wordId]: next }
    return save(lang, store.value)
  }

  function removeFromDeck(wordId: string, deckId: string) {
    const word = store.value[wordId]
    if (!word?.decks?.includes(deckId)) return 'ok' as const
    const next: SavedWord = { ...word, decks: word.decks.filter((d) => d !== deckId) }
    store.value = { ...store.value, [wordId]: next }
    return save(lang, store.value)
  }

  /**
   * Strips a deck from every word that referenced it.
   *
   * Called when a deck is forgotten. The words themselves are kept — a
   * deck is a grouping, and losing the grouping must not lose the
   * vocabulary or its review history.
   */
  function removeFromAllDecks(deckId: string) {
    const next: SavedMap = {}
    let changed = false
    for (const [id, w] of Object.entries(store.value)) {
      if (w.decks?.includes(deckId)) {
        next[id] = { ...w, decks: w.decks.filter((d) => d !== deckId) }
        changed = true
      } else {
        next[id] = w
      }
    }
    if (!changed) return 'ok' as const
    store.value = next
    return save(lang, store.value)
  }

  /** Deck ids a word is explicitly a member of. */
  function decksOf(wordId: string): string[] {
    return store.value[wordId]?.decks ?? []
  }

  function add(word: SavedWord) {
    store.value = { ...store.value, [word.id]: word }
    return save(lang, store.value)
  }

  function remove(wordId: string) {
    const next = { ...store.value }
    delete next[wordId]
    store.value = next
    return save(lang, store.value)
  }

  return {
    store,
    has,
    get,
    all,
    fromText,
    inDeck,
    addToDeck,
    removeFromDeck,
    removeFromAllDecks,
    decksOf,
    add,
    remove,
  }
}

/** Builds a SavedWord from a dictionary entry plus the reader's choices. */
export function toSavedWord(
  entry: Pick<ReadingEntry, 'lemma' | 'pos' | 'gender' | 'wordId' | 'surface'>,
  gloss: string,
  context?: WordContext,
  decks?: string[],
): SavedWord {
  return {
    id: entry.wordId,
    lemma: entry.lemma,
    pos: entry.pos,
    gender: entry.gender,
    gloss,
    surface: entry.surface !== entry.lemma ? entry.surface : undefined,
    context,
    savedAt: new Date().toISOString(),
    decks: decks?.length ? decks : undefined,
  }
}
