package es

// findVerb looks up a verb in the package dataset by infinitive. It panics
// if not found: "haber", "estar", "acabar", and "ir" must always be present
// in Verbs — see verbs.go — since every derived tense depends on at least
// one of them.
func findVerb(infinitive string) Verb {
	for _, v := range Verbs {
		if v.Infinitive == infinitive {
			return v
		}
	}
	panic("es: required auxiliary/helper verb not found in dataset: " + infinitive)
}

// setPerson returns a copy of c with the given person's slot set to form.
func setPerson(c Conjugation, p Person, form string) Conjugation {
	switch p {
	case Je:
		c.Je = form
	case Tu:
		c.Tu = form
	case Il:
		c.Il = form
	case Nous:
		c.Nous = form
	case Vous:
		c.Vous = form
	case Ils:
		c.Ils = form
	}
	return c
}

// buildForEachPerson fills all six slots of a Conjugation from a per-person
// formatter function, used by the periphrastic constructions below.
func buildForEachPerson(format func(Person) string) Conjugation {
	var c Conjugation
	for _, p := range ConjugablePersons {
		c = setPerson(c, p, format(p))
	}
	return c
}

// composeCompound combines haber's conjugation (in whichever simple tense
// the caller passed) with v's past participle into a full compound-tense
// Conjugation. Unlike French, the participle never agrees with the subject
// here — Spanish compound tenses with haber always use the invariable
// participle.
func composeCompound(auxiliar Conjugation, participle string) Conjugation {
	return Conjugation{
		Je:   auxiliar.Je + " " + participle,
		Tu:   auxiliar.Tu + " " + participle,
		Il:   auxiliar.Il + " " + participle,
		Nous: auxiliar.Nous + " " + participle,
		Vous: auxiliar.Vous + " " + participle,
		Ils:  auxiliar.Ils + " " + participle,
	}
}

// derivedConjugationFor computes every tense that isn't a hand-authored
// simple tense stored directly on v (see Verb.ConjugationFor, which handles
// those and delegates everything else here).
func derivedConjugationFor(v Verb, t TenseID, p Person) (string, bool) {
	haber := findVerb("haber")

	switch t {
	case PresentPerfect:
		c := composeCompound(haber.Present, v.PastParticiple)
		return c.Get(p), true

	case Pluperfect:
		c := composeCompound(haber.Imperfect, v.PastParticiple)
		return c.Get(p), true

	case FuturePerfect:
		c := composeCompound(haber.FutureSimple, v.PastParticiple)
		return c.Get(p), true

	case ConditionalPerfect:
		c := composeCompound(haber.ConditionalPresent, v.PastParticiple)
		return c.Get(p), true

	case SubjunctivePerfect:
		c := composeCompound(haber.SubjunctivePresent, v.PastParticiple)
		return c.Get(p), true

	case PresentProgressive:
		estar := findVerb("estar")
		c := buildForEachPerson(func(person Person) string {
			return estar.Present.Get(person) + " " + v.Gerundio
		})
		return c.Get(p), true

	case RecentPast:
		acabar := findVerb("acabar")
		c := buildForEachPerson(func(person Person) string {
			return acabar.Present.Get(person) + " de " + v.Infinitive
		})
		return c.Get(p), true

	case NearFuture:
		ir := findVerb("ir")
		c := buildForEachPerson(func(person Person) string {
			return ir.Present.Get(person) + " a " + v.Infinitive
		})
		return c.Get(p), true

	case PastInfinitive:
		return "haber " + v.PastParticiple, true

	case Gerund:
		return v.Gerundio, true

	case PerfectGerund:
		return "habiendo " + v.PastParticiple, true

	default:
		return "", false
	}
}
