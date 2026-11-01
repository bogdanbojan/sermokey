import { ref, computed, watch } from 'vue'
import type { LanguageCode, Language } from '../types/api'
import { plural, strings, type PluralForms, type UiLanguageCode } from './strings'
import { api, ApiError } from '../api/client'
import { navigate } from '../router/router'
import { read, write } from '../storage'

const STORAGE_KEY = 'sermokey:lang'
const UI_LANG_KEY = 'sermokey:ui-lang'

/** Study languages the frontend ships translations and content for. */
const BUILT_IN: Language[] = [
  { code: 'fr', name: 'Français' },
  { code: 'es', name: 'Español' },
  { code: 'de', name: 'Deutsch' },
]

function isLanguageCode(u: unknown): u is LanguageCode {
  return u === 'fr' || u === 'es' || u === 'de'
}

function parseLang(u: unknown): LanguageCode | null {
  return isLanguageCode(u) ? u : null
}

/** 'auto' means "whatever is being studied", which is the original behaviour. */
export type UiLangPreference = UiLanguageCode | 'auto'

function parseUiLang(u: unknown): UiLangPreference | null {
  if (u === 'auto' || u === 'en') return u
  return isLanguageCode(u) ? u : null
}

// Module-level singletons (same pattern as router.ts's currentRoute): one
// reactive language selection shared by the whole app.
const lang = ref<LanguageCode>(read(STORAGE_KEY, parseLang, 'fr'))
const languages = ref<Language[]>(BUILT_IN)

/**
 * Which language the *chrome* is rendered in.
 *
 * Defaults to 'auto', which reproduces the original immersive behaviour
 * exactly: the interface follows whatever is being studied. A learner who
 * cannot yet read that language can pin it to English instead, without
 * giving up the content language they came for.
 */
const uiLangPreference = ref<UiLangPreference>(read(UI_LANG_KEY, parseUiLang, 'auto'))

export const uiLang = computed<UiLanguageCode>(() =>
  uiLangPreference.value === 'auto' ? lang.value : uiLangPreference.value,
)

watch(lang, (value) => write(STORAGE_KEY, value))
watch(uiLangPreference, (value) => write(UI_LANG_KEY, value))

// The document language drives screen-reader pronunciation and, for the
// reading surface, which hyphenation dictionary the browser applies. It was
// hardcoded to French in index.html, so German prose was being hyphenated
// by French rules and announced in a French voice.
watch(
  uiLang,
  (value) => {
    document.documentElement.lang = value
  },
  { immediate: true },
)

// Best-effort refresh from the backend (in case a language is ever added
// without a frontend redeploy); the hardcoded fallback above means the
// toggle still renders correctly even if this fails or the backend isn't
// up yet.
api
  .getLanguages()
  .then((fetched) => {
    // Only languages this build can actually render. Without this filter
    // the feature defeats itself: a backend that adds "it" produces a
    // button that sets `lang` to a code with no strings entry, `t` becomes
    // undefined, the next render throws, and the bad code is persisted —
    // so the app is blank on every subsequent load too.
    const usable = fetched.filter((l) => isLanguageCode(l.code))
    if (usable.length > 0) languages.value = usable
  })
  .catch((e) => {
    if (e instanceof ApiError) {
      console.warn('Could not fetch /api/languages, using built-in defaults:', e.message)
    }
  })

/** Reactive UI chrome strings for the currently selected interface language. */
export const t = computed(() => strings[uiLang.value] ?? strings.en)

export function useLanguage() {
  return { lang, languages, uiLang, uiLangPreference }
}

export function setLanguage(code: LanguageCode) {
  if (code === lang.value) return
  // Tense IDs are not shared between languages (French "present" vs
  // Spanish "presente"), so any in-progress theory/quiz route would be
  // meaningless after switching — go back to a safe, valid screen instead
  // of guessing at a translation.
  //
  // Navigating *first* matters: `hashchange` is delivered as a task, not a
  // microtask, so a pre-flush watcher on `lang` would otherwise refetch the
  // current route under the new language and flash a 404 before the route
  // actually changed.
  navigate({ name: 'home' })
  lang.value = code
}

export function setUiLangPreference(pref: UiLangPreference) {
  uiLangPreference.value = pref
}

export { format } from './strings'

/** Plural-aware formatting bound to the current interface language. */
export function pluralize(forms: PluralForms, count: number): string {
  return plural(forms, count, uiLang.value)
}
export type { UiLanguageCode }
