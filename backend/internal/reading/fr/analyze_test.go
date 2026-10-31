package fr

import (
	"testing"

	"github.com/bogdanbojan/sermokey/backend/internal/reading"
)

func first(t *testing.T, surface string) reading.Analysis {
	t.Helper()
	as := Analyzer.Analyze(surface)
	if len(as) == 0 {
		t.Fatalf("Analyze(%q) returned nothing", surface)
	}
	return as[0]
}

func TestAnalyzeLemmas(t *testing.T) {
	tests := []struct{ surface, lemma, pos string }{
		{"maison", "maison", "noun"},
		{"Maison", "maison", "noun"}, // sentence-initial capital
		{"parler", "parler", "verb"},
		// "où" is a function word in the curriculum, and the curated
		// part of speech deliberately overrides the dictionary's finer
		// "adv" so the reader sees the same label as on its flashcard.
		{"où", "où", "other"},
	}
	for _, tc := range tests {
		t.Run(tc.surface, func(t *testing.T) {
			a := first(t, tc.surface)
			if a.Lemma != tc.lemma || a.POS != tc.pos {
				t.Errorf("got %s/%s, want %s/%s", a.Lemma, a.POS, tc.lemma, tc.pos)
			}
			if len(a.Senses) == 0 || a.Senses[0].Gloss == "" {
				t.Error("no gloss")
			}
		})
	}
}

// Deinflection is the whole point of the form index: a reader hovering
// "parlons" needs to be taken to parler, not told the word is unknown.
func TestAnalyzeInflectedForms(t *testing.T) {
	tests := []struct{ surface, lemma string }{
		{"parlons", "parler"},
		{"parlait", "parler"},
		{"suis", "être"},
		{"étaient", "être"},
		{"eu", "avoir"},
		{"ayant", "avoir"},
		{"vu", "voir"},
		{"allés", "aller"}, // two-hop chain: allés -> allé -> aller
		{"belles", "beau"},
		{"maisons", "maison"},
	}
	for _, tc := range tests {
		t.Run(tc.surface, func(t *testing.T) {
			a := first(t, tc.surface)
			if a.Lemma != tc.lemma {
				t.Errorf("Analyze(%q) resolved to %q, want %q", tc.surface, a.Lemma, tc.lemma)
			}
			if a.FormNote == "" {
				t.Errorf("Analyze(%q): no form note, so the UI cannot explain the inflection", tc.surface)
			}
		})
	}
}

// A surface that is both a headword and an inflection of another word
// ("mieux" is an adverb in its own right and the comparative of "bien")
// must offer both readings rather than silently picking one.
func TestAnalyzeOffersBothHeadwordAndInflection(t *testing.T) {
	as := Analyzer.Analyze("mieux")
	var sawSelf, sawBien bool
	for _, a := range as {
		switch a.Lemma {
		case "mieux":
			sawSelf = true
		case "bien":
			sawBien = true
		}
	}
	if !sawSelf || !sawBien {
		t.Errorf("Analyze(\"mieux\") gave %d readings (self=%v bien=%v), want both", len(as), sawSelf, sawBien)
	}
}

// "est" is simultaneously a curated noun ("east") and the single most
// common verb form in French. Ranking by curriculum tier must surface être
// first; getting this backwards would be a glaring error on almost every
// page of running text.
func TestAnalyzeRanksHomographsByCurriculumTier(t *testing.T) {
	as := Analyzer.Analyze("est")
	if len(as) < 2 {
		t.Fatalf("expected both readings of \"est\", got %d", len(as))
	}
	if as[0].Lemma != "être" {
		t.Errorf("first reading of \"est\" is %q, want \"être\"", as[0].Lemma)
	}
	var sawEast bool
	for _, a := range as {
		if a.Lemma == "est" {
			sawEast = true
		}
	}
	if !sawEast {
		t.Error("the noun reading of \"est\" was dropped entirely")
	}
}

// Words in the curriculum must resolve to the SRS id the vocabulary deck
// already uses. A mismatch would silently create a second card for the same
// word, splitting its review history.
func TestAnalyzeCuratedWordsCarryTheirDeckID(t *testing.T) {
	for _, tc := range []struct{ surface, wantID string }{
		{"maison", "maison|maison"},
		{"maisons", "maison|maison"}, // inflected form, same card
		{"parler", "parler|verbes"},
		{"parlons", "parler|verbes"},
	} {
		t.Run(tc.surface, func(t *testing.T) {
			a := first(t, tc.surface)
			if a.Source != reading.SourceCurated {
				t.Fatalf("source = %v, want curated", a.Source)
			}
			if a.WordID != tc.wantID {
				t.Errorf("WordID = %q, want %q", a.WordID, tc.wantID)
			}
		})
	}
}

