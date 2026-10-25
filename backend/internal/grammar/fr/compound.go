package fr

import "strings"

// findVerb looks up a verb in the package dataset by infinitive. It panics if
// not found: the four verbs it's used for (avoir, être, aller, venir) must
// always be present in Verbs — see verbs.go — since every derived tense
// depends on at least one of them.
func findVerb(infinitive string) Verb {
	for _, v := range Verbs {
		if v.Infinitive == infinitive {
			return v
		}
	}
	panic("fr: required auxiliary/helper verb not found in dataset: " + infinitive)
}

// auxVerb returns v's auxiliary verb (avoir or être), whose own simple-tense
// tables are reused to build every compound tense of v.
func (v Verb) auxVerb() Verb {
	return findVerb(v.Auxiliary)
}

func (v Verb) isEtre() bool {
	return v.Auxiliary == "être"
}

// presentParticipleOverrides holds the two verbs whose participe présent
// isn't the regular "(nous-form stem without -ons) + ant".
var presentParticipleOverrides = map[string]string{
	"avoir": "ayant",
	"être":  "étant",
}

// presentParticiple returns v's participe présent (used for the gérondif),
// e.g. parlant, finissant, prenant — or the irregular ayant/étant.
func presentParticiple(v Verb) string {
	if override, ok := presentParticipleOverrides[v.Infinitive]; ok {
		return override
	}
	return strings.TrimSuffix(v.Present.Nous, "ons") + "ant"
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

// composePersonal combines an auxiliary's conjugation (in whichever simple
// tense the caller passed) with v's past participle into a full
// compound-tense Conjugation. With être, the plural persons (nous/vous/ils)
// get the agreement -s on the participle — the simplified default
// (masculine) agreement documented on the passé composé theory page.
func composePersonal(aux Conjugation, participle string, isEtre bool) Conjugation {
	singular, plural := participle, participle
	if isEtre {
		plural = participle + "s"
	}
	return Conjugation{
		Je:   aux.Je + " " + singular,
		Tu:   aux.Tu + " " + singular,
		Il:   aux.Il + " " + singular,
		Nous: aux.Nous + " " + plural,
		Vous: aux.Vous + " " + plural,
		Ils:  aux.Ils + " " + plural,
	}
}

// composeImperative is composePersonal's equivalent for the three
// imperative persons (Tu/Nous/Vous; Je/Il/Ils stay empty).
func composeImperative(aux Conjugation, participle string, isEtre bool) Conjugation {
	plural := participle
	if isEtre {
		plural = participle + "s"
	}
	return Conjugation{
		Tu:   aux.Tu + " " + participle,
		Nous: aux.Nous + " " + plural,
		Vous: aux.Vous + " " + plural,
	}
}

// derivedConjugationFor computes every tense that isn't a hand-authored
// simple tense stored directly on v (see Verb.ConjugationFor, which handles
// those and delegates everything else here).
func derivedConjugationFor(v Verb, t TenseID, p Person) (string, bool) {
	switch t {
	case PasseCompose:
		c := composePersonal(v.auxVerb().Present, v.PastParticiple, v.isEtre())
		return c.Get(p), true

	case PlusQueParfait:
		c := composePersonal(v.auxVerb().Imparfait, v.PastParticiple, v.isEtre())
		return c.Get(p), true

	case FuturAnterieur:
		c := composePersonal(v.auxVerb().FutureSimple, v.PastParticiple, v.isEtre())
		return c.Get(p), true

	case ConditionnelPasse:
		c := composePersonal(v.auxVerb().ConditionnelPresent, v.PastParticiple, v.isEtre())
		return c.Get(p), true

	case SubjonctifPasse:
		c := composePersonal(v.auxVerb().SubjonctifPresent, v.PastParticiple, v.isEtre())
		return c.Get(p), true

	case ImperatifPasse:
		if !v.HasImperative {
			return "", false
		}
		imp := composeImperative(v.auxVerb().ImperatifPresent, v.PastParticiple, v.isEtre())
		form := imp.Get(p)
		return form, form != ""

	case PresentContinu:
		etre := findVerb("être")
		c := buildForEachPerson(func(person Person) string {
			return etre.Present.Get(person) + " en train de " + v.Infinitive
		})
		return c.Get(p), true

	case PasseRecent:
		venir := findVerb("venir")
		connector := "de "
		if needsElision(v.Infinitive) {
			connector = "d'"
		}
		c := buildForEachPerson(func(person Person) string {
			return venir.Present.Get(person) + " " + connector + v.Infinitive
		})
		return c.Get(p), true

	case FuturProche:
		aller := findVerb("aller")
		c := buildForEachPerson(func(person Person) string {
			return aller.Present.Get(person) + " " + v.Infinitive
		})
		return c.Get(p), true

	case InfinitifPasse:
		return v.auxVerb().Infinitive + " " + v.PastParticiple, true

	case GerondifPresent:
		return "en " + presentParticiple(v), true

	case GerondifPasse:
		return presentParticiple(v.auxVerb()) + " " + v.PastParticiple, true

	default:
		return "", false
	}
}
