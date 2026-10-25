// Tracks the last few things a learner actually touched (a tense practiced,
// a vocab category studied) so the Home screen can surface a "continue
// where you left off" shortcut row. This matters specifically because a
// spaced-repetition app is used *daily* — optimizing only for first-visit
// discovery and ignoring the returning-user path (by far the more common
// one) is a classic product-design miss.
import { ref, type Ref } from 'vue'
import type { LanguageCode } from './types/api'

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

function load(lang: LanguageCode): RecentItem[] {
  try {
    const raw = localStorage.getItem(storageKey(lang))
    return raw ? (JSON.parse(raw) as RecentItem[]) : []
  } catch {
    return []
  }
}

function persist(lang: LanguageCode, items: RecentItem[]) {
  try {
    localStorage.setItem(storageKey(lang), JSON.stringify(items))
  } catch {
    // best-effort; losing "recent" history silently is fine
  }
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

  return { items, record }
}
