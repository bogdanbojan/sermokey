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
//
// Three departures from naive SM-2, each for a reason the published
// algorithm leaves open:
//
//  - **Days are local days.** SM-2 says "days" without saying whose. Using
//    UTC to name the day while using local arithmetic to advance it makes
//    cards come due in the evening for anyone west of Greenwich, which
//    quietly compresses every interval in the deck.
//
//  - **Lateness is credited.** SM-2 assumes a review happens on the day it
//    was scheduled. When a card scheduled for 10 days is answered after 90,
//    the learner has demonstrated 90 days of retention, not 10, and the
//    next interval should reflect that. This is Anki's `delay` term.
//
//  - **Lapses are counted.** A card failed over and over is a "leech": it
//    is costing more review time than it is worth and usually means the
//    card itself is bad. Counting lapses lets the deck surface it instead
//    of cycling it forever.

export type Grade = 'again' | 'hard' | 'good' | 'easy'

/**
 * Where a card sits in the scheduling ladder.
 *
 * `relearning` exists specifically so a lapsed card is distinguishable from
 * a never-seen one. Collapsing the two (by testing `repetitions === 0`,
 * which a lapse resets) reported every failed card as brand new, so a deck
 * of 500 mature cards with 40 recent lapses claimed 40 untouched words.
 */
export type CardPhase = 'new' | 'learning' | 'review' | 'relearning'

export interface CardState {
  /** Consecutive successful reviews (resets to 0 on "again"). */
  repetitions: number
  /** How quickly this card's interval grows; 1.3 is SM-2's documented floor. */
  easeFactor: number
  /** Days until the next review is due. */
  intervalDays: number
  /** ISO date string (yyyy-mm-dd, local) this card is next due. */
  dueDate: string
  /** ISO date string of the last review, or undefined if never reviewed. */
  lastReviewed?: string
  /** How many times a *mature* card has been failed. Drives leech detection. */
  lapses: number
  /** Scheduling phase; see CardPhase. */
  phase: CardPhase
  /** Set once `lapses` reaches the leech threshold. Suspended cards are
   *  withheld from review queues until the learner acts on them. */
  suspended?: boolean
}

export const MIN_EASE_FACTOR = 1.3
/** Without a ceiling, each "easy" raises the factor *and* multiplies by it.
 *  Growth is super-exponential and the interval exceeds the range of a JS
 *  Date in under twenty reviews, at which point scheduling throws. */
export const MAX_EASE_FACTOR = 5.0
const DEFAULT_EASE_FACTOR = 2.5

/** ~100 years. Anki uses the same cap, for the same reason. */
export const MAX_INTERVAL_DAYS = 36500
/** Interval at or beyond which a card counts as "mastered" for stats. */
export const MASTERED_INTERVAL_DAYS = 30
/** Lapses before a card is flagged as a leech and suspended. Anki's default. */
export const LEECH_THRESHOLD = 8

const HARD_MULTIPLIER = 1.2
const EASY_BONUS = 1.3

function clampEase(ef: number): number {
  if (!Number.isFinite(ef)) return DEFAULT_EASE_FACTOR
  return Math.min(MAX_EASE_FACTOR, Math.max(MIN_EASE_FACTOR, ef))
}

function clampInterval(days: number): number {
  if (!Number.isFinite(days)) return 1
  return Math.min(MAX_INTERVAL_DAYS, Math.max(1, Math.round(days)))
}

/** A fresh card that has never been reviewed, due immediately. */
export function newCardState(today: Date = new Date()): CardState {
  return {
    repetitions: 0,
    easeFactor: DEFAULT_EASE_FACTOR,
    intervalDays: 0,
    dueDate: toISODate(today),
    lapses: 0,
    phase: 'new',
  }
}

/**
 * Formats a date as yyyy-mm-dd in the *local* calendar.
 *
 * Deliberately not `toISOString().slice(0, 10)`, which names the UTC day.
 * Paired with local `setDate` arithmetic that produced a day boundary that
 * belonged to neither calendar: at UTC-5, 20:00 on the 1st already reports
 * the 2nd, so everything scheduled for tomorrow fell due this evening.
 */
export function toISODate(d: Date): string {
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

/** Parses yyyy-mm-dd as local midnight, or null if malformed. */
function fromISODate(s: string): Date | null {
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(s)
  if (!m) return null
  const d = new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3]))
  return Number.isNaN(d.getTime()) ? null : d
}

function localMidnight(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate())
}

/**
 * Whole days from an ISO date to a Date, both taken at local midnight.
 *
 * Rounding absorbs the ±1h that a DST transition puts between the two
 * midnights, which would otherwise make intervals drift by a day twice a
 * year for everyone observing it.
 */
function daysBetween(fromISO: string, to: Date): number {
  const from = fromISODate(fromISO)
  if (!from) return 0
  return Math.round((localMidnight(to).getTime() - from.getTime()) / 86_400_000)
}

function addDays(date: Date, days: number): Date {
  const result = localMidnight(date)
  result.setDate(result.getDate() + Math.round(days))
  return result
}

/**
 * Upgrades a card read from storage to the current shape.
 *
 * Cards written before lapse tracking existed have neither `phase` nor
 * `lapses`. Both are reconstructed from what version 1 did record: it reset
 * `repetitions` to 0 on a lapse but left the interval behind, so a card
 * with no repetitions and a non-zero interval has certainly been failed at
 * least once, while one with neither has never been seen.
 */
