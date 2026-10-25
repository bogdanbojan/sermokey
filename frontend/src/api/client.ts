import type {
  Exercise,
  Theory,
  TenseSummary,
  VerbSummary,
  Language,
  LanguageCode,
  VocabCategory,
  VocabWord,
} from '../types/api'

const BASE_URL = (import.meta.env.VITE_API_BASE_URL as string | undefined) ?? 'http://localhost:8080'

class ApiError extends Error {}

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
}

export { ApiError }
