// How a word is rendered in running text, and how much of a text the
// reader already knows.
//
// Status is derived, never stored: it is a view over the SRS deck, the
// saved-word store and the known/ignored sets, all of which are owned
// elsewhere. Caching it would mean invalidating it on every review.
import { computed, type ComputedRef } from 'vue'
import type { LanguageCode, ReadingText } from '../types/api'
import { useSrsDeck, cardID, type Direction } from '../srs/store'
import { bucketOf } from '../srs/sm2'
import { useSavedWords } from './saved'
import { useWordStatuses } from './known'

/**
 * The visual states a word can be in on the page.
 *
 * Turning the page into an ambient progress display is the single feature
 * that makes reading feel like it is advancing the deck rather than running
 * beside it — it is why LingQ's colour-coding is the thing people remember
 * about it. The states are deliberately few: more than four becomes
 * decoration a reader stops parsing.
 */
export type WordStatus =
  /** Never seen, not in the deck. The thing to notice. */
  | 'unknown'
  /** In the deck and still maturing. */
  | 'learning'
  /** Mature in the deck, or explicitly marked known. No decoration. */
  | 'known'
  /** Explicitly dismissed: names, numerals. No decoration, no counting. */
  | 'ignored'

/** Which direction's scheduling decides "learning" vs "known". Recognition
 *  is the right one for reading: it is the skill being exercised. */
const READING_DIRECTION: Direction = 'recognition'

export function useWordStatus(lang: ComputedRef<LanguageCode> | { value: LanguageCode }) {
  const deck = computed(() => useSrsDeck(lang.value))
  const saved = computed(() => useSavedWords(lang.value))
  const statuses = computed(() => useWordStatuses(lang.value))

  function statusOf(wordId: string | undefined): WordStatus {
    if (!wordId) return 'ignored'
    if (statuses.value.isIgnored(wordId)) return 'ignored'
    if (statuses.value.isKnown(wordId)) return 'known'

    const state = deck.value.stateFor(cardID(wordId, READING_DIRECTION))
    if (state) {
      return bucketOf(state) === 'mastered' ? 'known' : 'learning'
    }
    // Saved but not yet reviewed: still learning, and visibly so, which is
    // the feedback that makes the save feel like it did something.
    return saved.value.has(wordId) ? 'learning' : 'unknown'
  }

  /**
   * Whether the reader can be expected to recognise a word while reading.
   *
   * This is deliberately a weaker test than statusOf's 'known', and the
   * difference matters a great deal. statusOf reports *mastery* — a card
   * with a month-long interval — which is the right thing to show on the
   * page, where the point is to see which words still need work.
   *
   * Coverage is asking something else: can this person read this text? A
   * word reviewed successfully even once is one they will recognise in
   * context, and that is what the comprehension thresholds count. Using
   * mastery here instead meant a learner who had worked through a
   * thousand words but had none of them mature yet scored barely above
   * someone who had studied nothing, and every text — including ones
   * written for beginners — was labelled too hard.
   *
   * A card reset to zero repetitions by an "again" grade is excluded: a
   * word just failed is not one being recognised.
   */
  function comprehensionOf(wordId: string | undefined): CoverageClass {
    if (!wordId) return 'excluded'
    if (statuses.value.isIgnored(wordId)) return 'excluded'
    if (statuses.value.isKnown(wordId)) return 'known'

    const state = deck.value.stateFor(cardID(wordId, READING_DIRECTION))
    if (state && state.repetitions >= 1) return 'known'
    return 'unknown'
  }

  return { statusOf, comprehensionOf }
}

/** How a word counts toward the coverage figure. */
export type CoverageClass = 'known' | 'unknown' | 'excluded'

/** How much of a text the reader already knows, token-weighted. */
export interface Coverage {
  /** Share of countable tokens that are known, 0..1. */
  known: number
  /** Distinct lemmas not yet known — the work the text represents. */
  newWords: number
  /** Tokens counted, i.e. excluding ignored words. */
  countedTokens: number
  band: CoverageBand
  /**
   * True when the reader has told us nothing yet — empty deck, nothing
   * marked known. Coverage is then 0 for every text, which is literally
   * true and completely uninformative: it labels three texts of very
   * different difficulty identically as "hard" and makes the feature look
   * broken on first run. Callers fall back to the author-assigned CEFR
   * level until there is something to measure against.
   */
  coldStart: boolean
}

/**
 * Hu & Nation (2000) put adequate unassisted comprehension at ~98% known
 * running words and assisted comprehension at ~95% (Laufer 1989). Those
 * thresholds are about *tokens*, not distinct words, which is why the
 * backend ships a token-weighted frequency table rather than a word list.
 *
 * Expressing difficulty this way means it is relative to the reader: the
 * same text is "comfortable" for one person and "hard" for another, which
 * is both true and far more useful than a fixed CEFR badge.
 */
export type CoverageBand = 'comfortable' | 'challenging' | 'hard'

export function bandFor(known: number): CoverageBand {
  if (known >= 0.98) return 'comfortable'
  if (known >= 0.95) return 'challenging'
  return 'hard'
}

export function coverageOf(
  text: Pick<ReadingText, 'wordFrequency' | 'functionTokenCount' | 'properTokenCount'>,
  comprehensionOf: (wordId: string) => CoverageClass,
): Coverage {
  let counted = 0
  let known = 0
  let newWords = 0
  let anyTouched = false
  for (const [wordId, n] of Object.entries(text.wordFrequency)) {
    const cls = comprehensionOf(wordId)
    // Excluded words — the ones the reader dismissed as names or numerals
    // — are left out of both sides rather than counted as known: a page
    // full of names should not flatter the score.
    if (cls === 'excluded') {
      anyTouched = true
      continue
    }
    counted += n
    if (cls === 'known') {
      known += n
      anyTouched = true
    } else {
      newWords++
    }
  }
  // Function words and proper nouns are counted on both sides rather than
  // skipped. Together they are around 40% of running French, and a reader
  // knows all of them — articles because they are grammar, names because a
  // name is simply read. Omitting them would deflate the ratio just as
  // badly as treating them as unknown vocabulary did.
  const alwaysKnown = text.functionTokenCount + text.properTokenCount
  counted += alwaysKnown
  known += alwaysKnown

  const ratio = counted === 0 ? 0 : known / counted
  return {
    known: ratio,
    newWords,
    countedTokens: counted,
    band: bandFor(ratio),
    coldStart: !anyTouched,
  }
}
