// Package es contains the domain model and reference data for Spanish verb
// conjugation: tenses, verbs, and the theory shown to learners. See the root
// grammar package for the types shared with the fr (French) package, and the
// fr package doc comment for why this is an independent package rather than
// a shared "Language" abstraction.
//
// Compared to French, Spanish simplifies compound-tense derivation
// considerably: every verb uses haber as its sole auxiliary (no avoir/être
// duality), and the past participle never agrees with the subject in
// compound tenses (invariable, unlike French être-verbs) — see compound.go.
// Spanish also has a grammatically distinct simple past, pretérito
// indefinido, with no French equivalent in this app (French's passé simple
// is omitted there as archaic/literary; Spanish's pretérito indefinido, by
// contrast, is the everyday spoken simple past and must be included).
//
// Spanish's imperativo afirmativo has five distinct forms (tú, usted,
// nosotros, vosotros, ustedes) rather than French's three (tu, nous, vous):
// "usted"/"ustedes" commands are conjugated exactly like 3rd-person
// indicative subjects conceptually, but their actual forms come from the
// subjuntivo presente, not the indicative. This maps cleanly onto the
// shared 6-slot Conjugation type (Je is simply never used for any
// imperative, in either language).
package es

import "github.com/bogdanbojan/sermokey/backend/internal/grammar"

// TenseID identifies one of the tenses/moods/non-finite forms covered by the
// app.
type TenseID = grammar.TenseID

const (
	// Tiempos simples (hand-authored).
	Present            TenseID = "presente"
	Imperfect          TenseID = "preterito-imperfecto"
	Preterite          TenseID = "preterito-indefinido"
	FutureSimple       TenseID = "futuro-simple"
	ConditionalPresent TenseID = "condicional-simple"
	SubjunctivePresent TenseID = "subjuntivo-presente"
	ImperativePresent  TenseID = "imperativo-afirmativo"

	// Tiempos compuestos (derived: haber conjugated in the matching simple tense + participio).
	PresentPerfect     TenseID = "preterito-perfecto-compuesto"  // haber presente + participio
	Pluperfect         TenseID = "pluscuamperfecto"              // haber imperfecto + participio
	FuturePerfect      TenseID = "futuro-perfecto"               // haber futuro + participio
	ConditionalPerfect TenseID = "condicional-compuesto"         // haber condicional + participio
	SubjunctivePerfect TenseID = "preterito-perfecto-subjuntivo" // haber subjuntivo presente + participio

	// Perífrasis verbales (derived from estar/acabar/ir presente).
	PresentProgressive TenseID = "presente-continuo"  // estar presente + gerundio
	RecentPast         TenseID = "preterito-reciente" // acabar presente + "de" + infinitivo
	NearFuture         TenseID = "futuro-proximo"     // ir presente + "a" + infinitivo

	// Formas no personales (derived, invariable — no person agreement).
	PastInfinitive TenseID = "infinitivo-compuesto" // haber (infinitivo) + participio
	Gerund         TenseID = "gerundio-simple"      // the gerundio itself
	PerfectGerund  TenseID = "gerundio-compuesto"   // habiendo + participio
)

// Category groups tenses for display (navigation sections).
type Category = grammar.Category

const (
	CategorySimple       Category = "Tiempos simples"
	CategoryCompound     Category = "Tiempos compuestos"
	CategoryPeriphrastic Category = "Perífrasis verbales"
	CategoryImpersonal   Category = "Formas no personales"
)

// TenseCategory maps each tense to its display category.
var TenseCategory = map[TenseID]Category{
	Present:            CategorySimple,
	Imperfect:          CategorySimple,
	Preterite:          CategorySimple,
	FutureSimple:       CategorySimple,
	ConditionalPresent: CategorySimple,
	SubjunctivePresent: CategorySimple,
	ImperativePresent:  CategorySimple,

	PresentPerfect:     CategoryCompound,
	Pluperfect:         CategoryCompound,
	FuturePerfect:      CategoryCompound,
	ConditionalPerfect: CategoryCompound,
	SubjunctivePerfect: CategoryCompound,

	PresentProgressive: CategoryPeriphrastic,
	RecentPast:         CategoryPeriphrastic,
	NearFuture:         CategoryPeriphrastic,

	PastInfinitive: CategoryImpersonal,
	Gerund:         CategoryImpersonal,
	PerfectGerund:  CategoryImpersonal,
}

// AllTenses lists every tense in the fixed display order used throughout the
// app, grouped by category.
var AllTenses = []TenseID{
	Present, Imperfect, Preterite, FutureSimple, ConditionalPresent, SubjunctivePresent, ImperativePresent,
	PresentPerfect, Pluperfect, FuturePerfect, ConditionalPerfect, SubjunctivePerfect,
	PresentProgressive, RecentPast, NearFuture,
	PastInfinitive, Gerund, PerfectGerund,
}

