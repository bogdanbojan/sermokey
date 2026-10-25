// Mirrors the JSON shapes returned by the Go backend (internal/httpapi/dto.go).
// Keep this in sync by hand — the dataset is small and static enough that a
// generated client would be more ceremony than value here.

export type LanguageCode = 'fr' | 'es'

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
  gender?: 'm' | 'f' | ''
  translationEn: string
  category: string
  tier: number
}
