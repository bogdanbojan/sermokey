package es

import (
	"math/rand/v2"
	"testing"
)

func TestKnownConjugations(t *testing.T) {
	byInf := map[string]Verb{}
	for _, v := range Verbs {
		byInf[v.Infinitive] = v
	}

	cases := []struct {
		infinitive string
		tense      TenseID
		person     Person
		want       string
	}{
		{"ser", Present, Je, "soy"},
		{"ser", Preterite, Je, "fui"},
		{"ir", Preterite, Je, "fui"}, // ser and ir share the preterite
		{"ir", Imperfect, Je, "iba"}, // ir's imperfecto is irregular, not "ía"
		{"estar", Preterite, Il, "estuvo"},
		{"tener", Present, Je, "tengo"},
		{"tener", FutureSimple, Je, "tendré"},
		{"hacer", Preterite, Il, "hizo"},
		{"hacer", ImperativePresent, Tu, "haz"},
		{"poder", FutureSimple, Je, "podré"},
		{"querer", FutureSimple, Je, "querré"},
		{"buscar", Preterite, Je, "busqué"}, // c→qu before é
		{"llegar", Preterite, Je, "llegué"}, // g→gu before é
		{"buscar", SubjunctivePresent, Nous, "busquemos"},
		{"pensar", Present, Nous, "pensamos"},            // unstressed stem, no e→ie
		{"pensar", Present, Je, "pienso"},                // stressed stem, e→ie
		{"volver", SubjunctivePresent, Nous, "volvamos"}, // unstressed, no o→ue
		{"pedir", Preterite, Il, "pidió"},                // -ir stem-changer: 3rd person only
		{"pedir", Preterite, Je, "pedí"},                 // yo unaffected
		{"pedir", SubjunctivePresent, Nous, "pidamos"},   // -ir stem-changer: keeps change even in nosotros
	}

	for _, tc := range cases {
		v, ok := byInf[tc.infinitive]
		if !ok {
			t.Fatalf("verb %q not found in dataset", tc.infinitive)
		}
		got, ok := v.ConjugationFor(tc.tense, tc.person)
		if !ok {
			t.Errorf("%s/%s/%s: ConjugationFor returned ok=false", tc.infinitive, tc.tense, tc.person)
			continue
		}
		if got != tc.want {
			t.Errorf("%s/%s/%s: got %q, want %q", tc.infinitive, tc.tense, tc.person, got, tc.want)
		}
	}
}

func TestDerivedCompoundTenses(t *testing.T) {
	byInf := map[string]Verb{}
	for _, v := range Verbs {
		byInf[v.Infinitive] = v
	}

	cases := []struct {
		infinitive string
		tense      TenseID
		person     Person
		want       string
	}{
		// Compound tenses: haber-only, participle never agrees.
		{"hablar", PresentPerfect, Je, "he hablado"},
		{"hablar", PresentPerfect, Ils, "han hablado"}, // no agreement, unlike French être-verbs
		{"volver", PresentPerfect, Il, "ha vuelto"},    // irregular participle
		{"hacer", Pluperfect, Nous, "habíamos hecho"},
		{"hablar", FuturePerfect, Je, "habré hablado"},
		{"hablar", ConditionalPerfect, Tu, "habrías hablado"},
		{"hablar", SubjunctivePerfect, Nous, "hayamos hablado"},

		// Periphrastic constructions.
		{"hablar", PresentProgressive, Je, "estoy hablando"},
		{"pedir", PresentProgressive, Il, "está pidiendo"}, // stem-changing gerundio
		{"hablar", RecentPast, Tu, "acabas de hablar"},
		{"hablar", NearFuture, Nous, "vamos a hablar"},

		// Modos impersonales (invariable).
		{"hablar", PastInfinitive, Je, "haber hablado"},
		{"volver", PastInfinitive, Je, "haber vuelto"},
		{"hablar", Gerund, Je, "hablando"},
		{"pedir", Gerund, Je, "pidiendo"},
		{"poder", Gerund, Je, "pudiendo"},
		{"ir", Gerund, Je, "yendo"},
		{"hablar", PerfectGerund, Je, "habiendo hablado"},
	}

	for _, tc := range cases {
		v, ok := byInf[tc.infinitive]
		if !ok {
			t.Fatalf("verb %q not found in dataset", tc.infinitive)
		}
		got, ok := v.ConjugationFor(tc.tense, tc.person)
		if !ok {
			t.Errorf("%s/%s/%s: ConjugationFor returned ok=false", tc.infinitive, tc.tense, tc.person)
			continue
		}
		if got != tc.want {
			t.Errorf("%s/%s/%s: got %q, want %q", tc.infinitive, tc.tense, tc.person, got, tc.want)
		}
	}
}