// invariableTenses have exactly one form regardless of grammatical person.
var invariableTenses = map[TenseID]bool{
	PastInfinitive: true,
	Gerund:         true,
	PerfectGerund:  true,
}

// IsInvariable reports whether t has a single form shared by every person.
func IsInvariable(t TenseID) bool {
	return invariableTenses[t]
}

// IsValidTense reports whether id is one of AllTenses.
func IsValidTense(id TenseID) bool {
	for _, t := range AllTenses {
		if t == id {
			return true
		}
	}
	return false
}

// Re-exported shared types/constants.
type Person = grammar.Person

var (
	Je   = grammar.Je
	Tu   = grammar.Tu
	Il   = grammar.Il
	Nous = grammar.Nous
	Vous = grammar.Vous
	Ils  = grammar.Ils
)

// Conjugation holds a verb's form for each of the six persons. It is reused
// for the imperativo afirmativo tense too, which populates five of the six
// slots (Tu/Il/Nous/Vous/Ils) — only Je stays empty. This type is
// deliberately NOT shared with the fr package (despite the identical
// shape): it's purely an internal implementation detail of this package —
// see fr.Conjugation's doc comment for why.
type Conjugation struct {
	Je, Tu, Il, Nous, Vous, Ils string
}

// Get returns the form for the given person, or "" if unset.
func (c Conjugation) Get(p Person) string {
	switch p {
	case Je:
		return c.Je
	case Tu:
		return c.Tu
	case Il:
		return c.Il
	case Nous:
		return c.Nous
	case Vous:
		return c.Vous
	case Ils:
		return c.Ils
	default:
		return ""
	}
}

// ConjugablePersons lists all six persons, used for every regular tense.
var ConjugablePersons = grammar.ConjugablePersons

// ImperativePersons lists the five persons Spanish's imperativo afirmativo
// uses: tú, usted (3rd singular slot), nosotros, vosotros, ustedes (3rd
// plural slot). Unlike French, there is no form that's skipped besides "yo".
var ImperativePersons = []Person{Tu, Il, Nous, Vous, Ils}

// PersonsFor returns the valid persons to draw from for the given tense.
func PersonsFor(t TenseID) []Person {
	switch {
	case t == ImperativePresent:
		return ImperativePersons
	case IsInvariable(t):
		return []Person{Je}
	default:
		return ConjugablePersons
	}
}

// Group identifies one of the three Spanish conjugation classes, named by
// infinitive ending. Stem-changing and irregular verbs cut across all three
// and are called out individually in the theory notes.
type Group string

const (
	GroupAR Group = "-ar"
	GroupER Group = "-er"
	GroupIR Group = "-ir"
)

// Verb is a single verb with its hand-authored simple-tense conjugation
// table, plus its past participle and gerund (used to derive every
// compound/periphrastic/non-finite form on demand — see compound.go).
type Verb struct {
	Infinitive    string
	Group         Group
	TranslationEn string
	// PastParticiple is the invariable past participle, e.g. "hablado",
	// "hecho", "vuelto". Unlike French, it never agrees with the subject in
	// compound tenses — Spanish participles are invariable with haber.
	PastParticiple string
	// Gerundio is the invariable gerund, e.g. "hablando", "pidiendo",
	// "yendo". Used to derive the présente continuo periphrasis.
	Gerundio string
	// HasImperative is false for verbs with no natural imperative (e.g.
	// haber, poder), excluded from imperativo-afirmativo exercises.
	HasImperative bool
	// Contexts are short sentence completions (no leading/trailing space,
	// no final punctuation) used to build natural exercise sentences.
	Contexts []string

	Present            Conjugation
	Imperfect          Conjugation
	Preterite          Conjugation
	FutureSimple       Conjugation
	ConditionalPresent Conjugation
	SubjunctivePresent Conjugation
	ImperativePresent  Conjugation // only Tu/Il/Nous/Vous/Ils populated
}

// ConjugationFor returns the conjugated form of v for the given tense and
// person, and whether that combination exists. Simple tenses are read
// directly off v; every other tense is derived — see compound.go.
func (v Verb) ConjugationFor(t TenseID, p Person) (string, bool) {
	switch t {
	case Present:
		return v.Present.Get(p), v.Present.Get(p) != ""
	case Imperfect:
		return v.Imperfect.Get(p), v.Imperfect.Get(p) != ""
	case Preterite:
		return v.Preterite.Get(p), v.Preterite.Get(p) != ""
	case FutureSimple:
		return v.FutureSimple.Get(p), v.FutureSimple.Get(p) != ""
	case ConditionalPresent:
		return v.ConditionalPresent.Get(p), v.ConditionalPresent.Get(p) != ""
	case SubjunctivePresent:
		return v.SubjunctivePresent.Get(p), v.SubjunctivePresent.Get(p) != ""
	case ImperativePresent:
		form := v.ImperativePresent.Get(p)
		return form, form != ""
	default:
		return derivedConjugationFor(v, t, p)
	}
}
