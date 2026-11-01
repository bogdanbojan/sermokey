// Named decks for words saved while reading.
//
// A deck here is a *view* over saved words, not a separate store. Every
// saved word always belongs to "My words"; a deck adds membership on top.
// That has three consequences worth stating, because they are the reason
// for the shape below:
//
//  - Deleting a deck never loses words. It removes a grouping, nothing
//    more, so the destructive-action question mostly disappears.
//  - A word can sit in several decks without being duplicated, so its
//    review history stays single-valued. The SRS card id is minted from
//    the word, not from the deck.
//  - A text's deck does not have to be created by anyone. "Words I met in
//    La Parure" is a fact about the word, already recorded in its context,
//    so a text deck is derivable for words saved before decks existed.
//
// Text decks and custom decks are deliberately the same type. The only
// difference is that one is created for you, named after the text, the
// first time you save a word while reading it.
import { computed, ref, type Ref } from 'vue'
import type { LanguageCode } from '../types/api'
import { isRecord, read, subscribe, write } from '../storage'

const DECKS_PREFIX = 'sermokey:decks:'
const TARGET_PREFIX = 'sermokey:deck-target:'

/** Prefix on a text deck's id. Also how `inDeck` knows to fall back to context. */
const TEXT_PREFIX = 'text:'
const CUSTOM_PREFIX = 'custom:'

export interface Deck {
  /** "text:la-parure" or "custom:k3f9a2". Namespaced so the two kinds
   *  share one id space and one storage key. */
  id: string
  name: string
  /** Set on a deck that stands for a text, so the reader can find it. */
  textId?: string
  createdAt: string
}

/** The id a text's deck has, whether or not it exists yet. */
export function textDeckID(textId: string): string {
  return TEXT_PREFIX + textId
}

export function isTextDeck(deckId: string): boolean {
  return deckId.startsWith(TEXT_PREFIX)
}

/** The text a deck stands for, or undefined for a custom deck. */
export function textIDOf(deckId: string): string | undefined {
  return isTextDeck(deckId) ? deckId.slice(TEXT_PREFIX.length) : undefined
}

/**
 * Where new saves go.
 *
 * 'auto' means "the text being read", which is what makes saving a word a
 * single tap in the common case: a reader working through a book almost
 * always wants that book's deck, and should not have to say so once per
 * word. An explicit id overrides it and sticks until changed.
 */
export type DeckTarget = 'auto' | string

function decksKey(lang: LanguageCode) {
  return DECKS_PREFIX + lang
}
function targetKey(lang: LanguageCode) {
  return TARGET_PREFIX + lang
}

function parseDecks(u: unknown): Deck[] | null {
  if (!Array.isArray(u)) return null
  return u.filter(
    (d): d is Deck => isRecord(d) && typeof d['id'] === 'string' && typeof d['name'] === 'string',
  )
}

const stores = new Map<LanguageCode, Ref<Deck[]>>()

function decksRef(lang: LanguageCode): Ref<Deck[]> {
  let existing = stores.get(lang)
  if (!existing) {
    existing = ref<Deck[]>(read(decksKey(lang), parseDecks, []))
    stores.set(lang, existing)
    subscribe(decksKey(lang), () => {
      existing!.value = read(decksKey(lang), parseDecks, [])
    })
  }
  return existing
}

const targets = new Map<LanguageCode, Ref<DeckTarget>>()

function targetRef(lang: LanguageCode): Ref<DeckTarget> {
  let existing = targets.get(lang)
  if (!existing) {
    existing = ref<DeckTarget>(
      read(targetKey(lang), (u) => (typeof u === 'string' ? u : null), 'auto' as DeckTarget),
    )
    targets.set(lang, existing)
  }
  return existing
}

let counter = 0
function newCustomID(): string {
  // Date plus a counter rather than a random id: ids are only ever local,
  // and a monotonic one keeps the stored list readable when debugging.
  counter += 1
  return CUSTOM_PREFIX + Date.now().toString(36) + counter.toString(36)
}

export function useDecks(lang: LanguageCode) {
  const decks = decksRef(lang)
  const target = targetRef(lang)

  const customDecks = computed(() => decks.value.filter((d) => !isTextDeck(d.id)))
  const textDecks = computed(() => decks.value.filter((d) => isTextDeck(d.id)))

  function get(id: string): Deck | undefined {
    return decks.value.find((d) => d.id === id)
  }

  function persist(next: Deck[]) {
    decks.value = next
    return write(decksKey(lang), next)
  }

  /** Creates a custom deck and returns it, or returns the existing one if
   *  the name is already taken — re-typing a name should not fork a deck. */
  function create(name: string): Deck {
    const trimmed = name.trim()
    const existing = decks.value.find(
      (d) => !isTextDeck(d.id) && d.name.toLowerCase() === trimmed.toLowerCase(),
    )
    if (existing) return existing

    const deck: Deck = { id: newCustomID(), name: trimmed, createdAt: new Date().toISOString() }
    persist([...decks.value, deck])
    return deck
  }

  /**
   * Returns the deck standing for a text, creating it on first use.
   *
   * Called when a word is saved while reading, not when a text is opened,
   * so merely browsing a book does not litter the library with empty decks.
   */
  function ensureTextDeck(textId: string, title: string): Deck {
    const id = textDeckID(textId)
    const existing = get(id)
    if (existing) {
      // The title can change when a book is re-imported with better
      // metadata; keep the deck's name in step.
      if (title && existing.name !== title) {
        persist(decks.value.map((d) => (d.id === id ? { ...d, name: title } : d)))
      }
      return get(id)!
    }
    const deck: Deck = { id, name: title || textId, textId, createdAt: new Date().toISOString() }
    persist([...decks.value, deck])
    return deck
  }

  function rename(id: string, name: string) {
    const trimmed = name.trim()
    if (!trimmed) return
    persist(decks.value.map((d) => (d.id === id ? { ...d, name: trimmed } : d)))
  }

  /**
   * Forgets a deck.
   *
   * Words are untouched: they stay in "My words" and keep their context,
   * so a text deck removed this way simply reappears the next time a word
   * is saved from that text. Membership cleanup is the caller's job — see
   * useSavedWords().removeFromAllDecks.
   */
  function remove(id: string) {
    persist(decks.value.filter((d) => d.id !== id))
    if (target.value === id) setTarget('auto')
  }

  function setTarget(next: DeckTarget) {
    target.value = next
    write(targetKey(lang), next)
  }

  /**
   * The deck a save would currently go to.
   *
   * `currentText` is supplied by the reader. Outside a text, 'auto' means
   * no deck beyond "My words", which is the right default for a word
   * looked up from the vocabulary screen.
   */
  function resolveTarget(currentText?: { id: string; title: string }): Deck | undefined {
    if (target.value === 'auto') {
      if (!currentText) return undefined
      return get(textDeckID(currentText.id)) ?? {
        id: textDeckID(currentText.id),
        name: currentText.title,
        textId: currentText.id,
        createdAt: '',
      }
    }
    return get(target.value)
  }

  return {
    decks,
    customDecks,
    textDecks,
    target,
    get,
    create,
    ensureTextDeck,
    rename,
    remove,
    setTarget,
    resolveTarget,
  }
}
