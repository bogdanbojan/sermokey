import { ref } from 'vue'

export type Route =
  | { name: 'home' }
  | { name: 'theory'; tense: string }
  | { name: 'quiz-setup'; tense: string }
  | { name: 'quiz'; tense: string; type: string; count: number }
  | { name: 'flashcards'; category: string }

function parseHash(hash: string): Route {
  const clean = hash.replace(/^#\/?/, '')
  const parts = clean.split('/').filter(Boolean).map(decodeURIComponent)

  if (parts[0] === 'theory' && parts[1]) return { name: 'theory', tense: parts[1] }
  if (parts[0] === 'quiz-setup' && parts[1]) return { name: 'quiz-setup', tense: parts[1] }
  if (parts[0] === 'quiz' && parts[1] && parts[2] && parts[3]) {
    const count = Number(parts[3])
    return { name: 'quiz', tense: parts[1], type: parts[2], count: Number.isFinite(count) ? count : 10 }
  }
  if (parts[0] === 'flashcards') return { name: 'flashcards', category: parts[1] ?? '' }
  return { name: 'home' }
}

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
    case 'flashcards':
      return `#/flashcards/${encodeURIComponent(route.category)}`
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
