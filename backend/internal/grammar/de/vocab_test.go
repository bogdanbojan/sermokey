package de

import (
	"strings"
	"testing"

	"github.com/bogdanbojan/sermokey/backend/internal/grammar"
)

// categoryBudget is the target size of each vocabulary category, chosen by
// communicative load rather than split evenly — see VocabCategories' doc
// comment in vocab.go.
var categoryBudget = map[string]int{
	"verben":       210,
	"funktion":     150,
	"adjektive":    150,
	"abstrakt":     130,
	"essen":        115,
	"haus":         95,
	"zeit":         90,
	"arbeit":       75,
	"gefuehle":     75,
	"orte":         70,
	"freizeit":     70,
	"natur":        68,
	"schule":       65,
	"tiere":        65,
	"gesellschaft": 65,
	"medien":       62,
	"berufe":       58,
	"geld":         58,
	"gesundheit":   58,
	"stadt":        58,
	"verkehr":      55,
	"kleidung":     53,
	"familie":      48,
	"zahlen":       48,
	"masse":        45,
	"koerper":      44,
	"farben":       20,
}

// mustHaveLemmas is the closed-class guarantee for German. Every item sits
// in roughly the top 300 lemmas by frequency (cf. the Goethe-Institut
// A1/A2/B1 Wortlisten), so a list omitting any of them is mis-targeted no
// matter how many rare words it carries.
//
// Nouns are omitted here on purpose: German noun lemmas in this dataset
// carry their article, which makes them awkward to assert by bare form,
// and the highest-frequency closed classes in German are all non-nouns
// anyway.
var mustHaveLemmas = []string{
	// Interrogatives.
	"was", "wer", "wie", "warum", "wo", "wann", "welcher", "wohin", "woher",
	// Negative and indefinite pronouns.
	"nichts", "niemand", "etwas", "jemand", "jeder", "alle", "andere", "beide",
	// Personal pronouns and possessives.
	"ich", "du", "er", "sie", "es", "wir", "ihr", "man", "mein", "dein", "sein",
	"unser", "dieser",
	// Conjunctions and subordinators — the backbone of German clause order.
	"und", "oder", "aber", "denn", "weil", "dass", "wenn", "als", "ob",
	"sondern", "damit", "obwohl", "während", "bevor", "nachdem", "falls",
	// Discourse adverbs.
	"auch", "schon", "noch", "erst", "immer", "nie", "oft", "manchmal",
	"vielleicht", "natürlich", "wirklich", "eigentlich", "leider", "gern",
	"trotzdem", "deshalb", "also", "sogar", "fast", "nur", "sehr", "zu",
	// Core prepositions.
	"in", "an", "auf", "unter", "über", "vor", "nach", "bei", "mit", "ohne",
	"für", "gegen", "durch", "von", "zwischen", "neben", "seit", "bis",
	// The top verbs.
	"sein", "haben", "werden", "können", "müssen", "wollen", "sollen",
	"dürfen", "mögen", "machen", "gehen", "kommen", "sehen", "sprechen",
	"sagen", "fragen", "antworten", "wissen", "kennen", "denken", "glauben",
	"verstehen", "lernen", "arbeiten", "spielen", "essen", "trinken",
	"schlafen", "wohnen", "leben", "lieben", "suchen", "finden", "kaufen",
	"geben", "nehmen", "bringen", "brauchen", "helfen", "bleiben", "fahren",
	"laufen", "lesen", "schreiben", "hören", "warten", "stehen", "liegen",
	"halten", "tragen", "zeigen", "heißen", "lassen", "gehören", "treffen",
	"versuchen", "vergessen", "verlieren", "gewinnen", "erklären", "bekommen",
}

func TestVocabDataIsWellFormed(t *testing.T) {
	for _, err := range grammar.ValidateVocab(grammar.VocabCheck{
		Lang:            "de",
		Words:           Vocabulary,
		Categories:      VocabCategories,
		Budget:          categoryBudget,
		BudgetTolerance: 0.30,
		MustHave:        mustHaveLemmas,
	}) {
		t.Error(err)
	}
}

// TestNounsCarryTheirArticle enforces the German-only lemma convention: a
// noun is taught together with its article, and the Gender column must
// agree with it. Plural-only nouns (die Eltern, die Kosten) take "die" with
// an empty gender, since they have no singular gender to learn.
func TestNounsCarryTheirArticle(t *testing.T) {
	articleGender := map[string]string{"der ": "m", "die ": "f", "das ": "n"}
	for _, w := range Vocabulary {
		if w.POS != "noun" {
			if strings.HasPrefix(w.Lemma, "der ") || strings.HasPrefix(w.Lemma, "die ") ||
				strings.HasPrefix(w.Lemma, "das ") {
				t.Errorf("%q: non-noun (pos=%s) should not carry an article", w.Lemma, w.POS)
			}
			continue
		}
		var matched bool
		for article, gender := range articleGender {
			if !strings.HasPrefix(w.Lemma, article) {
				continue
			}
			matched = true
			// "die X" is ambiguous: feminine singular or any plural.
			if article == "die " && w.Gender == "" {
				break // plural-only noun, fine
			}
			if w.Gender != gender {
				t.Errorf("%q: article %q implies gender %q but column says %q",
					w.Lemma, strings.TrimSpace(article), gender, w.Gender)
			}
			break
		}
		if !matched {
			t.Errorf("%q: noun lemma must start with der/die/das", w.Lemma)
		}
	}
}

func TestVocabSizeInB2Range(t *testing.T) {
	if n := len(Vocabulary); n < 1800 || n > 2800 {
		t.Errorf("de: %d words, want 1800..2800", n)
	}
}
