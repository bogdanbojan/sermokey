// Package grammar holds types shared by every language-specific conjugation
// engine (see the fr and es subpackages). Each language package owns its own
// data (verbs, theory content) and derivation/sentence-building logic, since
// those genuinely differ per language (e.g. French elision and avoir/être
// auxiliary duality vs Spanish's haber-only compound tenses and lack of
// elision) — see the fr/es package doc comments for why a shared "generic
// grammar engine" isn't the right abstraction here. What IS shared is the
// pure data shapes: the six grammatical persons, a six-slot conjugation
// table, and the DTO-ish shapes (Exercise, TenseTheory) that the HTTP layer
// and frontend consume identically regardless of which language produced
// them.
package grammar

// Person identifies one of the six grammatical persons common to French and
// Spanish conjugation: 1st/2nd/3rd singular, 1st/2nd/3rd plural. Display
// labels (e.g. French "il / elle / on" vs Spanish "él / ella / usted") are
// language-specific and live in each language package.
type Person string

const (
	Je   Person = "p1s" // 1st singular (je / yo)
	Tu   Person = "p2s" // 2nd singular (tu / tú)
	Il   Person = "p3s" // 3rd singular (il/elle/on / él,ella,usted)
	Nous Person = "p1p" // 1st plural (nous / nosotros)
	Vous Person = "p2p" // 2nd plural (vous / vosotros)
	Ils  Person = "p3p" // 3rd plural (ils/elles / ellos,ellas,ustedes)
)

// ConjugablePersons lists all six persons, in canonical display order.
var ConjugablePersons = []Person{Je, Tu, Il, Nous, Vous, Ils}

// TenseID identifies a tense/mood/non-finite form within one language's own
// tense set. Values are scoped per language package (fr.Present, es.Present,
// ...); the underlying string is only unique within a language.
type TenseID string

// Category groups a language's tenses for display (navigation sections),
// e.g. "Temps simples" (French) or "Tiempos simples" (Spanish).
type Category string

// ExerciseType distinguishes how the learner answers an exercise.
type ExerciseType string

const (
	MultipleChoice ExerciseType = "mcq"
	FreeText       ExerciseType = "text"
)

// GroupEndings describes one verb group/class's formation pattern within a
// tense, shown on the theory page.
type GroupEndings struct {
	GroupName    string
	Explanation  string
	Endings      [6]string // je/yo, tu, il/él, nous/nosotros, vous/vosotros, ils/ellos — "" if not applicable
	ExampleVerb  string
	ExampleForms [6]string
}

// TenseTheory is the full theory content for one tense: how it's used, how
// it's formed, and per-group terminaison tables with a worked example.
type TenseTheory struct {
	ID             TenseID
	NameNative     string // the tense's name in the language being studied
	NameEn         string // English gloss, shown as a parenthetical for learners
	Usage          string
	Formation      string
	Groups         []GroupEndings
	IrregularNotes string
}

// Exercise is one generated conjugation question, identical in shape
// regardless of which language produced it. Sentence is split into a Prefix
// (everything before the blank) and Suffix (everything after it) so the
// frontend only needs to render Prefix + <input/choices> + Suffix — all
// grammar concerns (elision, pronoun choice, mood-specific particles,
// capitalization) are resolved server-side, per language.
type Exercise struct {
	ID             string
	Tense          TenseID
	TenseName      string // the tense's name in the language being studied
	Type           ExerciseType
	Infinitive     string
	TranslationEn  string
	PersonLabel    string
	SentencePrefix string
	SentenceSuffix string
	Choices        []string // only populated for MultipleChoice
	Answer         string
}

// VocabWord is one vocabulary entry for the spaced-repetition flashcard
// system (see /vocab.go in each language package and the frontend's srs/
// module). Unlike the conjugation Exercise above, vocabulary review is a
// self-graded recall task (see a word, attempt to recall its meaning,
// reveal, rate yourself), not an auto-graded MCQ/text exercise — vocabulary
// at the scale needed for real coverage (literature converges on ~2000
// word-families for basic fluency; see Nation 2006, Schmitt 2000) is
// recognized/learned far faster than it can be meaningfully drilled in
// full sentences one grammatical form at a time.
type VocabWord struct {
	ID            string // stable identifier: "{lemma}|{category}", used as the SRS deck key
	Lemma         string
	POS           string // "noun" | "verb" | "adj" | "adv" | "other"
	Gender        string // "m" | "f" | "" (non-nouns, or invariant-gender items)
	TranslationEn string
	Category      string // language-scoped category ID, e.g. "famille" (fr) / "familia" (es)
	Tier          int    // rough frequency tier: 1 = most essential, 2, 3 = less common
}

// VocabCategoryInfo is the display metadata for one vocabulary category.
type VocabCategoryInfo struct {
	ID         string
	NameNative string
	NameEn     string
}
