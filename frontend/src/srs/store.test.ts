import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { Cardlike } from './store'
import { newCardState, review, toISODate } from './sm2'

const LANG = 'fr' as const
const KEY = 'sermokey:srs:fr'

// The per-language deck is a module-level singleton that loads on first
// access. That is right for the app — one reactive deck shared by every
// view — but it means storage has to be seeded *before* the module is
// imported, so each test gets a fresh module graph.
type Store = typeof import('./store')

async function freshStore(): Promise<Store> {
  vi.resetModules()
  return import('./store')
}

/** Seeds storage in the current envelope format. Call before `freshStore`. */
function seed(deck: Record<string, unknown>) {
  localStorage.setItem(KEY, JSON.stringify({ v: 2, d: deck }))
}

/** Convenience: fresh module, then the French deck. */
async function freshDeck() {
  return (await freshStore()).useSrsDeck(LANG)
}

const ids = (n: number): Cardlike[] => Array.from({ length: n }, (_, i) => ({ id: `w${i}` }))

const reviewCard = (over: Record<string, unknown> = {}) => ({
  ...newCardState(),
  repetitions: 3,
  intervalDays: 10,
  phase: 'review',
  lapses: 0,
  ...over,
})

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

describe('cardID', () => {
  it('keeps the two directions apart', async () => {
    const { cardID } = await freshStore()
    expect(cardID('maison|noms', 'recognition')).not.toBe(cardID('maison|noms', 'production'))
  })
})

describe('sessionQueue', () => {
  it('introduces new cards up to the limit', async () => {
    const d = await freshDeck()
    expect(d.sessionQueue(ids(100), { newCardLimit: 15 })).toHaveLength(15)
  })

  it('returns scheduled-but-not-due cards when asked to', async () => {
    // The regression this guards: "study anyway" raised the *new card*
    // limit, but every card already had state, so the queue came back
    // empty while the phase was set to 'studying'. No template branch
    // matched a studying phase with no current card, and the screen went
    // blank.
    const future = toISODate(new Date(Date.now() + 30 * 86_400_000))
    seed({ a: reviewCard({ dueDate: future }), b: reviewCard({ dueDate: future }) })
    const d = await freshDeck()
    const cards = [{ id: 'a' }, { id: 'b' }]

    expect(d.sessionQueue(cards)).toHaveLength(0)
    expect(d.sessionQueue(cards, { includeNotDue: true })).toHaveLength(2)
  })

  it('withholds suspended leeches', async () => {
    seed({
      a: reviewCard({ dueDate: '2000-01-01', suspended: true, lapses: 8 }),
      b: reviewCard({ dueDate: '2000-01-01' }),
    })
    const d = await freshDeck()
    expect(d.sessionQueue([{ id: 'a' }, { id: 'b' }]).map((c) => c.id)).toEqual(['b'])
  })

  it('caps the review pile so a long absence is not a wall of cards', async () => {
    const deck: Record<string, unknown> = {}
    for (let i = 0; i < 400; i++) deck[`w${i}`] = reviewCard({ dueDate: '2000-01-01' })
    seed(deck)
    const d = await freshDeck()
    expect(d.sessionQueue(ids(400), { reviewLimit: 50, newCardLimit: 0 })).toHaveLength(50)
  })

  it('orders due cards reproducibly', async () => {
    // A comparator that never returns 0 is not a valid ordering, and the
    // common case here is a tie: due dates are day-granular strings, so a
    // backlog shares one value and came out shuffled differently per call.
    const deck: Record<string, unknown> = {}
    for (let i = 0; i < 20; i++) deck[`w${i}`] = reviewCard({ dueDate: '2000-01-01' })
    seed(deck)
    const d = await freshDeck()
    const first = d.sessionQueue(ids(20), { newCardLimit: 0 }).map((c) => c.id)
    const second = d.sessionQueue(ids(20), { newCardLimit: 0 }).map((c) => c.id)
    expect(first).toEqual(second)
  })

  it('spreads new cards through the session rather than queueing them last', async () => {
    const deck: Record<string, unknown> = {}
    for (let i = 0; i < 10; i++) deck[`r${i}`] = reviewCard({ dueDate: '2000-01-01' })
    seed(deck)
    const d = await freshDeck()
    const cards = [...Array.from({ length: 10 }, (_, i) => ({ id: `r${i}` })), ...ids(4)]
    const q = d.sessionQueue(cards, { newCardLimit: 4 })

    const firstNew = q.findIndex((c) => c.id.startsWith('w'))
    expect(firstNew).toBeGreaterThan(0)
    expect(firstNew).toBeLessThan(q.length - 1)
  })
})

