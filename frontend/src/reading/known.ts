// Words the reader has declared they already know, and words they want the
// reader to stop asking about.
//
// Both exist because without them the coverage figure is worthless. A text
// is full of proper nouns ("Sophie"), numerals and words the reader has
// known for years; with no way to dismiss them they stay "unknown" forever,
// every page reports a pessimistic score, and the status colouring never
// settles. Lute and LWT both arrived at the same pair of statuses for
// exactly this reason.
//
// "Known" is kept as a plain set rather than as a fabricated SM-2 state
// with a distant due date. Faking review history to mean "do not ask me"
// would corrupt the deck statistics that the same history feeds.
import { ref, watch, type Ref } from 'vue'
import type { LanguageCode } from '../types/api'
import { read, subscribe, write } from '../storage'

const KNOWN_PREFIX = 'sermokey:known:'
const IGNORED_PREFIX = 'sermokey:ignored:'

interface Sets {
  known: Set<string>
  ignored: Set<string>
}

function parseIds(u: unknown): string[] | null {
  return Array.isArray(u) ? u.filter((x): x is string => typeof x === 'string') : null
}

function loadSet(key: string): Set<string> {
  return new Set(read(key, parseIds, []))
}

function saveSet(key: string, value: Set<string>) {
  return write(key, [...value])
}

const stores = new Map<LanguageCode, Ref<Sets>>()

function storeRef(lang: LanguageCode): Ref<Sets> {
  let existing = stores.get(lang)
  if (!existing) {
    existing = ref<Sets>({
      known: loadSet(KNOWN_PREFIX + lang),
      ignored: loadSet(IGNORED_PREFIX + lang),
    })
    stores.set(lang, existing)
    watch(existing, (v) => {
      saveSet(KNOWN_PREFIX + lang, v.known)
      saveSet(IGNORED_PREFIX + lang, v.ignored)
    })
    const reload = () => {
      existing!.value = {
        known: loadSet(KNOWN_PREFIX + lang),
        ignored: loadSet(IGNORED_PREFIX + lang),
      }
    }
    subscribe(KNOWN_PREFIX + lang, reload)
    subscribe(IGNORED_PREFIX + lang, reload)
  }
  return existing
}

export function useWordStatuses(lang: LanguageCode) {
  const store = storeRef(lang)

  // Sets are replaced rather than mutated so Vue sees the change; the
  // sets are small (a few thousand ids at most) and this happens only on
  // an explicit user action.
  function update(fn: (s: Sets) => Sets) {
    store.value = fn(store.value)
  }

  function isKnown(wordId: string) {
    return store.value.known.has(wordId)
  }
  function isIgnored(wordId: string) {
    return store.value.ignored.has(wordId)
  }

  /** Mark as already known: no card is created, and it counts as covered. */
  function markKnown(wordId: string) {
    update((s) => ({
      known: new Set(s.known).add(wordId),
      ignored: without(s.ignored, wordId),
    }))
  }

  /** Stop surfacing this word: proper nouns, numerals, names. It is
   *  excluded from coverage entirely rather than counted either way. */
  function markIgnored(wordId: string) {
    update((s) => ({
      known: without(s.known, wordId),
      ignored: new Set(s.ignored).add(wordId),
    }))
  }

  /** Back to untouched, so a mis-click is recoverable. */
  function clear(wordId: string) {
    update((s) => ({ known: without(s.known, wordId), ignored: without(s.ignored, wordId) }))
  }

  return { store, isKnown, isIgnored, markKnown, markIgnored, clear }
}

function without(set: Set<string>, id: string): Set<string> {
  if (!set.has(id)) return set
  const next = new Set(set)
  next.delete(id)
  return next
}
