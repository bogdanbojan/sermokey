// Tracks the last few things a learner actually touched (a tense practiced,
// a vocab category studied) so the Home screen can surface a "continue
// where you left off" shortcut row. This matters specifically because a
// spaced-repetition app is used *daily* — optimizing only for first-visit
// discovery and ignoring the returning-user path (by far the more common
// one) is a classic product-design miss.
import { ref, type Ref } from 'vue'
import type { LanguageCode } from './types/api'
import { isRecord, read, write } from './storage'

export type RecentKind = 'theory' | 'quiz' | 'vocab'

export interface RecentItem {
  kind: RecentKind
  id: string // tense id or vocab category id ("" = mixed)
  label: string
  at: number // epoch ms
}

const MAX_RECENTS = 3
const STORAGE_PREFIX = 'sermokey:recent:'

function storageKey(lang: LanguageCode): string {
  return STORAGE_PREFIX + lang
}

/** A stored object rather than an array used to survive the old cast and
 *  then throw inside `record`'s `.filter`, outside any try/catch. */
function parseRecents(u: unknown): RecentItem[] | null {
  if (!Array.isArray(u)) return null
  return u.filter(
    (it): it is RecentItem =>
      isRecord(it) && typeof it['id'] === 'string' && typeof it['label'] === 'string',
  )
}

function load(lang: LanguageCode): RecentItem[] {
  return read(storageKey(lang), parseRecents, [])
}

function persist(lang: LanguageCode, items: RecentItem[]) {
  return write(storageKey(lang), items)
}

const caches = new Map<LanguageCode, Ref<RecentItem[]>>()

function cacheFor(lang: LanguageCode): Ref<RecentItem[]> {
  let existing = caches.get(lang)
  if (!existing) {
    existing = ref(load(lang))
    caches.set(lang, existing)
  }
  return existing
}

export function useRecents(lang: LanguageCode) {
  const items = cacheFor(lang)

  function record(kind: RecentKind, id: string, label: string) {
    const next = items.value.filter((it) => !(it.kind === kind && it.id === id))
    next.unshift({ kind, id, label, at: Date.now() })
    items.value = next.slice(0, MAX_RECENTS)
    persist(lang, items.value)
  }

  /**
   * Drops any entry pointing at `id`.
   *
   * Called when the thing itself is deleted. Without it a removed text
   * kept its "continue" chip, which no longer matched any text and so fell
   * through to the vocabulary branch of the router — navigating to a
   * category that does not exist.
   */
  function forget(id: string) {
    const next = items.value.filter((it) => it.id !== id)
    if (next.length === items.value.length) return
    items.value = next
    persist(lang, next)
  }

  return { items, record, forget }
}
