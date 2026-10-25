package fr

import (
	"fmt"
	"math/rand/v2"
	"unicode"

	"github.com/bogdanbojan/sermokey/backend/internal/grammar"
)

// Exercise is a re-export of the shared exercise shape, so httpapi can
// consume French and Spanish exercises identically.
type Exercise = grammar.Exercise

const (
	MultipleChoice = grammar.MultipleChoice
	FreeText       = grammar.FreeText
)

// subjonctifTriggers are common impersonal/personal phrases that naturally
// introduce a subjonctif clause with "que". They never need to agree with
// the exercise's target person, since they belong to an independent main
// clause.
var subjonctifTriggers = []string{
	"Il faut",
	"Il est important",
	"Il vaudrait mieux",
	"Je voudrais",
	"Elle souhaite",
}

// needsElision reports whether a word beginning with this letter triggers
// French elision ("je" -> "j'", "que" -> "qu'") — i.e. it starts with a
// vowel sound (a, e, i, o, u, mute h, y).
func needsElision(s string) bool {
	if s == "" {
		return false
	}
	r := unicode.ToLower([]rune(s)[0])
	switch r {
	case 'a', 'e', 'i', 'o', 'u', 'h', 'y',
		'à', 'â', 'é', 'è', 'ê', 'ë', 'î', 'ï', 'ô', 'ù', 'û', 'ü':
		return true
	default:
		return false
	}
}

func personLabel(p Person) string {
	switch p {
	case Je:
		return "je"
	case Tu:
		return "tu"
	case Il:
		return "il / elle / on"
	case Nous:
		return "nous"
	case Vous:
		return "vous"
	case Ils:
		return "ils / elles"
	default:
		return ""
	}
}

// personLabelFor is personLabel, except invariable tenses (no subject
// agreement at all) always report "forme invariable" instead of whichever
// placeholder Person happened to be drawn internally.
func personLabelFor(tense TenseID, p Person) string {
	if IsInvariable(tense) {
		return "forme invariable"
	}
	return personLabel(p)
}

// pronounPrefix builds the subject + trailing space (or elided apostrophe,
// no space) that starts a non-subjonctif, non-imperative sentence.
func pronounPrefix(p Person, answer string) string {
	switch p {
	case Je:
		if needsElision(answer) {
			return "J'"
		}
		return "Je "
	case Tu:
		return "Tu "
	case Il:
		return "Il/Elle/On "
	case Nous:
		return "Nous "
	case Vous:
		return "Vous "
	case Ils:
		return "Ils/Elles "
	default:
		return ""
	}
}

// quePrefix builds the "que"/"qu'" + subject pronoun used before a
// subjonctif clause. "que" itself only elides before il/ils (vowel-initial
// pronouns); "je" never elides with "que" but the verb that follows it may
// still elide with "je".
func quePrefix(p Person, answer string) string {
	switch p {
	case Je:
		if needsElision(answer) {
			return "que j'"
		}
		return "que je "
	case Tu:
		return "que tu "
	case Il:
		return "qu'il/elle/on "
	case Nous:
		return "que nous "
	case Vous:
		return "que vous "
	case Ils:
		return "qu'ils/elles "
	default:
		return ""
	}
}

func buildSentence(tense TenseID, person Person, answer, context string, rng *rand.Rand) (prefix, suffix string) {
	switch tense {
	case ImperatifPresent, ImperatifPasse:
		return "", " " + context + " !"
	case SubjonctifPresent, SubjonctifPasse:
		trigger := subjonctifTriggers[rng.IntN(len(subjonctifTriggers))]
		return trigger + " " + quePrefix(person, answer), " " + context + "."
	case InfinitifPasse:
		// "Après avoir parlé ..." / "Après être allé ..." — the infinitif
		// passé is almost always introduced by "après".
		return "Après ", " " + context + "."
	case GerondifPresent, GerondifPasse:
		// The answer itself is the full invariable phrase ("en parlant",
		// "ayant parlé"); it reads naturally as a sentence-initial clause.
		return "", " " + context + "."
	default:
		return pronounPrefix(person, answer), " " + context + "."
	}
}

// buildChoices returns 4 shuffled unique answer options including the
// correct one. Distractors are preferentially other persons' forms of the
// *same verb* (across any tense), which keeps the question focused on
// person/tense agreement rather than vocabulary; if the dataset can't
// supply enough unique distractors from the same verb it falls back to
// other verbs' forms for the same tense and person.
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
// exercise type.
//
// tense is either "mixed" (draw from every tense) or one TenseID value.
// exerciseType is "mixed" (random per-question), "mcq", or "text".
// rng must be non-nil; callers that need determinism (tests) can pass a
// seeded *rand.Rand, production code can use rand.New(rand.NewPCG(seed...))
// seeded from crypto randomness or time.
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
		if t == ImperatifPresent || t == ImperatifPasse {
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
		prefix, suffix := buildSentence(t, person, answer, context, rng)

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
		if (tense == ImperatifPresent || tense == ImperatifPasse) && !v.HasImperative {
			continue
		}
		total += len(v.Contexts) * len(persons)
	}
	return total
}
