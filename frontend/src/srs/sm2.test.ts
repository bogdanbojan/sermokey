import { describe, expect, it } from 'vitest'
import {
  LEECH_THRESHOLD,
  MAX_EASE_FACTOR,
  MAX_INTERVAL_DAYS,
  MIN_EASE_FACTOR,
  bucketOf,
  isDue,
  isLeech,
  newCardState,
  normalizeState,
  previewIntervals,
  review,
  toISODate,
  type CardState,
  type Grade,
} from './sm2'

const at = (iso: string, h = 12) => {
  const [y, m, d] = iso.split('-').map(Number)
  return new Date(y!, m! - 1, d!, h)
}

/** Grades a fresh card through a sequence, one day apart. */
function run(grades: Grade[], start = '2026-01-01'): CardState {
  let state = newCardState(at(start))
  let day = at(start)
  for (const g of grades) {
    state = review(state, g, day)
    day = new Date(day.getTime() + 86_400_000)
  }
  return state
}

describe('toISODate', () => {
  it('names the local day, not the UTC day', () => {
    // 23:30 local on the 1st is the 2nd in UTC for anything east of
    // Greenwich and still the 1st west of it. Either way the learner is
    // looking at the 1st, so that is the day the card belongs to.
    expect(toISODate(at('2026-03-01', 23))).toBe('2026-03-01')
    expect(toISODate(at('2026-03-01', 0))).toBe('2026-03-01')
  })

  it('zero-pads month and day', () => {
    expect(toISODate(at('2026-07-04'))).toBe('2026-07-04')
  })
})

describe('isDue', () => {
  it('is true on the due day and false before it', () => {
    const s = review(newCardState(at('2026-01-01')), 'good', at('2026-01-01'))
    expect(s.dueDate).toBe('2026-01-02')
    expect(isDue(s, at('2026-01-01'))).toBe(false)
    expect(isDue(s, at('2026-01-02'))).toBe(true)
    expect(isDue(s, at('2026-01-03'))).toBe(true)
  })

  it('does not fall due in the evening of the day before', () => {
    // The regression this guards: with a UTC-named day, 20:00 local on the
    // 1st reported the 2nd for any negative-offset zone, pulling the whole
    // of tomorrow's queue forward every evening.
    const s = review(newCardState(at('2026-01-01')), 'good', at('2026-01-01'))
    expect(isDue(s, at('2026-01-01', 20))).toBe(false)
    expect(isDue(s, at('2026-01-01', 23))).toBe(false)
  })

  it('treats a card with a corrupt due date as due rather than retiring it', () => {
    const broken = { ...newCardState(), dueDate: undefined } as unknown as CardState
    expect(isDue(broken, at('2026-01-01'))).toBe(true)
  })

  it('withholds suspended cards', () => {
    const s = { ...newCardState(at('2026-01-01')), suspended: true }
    expect(isDue(s, at('2026-06-01'))).toBe(false)
  })
})

describe('the published SM-2 ladder', () => {
  it('runs 1 then 6 then interval * ease for "good"', () => {
    expect(run(['good']).intervalDays).toBe(1)
    expect(run(['good', 'good']).intervalDays).toBe(6)
    // 6 * 2.5 = 15, answered on schedule so no lateness credit.
    expect(run(['good', 'good', 'good']).intervalDays).toBe(15)
  })

  it('leaves the ease factor alone on "good"', () => {
    expect(run(['good', 'good', 'good']).easeFactor).toBeCloseTo(2.5)
  })

  it('erodes ease on "hard" and raises it on "easy"', () => {
    expect(run(['hard']).easeFactor).toBeCloseTo(2.35)
    expect(run(['easy']).easeFactor).toBeCloseTo(2.65)
  })

  it('drops ease by 0.2 on a lapse', () => {
    expect(run(['good', 'good', 'good', 'again']).easeFactor).toBeCloseTo(2.3)
  })
})

describe('bounds', () => {
  it('never lets the ease factor fall below the documented floor', () => {
    const s = run(Array<Grade>(40).fill('again'))
    expect(s.easeFactor).toBe(MIN_EASE_FACTOR)
  })

  it('caps the ease factor so growth cannot run away', () => {
    const s = run(Array<Grade>(40).fill('easy'))
    expect(s.easeFactor).toBeLessThanOrEqual(MAX_EASE_FACTOR)
  })

  it('caps the interval and keeps the due date a valid date', () => {
    // Before the cap this overflowed the range of a JS Date at around the
    // seventeenth "easy": addDays produced an Invalid Date and formatting
    // it threw a RangeError straight out of the grading handler.
    const s = run(Array<Grade>(40).fill('easy'))
    expect(s.intervalDays).toBeLessThanOrEqual(MAX_INTERVAL_DAYS)
    expect(s.dueDate).toMatch(/^\d{4}-\d{2}-\d{2}$/)
  })

  it('survives 200 random grades without producing a non-finite state', () => {
    const grades: Grade[] = ['again', 'hard', 'good', 'easy']
    let state = newCardState(at('2026-01-01'))
    let day = at('2026-01-01')
    for (let i = 0; i < 200; i++) {
      state = review(state, grades[i % 4]!, day)
      expect(Number.isFinite(state.intervalDays)).toBe(true)
      expect(Number.isFinite(state.easeFactor)).toBe(true)
      expect(state.dueDate).toMatch(/^\d{4}-\d{2}-\d{2}$/)
      day = new Date(day.getTime() + 86_400_000)
    }
  })

  it('always stores a whole number of days', () => {
    // "hard" used to multiply by 1.2 without rounding, so intervals
    // accumulated a fractional tail that every consumer then had to guess
    // about.
    const s = run(['good', 'good', 'hard', 'hard', 'hard'])
    expect(Number.isInteger(s.intervalDays)).toBe(true)
  })
})

