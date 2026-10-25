package httpapi

import (
	"math/rand/v2"

	"github.com/bogdanbojan/sermokey/backend/internal/grammar"
	"github.com/bogdanbojan/sermokey/backend/internal/grammar/es"
	"github.com/bogdanbojan/sermokey/backend/internal/grammar/fr"
)

// languageProvider is the small, consumer-side interface httpapi uses to
// treat fr and es uniformly. It is intentionally NOT implemented by the fr
// or es packages themselves (see fr's package doc comment for why those two
// packages are independent rather than built behind a shared abstraction) —
// each is wrapped here by a thin adapter that calls its package-level
// functions and maps its package-local Verb type to VerbSummaryDTO.
type languageProvider interface {
	Code() string
	Name() string
	AllTenses() []grammar.TenseID
	TenseCategory(grammar.TenseID) grammar.Category
	Theory(grammar.TenseID) (grammar.TenseTheory, bool)
	Verbs() []VerbSummaryDTO
	GenerateExercises(tense, exerciseType string, count int, rng *rand.Rand) ([]grammar.Exercise, error)
	VocabCategories() []grammar.VocabCategoryInfo
	VocabWords() []grammar.VocabWord
}

type frProvider struct{}

func (frProvider) Code() string                                      { return "fr" }
func (frProvider) Name() string                                      { return "Français" }
func (frProvider) AllTenses() []grammar.TenseID                      { return fr.AllTenses }
func (frProvider) TenseCategory(id grammar.TenseID) grammar.Category { return fr.TenseCategory[id] }
func (frProvider) Theory(id grammar.TenseID) (grammar.TenseTheory, bool) {
	th, ok := fr.Theories[id]
	return th, ok
}
func (frProvider) Verbs() []VerbSummaryDTO {
	out := make([]VerbSummaryDTO, 0, len(fr.Verbs))
	for _, v := range fr.Verbs {
		out = append(out, VerbSummaryDTO{Infinitive: v.Infinitive, Group: string(v.Group), TranslationEn: v.TranslationEn})
	}
	return out
}
func (frProvider) GenerateExercises(tense, exerciseType string, count int, rng *rand.Rand) ([]grammar.Exercise, error) {
	return fr.GenerateExercises(tense, exerciseType, count, rng)
}
func (frProvider) VocabCategories() []grammar.VocabCategoryInfo { return fr.VocabCategories }
func (frProvider) VocabWords() []grammar.VocabWord              { return fr.Vocabulary }

type esProvider struct{}

func (esProvider) Code() string                                      { return "es" }
func (esProvider) Name() string                                      { return "Español" }
func (esProvider) AllTenses() []grammar.TenseID                      { return es.AllTenses }
func (esProvider) TenseCategory(id grammar.TenseID) grammar.Category { return es.TenseCategory[id] }
func (esProvider) Theory(id grammar.TenseID) (grammar.TenseTheory, bool) {
	th, ok := es.Theories[id]
	return th, ok
}
func (esProvider) Verbs() []VerbSummaryDTO {
	out := make([]VerbSummaryDTO, 0, len(es.Verbs))
	for _, v := range es.Verbs {
		out = append(out, VerbSummaryDTO{Infinitive: v.Infinitive, Group: string(v.Group), TranslationEn: v.TranslationEn})
	}
	return out
}
func (esProvider) GenerateExercises(tense, exerciseType string, count int, rng *rand.Rand) ([]grammar.Exercise, error) {
	return es.GenerateExercises(tense, exerciseType, count, rng)
}
func (esProvider) VocabCategories() []grammar.VocabCategoryInfo { return es.VocabCategories }
func (esProvider) VocabWords() []grammar.VocabWord              { return es.Vocabulary }

// languages is the registry of every supported language, in the order shown
// by the GET /api/languages endpoint and used as the frontend's toggle
// order.
var languages = []languageProvider{frProvider{}, esProvider{}}

const defaultLanguageCode = "fr"

// resolveLanguage looks up a languageProvider by its "fr"/"es" code,
// defaulting to defaultLanguageCode when code is empty.
func resolveLanguage(code string) (languageProvider, bool) {
	if code == "" {
		code = defaultLanguageCode
	}
	for _, l := range languages {
		if l.Code() == code {
			return l, true
		}
	}
	return nil, false
}