func TestHaberAndPoderHaveNoImperative(t *testing.T) {
	byInf := map[string]Verb{}
	for _, v := range Verbs {
		byInf[v.Infinitive] = v
	}
	for _, inf := range []string{"haber", "poder"} {
		v := byInf[inf]
		if v.HasImperative {
			t.Errorf("%s should have HasImperative=false", inf)
		}
		if _, ok := v.ConjugationFor(ImperativePresent, Tu); ok {
			t.Errorf("%s should have no imperativo-afirmativo form", inf)
		}
	}
}

func TestImperativeUsesFiveSlots(t *testing.T) {
	v := Verbs[0] // hablar
	for _, p := range ImperativePersons {
		if form, ok := v.ConjugationFor(ImperativePresent, p); !ok || form == "" {
			t.Errorf("hablar: missing imperativo-afirmativo/%s", p)
		}
	}
	if form, ok := v.ConjugationFor(ImperativePresent, Je); ok || form != "" {
		t.Errorf("hablar: imperativo-afirmativo should have no yo form, got %q (ok=%v)", form, ok)
	}
}

func TestInvariableTensesIgnorePerson(t *testing.T) {
	byInf := map[string]Verb{}
	for _, v := range Verbs {
		byInf[v.Infinitive] = v
	}
	v := byInf["hablar"]
	for _, tense := range []TenseID{PastInfinitive, Gerund, PerfectGerund} {
		if !IsInvariable(tense) {
			t.Errorf("%s should be marked invariable", tense)
		}
		want, ok := v.ConjugationFor(tense, Je)
		if !ok || want == "" {
			t.Fatalf("%s: no form for hablar/yo", tense)
		}
		for _, p := range ConjugablePersons {
			got, ok := v.ConjugationFor(tense, p)
			if !ok || got != want {
				t.Errorf("%s/%s: got %q, want %q (same form for every person)", tense, p, got, want)
			}
		}
	}
}

func TestEveryVerbHasCompleteData(t *testing.T) {
	for _, v := range Verbs {
		if len(v.Contexts) == 0 {
			t.Errorf("%s: no sentence contexts", v.Infinitive)
		}
		if v.PastParticiple == "" {
			t.Errorf("%s: missing past participle", v.Infinitive)
		}
		if v.Gerundio == "" {
			t.Errorf("%s: missing gerundio", v.Infinitive)
		}
		for _, tense := range AllTenses {
			if tense == ImperativePresent {
				continue // checked separately below, different person set
			}
			for _, p := range PersonsFor(tense) {
				if form, ok := v.ConjugationFor(tense, p); !ok || form == "" {
					t.Errorf("%s: missing %s/%s", v.Infinitive, tense, p)
				}
			}
		}
		if v.HasImperative {
			for _, p := range ImperativePersons {
				if form, ok := v.ConjugationFor(ImperativePresent, p); !ok || form == "" {
					t.Errorf("%s: missing imperativo-afirmativo/%s", v.Infinitive, p)
				}
			}
		}
	}
}

func TestGenerateExercisesBasic(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	exercises, err := GenerateExercises("mixed", "mixed", 25, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exercises) != 25 {
		t.Fatalf("got %d exercises, want 25", len(exercises))
	}
	for _, ex := range exercises {
		if ex.Answer == "" {
			t.Errorf("exercise %s has empty answer", ex.ID)
		}
		if ex.Type == MultipleChoice {
			if len(ex.Choices) != 4 {
				t.Errorf("exercise %s: got %d choices, want 4", ex.ID, len(ex.Choices))
			}
			found := false
			for _, c := range ex.Choices {
				if c == ex.Answer {
					found = true
				}
			}
			if !found {
				t.Errorf("exercise %s: answer %q not among choices %v", ex.ID, ex.Answer, ex.Choices)
			}
		}
	}
}

func TestGenerateExercisesRejectsUnknownTense(t *testing.T) {
	rng := rand.New(rand.NewPCG(5, 6))
	if _, err := GenerateExercises("not-a-tense", "mcq", 5, rng); err == nil {
		t.Error("expected error for unknown tense, got nil")
	}
}

func TestPoolSizeAtLeast100PerTense(t *testing.T) {
	for _, tense := range AllTenses {
		if got := PoolSize(tense); got < 100 {
			t.Errorf("tense %s: pool size %d, want at least 100 distinct sentences", tense, got)
		}
	}
}
