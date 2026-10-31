// Validation harness shared by the fr/es/de vocabulary datasets.
//
// The vocabulary lists are hand-curated content, which makes them exactly
// the kind of data that silently rots: a word gets added twice under two
// categories, a translation gains a comma and breaks the parser, a
// category drifts back toward padding with rare words, or a high-frequency
// function word quietly goes missing. None of that breaks a build, and
// none of it is visible without reading 2000+ lines.
//
// ValidateVocab is the regression guard. It returns problems rather than
// taking a *testing.T so this file stays free of a testing import (which
// would otherwise be linked into the production server binary); each
// language package calls it from its own test — see fr/vocab_test.go.
package grammar

import (
	"fmt"
	"strings"
)

// VocabCheck is one language's validation inputs.
type VocabCheck struct {
	// Lang is the language code, used only in failure messages.
	Lang string
	// Words is the parsed dataset under test.
	Words []VocabWord
	// Categories is the display registry the dataset must agree with.
	Categories []VocabCategoryInfo
	// Budget is the expected size of each category, keyed by category ID.
	// Actual counts may deviate by up to BudgetTolerance in either
	// direction — the point is to catch a category drifting an order of
	// magnitude off (the old "nombres" held 115 words, over half of them
	// quantifiers, measurements and verbs), not to freeze exact counts and
	// make every content edit a test edit.
	Budget map[string]int
	// BudgetTolerance is the allowed fractional deviation from Budget,
	// e.g. 0.30 for ±30%.
	BudgetTolerance float64
	// MustHave lists lemmas that must be present somewhere in the dataset,
	// regardless of category. This is the closed-class guarantee: the
	// pronouns, interrogatives, determiners, conjunctions and top verbs
	// that carry the most tokens in real speech. These are finite sets, so
	// "all of them are present" is a checkable property — unlike open
	// classes, where no such list could ever be complete.
	MustHave []string
}

var validPOS = map[string]bool{
	"noun": true, "verb": true, "adj": true, "adv": true, "other": true,
}

// validGender holds every accepted value of VocabWord.Gender. "n" is
// German-only (neuter); "" covers non-nouns and plural-only nouns such as
// "die Eltern", which have no singular gender to teach.
var validGender = map[string]bool{"m": true, "f": true, "n": true, "": true}

// ValidateVocab checks every structural invariant the vocabulary datasets
// must hold, returning one error per problem found (empty slice if clean).
// It is deliberately strict about the things that are cheap to get wrong by
// hand and expensive to notice later.
func ValidateVocab(c VocabCheck) []error {
	var errs []error
	report := func(format string, args ...any) {
		errs = append(errs, fmt.Errorf(format, args...))
	}

	known := make(map[string]bool, len(c.Categories))
	for _, cat := range c.Categories {
		if known[cat.ID] {
			report("%s: category %q declared twice in VocabCategories", c.Lang, cat.ID)
		}
		known[cat.ID] = true
	}

	seenID := make(map[string]string, len(c.Words))
	seenLemmaGloss := make(map[string]string, len(c.Words))
	counts := make(map[string]int, len(c.Categories))

	for _, w := range c.Words {
		where := fmt.Sprintf("%s: %q (%s)", c.Lang, w.Lemma, w.Category)

		// The CSV parser is a plain strings.Split on "," with no quoting
		// support, so a comma anywhere in the data silently shifts every
		// later field — or panics the whole package at init.
		for _, field := range []struct{ name, val string }{
			{"lemma", w.Lemma}, {"translation", w.TranslationEn},
			{"pos", w.POS}, {"category", w.Category},
		} {
			if strings.Contains(field.val, ",") {
				report("%s: %s contains a comma, which the CSV parser cannot represent", where, field.name)
			}
			if strings.TrimSpace(field.val) != field.val {
				report("%s: %s has leading/trailing whitespace", where, field.name)
			}
		}
		if w.Lemma == "" || w.TranslationEn == "" {
			report("%s: empty lemma or translation", where)
		}

		// ID is the SRS deck key. A collision means two different words
		// share review state and one of them is unreachable.
		if prev, dup := seenID[w.ID]; dup {
			report("%s: duplicate SRS id %q (previously %s)", where, w.ID, prev)
		}
		seenID[w.ID] = where

		// Same word + same meaning filed under two categories produces two
		// identical flashcards, which just wastes the learner's reviews.
		// Genuine homonyms differ in gloss and so pass.
		lg := w.Lemma + "\x00" + w.TranslationEn
		if prev, dup := seenLemmaGloss[lg]; dup {
			report("%s: %q/%q duplicated across categories (also %s)", where, w.Lemma, w.TranslationEn, prev)
		}
		seenLemmaGloss[lg] = where

		if !validPOS[w.POS] {
			report("%s: invalid pos %q", where, w.POS)
		}
		if !validGender[w.Gender] {
			report("%s: invalid gender %q", where, w.Gender)
		}
		if w.Gender != "" && w.POS != "noun" {
			report("%s: gender %q set on a non-noun (pos=%s)", where, w.Gender, w.POS)
		}
		if w.Tier < 1 || w.Tier > 3 {
			report("%s: tier %d out of range 1..3", where, w.Tier)
		}
		if !known[w.Category] {
			report("%s: category not declared in VocabCategories", where)
		}
		counts[w.Category]++
	}

	// Every declared category must actually contain words, or the Home
	// page renders an empty card.
	for _, cat := range c.Categories {
		if counts[cat.ID] == 0 {
			report("%s: category %q is declared but has no words", c.Lang, cat.ID)
		}
	}

	for id, want := range c.Budget {
		got := counts[id]
		lo := int(float64(want) * (1 - c.BudgetTolerance))
		hi := int(float64(want) * (1 + c.BudgetTolerance))
		if got < lo || got > hi {
			report("%s: category %q has %d words, want %d (±%.0f%%, i.e. %d..%d)",
				c.Lang, id, got, want, c.BudgetTolerance*100, lo, hi)
		}
	}
	for id := range counts {
		if _, ok := c.Budget[id]; !ok {
			report("%s: category %q has no budget entry", c.Lang, id)
		}
	}

	have := make(map[string]bool, len(c.Words))
	for _, w := range c.Words {
		have[w.Lemma] = true
	}
	var missing []string
	for _, lemma := range c.MustHave {
		if !have[lemma] {
			missing = append(missing, lemma)
		}
	}
	if len(missing) > 0 {
		report("%s: %d required high-frequency lemma(s) missing: %s",
			c.Lang, len(missing), strings.Join(missing, " "))
	}

	return errs
}
