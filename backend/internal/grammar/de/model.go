// Package de contains the domain model and reference data for German verb
// conjugation: tenses, verbs, and the theory shown to learners. See the root
// grammar package for the types shared with the fr (French) and es
// (Spanish) packages, and the fr package doc comment for why this is an
// independent package rather than a shared "Language" abstraction.
//
// German turns out to sit structurally between French and Spanish: like
// French (and unlike Spanish), it has a genuine haben/sein auxiliary
// duality — sein is used for intransitive verbs of motion or change of
// state (gehen, kommen, fahren, bleiben, werden, and sein itself), haben
// for everything else — see compound.go. But like Spanish (and unlike
// French), the past participle in a compound tense never agrees with the
// subject: it is always invariable.
//
// German also forces a genuinely new problem French/Spanish never raised:
// word order. In a German main clause, the finite (conjugated) verb sits in
// position 2, while any non-finite part — a past participle, a bare
// infinitive, or an infinitive+participle pair — is pushed all the way to
// the END of the clause (the "Satzklammer"/clause bracket), with every
// object and adverbial in between. "Ich habe mit meiner Schwester
// telefoniert" is grammatical; "Ich habe telefoniert mit meiner Schwester"
// is not. Because this app's generated exercises wrap each answer in a
// natural sentence with a context phrase, French/Spanish can safely return
// a compound tense's ENTIRE "aux + participle" string as one answer (the
// context phrase is simply appended after it). German cannot: once a
// context phrase sits between the finite auxiliary and the clause-final
// non-finite part, a single contiguous answer string would produce an
// ungrammatical sentence. So here, uniquely among the three language
// packages, Verb.ConjugationFor for a compound tense returns ONLY the
// finite part (the conjugated auxiliary/werden), and the invariant
// non-finite tail that belongs at the very end of the clause is supplied
// separately by NonfiniteTail (see compound.go) for exercise.go to place
// correctly. This also means a German compound-tense exercise specifically
// drills auxiliary selection (haben or sein?) and its correct conjugation —
// arguably the single most German-specific skill in this whole tense
// system — rather than bundling that together with participle recall the
// way French/Spanish do.
//
// This package's tense categories follow how German is actually taught
// (Indikativ / Konjunktiv / Imperativ / nicht-finite Formen), rather than
// forcing fr/es's "simple vs. compound vs. periphrastic" structural split:
// German's "going to" future is simply Futur I (there is no distinct
// French-style futur proche/passé récent periphrasis in standard German),
// while reported speech (Konjunktiv I) has no French/Spanish equivalent at
// all and deserves its own category. Separable-prefix verbs (trennbare
// Verben, e.g. aufstehen) are deliberately out of scope for this first
// version: their prefix detaches and moves to the clause-final position
// independently of the participle/infinitive, which would compound the
// word-order handling above considerably — every verb in Verbs (see
// verbs.go) is non-separable.
package de

import "github.com/bogdanbojan/sermokey/backend/internal/grammar"

// TenseID identifies one of the tenses/moods/non-finite forms covered by
// the app.
type TenseID = grammar.TenseID

const (
	// Indikativ (Präsens/Präteritum hand-authored; the rest derived from
	// the auxiliary's own simple-tense table + the verb's past participle
	// or infinitive — see compound.go).
	Prasens         TenseID = "praesens"
	Prateritum      TenseID = "praeteritum"
	Perfekt         TenseID = "perfekt"         // haben/sein präsens + Partizip II
	Plusquamperfekt TenseID = "plusquamperfekt" // haben/sein präteritum + Partizip II
	FuturI          TenseID = "futur-1"         // werden präsens + Infinitiv
	FuturII         TenseID = "futur-2"         // werden präsens + Partizip II + haben/sein (Infinitiv)

	// Konjunktiv (Präsens forms hand-authored; Perfekt forms derived).
	KonjunktivIIPrasens TenseID = "konjunktiv-2-praesens" // hypothetical, polite, irrealis
	KonjunktivIPrasens  TenseID = "konjunktiv-1-praesens" // indirekte Rede (reported speech)
	KonjunktivIIPerfekt TenseID = "konjunktiv-2-perfekt"  // hätte/wäre + Partizip II
	KonjunktivIPerfekt  TenseID = "konjunktiv-1-perfekt"  // habe/sei + Partizip II

	// Imperativ (hand-authored).
	Imperativ TenseID = "imperativ"

	// Nicht-finite Formen (derived, invariable — no person agreement).
	InfinitivPerfekt TenseID = "infinitiv-perfekt" // Partizip II + zu + haben/sein (Infinitiv)
	PartizipI        TenseID = "partizip-1"        // Infinitiv + d
)

// Category groups tenses for display (navigation sections).
type Category = grammar.Category

const (
	CategoryIndikativ  Category = "Indikativ"
	CategoryKonjunktiv Category = "Konjunktiv"
	CategoryImperativ  Category = "Imperativ"
	CategoryNichtFinit Category = "Nicht-finite Formen"
)

// TenseCategory maps each tense to its display category.
var TenseCategory = map[TenseID]Category{
	Prasens:         CategoryIndikativ,
	Prateritum:      CategoryIndikativ,
	Perfekt:         CategoryIndikativ,
	Plusquamperfekt: CategoryIndikativ,
	FuturI:          CategoryIndikativ,
	FuturII:         CategoryIndikativ,

	KonjunktivIIPrasens: CategoryKonjunktiv,
	KonjunktivIPrasens:  CategoryKonjunktiv,
	KonjunktivIIPerfekt: CategoryKonjunktiv,
	KonjunktivIPerfekt:  CategoryKonjunktiv,

	Imperativ: CategoryImperativ,

	InfinitivPerfekt: CategoryNichtFinit,
	PartizipI:        CategoryNichtFinit,
}

