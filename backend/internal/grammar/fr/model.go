// Package fr contains the domain model and reference data for French verb
// conjugation: tenses, verbs, and the theory ("terminaisons") shown to
// learners. See the root grammar package for the types shared with the es
// (Spanish) package.
//
// Simple tenses (présent, imparfait, futur simple, conditionnel présent,
// subjonctif présent, impératif présent) are authored by hand and kept as
// plain data (not derived by rule-based code generation) because correctness
// of the French is the entire value of this application, and irregular verbs
// are far riskier to generate than to review as an explicit table.
//
// Compound tenses, periphrastic constructions, and non-finite forms
// (passé composé, plus-que-parfait, futur antérieur, conditionnel passé,
// subjonctif passé, impératif passé, présent continu, passé récent, futur
// proche, infinitif passé, gérondif présent, gérondif passé) are instead
// *derived* at lookup time from: the verb's auxiliary choice (avoir/être),
// its past participle, and — since avoir/être/aller/venir are verbs in this
// same dataset — the already-reviewed simple-tense tables of those four
// auxiliary verbs. See compound.go. This means adding a new simple verb
// automatically gets all twelve derived tenses for free and correct by
// construction, instead of needing ~70 additional hand-typed forms per verb.
//
// This package is intentionally NOT built behind a generic "Language"
// interface shared with es: French's avoir/être auxiliary duality, elision
// rules, and "que"-clause contraction have no structural equivalent in
// Spanish (which uses a single haber auxiliary and has no elision), so
// forcing a shared abstraction would either leak French-only concepts into
// the interface or require per-language escape hatches that defeat the
// point of the abstraction. Instead, fr and es are independent packages with
// parallel shapes, consumed by internal/httpapi through a small local
// interface — see httpapi/languages.go.
package fr

import "github.com/bogdanbojan/sermokey/backend/internal/grammar"

// TenseID identifies one of the tenses/moods/non-finite forms covered by the
// app.
type TenseID = grammar.TenseID

const (
	// Temps simples (hand-authored).
	Present             TenseID = "present"
	Imparfait           TenseID = "imparfait"
	FutureSimple        TenseID = "futur-simple"
	ConditionnelPresent TenseID = "conditionnel-present"
	SubjonctifPresent   TenseID = "subjonctif-present"
	ImperatifPresent    TenseID = "imperatif-present"

	// Temps composés (derived: aux conjugated in the matching simple tense + participle).
	PasseCompose      TenseID = "passe-compose"      // avoir/être présent + participe
	PlusQueParfait    TenseID = "plus-que-parfait"   // avoir/être imparfait + participe
	FuturAnterieur    TenseID = "futur-anterieur"    // avoir/être futur simple + participe
	ConditionnelPasse TenseID = "conditionnel-passe" // avoir/être conditionnel présent + participe
	SubjonctifPasse   TenseID = "subjonctif-passe"   // avoir/être subjonctif présent + participe
	ImperatifPasse    TenseID = "imperatif-passe"    // avoir/être impératif présent + participe

	// Constructions périphrastiques (derived from être/venir/aller présent).
	PresentContinu TenseID = "present-continu" // être présent + "en train de" + infinitif
	PasseRecent    TenseID = "passe-recent"    // venir présent + "de" + infinitif
	FuturProche    TenseID = "futur-proche"    // aller présent + infinitif

	// Modes impersonnels (derived, invariable — no person agreement).
	InfinitifPasse  TenseID = "infinitif-passe"  // avoir/être (infinitif) + participe
	GerondifPresent TenseID = "gerondif-present" // "en" + participe présent
	GerondifPasse   TenseID = "gerondif-passe"   // ayant/étant + participe
)

// Category groups tenses for display (navigation sections).
type Category = grammar.Category

const (
	CategorySimple       Category = "Temps simples"
	CategoryCompound     Category = "Temps composés"
	CategoryPeriphrastic Category = "Constructions périphrastiques"
	CategoryImpersonal   Category = "Modes impersonnels"
)

// TenseCategory maps each tense to its display category.
var TenseCategory = map[TenseID]Category{
	Present:             CategorySimple,
	Imparfait:           CategorySimple,
	FutureSimple:        CategorySimple,
	ConditionnelPresent: CategorySimple,
	SubjonctifPresent:   CategorySimple,
	ImperatifPresent:    CategorySimple,

	PasseCompose:      CategoryCompound,
	PlusQueParfait:    CategoryCompound,
	FuturAnterieur:    CategoryCompound,
	ConditionnelPasse: CategoryCompound,
	SubjonctifPasse:   CategoryCompound,
	ImperatifPasse:    CategoryCompound,

	PresentContinu: CategoryPeriphrastic,
	PasseRecent:    CategoryPeriphrastic,
	FuturProche:    CategoryPeriphrastic,

	InfinitifPasse:  CategoryImpersonal,
	GerondifPresent: CategoryImpersonal,
	GerondifPasse:   CategoryImpersonal,
}

