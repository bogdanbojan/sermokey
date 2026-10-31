import type {
  Exercise,
  Theory,
  TenseSummary,
  VerbSummary,
  Language,
  LanguageCode,
  VocabCategory,
  VocabWord,
  ReadingText,
  ReadingSection,
  ReadingEntry,
} from '../types/api'

const BASE_URL = (import.meta.env.VITE_API_BASE_URL as string | undefined) ?? 'http://localhost:8080'

class ApiError extends Error {}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let res: Response
  try {
    res = await fetch(`${BASE_URL}${path}`, init)
  } catch {
    throw new ApiError(
      `Could not reach the backend at ${BASE_URL}. Is it running? (cd backend && go run ./cmd/server)`,
    )
  }
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { error?: string } | null
    throw new ApiError(body?.error ?? `Request to ${path} failed with status ${res.status}`)
  }
  if (res.status === 204) return undefined as T
  return (await res.json()) as T
}

async function getJSON<T>(path: string): Promise<T> {
  let res: Response
  try {
    res = await fetch(`${BASE_URL}${path}`)
  } catch {
    throw new ApiError(
      `Could not reach the backend at ${BASE_URL}. Is it running? (cd backend && go run ./cmd/server)`,
    )
  }
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { error?: string } | null
    throw new ApiError(body?.error ?? `Request to ${path} failed with status ${res.status}`)
  }
  return (await res.json()) as T
}

export const api = {
  getLanguages: () => getJSON<Language[]>('/api/languages'),
  getTenses: (lang: LanguageCode) => getJSON<TenseSummary[]>(`/api/tenses?lang=${lang}`),
  getTheory: (lang: LanguageCode, tenseId: string) =>
    getJSON<Theory>(`/api/theory/${encodeURIComponent(tenseId)}?lang=${lang}`),
  getVerbs: (lang: LanguageCode) => getJSON<VerbSummary[]>(`/api/verbs?lang=${lang}`),
  getExercises: (lang: LanguageCode, tense: string, type: string, count: number) =>
    getJSON<Exercise[]>(
      `/api/exercises?lang=${lang}&tense=${encodeURIComponent(tense)}&type=${encodeURIComponent(type)}&count=${count}`,
    ),
  getVocabCategories: (lang: LanguageCode) => getJSON<VocabCategory[]>(`/api/vocab-categories?lang=${lang}`),
  getVocabulary: (lang: LanguageCode, category?: string) =>
    getJSON<VocabWord[]>(
      `/api/vocabulary?lang=${lang}${category ? `&category=${encodeURIComponent(category)}` : ''}`,
    ),

  // Reading. Note that these can legitimately 404 for a language that has
  // grammar and vocabulary but no corpus yet — reading ships for a subset.
  getTexts: (lang: LanguageCode) => getJSON<ReadingText[]>(`/api/texts?lang=${lang}`),
  getTextSection: (lang: LanguageCode, textId: string, section: number) =>
    getJSON<ReadingSection>(
      `/api/texts/${encodeURIComponent(textId)}?lang=${lang}&section=${section}`,
    ),
  /** Resolves an arbitrary surface form or multi-word phrase. Used for
   *  drag-selected spans, which are not single tokens and so cannot come
   *  from the section lexicon. Returns [] when nothing matches — that is a
   *  normal outcome, not an error. */
  lookup: (lang: LanguageCode, q: string) =>
    getJSON<ReadingEntry[]>(`/api/lookup?lang=${lang}&q=${encodeURIComponent(q)}`),

  /** Uploads an EPUB and returns it as a readable text. The only
   *  non-GET call in the app: a book is content rather than user state,
   *  and has to be tokenized by the same engine that serves the bundled
   *  corpus. */
  importText: (lang: LanguageCode, file: File) => {
    const body = new FormData()
    body.append('file', file)
    return request<ReadingText>(`/api/texts/import?lang=${lang}`, { method: 'POST', body })
  },

  deleteText: (lang: LanguageCode, id: string) =>
    request<void>(`/api/texts/${encodeURIComponent(id)}?lang=${lang}`, { method: 'DELETE' }),
}

export { ApiError }
