package es

import (
	"fmt"
	"math/rand/v2"

	"github.com/bogdanbojan/sermokey/backend/internal/grammar"
)

// Exercise is a re-export of the shared exercise shape, so httpapi can
// consume French and Spanish exercises identically.
type Exercise = grammar.Exercise

const (
	MultipleChoice = grammar.MultipleChoice
	FreeText       = grammar.FreeText
)

// subjunctiveTriggers are common phrases that naturally introduce a
// subjuntivo clause with "que". Unlike French, "que" never contracts in
// Spanish — there is no elision to handle at all here.
var subjunctiveTriggers = []string{
	"Espero",
	"Es importante",
	"Quiero",
	"Dudo",
	"Ojalá",
}

func personLabel(p Person) string {
	switch p {
	case Je:
		return "yo"
	case Tu:
		return "tú"
	case Il:
		return "él / ella / usted"
	case Nous:
		return "nosotros/as"
	case Vous:
		return "vosotros/as"
	case Ils:
		return "ellos / ellas / ustedes"
	default:
		return ""
	}
}

// personLabelFor is personLabel, except invariable tenses report "forma
// invariable", and the imperativo's Il/Ils slots — conjugated like 3rd
// person but addressing "you" formally — are labelled "usted"/"ustedes"
// instead of the ambiguous "él / ella / usted" / "ellos / ellas / ustedes".
func personLabelFor(tense TenseID, p Person) string {
	if IsInvariable(tense) {
		return "forma invariable"
	}
	if tense == ImperativePresent {
		switch p {
		case Il:
			return "usted"
		case Ils:
			return "ustedes"
		}
	}
	return personLabel(p)
}

// pronounPrefix builds the subject pronoun + trailing space that starts a
// non-subjunctive, non-imperative sentence. Spanish never elides, so this is
// simpler than French's equivalent.
func pronounPrefix(p Person) string {
	switch p {
	case Je:
		return "Yo "
	case Tu:
		return "Tú "
	case Il:
		return "Él/Ella/Usted "
	case Nous:
		return "Nosotros "
	case Vous:
		return "Vosotros "
	case Ils:
		return "Ellos/Ellas/Ustedes "
	default:
		return ""
	}
}

// quePrefix builds the "que" + subject pronoun used before a subjuntivo
// clause. "que" never contracts in Spanish, unlike French.
func quePrefix(p Person) string {
	switch p {
	case Je:
		return "que yo "
	case Tu:
		return "que tú "
	case Il:
		return "que él/ella/usted "
	case Nous:
		return "que nosotros "
	case Vous:
		return "que vosotros "
	case Ils:
		return "que ellos/ellas/ustedes "
	default:
		return ""
	}
}

func buildSentence(tense TenseID, person Person, context string, rng *rand.Rand) (prefix, suffix string) {
	switch tense {
	case ImperativePresent:
		// Spanish wraps exclamations in inverted + regular marks: ¡...!
		return "¡", " " + context + "!"
	case SubjunctivePresent, SubjunctivePerfect:
		trigger := subjunctiveTriggers[rng.IntN(len(subjunctiveTriggers))]
		return trigger + " " + quePrefix(person), " " + context + "."
	case PastInfinitive:
		// "Después de haber hablado ..." — the infinitivo compuesto is
		// almost always introduced by "después de".
		return "Después de ", " " + context + "."
	case Gerund, PerfectGerund:
		// The answer itself is the full invariable phrase ("hablando",
		// "habiendo hablado"); it reads naturally as a sentence-initial
		// clause, with no "en"-style prefix (unlike French).
		return "", " " + context + "."
	default:
		return pronounPrefix(person), " " + context + "."
	}
}

