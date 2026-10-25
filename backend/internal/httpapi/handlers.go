package httpapi

import (
	"encoding/json"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"

	"github.com/bogdanbojan/sermokey/backend/internal/grammar"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("httpapi: failed to encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorDTO{Error: message})
}

// languageFromRequest resolves the ?lang= query param to a languageProvider,
// writing a 400 response and returning ok=false if it names an unknown
// language.
func languageFromRequest(w http.ResponseWriter, r *http.Request) (languageProvider, bool) {
	code := r.URL.Query().Get("lang")
	lang, ok := resolveLanguage(code)
	if !ok {
		writeError(w, http.StatusBadRequest, "unknown language: "+code)
		return nil, false
	}
	return lang, true
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleLanguages(w http.ResponseWriter, r *http.Request) {
	out := make([]LanguageDTO, 0, len(languages))
	for _, l := range languages {
		out = append(out, LanguageDTO{Code: l.Code(), Name: l.Name()})
	}
	writeJSON(w, http.StatusOK, out)
}

func handleTenses(w http.ResponseWriter, r *http.Request) {
	lang, ok := languageFromRequest(w, r)
	if !ok {
		return
	}
	tenses := lang.AllTenses()
	out := make([]TenseSummaryDTO, 0, len(tenses))
	for _, id := range tenses {
		th, _ := lang.Theory(id)
		out = append(out, TenseSummaryDTO{
			ID:         string(id),
			Category:   string(lang.TenseCategory(id)),
			NameNative: th.NameNative,
			NameEn:     th.NameEn,
			Usage:      th.Usage,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func handleTheory(w http.ResponseWriter, r *http.Request) {
	lang, ok := languageFromRequest(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	th, ok := lang.Theory(grammar.TenseID(id))
	if !ok {
		writeError(w, http.StatusNotFound, "unknown tense: "+id)
		return
	}
	writeJSON(w, http.StatusOK, toTheoryDTO(th, lang.TenseCategory(th.ID)))
}

func toTheoryDTO(th grammar.TenseTheory, category grammar.Category) TheoryDTO {
	groups := make([]GroupEndingsDTO, 0, len(th.Groups))
	for _, g := range th.Groups {
		groups = append(groups, GroupEndingsDTO{
			GroupName:    g.GroupName,
			Explanation:  g.Explanation,
			Endings:      g.Endings,
			ExampleVerb:  g.ExampleVerb,
			ExampleForms: g.ExampleForms,
		})
	}
	return TheoryDTO{
		ID:             string(th.ID),
		Category:       string(category),
		NameNative:     th.NameNative,
		NameEn:         th.NameEn,
		Usage:          th.Usage,
		Formation:      th.Formation,
		Groups:         groups,
		IrregularNotes: th.IrregularNotes,
	}
}

func handleVerbs(w http.ResponseWriter, r *http.Request) {
	lang, ok := languageFromRequest(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, lang.Verbs())
}

func handleExercises(w http.ResponseWriter, r *http.Request) {
	lang, ok := languageFromRequest(w, r)
	if !ok {
		return
	}

	q := r.URL.Query()

	tense := q.Get("tense")
	if tense == "" {
		tense = "mixed"
	}
	exerciseType := q.Get("type")
	if exerciseType == "" {
		exerciseType = "mixed"
	}

	count := 10
	if raw := q.Get("count"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "count must be an integer")
			return
		}
		count = n
	}

	// Seed a fresh request-scoped RNG from the auto-seeded global source
	// (math/rand/v2's top-level functions are seeded securely at process
	// start), so each request gets an independent, non-reproducible draw.
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))

	exercises, err := lang.GenerateExercises(tense, exerciseType, count, rng)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out := make([]ExerciseDTO, 0, len(exercises))
	for _, ex := range exercises {
		out = append(out, ExerciseDTO{
			ID:             ex.ID,
			Tense:          string(ex.Tense),
			TenseName:      ex.TenseName,
			Type:           string(ex.Type),
			Infinitive:     ex.Infinitive,
			TranslationEn:  ex.TranslationEn,
			PersonLabel:    ex.PersonLabel,
			SentencePrefix: ex.SentencePrefix,
			SentenceSuffix: ex.SentenceSuffix,
			Choices:        ex.Choices,
			Answer:         ex.Answer,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func handleVocabCategories(w http.ResponseWriter, r *http.Request) {
	lang, ok := languageFromRequest(w, r)
	if !ok {
		return
	}
	counts := make(map[string]int)
	for _, word := range lang.VocabWords() {
		counts[word.Category]++
	}
	cats := lang.VocabCategories()
	out := make([]VocabCategoryDTO, 0, len(cats))
	for _, c := range cats {
		out = append(out, VocabCategoryDTO{
			ID:         c.ID,
			NameNative: c.NameNative,
			NameEn:     c.NameEn,
			WordCount:  counts[c.ID],
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func handleVocabulary(w http.ResponseWriter, r *http.Request) {
	lang, ok := languageFromRequest(w, r)
	if !ok {
		return
	}
	category := r.URL.Query().Get("category")

	words := lang.VocabWords()
	out := make([]VocabWordDTO, 0, len(words))
	for _, word := range words {
		if category != "" && word.Category != category {
			continue
		}
		out = append(out, VocabWordDTO{
			ID:            word.ID,
			Lemma:         word.Lemma,
			POS:           word.POS,
			Gender:        word.Gender,
			TranslationEn: word.TranslationEn,
			Category:      word.Category,
			Tier:          word.Tier,
		})
	}
	writeJSON(w, http.StatusOK, out)
}
