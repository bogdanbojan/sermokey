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
import { ref, watch, type Ref } from 'vue'
import type { LanguageCode, ReadingEntry } from '../types/api'

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
}

type SavedMap = Record<string, SavedWord>

function storageKey(lang: LanguageCode): string {
  return STORAGE_PREFIX + lang
}

function load(lang: LanguageCode): SavedMap {
  try {
    const raw = localStorage.getItem(storageKey(lang))
    return raw ? (JSON.parse(raw) as SavedMap) : {}
  } catch {
    return {}
  }
}

function save(lang: LanguageCode, value: SavedMap) {
  try {
    localStorage.setItem(storageKey(lang), JSON.stringify(value))
  } catch (e) {
    console.warn('Could not persist saved words to localStorage:', e)
  }
}

// One reactive store per language, lazily loaded — the same pattern as
// srs/store.ts, so French and Spanish never mix.
const stores = new Map<LanguageCode, Ref<SavedMap>>()

function storeRef(lang: LanguageCode): Ref<SavedMap> {
  let existing = stores.get(lang)
  if (!existing) {
    existing = ref(load(lang))
    watch(existing, (v) => save(lang, v), { deep: false })
    stores.set(lang, existing)
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

  function add(word: SavedWord) {
    store.value = { ...store.value, [word.id]: word }
  }

  function remove(wordId: string) {
    const next = { ...store.value }
    delete next[wordId]
    store.value = next
  }

  return { store, has, get, all, fromText, add, remove }
}

/** Builds a SavedWord from a dictionary entry plus the reader's choices. */
export function toSavedWord(
  entry: Pick<ReadingEntry, 'lemma' | 'pos' | 'gender' | 'wordId' | 'surface'>,
  gloss: string,
  context?: WordContext,
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
  }
}