describe('lapses and leeches', () => {
  it('does not count failing a card that is still being learned', () => {
    expect(run(['again']).lapses).toBe(0)
    expect(run(['good', 'again']).lapses).toBe(0)
  })

  it('counts failing a card that had graduated', () => {
    expect(run(['good', 'good', 'again']).lapses).toBe(1)
  })

  it('suspends a card once it has been failed too often', () => {
    let state = newCardState(at('2026-01-01'))
    let day = at('2026-01-01')
    for (let i = 0; i < LEECH_THRESHOLD; i++) {
      // Graduate, then fail: one lapse per cycle.
      for (const g of ['good', 'good', 'again'] as Grade[]) {
        state = review(state, g, day)
        day = new Date(day.getTime() + 86_400_000)
      }
    }
    expect(state.lapses).toBe(LEECH_THRESHOLD)
    expect(isLeech(state)).toBe(true)
    expect(state.suspended).toBe(true)
    expect(isDue(state, at('2027-01-01'))).toBe(false)
  })
})

describe('lateness credit', () => {
  it('schedules an overdue card from the retention it actually demonstrated', () => {
    // A 6-day card answered 60 days late. The old formula multiplied the
    // scheduled 6 days and produced 15, ignoring that the learner had just
    // proven two months of retention.
    const onLadder = run(['good', 'good'])
    expect(onLadder.intervalDays).toBe(6)

    const late = review(onLadder, 'good', at('2026-04-01'))
    const onTime = review(onLadder, 'good', at('2026-01-09'))
    expect(late.intervalDays).toBeGreaterThan(onTime.intervalDays)
  })

  it('gives no credit to a card that was answered early', () => {
    const onLadder = run(['good', 'good'])
    const early = review(onLadder, 'good', at('2026-01-04'))
    expect(early.intervalDays).toBe(15)
  })

  it('gives no credit to a card still in the learning ladder', () => {
    const fresh = newCardState(at('2026-01-01'))
    const late = review(fresh, 'good', at('2026-06-01'))
    expect(late.intervalDays).toBe(1)
  })
})

describe('bucketOf', () => {
  it('reports a never-seen card as new', () => {
    expect(bucketOf(undefined)).toBe('new')
    expect(bucketOf(newCardState())).toBe('new')
  })

  it('reports a lapsed card as learning, not new', () => {
    // The regression this guards: a lapse resets `repetitions` to 0, and
    // the bucket was derived from that, so a card with forty reviews
    // behind it was counted among the untouched ones the moment it was
    // failed once. Every "N new" figure in the UI was wrong.
    const lapsed = run(['good', 'good', 'good', 'again'])
    expect(lapsed.repetitions).toBe(0)
    expect(lapsed.phase).toBe('relearning')
    expect(bucketOf(lapsed)).toBe('learning')
  })

  it('reports a long-interval card as mastered', () => {
    const s = run(['easy', 'easy', 'easy', 'easy'])
    expect(s.intervalDays).toBeGreaterThanOrEqual(30)
    expect(bucketOf(s)).toBe('mastered')
  })
})

describe('normalizeState', () => {
  it('reconstructs phase and lapses for a card stored before they existed', () => {
    const v1 = {
      repetitions: 0,
      easeFactor: 2.3,
      intervalDays: 10,
      dueDate: '2026-01-05',
    } as unknown as CardState

    const migrated = normalizeState(v1)
    // Repetitions zero but a real interval behind it: version 1 only
    // produced that by failing a mature card.
    expect(migrated.phase).toBe('relearning')
    expect(migrated.lapses).toBe(1)
  })

  it('reads an untouched version 1 card as new', () => {
    const v1 = {
      repetitions: 0,
      easeFactor: 2.5,
      intervalDays: 0,
      dueDate: '2026-01-01',
    } as unknown as CardState
    expect(normalizeState(v1).phase).toBe('new')
    expect(normalizeState(v1).lapses).toBe(0)
  })

  it('reads a mature version 1 card as review', () => {
    const v1 = {
      repetitions: 5,
      easeFactor: 2.5,
      intervalDays: 40,
      dueDate: '2026-01-01',
    } as unknown as CardState
    expect(normalizeState(v1).phase).toBe('review')
  })

  it('repairs a non-finite ease factor rather than propagating it', () => {
    const broken = { ...newCardState(), easeFactor: NaN }
    expect(Number.isFinite(normalizeState(broken).easeFactor)).toBe(true)
  })
})

describe('previewIntervals', () => {
  it('orders the four grades by how long they defer the card', () => {
    const s = run(['good', 'good'])
    const p = previewIntervals(s, at('2026-01-09'))
    expect(p.again).toBeLessThanOrEqual(p.hard)
    expect(p.hard).toBeLessThan(p.good)
    expect(p.good).toBeLessThan(p.easy)
  })
})

describe('DST', () => {
  it('advances by whole days across a spring-forward boundary', () => {
    // US DST begins 2026-03-08. A naive millisecond difference puts 23h
    // between the two local midnights and rounds the interval down.
    const s = review(newCardState(at('2026-03-06')), 'good', at('2026-03-06'))
    expect(s.dueDate).toBe('2026-03-07')
    const next = review({ ...s, repetitions: 1 }, 'good', at('2026-03-07'))
    expect(next.dueDate).toBe('2026-03-13')
  })
})
