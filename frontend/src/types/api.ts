// Mirrors the JSON shapes returned by the Go backend (internal/httpapi/dto.go).
// Keep this in sync by hand — the dataset is small and static enough that a
// generated client would be more ceremony than value here.

export type LanguageCode = 'fr' | 'es' | 'de'

export interface Language {
  code: LanguageCode
  name: string // native endonym, e.g. "Français" / "Español"
}

export interface TenseSummary {
  id: string
  category: string
  nameNative: string // the tense's name in the language being studied
  nameEn: string // English gloss
  usage: string
}

export interface GroupEndings {
  groupName: string
  explanation: string
  endings: string[]
  exampleVerb: string
  exampleForms: string[]
}

export interface Theory {
  id: string
  category: string
  nameNative: string
  nameEn: string
  usage: string
  formation: string
  groups: GroupEndings[]
  irregularNotes: string
}

export interface VerbSummary {
  infinitive: string
  group: string
  translationEn: string
}

export type ExerciseType = 'mcq' | 'text'

export interface Exercise {
  id: string
  tense: string
  tenseName: string
  type: ExerciseType
  infinitive: string
  translationEn: string
  personLabel: string
  sentencePrefix: string
  sentenceSuffix: string
  choices?: string[]
  answer: string
}

export interface VocabCategory {
  id: string
  nameNative: string
  nameEn: string
  wordCount: number
}

export interface VocabWord {
  id: string
  lemma: string
  pos: 'noun' | 'verb' | 'adj' | 'adv' | 'other'
  gender?: 'm' | 'f' | 'n' | ''
  translationEn: string
  category: string
  tier: number
}

// --- Reading -------------------------------------------------------------
// Mirrors internal/httpapi/reading_dto.go.

export type ReadingLevel = 'A2' | 'B1' | 'B2'

/** One text's library card. Carries no content. */
export interface ReadingText {
  id: string
  title: string
  author?: string
  level: ReadingLevel
  blurb: string
  source?: string
  license: string
  sectionCount: number
  tokenCount: number
  /** How many tokens are closed-class grammatical words, which
   *  `wordFrequency` deliberately omits. Added to both sides of the
   *  coverage ratio: the 95%/98% comprehension thresholds are defined over
   *  all running words, articles included, but "le" is grammar rather than
   *  vocabulary to learn. */
  functionTokenCount: number
  /** How many tokens are proper nouns. Likewise omitted from
   *  `wordFrequency` and likewise counted as known: coverage studies treat
   *  names as known words (Nation 2006), because a reader simply reads
   *  them rather than needing to have learned them. */
  properTokenCount: number
  /** A book the reader added themselves rather than a bundled text.
   *  Only these can be removed. */
  imported?: boolean
  /**
   * SRS card id -> number of word tokens resolving to it. Intersected with
   * the reader's own deck and known-word set to produce the coverage
   * figure on the card, which is a far better difficulty signal for a
   * given reader than the static CEFR `level`.
   *
   * Keyed by card id rather than lemma so the intersection is direct: a
   * curated word's id is "{lemma}|{category}" and cannot be rebuilt from
   * the lemma alone.
   */
  wordFrequency: Record<string, number>
}

/** One English gloss. */
export interface ReadingSense {
  gloss: string
}

/** A secondary reading of a surface form, offered so a reader can correct
 *  a mis-ranked homograph ("est" as east rather than as a form of être). */
export interface ReadingAlt {
  lemma: string
  pos: string
  gender?: string
  formNote?: string
  senses: ReadingSense[]
  wordId: string
  curated: boolean
}

/** One dictionary entry as the reader shows it. */
export interface ReadingEntry {
  surface: string
  lemma: string
  pos: string
  gender?: string
  /** Explains the inflection, e.g. "first-person plural present of parler".
   *  Empty when the surface is already the lemma. */
  formNote?: string
  senses: ReadingSense[]
  /** The SRS card id to save under. Always minted server-side so that a
   *  word saved while reading touches the same card the vocabulary tab
   *  schedules, never a duplicate with its own review history. */
  wordId: string
  /** Part of the taught curriculum, as opposed to merely present in the
   *  dictionary. */
  curated: boolean
  /** A closed-class grammatical word: rendered as known, left undecorated. */
  function?: boolean
  /** A proper noun: likewise left undecorated. */
  proper?: boolean
  /** The analysis was guessed by rule-based fallback; present it as a
   *  guess rather than as a dictionary fact. */
  uncertain?: boolean
  alternatives?: ReadingAlt[]
}

/** Number of integers per token in ReadingParagraph.spans. */
export const SPAN_STRIDE = 3
/** The spans lexicon index meaning "nothing to look up here". */
export const NO_ENTRY = -1

export interface ReadingParagraph {
  text: string
  /**
   * Flat triples of [startByte, endByte, lexiconIndex], one per token,
   * covering `text` contiguously. lexiconIndex is NO_ENTRY for punctuation,
   * whitespace and words with no dictionary entry.
   *
   * Flat integers over the verbatim text rather than pre-rendered token
   * objects: rendering by slicing keeps typography intact, and it puts one
   * span per word in the DOM over the real text, so the browser's native
   * selection maps straight onto token indices. That is what makes
   * drag-selecting a phrase possible.
   */
  spans: number[]
}

export interface ReadingSection {
  textId: string
  title: string
  sectionTitle: string
  index: number
  sectionCount: number
  license: string
  source?: string
  paragraphs: ReadingParagraph[]
  /** Every distinct analysis the section's tokens reference. Shipped with
   *  the section so hovering never waits on the network. */
  lexicon: ReadingEntry[]
}
