package fr

// closedClass is the set of French grammatical words the reader treats as
// known rather than as vocabulary to learn.
//
// Why a hand-written list is the right answer here, and not a cop-out:
// these are closed classes. Unlike nouns or verbs, the set of French
// articles, clitic pronouns and negation particles does not grow, so
// enumerating it is complete by construction — the same reasoning the
// vocabulary suite already uses for its closed-class guarantee (see
// mustHaveLemmas in grammar/fr/vocab_test.go).
//
// Why they are excluded from vocabulary at all: the curated deck
// deliberately omits articles and clitics because "le → the" is not a
// flashcard anybody needs; they are acquired through grammar in the first
// week. But they are roughly a quarter of all running tokens, so counting
// them as unknown made every text — including deliberately simple ones —
// score around 70% and come out labelled "hard". The thresholds that
// number is compared against (Hu & Nation 2000) are defined over all
// running words, articles included, so the fix is to count them as known,
// not to lower the thresholds or to pad the curriculum.
//
// Membership is tested against the lemma, so inflected and elided forms
// reach it through the normal analysis path: "l'" resolves to "le", "suis"
// to "être".
var closedClass = map[string]bool{
	// Articles and contracted articles.
	"le": true, "la": true, "les": true, "l'": true,
	"un": true, "une": true, "des": true,
	"du": true, "au": true, "aux": true,

	// Subject, object, reflexive and adverbial pronouns.
	"je": true, "tu": true, "il": true, "elle": true, "on": true,
	"nous": true, "vous": true, "ils": true, "elles": true,
	"me": true, "te": true, "se": true, "lui": true, "leur": true,
	"y": true, "en": true, "soi": true,

	// Possessive determiners.
	"mon": true, "ma": true, "mes": true,
	"ton": true, "ta": true, "tes": true,
	"son": true, "sa": true, "ses": true,
	"notre": true, "nos": true, "votre": true, "vos": true,

	// Demonstratives.
	"ce": true, "cet": true, "cette": true, "ces": true,

	// Negation and the complementiser.
	"ne": true, "pas": true, "que": true, "qu'": true,

	// The handful of prepositions and conjunctions so frequent that
	// treating them as vocabulary would be noise. Deliberately short:
	// anything with real semantic content ("pendant", "malgré", "parmi")
	// stays a vocabulary item.
	"de": true, "à": true, "et": true, "ou": true, "en tant que": true,
}

// isFunctionWord reports whether a lemma is closed-class grammar rather
// than vocabulary. Used by the corpus tests.
func isFunctionWord(lemma string) bool {
	return closedClass[lemma]
}

// elisions maps an elided clitic to the word it stands for. The tokenizer
// emits "l'" as its own token, and that token must be glossable: a reader
// should be able to hover the "l'" in "l'homme" and be told it is the
// definite article, not meet an unexplained dead span at the start of a
// great many words.
//
// Stripping the apostrophe and looking up the bare letter is not a
// substitute. "l" and "j" are not French words at all, and "n" and "d"
// match unrelated junk, so the answer has to be lexical.
var elisions = map[string]string{
	"l'":      "le",
	"d'":      "de",
	"j'":      "je",
	"n'":      "ne",
	"m'":      "me",
	"t'":      "te",
	"s'":      "se",
	"c'":      "ce",
	"qu'":     "que",
	"jusqu'":  "jusque",
	"lorsqu'": "lorsque",
	"puisqu'": "puisque",
	"quoiqu'": "quoique",
	"entr'":   "entre",
	"presqu'": "presque",
}