// Elided clitics are their own tokens and must be glossable, or the reader
// sees an unexplained dead span at the start of very common words.
func TestAnalyzeElidedClitics(t *testing.T) {
	for _, surface := range []string{"l'", "qu'", "j'", "n'", "d'"} {
		if as := Analyzer.Analyze(surface); len(as) == 0 {
			t.Errorf("Analyze(%q) returned nothing", surface)
		}
	}
}

func TestAnalyzeUnknownReturnsNothing(t *testing.T) {
	for _, surface := range []string{"", "   ", "zzzqqqxx"} {
		if as := Analyzer.Analyze(surface); len(as) != 0 {
			t.Errorf("Analyze(%q) = %v, want nothing", surface, as)
		}
	}
}

// Multi-word entries matter because phrase selection in the reader looks up
// the joined surface, and the curated CSV itself contains items like
// "salle de sport".
func TestAnalyzeMultiWordEntries(t *testing.T) {
	for _, surface := range []string{"tout à coup", "salle de sport"} {
		if as := Analyzer.Analyze(surface); len(as) == 0 {
			t.Errorf("Analyze(%q) returned nothing", surface)
		}
	}
}

// French lexicalises past participles as nouns on a large scale, and for
// those the headword reading is normally the right one even though the
// verb behind it is the more essential word. "Le marché du samedi" is a
// market, not something walked.
//
// Finite forms must keep the opposite behaviour, which TestAnalyzeRanks-
// HomographsByCurriculumTier pins down for "est".
func TestAnalyzePrefersHeadwordOverLexicalisedParticiple(t *testing.T) {
	for _, tc := range []struct{ surface, wantLemma string }{
		{"marché", "marché"}, // vs. past participle of marcher
		{"entrée", "entrée"}, // vs. feminine past participle of entrer
		{"employé", "employé"},
	} {
		t.Run(tc.surface, func(t *testing.T) {
			as := Analyzer.Analyze(tc.surface)
			if len(as) == 0 {
				t.Fatalf("Analyze(%q) returned nothing", tc.surface)
			}
			if as[0].Lemma != tc.wantLemma {
				t.Errorf("primary reading of %q is %q, want %q", tc.surface, as[0].Lemma, tc.wantLemma)
			}
			// The verbal reading must remain reachable, so a reader can
			// correct the choice where the participle really was meant.
			if len(as) < 2 {
				t.Errorf("Analyze(%q) offers no alternative reading", tc.surface)
			}
		})
	}
}

// Closed-class words must be reachable through the normal analysis path,
// including from their elided and inflected forms, because that is how the
// reader sees them in running text. If "l'" or "suis" failed to inherit the
// flag, roughly a quarter of every page would be counted as unknown
// vocabulary and every text would be labelled too hard.
func TestAnalyzeMarksFunctionWords(t *testing.T) {
	for _, surface := range []string{"le", "la", "les", "l'", "du", "je", "il", "se", "ne", "que", "qu'", "de", "ce"} {
		t.Run(surface, func(t *testing.T) {
			as := Analyzer.Analyze(surface)
			if len(as) == 0 {
				t.Fatalf("Analyze(%q) returned nothing", surface)
			}
			if !as[0].Function {
				t.Errorf("Analyze(%q) -> %q is not marked as a function word", surface, as[0].Lemma)
			}
		})
	}
	// Content words must not be swept up by the same rule.
	for _, surface := range []string{"maison", "parler", "marché", "pendant", "beau"} {
		if as := Analyzer.Analyze(surface); len(as) > 0 && as[0].Function {
			t.Errorf("Analyze(%q) -> %q is wrongly marked as a function word", surface, as[0].Lemma)
		}
	}
}

