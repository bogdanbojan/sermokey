package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func doRequest(t *testing.T, handler http.Handler, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestHandleLanguages(t *testing.T) {
	rec := doRequest(t, NewRouter(), http.MethodGet, "/api/languages")
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
	var out []LanguageDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("got %d languages, want 2", len(out))
	}
	if out[0].Code != "fr" || out[1].Code != "es" {
		t.Errorf("got codes %q, %q; want fr, es", out[0].Code, out[1].Code)
	}
}

func TestHandleTensesDefaultsToFrench(t *testing.T) {
	rec := doRequest(t, NewRouter(), http.MethodGet, "/api/tenses")
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
	var out []TenseSummaryDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out) != 18 {
		t.Fatalf("got %d tenses, want 18", len(out))
	}
	if out[0].ID != "present" {
		t.Errorf("got first tense %q, want present", out[0].ID)
	}
	if out[0].NameNative != "Présent" {
		t.Errorf("got NameNative %q, want Présent", out[0].NameNative)
	}
}

func TestHandleTensesSpanish(t *testing.T) {
	rec := doRequest(t, NewRouter(), http.MethodGet, "/api/tenses?lang=es")
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
	var out []TenseSummaryDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out) != 18 {
		t.Fatalf("got %d tenses, want 18", len(out))
	}
	if out[0].NameNative != "Presente" {
		t.Errorf("got NameNative %q, want Presente", out[0].NameNative)
	}
}

func TestHandleTensesUnknownLanguage(t *testing.T) {
	rec := doRequest(t, NewRouter(), http.MethodGet, "/api/tenses?lang=de")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", rec.Code)
	}
}

func TestHandleTheoryFound(t *testing.T) {
	rec := doRequest(t, NewRouter(), http.MethodGet, "/api/theory/present")
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
	var out TheoryDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if out.NameNative != "Présent" {
		t.Errorf("got NameNative %q, want Présent", out.NameNative)
	}
	if len(out.Groups) == 0 {
		t.Errorf("expected at least one group")
	}
}

func TestHandleTheorySpanish(t *testing.T) {
	rec := doRequest(t, NewRouter(), http.MethodGet, "/api/theory/preterito-indefinido?lang=es")
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
	var out TheoryDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if out.NameNative != "Pretérito indefinido" {
		t.Errorf("got NameNative %q, want Pretérito indefinido", out.NameNative)
	}
}

func TestHandleTheoryNotFound(t *testing.T) {
	rec := doRequest(t, NewRouter(), http.MethodGet, "/api/theory/not-a-tense")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404", rec.Code)
	}
}

func TestHandleTheoryWrongLanguageForTenseID(t *testing.T) {
	// "present" is French's ID; Spanish's equivalent is "presente".
	rec := doRequest(t, NewRouter(), http.MethodGet, "/api/theory/present?lang=es")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404", rec.Code)
	}
}

func TestHandleVerbs(t *testing.T) {
	rec := doRequest(t, NewRouter(), http.MethodGet, "/api/verbs")
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
	var out []VerbSummaryDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out) != 16 {
		t.Fatalf("got %d verbs, want 16", len(out))
	}
}

func TestHandleVerbsSpanish(t *testing.T) {
	rec := doRequest(t, NewRouter(), http.MethodGet, "/api/verbs?lang=es")
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
	var out []VerbSummaryDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out) != 17 {
		t.Fatalf("got %d verbs, want 17", len(out))
	}
}

func TestHandleExercisesDefaults(t *testing.T) {
	rec := doRequest(t, NewRouter(), http.MethodGet, "/api/exercises")
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
	var out []ExerciseDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out) != 10 {
		t.Fatalf("got %d exercises, want 10 (default count)", len(out))
	}
}

func TestHandleExercisesParams(t *testing.T) {
	rec := doRequest(t, NewRouter(), http.MethodGet, "/api/exercises?tense=imperatif-present&type=mcq&count=5")
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
	var out []ExerciseDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out) != 5 {
		t.Fatalf("got %d exercises, want 5", len(out))
	}
	for _, ex := range out {
		if ex.Tense != "imperatif-present" {
			t.Errorf("got tense %q, want imperatif-present", ex.Tense)
		}
		if ex.Type != "mcq" {
			t.Errorf("got type %q, want mcq", ex.Type)
		}
		if len(ex.Choices) != 4 {
			t.Errorf("got %d choices, want 4", len(ex.Choices))
		}
	}
}

func TestHandleExercisesSpanish(t *testing.T) {
	rec := doRequest(t, NewRouter(), http.MethodGet, "/api/exercises?lang=es&tense=imperativo-afirmativo&type=mcq&count=5")
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
	var out []ExerciseDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out) != 5 {
		t.Fatalf("got %d exercises, want 5", len(out))
	}
	for _, ex := range out {
		if ex.Tense != "imperativo-afirmativo" {
			t.Errorf("got tense %q, want imperativo-afirmativo", ex.Tense)
		}
		if len(ex.Choices) != 4 {
			t.Errorf("got %d choices, want 4", len(ex.Choices))
		}
	}
}

func TestHandleExercisesInvalidTense(t *testing.T) {
	rec := doRequest(t, NewRouter(), http.MethodGet, "/api/exercises?tense=nope")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", rec.Code)
	}
}

func TestHandleExercisesInvalidCount(t *testing.T) {
	rec := doRequest(t, NewRouter(), http.MethodGet, "/api/exercises?count=abc")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", rec.Code)
	}
}

func TestHandleExercisesInvalidLanguage(t *testing.T) {
	rec := doRequest(t, NewRouter(), http.MethodGet, "/api/exercises?lang=de")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", rec.Code)
	}
}

func TestCORSHeaders(t *testing.T) {
	rec := doRequest(t, NewRouter(), http.MethodGet, "/api/health")
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("got CORS origin %q, want http://localhost:5173", got)
	}
}
