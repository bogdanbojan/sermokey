// Package httpapi exposes the grammar package over a small JSON REST API.
//
// Design note: exercises are generated with the correct answer embedded in
// the JSON response, and the frontend grades answers locally instead of
// calling a server-side "check" endpoint. For a multi-user graded product
// this would be a real vulnerability (a curious user can read the answer in
// the network tab); for a local, single-user practice tool it is a
// deliberate simplicity trade-off — building session/token based grading
// would add real complexity for no actual benefit here.
package httpapi

// TenseSummaryDTO is a compact tense listing, used for navigation.
type TenseSummaryDTO struct {
	ID         string `json:"id"`
	Category   string `json:"category"`
	NameNative string `json:"nameNative"` // the tense's name in the language being studied
	NameEn     string `json:"nameEn"`     // English gloss
	Usage      string `json:"usage"`
}

// GroupEndingsDTO describes one verb group's formation pattern within a tense.
type GroupEndingsDTO struct {
	GroupName    string    `json:"groupName"`
	Explanation  string    `json:"explanation"`
	Endings      [6]string `json:"endings"`
	ExampleVerb  string    `json:"exampleVerb"`
	ExampleForms [6]string `json:"exampleForms"`
}

// TheoryDTO is the full theory page content for one tense.
type TheoryDTO struct {
	ID             string            `json:"id"`
	Category       string            `json:"category"`
	NameNative     string            `json:"nameNative"`
	NameEn         string            `json:"nameEn"`
	Usage          string            `json:"usage"`
	Formation      string            `json:"formation"`
	Groups         []GroupEndingsDTO `json:"groups"`
	IrregularNotes string            `json:"irregularNotes"`
}

// VerbSummaryDTO is a compact verb listing, used for a verb reference page.
type VerbSummaryDTO struct {
	Infinitive    string `json:"infinitive"`
	Group         string `json:"group"`
	TranslationEn string `json:"translationEn"`
}

// ExerciseDTO is one generated conjugation question.
type ExerciseDTO struct {
	ID             string   `json:"id"`
	Tense          string   `json:"tense"`
	TenseName      string   `json:"tenseName"`
	Type           string   `json:"type"`
	Infinitive     string   `json:"infinitive"`
	TranslationEn  string   `json:"translationEn"`
	PersonLabel    string   `json:"personLabel"`
	SentencePrefix string   `json:"sentencePrefix"`
	SentenceSuffix string   `json:"sentenceSuffix"`
	Choices        []string `json:"choices,omitempty"`
	Answer         string   `json:"answer"`
}

// LanguageDTO is one entry in the language switcher.
type LanguageDTO struct {
	Code string `json:"code"` // "fr" | "es"
	Name string `json:"name"` // native endonym, e.g. "Français" / "Español"
}

// VocabCategoryDTO is one vocabulary theme, used for the flashcard browse UI.
type VocabCategoryDTO struct {
	ID         string `json:"id"`
	NameNative string `json:"nameNative"`
	NameEn     string `json:"nameEn"`
	WordCount  int    `json:"wordCount"`
}

// VocabWordDTO is one flashcard-able vocabulary entry.
type VocabWordDTO struct {
	ID            string `json:"id"`
	Lemma         string `json:"lemma"`
	POS           string `json:"pos"`
	Gender        string `json:"gender,omitempty"`
	TranslationEn string `json:"translationEn"`
	Category      string `json:"category"`
	Tier          int    `json:"tier"`
}

// ErrorDTO is the body returned for any non-2xx JSON response.
type ErrorDTO struct {
	Error string `json:"error"`
}
