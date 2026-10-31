package de

import (
	"fmt"
	"math/rand/v2"

	"github.com/bogdanbojan/sermokey/backend/internal/grammar"
)

// Exercise is a re-export of the shared exercise shape, so httpapi can
// consume French, Spanish, and German exercises identically.
type Exercise = grammar.Exercise

const (
	MultipleChoice = grammar.MultipleChoice
	FreeText       = grammar.FreeText
)

// indirekteRedeTriggers are common reporting phrases that naturally
// introduce Konjunktiv I indirekte Rede (reported speech) WITHOUT a "dass"
// conjunction. This deliberately avoids the "dass"-clause variant, which
// would push the finite verb to the very end of the clause (German's other
// major verb-position rule, alongside the compound-tense bracket — see the
// package doc comment in model.go) and so would need the same
// finite/non-finite split NonfiniteTail exists for, just triggered by a
// different conjunction. The "dass"-free variant used here is also the
// more common register in contemporary journalistic German.
var indirekteRedeTriggers = []string{
	"Er sagt,",
	"Sie behauptet,",
	"Man vermutet,",
	"Es heißt,",
	"Die Zeitung schreibt,",
}

// personLabel is the plain subject-pronoun label for a person, independent
// of tense. sie/Sie share one label because they are the same verb form
// (see Conjugation's doc comment in model.go).
func personLabel(p Person) string {
	switch p {
	case Je:
		return "ich"
	case Tu:
		return "du"
	case Il:
		return "er / sie / es"
	case Nous:
		return "wir"
	case Vous:
		return "ihr"
	case Ils:
		return "sie / Sie"
	default:
		return ""
	}
}

// lowerPronoun is personLabel's single-word, lowercase form, used mid-
// sentence (e.g. right after an indirekte-Rede trigger's comma).
func lowerPronoun(p Person) string {
	switch p {
	case Je:
		return "ich"
	case Tu:
		return "du"
	case Il:
		return "er"
	case Nous:
		return "wir"
	case Vous:
		return "ihr"
	case Ils:
		return "sie"
	default:
		return ""
	}
}

// personLabelFor is personLabel, except invariable tenses (no subject
// agreement at all) always report "unveränderliche Form", and the
// Imperativ's du/wir/ihr/Sie persons get their actual imperative pronoun
// instead of the generic (and here wrong) "er / sie / es" for the formal
// Sie slot.
func personLabelFor(tense TenseID, p Person) string {
	if IsInvariable(tense) {
		return "unveränderliche Form"
	}
	if tense == Imperativ {
		switch p {
		case Tu:
			return "du"
		case Nous:
			return "wir"
		case Vous:
			return "ihr"
		case Il:
			return "Sie"
		}
	}
	return personLabel(p)
}

// pronounPrefix builds the subject pronoun + trailing space that starts a
// plain declarative sentence. German never elides a subject pronoun (no
// French-style "je"→"j'"), so this is simpler than French's equivalent.
func pronounPrefix(p Person) string {
	switch p {
	case Je:
		return "Ich "
	case Tu:
		return "Du "
	case Il:
		return "Er/Sie/Es "
	case Nous:
		return "Wir "
	case Vous:
		return "Ihr "
	case Ils:
		return "Sie "
	default:
		return ""
	}
}

// tailFor returns NonfiniteTail prefixed with a space (ready to splice into
// a sentence suffix), or "" when the tense has no clause-final remainder.
func tailFor(v Verb, t TenseID) string {
	tail := NonfiniteTail(v, t)
	if tail == "" {
		return ""
	}
	return " " + tail
}

// buildSentence returns the text surrounding the blank for one exercise.
// Most tenses are a plain "Pronoun [blank] context." declarative, with the
// compound tenses' invariant participle/infinitive appended at the end of
// context via tailFor — exactly where German grammar requires it (see the
// package doc comment in model.go). A handful of tenses need their own
// shape:
//
//   - Imperativ fronts the verb with no subject pronoun, like French/
//     Spanish — except the formal Sie-form, which (uniquely among German's
//     four imperative persons) keeps its pronoun, placed right after the
//     verb: "Sprechen [Sie] bitte langsamer!".
//   - Konjunktiv I (Präsens and Perfekt) is wrapped in a "dass"-free
//     reported-speech frame — see indirekteRedeTriggers.
//   - Infinitiv Perfekt is introduced by "ohne" ("without having done ..."),
//     with the context placed BEFORE the invariant "... zu haben/sein"
//     answer, since it too sits at the end of its (subordinate-ish) clause.
//   - Partizip I stands alone as a sentence-initial adverbial fragment,
//     exactly as fr/es already do for their equivalent non-finite forms.
func buildSentence(tense TenseID, person Person, verb Verb, context string, rng *rand.Rand) (prefix, suffix string) {
	switch tense {
	case Imperativ:
		if person == Il {
			return "", " Sie " + context + "!"
		}
		return "", " " + context + "!"
	case KonjunktivIPrasens:
		trigger := indirekteRedeTriggers[rng.IntN(len(indirekteRedeTriggers))]
		return trigger + " " + lowerPronoun(person) + " ", " " + context + "."
	case KonjunktivIPerfekt:
		trigger := indirekteRedeTriggers[rng.IntN(len(indirekteRedeTriggers))]
		return trigger + " " + lowerPronoun(person) + " ", " " + context + tailFor(verb, tense) + "."
	case InfinitivPerfekt:
		return "Ohne " + context + " ", "."
	case PartizipI:
		return "", " " + context + "."
	default:
		return pronounPrefix(person), " " + context + tailFor(verb, tense) + "."
	}
}

// buildChoices returns 4 shuffled unique answer options including the
// correct one. Distractors are preferentially other persons' forms of the
// *same verb* (across any tense), which keeps the question focused on
// person/tense agreement rather than vocabulary; if the dataset can't
// supply enough unique distractors from the same verb it falls back to
// other verbs' forms for the same tense and person. For compound tenses,
// where the answer is the finite auxiliary alone (see the package doc
// comment in model.go), this naturally pulls distractors like "hatte" or
// "wäre" for an answer of "habe" — testing exactly the auxiliary-choice and
// mood/tense selection that matters most for those tenses.
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
// identical between all three languages.
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
		if t == Imperativ {
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
		prefix, suffix := buildSentence(t, person, verb, context, rng)

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
// combinations GenerateExercises can draw from for the given tense — i.e.
// how many genuinely different exercise sentences exist before any start
// repeating. Used to guarantee a minimum amount of variety per tense (see
// exercise_test.go).
func PoolSize(tense TenseID) int {
	persons := PersonsFor(tense)
	total := 0
	for _, v := range Verbs {
		if tense == Imperativ && !v.HasImperative {
			continue
		}
		total += len(v.Contexts) * len(persons)
	}
	return total
}
