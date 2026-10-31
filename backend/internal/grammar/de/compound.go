package de

// findVerb looks up a verb in the package dataset by infinitive. It panics
// if not found: "haben", "sein", and "werden" must always be present in
// Verbs — see verbs.go — since every derived tense depends on at least one
// of them.
func findVerb(infinitive string) Verb {
	for _, v := range Verbs {
		if v.Infinitive == infinitive {
			return v
		}
	}
	panic("de: required auxiliary/helper verb not found in dataset: " + infinitive)
}

// auxVerb returns v's auxiliary verb (haben or sein), whose own simple-tense
// tables are reused to build every compound tense of v.
func (v Verb) auxVerb() Verb {
	return findVerb(v.Auxiliary)
}

// partizipIOverrides holds the one verb in this dataset whose Partizip I
// isn't the regular "infinitive + d": sein's regular form would be "seind",
// but the attested form inserts an extra e, "seiend".
var partizipIOverrides = map[string]string{
	"sein": "seiend",
}

// partizipI returns v's Partizip I (present participle), e.g. "sprechend",
// "machend" — or the irregular "seiend".
func partizipI(v Verb) string {
	if override, ok := partizipIOverrides[v.Infinitive]; ok {
		return override
	}
	return v.Infinitive + "d"
}

// derivedConjugationFor computes every tense that isn't a hand-authored
// simple tense stored directly on v (see Verb.ConjugationFor, which handles
// those and delegates everything else here).
//
// Unlike fr/es, a compound tense here returns ONLY its finite part — see
// the package doc comment in model.go for why German's clause-final word
// order makes that necessary. Callers that need the invariant clause-final
// remainder (the participle, the infinitive, or both) call NonfiniteTail
// separately — see exercise.go's buildSentence.
func derivedConjugationFor(v Verb, t TenseID, p Person) (string, bool) {
	switch t {
	case Perfekt:
		form := v.auxVerb().Prasens.Get(p)
		return form, form != ""

	case Plusquamperfekt:
		form := v.auxVerb().Prateritum.Get(p)
		return form, form != ""

	case FuturI, FuturII:
		form := findVerb("werden").Prasens.Get(p)
		return form, form != ""

	case KonjunktivIIPerfekt:
		form := v.auxVerb().KonjunktivIIPrasens.Get(p)
		return form, form != ""

	case KonjunktivIPerfekt:
		form := v.auxVerb().KonjunktivIPrasens.Get(p)
		return form, form != ""

	case InfinitivPerfekt:
		// "... gesprochen zu haben" / "... gegangen zu sein" — a single
		// invariable form, no person agreement at all.
		return v.PastParticiple + " zu " + v.Auxiliary, true

	case PartizipI:
		return partizipI(v), true

	default:
		return "", false
	}
}

// NonfiniteTail returns the invariant, clause-final chunk of text that
// completes a compound tense's meaning once its finite part (returned by
// Verb.ConjugationFor) has been conjugated and placed earlier in the
// sentence, e.g. "gesprochen" for Perfekt, or "gesprochen haben" for Futur
// II. Returns "" for every tense with no such clause-final remainder
// (simple tenses, Imperativ, and the nicht-finite forms, which are each a
// single complete, invariable answer already).
func NonfiniteTail(v Verb, t TenseID) string {
	switch t {
	case Perfekt, Plusquamperfekt, KonjunktivIIPerfekt, KonjunktivIPerfekt:
		return v.PastParticiple
	case FuturI:
		return v.Infinitive
	case FuturII:
		return v.PastParticiple + " " + v.Auxiliary
	default:
		return ""
	}
}
