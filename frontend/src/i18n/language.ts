import { ref, computed, watch } from 'vue'
import type { LanguageCode, Language } from '../types/api'
import { strings } from './strings'
import { api, ApiError } from '../api/client'
import { navigate } from '../router/router'

const STORAGE_KEY = 'sermokey:lang'

function loadInitialLang(): LanguageCode {
  const stored = localStorage.getItem(STORAGE_KEY)
  return stored === 'es' ? 'es' : 'fr'
}

// Module-level singletons (same pattern as router.ts's currentRoute): one
// reactive language selection shared by the whole app.
const lang = ref<LanguageCode>(loadInitialLang())
const languages = ref<Language[]>([
  { code: 'fr', name: 'Français' },
  { code: 'es', name: 'Español' },
])

watch(lang, (value) => {
  localStorage.setItem(STORAGE_KEY, value)
})

// Best-effort refresh from the backend (in case a language is ever added
// without a frontend redeploy); the hardcoded fallback above means the
// toggle still renders correctly even if this fails or the backend isn't
// up yet.
api
  .getLanguages()
  .then((fetched) => {
    if (fetched.length > 0) languages.value = fetched
  })
  .catch((e) => {
    if (e instanceof ApiError) {
      console.warn('Could not fetch /api/languages, using built-in defaults:', e.message)
    }
  })

/** Reactive UI chrome strings for the currently selected language. */
export const t = computed(() => strings[lang.value])

export function useLanguage() {
  return { lang, languages }
}

export function setLanguage(code: LanguageCode) {
  if (code === lang.value) return
  lang.value = code
  // Tense IDs are not shared between languages (French "present" vs
  // Spanish "presente"), so any in-progress theory/quiz route would be
  // meaningless after switching — go back to a safe, valid screen instead
  // of guessing at a translation.
  navigate({ name: 'home' })
}

export { format } from './strings'
