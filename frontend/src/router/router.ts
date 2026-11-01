import { ref } from 'vue'

/**
 * Which set of words a flashcard session draws from.
 *
 * This is a tagged union rather than a bare category string because there
 * are now three genuinely different sources, and only one of them is a
 * category: words saved while reading belong to no category, and a
 * per-text session is scoped by text id. Encoding those as reserved
 * prefixes inside the category string ("saved:", "text:") would make every
 * consumer parse it and would break the moment a real category shared a
 * prefix.
 */
export type DeckRef =
  /** One vocabulary category, or all of them when id is empty. */
  | { kind: 'category'; id: string }
  /** Every word saved while reading, across all texts. */
  | { kind: 'saved' }
  /** Just the words saved from one text — the end-of-text review. */
  | { kind: 'text'; id: string }
  /** A named deck: one made by the reader, or one standing for a text. */
  | { kind: 'deck'; id: string }

export type Route =
  | { name: 'home' }
  | { name: 'theory'; tense: string }
  | { name: 'quiz-setup'; tense: string }
  | { name: 'quiz'; tense: string; type: string; count: number }
  | { name: 'flashcards'; deck: DeckRef }
  | { name: 'read'; textId: string; section: number }

/** A stable key for a DeckRef, for use in storage keys and v-for keys. */
export function deckKey(deck: DeckRef): string {
  switch (deck.kind) {
    case 'category':
      return `category:${deck.id}`
    case 'saved':
      return 'saved'
    case 'text':
      return `text:${deck.id}`
    case 'deck':
      return `deck:${deck.id}`
  }
}

/** Bounds on the quiz length, so a hand-edited URL cannot ask for 100,000. */
export const MIN_QUIZ_COUNT = 1
export const MAX_QUIZ_COUNT = 50
const DEFAULT_QUIZ_COUNT = 10

/**
 * `decodeURIComponent` throws a URIError on a malformed percent escape —
 * a truncated shared link is enough. This runs at module scope to seed the
 * initial route, so an unguarded throw happened before Vue even mounted:
 * a blank page with no UI and no way to recover.
 */
function safeDecode(part: string): string {
  try {
    return decodeURIComponent(part)
  } catch {
    return part
  }
}

function parseHash(hash: string): Route {
  const clean = hash.replace(/^#\/?/, '')
  const parts = clean.split('/').filter(Boolean).map(safeDecode)

  if (parts[0] === 'theory' && parts[1]) return { name: 'theory', tense: parts[1] }
  if (parts[0] === 'quiz-setup' && parts[1]) return { name: 'quiz-setup', tense: parts[1] }
  if (parts[0] === 'quiz' && parts[1] && parts[2] && parts[3]) {
    // `Number('')` is 0 and `Number.isFinite(0)` is true, so an empty
    // segment used to request a zero-question quiz — which rendered as
    // "no exercises could be generated".
    const parsed = Number(parts[3])
    const count = Number.isFinite(parsed)
      ? Math.min(MAX_QUIZ_COUNT, Math.max(MIN_QUIZ_COUNT, Math.round(parsed)))
      : DEFAULT_QUIZ_COUNT
    return { name: 'quiz', tense: parts[1], type: parts[2], count }
  }
  if (parts[0] === 'read' && parts[1]) {
    const section = Number(parts[2])
    return { name: 'read', textId: parts[1], section: Number.isInteger(section) && section >= 0 ? section : 0 }
  }
  if (parts[0] === 'flashcards') {
    // #/flashcards            -> all vocabulary
    // #/flashcards/{category} -> one category
    // #/flashcards/saved      -> words saved while reading
    // #/flashcards/text/{id}  -> words saved from one text
    if (parts[1] === 'saved') return { name: 'flashcards', deck: { kind: 'saved' } }
    if (parts[1] === 'text' && parts[2]) {
      return { name: 'flashcards', deck: { kind: 'text', id: parts[2] } }
    }
    // Deck ids are themselves namespaced ("custom:abc", "text:la-parure"),
    // so the remaining segments are rejoined rather than taking parts[2].
    if (parts[1] === 'deck' && parts[2]) {
      return { name: 'flashcards', deck: { kind: 'deck', id: parts.slice(2).join('/') } }
    }
    return { name: 'flashcards', deck: { kind: 'category', id: parts[1] ?? '' } }
  }
  return { name: 'home' }
}

/** Exposed for tests; the parser is pure and worth covering directly. */
export const __parseHashForTests = parseHash

function routeToHash(route: Route): string {
  switch (route.name) {
    case 'home':
      return '#/'
    case 'theory':
      return `#/theory/${encodeURIComponent(route.tense)}`
    case 'quiz-setup':
      return `#/quiz-setup/${encodeURIComponent(route.tense)}`
    case 'quiz':
      return `#/quiz/${encodeURIComponent(route.tense)}/${encodeURIComponent(route.type)}/${route.count}`
    case 'read':
      return `#/read/${encodeURIComponent(route.textId)}/${route.section}`
    case 'flashcards':
      switch (route.deck.kind) {
        case 'saved':
          return '#/flashcards/saved'
        case 'text':
          return `#/flashcards/text/${encodeURIComponent(route.deck.id)}`
        case 'deck':
          // The id contains a ':' separator that must survive, so each
          // segment is encoded rather than the whole id.
          return `#/flashcards/deck/${route.deck.id.split('/').map(encodeURIComponent).join('/')}`
        case 'category':
          return `#/flashcards/${encodeURIComponent(route.deck.id)}`
      }
  }
}

const currentRoute = ref<Route>(parseHash(window.location.hash))

window.addEventListener('hashchange', () => {
  currentRoute.value = parseHash(window.location.hash)
})

export function useRoute() {
  return currentRoute
}

export function navigate(route: Route) {
  window.location.hash = routeToHash(route)
}