// AllTenses lists every tense in the fixed display order used throughout
// the app (navigation, theory index, "mixed" exercise pool, ...), grouped
// by category.
var AllTenses = []TenseID{
	Prasens, Prateritum, Perfekt, Plusquamperfekt, FuturI, FuturII,
	KonjunktivIIPrasens, KonjunktivIPrasens, KonjunktivIIPerfekt, KonjunktivIPerfekt,
	Imperativ,
	InfinitivPerfekt, PartizipI,
}

// invariableTenses have exactly one form regardless of grammatical person
// (no subject agreement at all): the two nicht-finite Formen.
var invariableTenses = map[TenseID]bool{
	InfinitivPerfekt: true,
	PartizipI:        true,
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

// Conjugation holds a verb's form for each of the six persons: Je=ich,
// Tu=du, Il=er/sie/es, Nous=wir, Vous=ihr, Ils=sie (3rd plural "they" —
// which happens to share its verb form with the formal "Sie", so no
// seventh slot is needed; see personLabel in exercise.go). It is reused for
// the Imperativ tense too, which only populates Tu/Nous/Vous/Il (du, wir,
// ihr, and the formal Sie borrowed from the Il slot) — Je/Ils stay empty.
// This type is deliberately NOT shared with the fr/es packages (despite the
// identical shape): it's purely an internal implementation detail of this
// package — see fr.Conjugation's doc comment for why.
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

// ImperativePersons lists the four persons German's Imperativ uses: du
// (Tu), wir (Nous, the "let's ..." form), ihr (Vous), and the formal Sie
// (borrowing the Il slot — Sie is grammatically a 3rd-person-plural form,
// and conjugates identically to "sie"/they in every tense in this app).
var ImperativePersons = []Person{Tu, Nous, Vous, Il}

// PersonsFor returns the valid persons to draw from for the given tense:
// ImperativePersons for Imperativ, a single placeholder person for
// invariable tenses (the stored form is identical for all six — see
// IsInvariable), or ConjugablePersons otherwise.
func PersonsFor(t TenseID) []Person {
	switch {
	case t == Imperativ:
		return ImperativePersons
	case IsInvariable(t):
		return []Person{Je}
	default:
		return ConjugablePersons
	}
}

// Group identifies which traditional class of German verb a verb belongs
// to. This drives which pattern is shown as theory. Unlike French's three
// groups or Spanish's three conjugation classes, German pedagogy
// additionally singles out sein/haben/werden as fully irregular, and modal
// verbs (können, ...) as their own small, structurally distinct class (a
// special present-tense paradigm with no -e/-st ending in ich/er, and a
// "double infinitive" construction in the Perfekt when paired with another
// verb — simplified away in this app; see können's IrregularNotes).
type Group string

const (
	GroupWeak      Group = "schwach"      // regelmäßig: machen, spielen, ...
	GroupStrong    Group = "stark"        // Ablaut: sprechen→sprach→gesprochen
	GroupMixed     Group = "gemischt"     // Ablaut + schwache Endungen: denken→dachte→gedacht
	GroupIrregular Group = "unregelmäßig" // sein, haben, werden
	GroupModal     Group = "Modalverb"    // können, ...
)

// Verb is a single verb with its hand-authored simple-tense conjugation
// table, plus the metadata (auxiliary, past participle) needed to derive
// every compound/nicht-finite form on demand — see compound.go.
type Verb struct {
	Infinitive    string
	Group         Group
	TranslationEn string
	// Auxiliary is "haben" or "sein": which one this verb uses to form
	// every compound tense. sein itself uses "sein" as its own auxiliary
	// (ich bin gewesen); haben likewise uses "haben" (ich habe gehabt).
	Auxiliary string
	// PastParticiple is the invariable Partizip II, e.g. "gesprochen",
	// "gegangen". Unlike French, it never agrees with the subject — German
	// participles in compound tenses are always invariable, like Spanish.
	PastParticiple string
	// HasImperative is false for verbs with no natural imperative (modal
	// verbs such as können), excluded from Imperativ exercises.
	HasImperative bool
	// Contexts are short sentence completions (no leading/trailing space,
	// no final punctuation) used to build natural exercise sentences, e.g.
	// "mit ihrer Schwester am Telefon" for sprechen.
	Contexts []string

	Prasens             Conjugation
	Prateritum          Conjugation
	KonjunktivIIPrasens Conjugation
	KonjunktivIPrasens  Conjugation
	Imperativ           Conjugation // only Tu/Nous/Vous/Il populated
}

// ConjugationFor returns v's form for the given tense and person, and
// whether that combination exists. Simple tenses are read directly off v;
// every other tense is derived — see compound.go. For a compound tense,
// the returned string is ONLY the finite part (see the package doc comment
// above) — use NonfiniteTail for the invariant clause-final remainder.
func (v Verb) ConjugationFor(t TenseID, p Person) (string, bool) {
	switch t {
	case Prasens:
		return v.Prasens.Get(p), v.Prasens.Get(p) != ""
	case Prateritum:
		return v.Prateritum.Get(p), v.Prateritum.Get(p) != ""
	case KonjunktivIIPrasens:
		return v.KonjunktivIIPrasens.Get(p), v.KonjunktivIIPrasens.Get(p) != ""
	case KonjunktivIPrasens:
		return v.KonjunktivIPrasens.Get(p), v.KonjunktivIPrasens.Get(p) != ""
	case Imperativ:
		form := v.Imperativ.Get(p)
		return form, form != ""
	default:
		return derivedConjugationFor(v, t, p)
	}
}