// buildChoices returns 4 shuffled unique answer options including the
// correct one, preferring other persons' forms of the same verb before
// falling back to other verbs.
func buildChoices(answer string, tense TenseID, person Person, verb Verb, allVerbs []Verb, rng *rand.Rand) []string {
	seen := map[string]bool{answer: true}
	var pool []string

	addIfNew := func(form string) {
		if form == "" || seen[form] {
			return
		}
		seen[form] = true
		pool = append(pool, form)
	}

	for _, t := range AllTenses {
		for _, p := range PersonsFor(t) {
			if form, ok := verb.ConjugationFor(t, p); ok {
				addIfNew(form)
			}
		}
	}

	rng.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })

	choices := []string{answer}
	for _, f := range pool {
		if len(choices) == 4 {
			break
		}
		choices = append(choices, f)
	}

	if len(choices) < 4 {
		others := make([]Verb, len(allVerbs))
		copy(others, allVerbs)
		rng.Shuffle(len(others), func(i, j int) { others[i], others[j] = others[j], others[i] })
		for _, v2 := range others {
			if v2.Infinitive == verb.Infinitive {
				continue
			}
			if form, ok := v2.ConjugationFor(tense, person); ok && !seen[form] {
				seen[form] = true
				choices = append(choices, form)
			}
			if len(choices) == 4 {
				break
			}
		}
	}

	rng.Shuffle(len(choices), func(i, j int) { choices[i], choices[j] = choices[j], choices[i] })
	return choices
}

// GenerateExercises builds `count` exercises for the requested tense and
// exercise type. See fr.GenerateExercises for the parameter contract — it's
// identical between languages.
func GenerateExercises(tense, exerciseType string, count int, rng *rand.Rand) ([]Exercise, error) {
	if count < 1 {
		count = 1
	}
	if count > 30 {
		count = 30
	}

	var candidateTenses []TenseID
	if tense == "mixed" || tense == "" {
		candidateTenses = AllTenses
	} else if IsValidTense(TenseID(tense)) {
		candidateTenses = []TenseID{TenseID(tense)}
	} else {
		return nil, fmt.Errorf("unknown tense %q", tense)
	}

	switch exerciseType {
	case "mixed", "mcq", "text", "":
	default:
		return nil, fmt.Errorf("unknown exercise type %q", exerciseType)
	}

	exercises := make([]Exercise, 0, count)
	for i := 0; i < count; i++ {
		t := candidateTenses[rng.IntN(len(candidateTenses))]

		persons := PersonsFor(t)

		eligible := Verbs
		if t == ImperativePresent {
			eligible = make([]Verb, 0, len(Verbs))
			for _, v := range Verbs {
				if v.HasImperative {
					eligible = append(eligible, v)
				}
			}
		}

		verb := eligible[rng.IntN(len(eligible))]
		person := persons[rng.IntN(len(persons))]
		answer, ok := verb.ConjugationFor(t, person)
		if !ok {
			continue
		}

		qType := exerciseType
		if qType == "mixed" || qType == "" {
			if rng.IntN(2) == 0 {
				qType = string(MultipleChoice)
			} else {
				qType = string(FreeText)
			}
		}

		context := verb.Contexts[rng.IntN(len(verb.Contexts))]
		prefix, suffix := buildSentence(t, person, context, rng)

		ex := Exercise{
			ID:             fmt.Sprintf("%s-%s-%s-%d", t, verb.Infinitive, person, rng.Int()),
			Tense:          t,
			TenseName:      Theories[t].NameNative,
			Type:           grammar.ExerciseType(qType),
			Infinitive:     verb.Infinitive,
			TranslationEn:  verb.TranslationEn,
			PersonLabel:    personLabelFor(t, person),
			SentencePrefix: prefix,
			SentenceSuffix: suffix,
			Answer:         answer,
		}
		if grammar.ExerciseType(qType) == MultipleChoice {
			ex.Choices = buildChoices(answer, t, person, verb, Verbs, rng)
		}
		exercises = append(exercises, ex)
	}

	return exercises, nil
}

// PoolSize returns the number of distinct (verb, context, person) sentence
// combinations GenerateExercises can draw from for the given tense.
func PoolSize(tense TenseID) int {
	persons := PersonsFor(tense)
	total := 0
	for _, v := range Verbs {
		if tense == ImperativePresent && !v.HasImperative {
			continue
		}
		total += len(v.Contexts) * len(persons)
	}
	return total
}