export function normalizeState(state: CardState): CardState {
  const repetitions = Number.isFinite(state.repetitions) ? state.repetitions : 0
  const intervalDays = Number.isFinite(state.intervalDays) ? state.intervalDays : 0
  const easeFactor = clampEase(state.easeFactor)

  // Already current *and* numerically sound: nothing to do. The numeric
  // check has to come first, because a NaN ease factor is just as fatal in
  // a card that carries a phase as in one that does not.
  if (state.phase && typeof state.lapses === 'number' && easeFactor === state.easeFactor) {
    return state
  }

  let phase: CardPhase
  let lapses = 0
  if (repetitions === 0 && intervalDays === 0) {
    phase = 'new'
  } else if (repetitions === 0) {
    phase = 'relearning'
    lapses = 1
  } else if (repetitions < 2) {
    phase = 'learning'
  } else {
    phase = 'review'
  }

  return {
    ...state,
    repetitions,
    intervalDays,
    easeFactor,
    lapses: typeof state.lapses === 'number' ? state.lapses : lapses,
    phase: state.phase ?? phase,
  }
}

/**
 * Applies one review's grade to a card's state, returning the new state.
 * Does not mutate the input.
 */
export function review(input: CardState, grade: Grade, today: Date = new Date()): CardState {
  const state = normalizeState(input)
  let { repetitions, easeFactor, intervalDays, lapses, phase } = state

  // How late this review is. Only meaningful for a card that was actually
  // sitting at a mature interval; a card still in the learning ladder has
  // no retention claim to credit.
  const elapsed = state.lastReviewed ? daysBetween(state.lastReviewed, today) : 0
  const delay = phase === 'review' ? Math.max(0, elapsed - intervalDays) : 0

  if (grade === 'again') {
    // A lapse: start over, and the card is penalized slightly so it grows
    // more cautiously next time around. Only a card that had graduated
    // counts as a lapse — failing a card still being learned is the
    // learning process, not a regression.
    if (phase === 'review') {
      lapses += 1
      phase = 'relearning'
    } else if (phase === 'new') {
      phase = 'learning'
    }
    repetitions = 0
    intervalDays = 1
    easeFactor = clampEase(easeFactor - 0.2)
  } else {
    repetitions += 1

    if (grade === 'hard') {
      // Still correct, but effortful: grow the interval only a little, and
      // erode the ease factor a bit so future intervals grow more slowly.
      intervalDays =
        repetitions === 1 ? 1 : Math.max(intervalDays + 1, (intervalDays + delay / 4) * HARD_MULTIPLIER)
      easeFactor = clampEase(easeFactor - 0.15)
    } else if (grade === 'good') {
      // The classic SM-2 progression: 1 day, then 6 days, then
      // previousInterval * easeFactor — ease factor unchanged. Half the
      // lateness is credited, which is Anki's treatment.
      if (repetitions === 1) intervalDays = 1
      else if (repetitions === 2) intervalDays = 6
      else intervalDays = (intervalDays + delay / 2) * easeFactor
    } else {
      // "easy": same progression as "good" but with a bonus multiplier and
      // a small ease-factor boost, since the card is clearly well-known.
      // All of the lateness is credited.
      if (repetitions === 1) intervalDays = 4
      else if (repetitions === 2) intervalDays = 8
      else intervalDays = (intervalDays + delay) * easeFactor * EASY_BONUS
      easeFactor = clampEase(easeFactor + 0.15)
    }

    // Past the one-day step the card is on the review ladder proper. This
    // is the same boundary the old display buckets used, so existing decks
    // keep the stats they had.
    phase = repetitions >= 2 ? 'review' : 'learning'
  }

  intervalDays = clampInterval(intervalDays)

  return {
    repetitions,
    easeFactor,
    intervalDays,
    dueDate: toISODate(addDays(today, intervalDays)),
    lastReviewed: toISODate(today),
    lapses,
    phase,
    // A leech has proven it is not being learned by repetition. Surfacing
    // it is more useful than showing it a ninth time.
    suspended: lapses >= LEECH_THRESHOLD ? true : undefined,
  }
}

/**
 * Previews the interval each grade would produce, so the buttons can say
 * what they will do. Grading blind is the main reason learners mis-calibrate
 * "hard" against "good".
 */
export function previewIntervals(state: CardState, today: Date = new Date()): Record<Grade, number> {
  return {
    again: review(state, 'again', today).intervalDays,
    hard: review(state, 'hard', today).intervalDays,
    good: review(state, 'good', today).intervalDays,
    easy: review(state, 'easy', today).intervalDays,
  }
}

/** True if a card's dueDate is today or earlier. */
export function isDue(state: CardState, today: Date = new Date()): boolean {
  if (state.suspended) return false
  // A card whose due date did not survive storage must not become
  // invisible: `undefined <= "2026-01-01"` is false, which would retire it
  // from the rotation permanently and silently. Showing it is recoverable;
  // losing it is not.
  if (typeof state.dueDate !== 'string') return true
  return state.dueDate <= toISODate(today)
}

/** True once a card has failed often enough to be worth rebuilding. */
export function isLeech(state: CardState): boolean {
  return (state.lapses ?? 0) >= LEECH_THRESHOLD
}

/** Rough bucket for progress stats: new / learning / review / mastered. */
export type CardBucket = 'new' | 'learning' | 'review' | 'mastered'

export function bucketOf(state: CardState | undefined): CardBucket {
  if (!state) return 'new'
  const s = normalizeState(state)
  if (s.phase === 'new') return 'new'
  // A lapsed card is back in the learning pile, not back to being unseen.
  if (s.phase === 'learning' || s.phase === 'relearning') return 'learning'
  if (s.intervalDays >= MASTERED_INTERVAL_DAYS) return 'mastered'
  return 'review'
}