describe('recordReview', () => {
  it('persists and reports success', async () => {
    const d = await freshDeck()
    const { saved, previous } = d.recordReview('a', 'good')
    expect(saved).toBe('ok')
    expect(previous).toBeUndefined()
    expect(d.stateFor('a')?.repetitions).toBe(1)
  })

  it('reports a failed write instead of pretending the review landed', async () => {
    const d = await freshDeck()
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new DOMException('full', 'QuotaExceededError')
    })
    expect(d.recordReview('a', 'good').saved).toBe('quota')
  })

  it('hands back the prior state so a grade can be undone', async () => {
    const d = await freshDeck()
    d.recordReview('a', 'good')
    const before = d.stateFor('a')

    const { previous } = d.recordReview('a', 'again')
    expect(previous).toEqual(before)

    d.restore('a', previous)
    expect(d.stateFor('a')).toEqual(before)
  })

  it('removes the card entirely when undoing its first ever review', async () => {
    const d = await freshDeck()
    const { previous } = d.recordReview('a', 'good')
    d.restore('a', previous)
    expect(d.stateFor('a')).toBeUndefined()
  })
})

describe('unsuspend', () => {
  it('returns a leech to the rotation', async () => {
    seed({ a: reviewCard({ dueDate: '2000-01-01', suspended: true, lapses: 8 }) })
    const d = await freshDeck()
    expect(d.sessionQueue([{ id: 'a' }])).toHaveLength(0)

    d.unsuspend('a')
    expect(d.sessionQueue([{ id: 'a' }])).toHaveLength(1)
  })
})

describe('stats', () => {
  it('counts a lapsed card as learning, not as new', async () => {
    let s = newCardState(new Date(2026, 0, 1))
    for (const g of ['good', 'good', 'good', 'again'] as const) s = review(s, g, new Date(2026, 0, 1))
    seed({ a: s })
    const d = await freshDeck()

    const out = d.stats([{ id: 'a' }])
    expect(out.new).toBe(0)
    expect(out.learning).toBe(1)
  })

  it('counts leeches separately', async () => {
    seed({ a: reviewCard({ lapses: 8, suspended: true }) })
    const d = await freshDeck()
    expect(d.stats([{ id: 'a' }]).leeches).toBe(1)
  })
})

describe('corrupt storage', () => {
  it('does not crash on the literal string "null"', async () => {
    // Previously: JSON.parse('null') returned null, the `as Deck` cast hid
    // it, and the first `deck.value[id]` threw on every page load.
    localStorage.setItem(KEY, 'null')
    const d = await freshDeck()
    expect(() => d.stats([{ id: 'a' }])).not.toThrow()
  })

  it('drops only the corrupt cards, keeping the rest of the deck', async () => {
    seed({ good: reviewCard(), bad: 'not an object', alsoBad: { nonsense: true } })
    const d = await freshDeck()
    expect(d.stateFor('good')).toBeDefined()
    expect(d.stateFor('bad')).toBeUndefined()
    expect(d.stateFor('alsoBad')).toBeUndefined()
  })

  it('reads a deck written before envelopes and lapse tracking existed', async () => {
    localStorage.setItem(
      KEY,
      JSON.stringify({
        legacy: { repetitions: 0, easeFactor: 2.3, intervalDays: 10, dueDate: '2026-01-05' },
      }),
    )
    const d = await freshDeck()
    const state = d.stateFor('legacy')
    expect(state?.phase).toBe('relearning')
    expect(state?.lapses).toBe(1)
  })
})