// AllTenses lists every tense in the fixed display order used throughout the
// app (navigation, theory index, "mixed" exercise pool, ...), grouped by
// category.
var AllTenses = []TenseID{
	Present, Imparfait, FutureSimple, ConditionnelPresent, SubjonctifPresent, ImperatifPresent,
	PasseCompose, PlusQueParfait, FuturAnterieur, ConditionnelPasse, SubjonctifPasse, ImperatifPasse,
	PresentContinu, PasseRecent, FuturProche,
	InfinitifPasse, GerondifPresent, GerondifPasse,
}

// invariableTenses have exactly one form regardless of grammatical person
// (no subject agreement at all): the non-finite "modes impersonnels".
var invariableTenses = map[TenseID]bool{
	InfinitifPasse:  true,
	GerondifPresent: true,
	GerondifPasse:   true,
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

// Re-exported shared types/constants, so the rest of this package (and its
// tests) can refer to them without an extra import alias.
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
// for the impératif tense too, which only populates Tu/Nous/Vous — Je/Il/Ils
// stay empty. This type is deliberately NOT shared with the es package
// (despite having the identical shape): it's purely an internal
// implementation detail of this package, and keeping it local lets verbs.go
// use terse positional struct literals without tripping go vet's
// unkeyed-fields-from-another-package check.
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

// ImperativePersons lists the three persons French's imperative moods use.
var ImperativePersons = []Person{Tu, Nous, Vous}

// PersonsFor returns the valid persons to draw from for the given tense:
// ImperativePersons for the two imperative tenses, a single placeholder
// person for invariable tenses (the stored form is identical for all six —
// see IsInvariable), or ConjugablePersons otherwise.
func PersonsFor(t TenseID) []Person {
	switch {
	case t == ImperatifPresent || t == ImperatifPasse:
		return ImperativePersons
	case IsInvariable(t):
		return []Person{Je}
	default:
		return ConjugablePersons
	}
}

// Group identifies which of the three traditional French verb groups a verb
// belongs to. This drives which terminaisons table is shown as theory.
type Group string

const (
	GroupFirst  Group = "1er groupe" // -er verbs, e.g. parler
	GroupSecond Group = "2e groupe"  // regular -ir verbs, e.g. finir
	GroupThird  Group = "3e groupe"  // irregular verbs, e.g. être, aller, prendre
)

// Verb is a single verb with its hand-authored simple-tense conjugation
// table, plus the metadata (auxiliary, past participle) needed to derive
// every compound/periphrastic/non-finite form on demand — see compound.go.
type Verb struct {
	Infinitive    string
	Group         Group
	TranslationEn string
	// Auxiliary is "avoir" or "être": which one this verb uses to form every
	// compound tense, and whether the past participle agrees in number with
	// a plural subject (être verbs pluralize it; see compound.go).
	Auxiliary string
	// PastParticiple is the invariable (masculine singular) past participle,
	// e.g. "parlé", "pris", "allé". Used to derive every compound tense.
	PastParticiple string
	// HasImperative is false for verbs with no natural imperative (e.g.
	// pouvoir), which are excluded from imperatif-present/imperatif-passe
	// exercises.
	HasImperative bool
	// Contexts are short sentence completions (no leading/trailing space,
	// no final punctuation) used to build natural exercise sentences, e.g.
	// "au téléphone avec sa sœur" for parler.
	Contexts []string

	Present             Conjugation
	Imparfait           Conjugation
	FutureSimple        Conjugation
	ConditionnelPresent Conjugation
	SubjonctifPresent   Conjugation
	ImperatifPresent    Conjugation // only Tu/Nous/Vous populated
}

// ConjugationFor returns the conjugated form (or imperative form) of v for
// the given tense and person, and whether that combination exists. Simple
// tenses are read directly off v; every other tense is derived — see
// compound.go.
func (v Verb) ConjugationFor(t TenseID, p Person) (string, bool) {
	switch t {
	case Present:
		return v.Present.Get(p), v.Present.Get(p) != ""
	case Imparfait:
		return v.Imparfait.Get(p), v.Imparfait.Get(p) != ""
	case FutureSimple:
		return v.FutureSimple.Get(p), v.FutureSimple.Get(p) != ""
	case ConditionnelPresent:
		return v.ConditionnelPresent.Get(p), v.ConditionnelPresent.Get(p) != ""
	case SubjonctifPresent:
		return v.SubjonctifPresent.Get(p), v.SubjonctifPresent.Get(p) != ""
	case ImperatifPresent:
		form := v.ImperatifPresent.Get(p)
		return form, form != ""
	default:
		return derivedConjugationFor(v, t, p)
	}
}
