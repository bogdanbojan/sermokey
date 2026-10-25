package fr

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
		{"être", Present, Je, "suis"},
		{"être", Present, Vous, "êtes"},
		{"avoir", SubjonctifPresent, Je, "aie"},
		{"aller", SubjonctifPresent, Je, "aille"},
		{"aller", SubjonctifPresent, Nous, "allions"},
		{"aller", FutureSimple, Je, "irai"},
		{"faire", Present, Ils, "font"},
		{"faire", Present, Vous, "faites"},
		{"vouloir", SubjonctifPresent, Nous, "voulions"},
		{"venir", SubjonctifPresent, Je, "vienne"},
		{"venir", PasseCompose, Nous, "sommes venus"},
		{"parler", PasseCompose, Je, "ai parlé"},
		{"manger", Present, Nous, "mangeons"},
		{"manger", Imparfait, Nous, "mangions"},
		{"finir", Present, Nous, "finissons"},
		{"prendre", SubjonctifPresent, Ils, "prennent"},
		{"prendre", SubjonctifPresent, Nous, "prenions"},

		// Derived compound tenses (compound.go) — spot-check against
		// hand-verified French rather than trusting the derivation blindly.
		{"parler", FuturAnterieur, Je, "aurai parlé"},
		{"aller", FuturAnterieur, Nous, "serons allés"},
		{"parler", ConditionnelPasse, Il, "aurait parlé"},
		{"aller", ConditionnelPasse, Ils, "seraient allés"},
		{"parler", SubjonctifPasse, Nous, "ayons parlé"},
		{"aller", SubjonctifPasse, Je, "sois allé"},
		{"parler", ImperatifPasse, Tu, "aie parlé"},
		{"aller", ImperatifPasse, Vous, "soyez allés"},

		// Periphrastic constructions.
		{"parler", PresentContinu, Je, "suis en train de parler"},
		{"parler", PresentContinu, Ils, "sont en train de parler"},
		{"parler", PasseRecent, Tu, "viens de parler"},
		{"aller", PasseRecent, Je, "viens d'aller"}, // elision: aller starts with a vowel
		{"parler", FuturProche, Nous, "allons parler"},

		// Modes impersonnels (invariable — no person agreement at all).
		{"parler", InfinitifPasse, Je, "avoir parlé"},
		{"parler", InfinitifPasse, Ils, "avoir parlé"}, // same form regardless of "person"
		{"aller", InfinitifPasse, Je, "être allé"},
		{"parler", GerondifPresent, Je, "en parlant"},
		{"finir", GerondifPresent, Je, "en finissant"},
		{"être", GerondifPresent, Je, "en étant"},  // irregular participe présent
		{"avoir", GerondifPresent, Je, "en ayant"}, // irregular participe présent
		{"parler", GerondifPasse, Je, "ayant parlé"},
		{"aller", GerondifPasse, Je, "étant allé"},
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

func TestPouvoirHasNoImperative(t *testing.T) {
	for _, v := range Verbs {
		if v.Infinitive == "pouvoir" {
			if v.HasImperative {
				t.Errorf("pouvoir should have HasImperative=false")
			}
			if _, ok := v.ConjugationFor(ImperatifPresent, Tu); ok {
				t.Errorf("pouvoir should have no imperatif-present form")
			}
			if _, ok := v.ConjugationFor(ImperatifPasse, Tu); ok {
				t.Errorf("pouvoir should have no imperatif-passe form either")
			}
			return
		}
	}
	t.Fatal("pouvoir not found in dataset")
}

func TestInvariableTensesIgnorePerson(t *testing.T) {
	byInf := map[string]Verb{}
	for _, v := range Verbs {
		byInf[v.Infinitive] = v
	}
	v := byInf["parler"]
	for _, tense := range []TenseID{InfinitifPasse, GerondifPresent, GerondifPasse} {
		if !IsInvariable(tense) {
			t.Errorf("%s should be marked invariable", tense)
		}
		want, ok := v.ConjugationFor(tense, Je)
		if !ok || want == "" {
			t.Fatalf("%s: no form for parler/je", tense)
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
			if tense == ImperatifPresent || tense == ImperatifPasse {
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
				if form, ok := v.ConjugationFor(ImperatifPresent, p); !ok || form == "" {
					t.Errorf("%s: missing imperatif-present/%s", v.Infinitive, p)
				}
				if form, ok := v.ConjugationFor(ImperatifPasse, p); !ok || form == "" {
					t.Errorf("%s: missing imperatif-passe/%s", v.Infinitive, p)
				}
			}
		}
	}
}

func TestNeedsElision(t *testing.T) {
	cases := map[string]bool{
		"ai": true, "étais": true, "habite": true, "irai": true,
		"suis": false, "pars": false, "fais": false,
	}
	for word, want := range cases {
		if got := needsElision(word); got != want {
			t.Errorf("needsElision(%q) = %v, want %v", word, got, want)
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
	exercises, err := GenerateExercises(string(ImperatifPresent), "mcq", 10, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, ex := range exercises {
		if ex.Tense != ImperatifPresent {
			t.Errorf("got tense %s, want %s", ex.Tense, ImperatifPresent)
		}
		if ex.SentencePrefix != "" {
			t.Errorf("imperative exercise should have empty prefix, got %q", ex.SentencePrefix)
		}
		if !strings.HasSuffix(ex.SentenceSuffix, " !") {
			t.Errorf("imperative exercise suffix should end with ' !', got %q", ex.SentenceSuffix)
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

func TestSubjonctifElisionRules(t *testing.T) {
	// "que" elides before il/ils (vowel-initial pronoun) but never before je;
	// the verb after "que je" may still elide with je itself.
	if p := quePrefix(Il, "aille"); p != "qu'il/elle/on " {
		t.Errorf("quePrefix(Il) = %q", p)
	}
	if p := quePrefix(Je, "aille"); p != "que j'" {
		t.Errorf("quePrefix(Je, vowel) = %q", p)
	}
	if p := quePrefix(Je, "finisse"); p != "que je " {
		t.Errorf("quePrefix(Je, consonant) = %q", p)
	}
}
