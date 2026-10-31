package de

import (
	"math/rand/v2"
	"strings"
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
		{"sein", Prasens, Je, "bin"},
		{"sein", Prasens, Vous, "seid"},
		{"haben", Prasens, Tu, "hast"},
		{"werden", Prasens, Tu, "wirst"},
		{"sprechen", Prasens, Tu, "sprichst"},
		{"sprechen", Prasens, Il, "spricht"},
		{"nehmen", Prasens, Tu, "nimmst"},
		{"fahren", Prasens, Tu, "fährst"},
		{"essen", Prateritum, Tu, "aßt"},
		{"lesen", Prateritum, Tu, "last"},
		{"denken", Prateritum, Je, "dachte"},
		{"machen", Prasens, Nous, "machen"},

		// Konjunktiv I Präsens: always the plain infinitive stem + weak
		// endings, regardless of how irregular the Präsens is.
		{"sprechen", KonjunktivIPrasens, Il, "spreche"},
		{"haben", KonjunktivIPrasens, Tu, "habest"},
		{"sein", KonjunktivIPrasens, Il, "sei"},
		{"sein", KonjunktivIPrasens, Nous, "seien"},

		// Konjunktiv II Präsens: synthetic for sein/haben/werden/modals,
		// "würde + Infinitiv" for everything else.
		{"sein", KonjunktivIIPrasens, Je, "wäre"},
		{"können", KonjunktivIIPrasens, Il, "könnte"},
		{"machen", KonjunktivIIPrasens, Nous, "würden machen"},

		// Imperativ: du/wir/ihr/Sie (Sie borrows the Il slot).
		{"sprechen", Imperativ, Tu, "sprich"},
		{"sein", Imperativ, Tu, "sei"},
		{"sein", Imperativ, Il, "seien"},
		{"haben", Imperativ, Tu, "hab"},

		// Derived compound tenses (compound.go) — ONLY the finite part; see
		// TestNonfiniteTail below for the clause-final remainder.
		{"sprechen", Perfekt, Je, "habe"}, // haben: haben-Verb
		{"gehen", Perfekt, Je, "bin"},     // sein: sein-Verb
		{"gehen", Perfekt, Nous, "sind"},
		{"machen", Plusquamperfekt, Je, "hatte"},
		{"gehen", Plusquamperfekt, Il, "war"},
		{"machen", FuturI, Il, "wird"},
		{"gehen", FuturII, Nous, "werden"},
		{"machen", KonjunktivIIPerfekt, Je, "hätte"},
		{"gehen", KonjunktivIIPerfekt, Je, "wäre"},
		{"machen", KonjunktivIPerfekt, Il, "habe"},
		{"gehen", KonjunktivIPerfekt, Il, "sei"},

		// Nicht-finite Formen (invariable — no person agreement at all).
		{"machen", InfinitivPerfekt, Je, "gemacht zu haben"},
		{"gehen", InfinitivPerfekt, Je, "gegangen zu sein"},
		{"machen", InfinitivPerfekt, Ils, "gemacht zu haben"}, // same form regardless of "person"
		{"sprechen", PartizipI, Je, "sprechend"},
		{"sein", PartizipI, Je, "seiend"},
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

func TestNonfiniteTail(t *testing.T) {
	byInf := map[string]Verb{}
	for _, v := range Verbs {
		byInf[v.Infinitive] = v
	}
	machen, gehen := byInf["machen"], byInf["gehen"]

	cases := []struct {
		v    Verb
		t    TenseID
		want string
	}{
		{machen, Perfekt, "gemacht"},
		{gehen, Perfekt, "gegangen"},
		{machen, Plusquamperfekt, "gemacht"},
		{machen, FuturI, "machen"},
		{machen, FuturII, "gemacht haben"},
		{gehen, FuturII, "gegangen sein"},
		{machen, KonjunktivIIPerfekt, "gemacht"},
		{machen, KonjunktivIPerfekt, "gemacht"},
		{machen, Prasens, ""}, // simple tenses have no clause-final remainder
		{machen, Imperativ, ""},
	}
	for _, tc := range cases {
		if got := NonfiniteTail(tc.v, tc.t); got != tc.want {
			t.Errorf("NonfiniteTail(%s, %s) = %q, want %q", tc.v.Infinitive, tc.t, got, tc.want)
		}
	}
}

func TestKoennenHasNoImperative(t *testing.T) {
	for _, v := range Verbs {
		if v.Infinitive == "können" {
			if v.HasImperative {
				t.Errorf("können should have HasImperative=false")
			}
			if _, ok := v.ConjugationFor(Imperativ, Tu); ok {
				t.Errorf("können should have no imperativ form")
			}
			return
		}
	}
	t.Fatal("können not found in dataset")
}

func TestInvariableTensesIgnorePerson(t *testing.T) {
	byInf := map[string]Verb{}
	for _, v := range Verbs {
		byInf[v.Infinitive] = v
	}
	v := byInf["machen"]
	for _, tense := range []TenseID{InfinitivPerfekt, PartizipI} {
		if !IsInvariable(tense) {
			t.Errorf("%s should be marked invariable", tense)
		}
		want, ok := v.ConjugationFor(tense, Je)
		if !ok || want == "" {
			t.Fatalf("%s: no form for machen/je", tense)
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
		for _, tense := range AllTenses {
			if tense == Imperativ {
				continue // checked separately below, since persons differ
			}
			for _, p := range PersonsFor(tense) {
				if form, ok := v.ConjugationFor(tense, p); !ok || form == "" {
					t.Errorf("%s: missing %s/%s", v.Infinitive, tense, p)
				}
			}
		}
		if v.HasImperative {
			for _, p := range ImperativePersons {
				if form, ok := v.ConjugationFor(Imperativ, p); !ok || form == "" {
					t.Errorf("%s: missing imperativ/%s", v.Infinitive, p)
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
			seen := map[string]int{}
			for _, c := range ex.Choices {
				seen[c]++
				if c == ex.Answer {
					found = true
				}
			}
			if !found {
				t.Errorf("exercise %s: answer %q not among choices %v", ex.ID, ex.Answer, ex.Choices)
			}
			for c, n := range seen {
				if n > 1 {
					t.Errorf("exercise %s: choice %q duplicated", ex.ID, c)
				}
			}
		}
		if ex.Type == FreeText && len(ex.Choices) != 0 {
			t.Errorf("exercise %s: text exercise should not have choices", ex.ID)
		}
	}
}

func TestGenerateExercisesSingleTense(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	exercises, err := GenerateExercises(string(Imperativ), "mcq", 10, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, ex := range exercises {
		if ex.Tense != Imperativ {
			t.Errorf("got tense %s, want %s", ex.Tense, Imperativ)
		}
		if ex.SentencePrefix != "" {
			t.Errorf("imperativ exercise should have empty prefix, got %q", ex.SentencePrefix)
		}
		if !strings.HasSuffix(ex.SentenceSuffix, "!") {
			t.Errorf("imperativ exercise suffix should end with '!', got %q", ex.SentenceSuffix)
		}
	}
}

func TestGenerateExercisesRejectsUnknownTense(t *testing.T) {
	rng := rand.New(rand.NewPCG(5, 6))
	if _, err := GenerateExercises("not-a-tense", "mcq", 5, rng); err == nil {
		t.Error("expected error for unknown tense, got nil")
	}
}

func TestGenerateExercisesClampsCount(t *testing.T) {
	rng := rand.New(rand.NewPCG(7, 8))
	exercises, err := GenerateExercises("mixed", "text", 1000, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exercises) != 30 {
		t.Fatalf("got %d exercises, want clamped to 30", len(exercises))
	}
}

func TestPoolSizeAtLeast100PerTense(t *testing.T) {
	for _, tense := range AllTenses {
		if got := PoolSize(tense); got < 100 {
			t.Errorf("tense %s: pool size %d, want at least 100 distinct sentences", tense, got)
		}
	}
}
