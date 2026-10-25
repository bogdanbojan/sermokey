// SM-2 spaced-repetition scheduling, the algorithm behind SuperMemo (Wozniak,
// 1987) and — in this same 4-button "Again/Hard/Good/Easy" form — Anki,
// still the most evidence-backed tool for large-vocabulary retention
// decades later. This is a from-scratch TypeScript implementation of the
// published algorithm; the formulas below are the standard, publicly
// documented SM-2 update rules (not copied from any specific codebase).
//
// Core idea: every reviewed card has an `interval` (days until next due)
// and an `easeFactor` (how quickly intervals grow for this card, starting
// at 2.5 and adjusted up/down based on how easy each review felt). A
// correct, easy review grows the interval a lot; a lapse resets it near to
// zero and nudges the ease factor down — this is what produces the
// "spaced" in spaced repetition: well-known cards are shown rarely,
// struggling cards are shown often, automatically.

export type Grade = 'again' | 'hard' | 'good' | 'easy'

export interface CardState {
  /** Consecutive successful reviews (resets to 0 on "again"). */
  repetitions: number
  /** How quickly this card's interval grows; 1.3 is SM-2's documented floor. */
  easeFactor: number
  /** Days until the next review is due. */
  intervalDays: number
  /** ISO date string (yyyy-mm-dd) this card is next due. */
  dueDate: string
  /** ISO date string of the last review, or undefined if never reviewed. */
  lastReviewed?: string
}

export const MIN_EASE_FACTOR = 1.3
const DEFAULT_EASE_FACTOR = 2.5

/** A fresh card that has never been reviewed, due immediately. */
export function newCardState(today: Date = new Date()): CardState {
  return {
    repetitions: 0,
    easeFactor: DEFAULT_EASE_FACTOR,
    intervalDays: 0,
    dueDate: toISODate(today),
  }
}

export function toISODate(d: Date): string {
  return d.toISOString().slice(0, 10)
}

function addDays(date: Date, days: number): Date {
  const result = new Date(date)
  result.setDate(result.getDate() + Math.round(days))
  return result
}

/**
 * Applies one review's grade to a card's state, returning the new state.
 * Does not mutate the input.
 */
export function review(state: CardState, grade: Grade, today: Date = new Date()): CardState {
  let { repetitions, easeFactor, intervalDays } = state

  if (grade === 'again') {
    // A lapse: start over, and the card is penalized slightly so it grows
    // more cautiously next time around.
    repetitions = 0
    intervalDays = 1
    easeFactor = Math.max(MIN_EASE_FACTOR, easeFactor - 0.2)
  } else {
    repetitions += 1

    if (grade === 'hard') {
      // Still correct, but effortful: grow the interval only a little, and
      // erode the ease factor a bit so future intervals grow more slowly.
      intervalDays = repetitions === 1 ? 1 : Math.max(1, intervalDays * 1.2)
      easeFactor = Math.max(MIN_EASE_FACTOR, easeFactor - 0.15)
    } else if (grade === 'good') {
      // The classic SM-2 progression: 1 day, then 6 days, then
      // previousInterval * easeFactor — ease factor unchanged.
      if (repetitions === 1) intervalDays = 1
      else if (repetitions === 2) intervalDays = 6
      else intervalDays = Math.round(intervalDays * easeFactor)
    } else {
      // "easy": same progression as "good" but with a bonus multiplier and
      // a small ease-factor boost, since the card is clearly well-known.
      if (repetitions === 1) intervalDays = 4
      else if (repetitions === 2) intervalDays = 8
      else intervalDays = Math.round(intervalDays * easeFactor * 1.3)
      easeFactor = easeFactor + 0.15
    }
  }

  return {
    repetitions,
    easeFactor,
    intervalDays,
    dueDate: toISODate(addDays(today, intervalDays)),
    lastReviewed: toISODate(today),
  }
}

/** True if a card's dueDate is today or earlier. */
export function isDue(state: CardState, today: Date = new Date()): boolean {
  return state.dueDate <= toISODate(today)
}

/** Rough bucket for progress stats: new / learning / review / mastered. */
export type CardBucket = 'new' | 'learning' | 'review' | 'mastered'

export function bucketOf(state: CardState | undefined): CardBucket {
  if (!state || state.repetitions === 0) return 'new'
  if (state.repetitions < 2) return 'learning'
  if (state.intervalDays >= 30) return 'mastered'
  return 'review'
}