// French capitalises no common noun mid-sentence, so a capitalised surface
// with a proper-noun reading is strong evidence of a name. That evidence
// has to outweigh the curriculum: "Claire" is a person, even though
// "clair" is a tier-1 curated adjective whose feminine form is spelled the
// same way.
func TestAnalyzeUsesCapitalisationForProperNouns(t *testing.T) {
	for _, tc := range []struct{ surface, wantLemma string }{
		{"Claire", "Claire"},
		{"Sophie", "Sophie"},
		{"Marseille", "Marseille"},
	} {
		t.Run(tc.surface, func(t *testing.T) {
			as := Analyzer.Analyze(tc.surface)
			if len(as) == 0 {
				t.Fatalf("Analyze(%q) returned nothing", tc.surface)
			}
			if as[0].Lemma != tc.wantLemma {
				t.Errorf("primary reading of %q is %q, want %q", tc.surface, as[0].Lemma, tc.wantLemma)
			}
			if !as[0].Proper {
				t.Errorf("%q is not marked as a proper noun", tc.surface)
			}
		})
	}

	// Lower case is the ordinary word, with no proper-noun promotion.
	as := Analyzer.Analyze("claire")
	if len(as) == 0 || as[0].Lemma != "clair" {
		t.Errorf("Analyze(\"claire\") = %v, want the adjective clair", as)
	}

	// Sentence-initial function words must be immune: Wiktionary records
	// surnames for "Le" and "Du", and promoting those would mangle the
	// commonest words in the language.
	for _, surface := range []string{"Le", "La", "Du", "Il", "Ce"} {
		as := Analyzer.Analyze(surface)
		if len(as) == 0 {
			t.Fatalf("Analyze(%q) returned nothing", surface)
		}
		if as[0].Proper {
			t.Errorf("Analyze(%q) -> %q was promoted to a proper noun", surface, as[0].Lemma)
		}
		if !as[0].Function {
			t.Errorf("Analyze(%q) -> %q lost its function-word flag", surface, as[0].Lemma)
		}
	}
}

// Capitalisation alone must not promote a surname over a common noun:
// every sentence-initial word is capitalised, and Wiktionary records
// surnames for a great many ordinary words.
func TestAnalyzeDoesNotPromoteSentenceInitialCommonNouns(t *testing.T) {
	for _, tc := range []struct{ surface, wantLemma string }{
		{"Maison", "maison"},
		{"Marché", "marché"},
		{"Chaque", "chaque"},
		{"Elle", "elle"},
	} {
		as := Analyzer.Analyze(tc.surface)
		if len(as) == 0 {
			t.Fatalf("Analyze(%q) returned nothing", tc.surface)
		}
		if as[0].Lemma != tc.wantLemma {
			t.Errorf("Analyze(%q) -> %q, want %q", tc.surface, as[0].Lemma, tc.wantLemma)
		}
		if as[0].Proper {
			t.Errorf("Analyze(%q) was wrongly promoted to a proper noun", tc.surface)
		}
	}
}

// French derives adjectives from past participles wholesale, which leaves
// a two-step chain between an inflected participle and the verb a learner
// actually has on a flashcard: "oubliée" -> "oublié" (adj, a real
// headword) -> "oublier" (tier-1 curated verb). Stopping at the adjective
// shows "forgotten" and quietly severs the link to the deck.
func TestAnalyzeFollowsParticipleChainsToTheCuratedVerb(t *testing.T) {
	for _, tc := range []struct{ surface, wantLemma, wantID string }{
		{"oubliée", "oublier", "oublier|verbes"},
		{"aimée", "aimer", "aimer|verbes"},
		{"oublié", "oublier", "oublier|verbes"},
	} {
		t.Run(tc.surface, func(t *testing.T) {
			a := first(t, tc.surface)
			if a.Lemma != tc.wantLemma {
				t.Errorf("Analyze(%q) -> %q, want %q", tc.surface, a.Lemma, tc.wantLemma)
			}
			if a.WordID != tc.wantID {
				t.Errorf("Analyze(%q) WordID = %q, want %q", tc.surface, a.WordID, tc.wantID)
			}
		})
	}

	// The extra hop must not override a headword the curriculum teaches in
	// its own right: "marché" is a curated noun, not something walked.
	if a := first(t, "marché"); a.Lemma != "marché" {
		t.Errorf("Analyze(\"marché\") -> %q, want the noun marché", a.Lemma)
	}
}

// A merged phrase has to be analysable in the inflected shape the text
// used, not only in its citation form. "se lève" is what appears on the
// page; "se lever" is what the dictionary lists, and no source conjugates
// reflexive verbs as multi-word entries.
func TestAnalyzeInflectedPhrases(t *testing.T) {
	for _, tc := range []struct{ surface, wantLemma string }{
		{"se lève", "se lever"},
		{"se couche", "se coucher"},
		{"pommes de terre", "pomme de terre"},
		{"salle de bains", "salle de bains"},
	} {
		t.Run(tc.surface, func(t *testing.T) {
			a := first(t, tc.surface)
			if a.Lemma != tc.wantLemma {
				t.Errorf("Analyze(%q) -> %q, want %q", tc.surface, a.Lemma, tc.wantLemma)
			}
			if len(a.Senses) == 0 {
				t.Errorf("Analyze(%q) has no gloss", tc.surface)
			}
		})
	}

	// An arbitrary word sequence is still not a phrase.
	if as := Analyzer.Analyze("le chien bleu"); len(as) != 0 {
		t.Errorf("Analyze(\"le chien bleu\") = %v, want nothing", as)
	}
}
