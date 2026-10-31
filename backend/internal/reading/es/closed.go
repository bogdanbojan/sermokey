package es

// closedClass is the set of Spanish grammatical words the reader treats as
// known rather than as vocabulary to learn.
//
// A hand-written list is complete by construction here: these are closed
// classes, and unlike nouns or verbs the set of Spanish articles, clitic
// pronouns and possessive determiners does not grow.
//
// They are excluded from the curated deck on purpose — "el → the" is not a
// flashcard anybody needs — but they are roughly a third of running tokens,
// so counting them as unknown would make every text, however simple, score
// around 70% and come out labelled hard. The thresholds that figure is
// compared against (Hu & Nation 2000) are defined over all running words,
// articles included, so they are counted as known.
//
// Membership is tested against the lemma, so inflected forms reach it
// through the normal analysis path.
var closedClass = map[string]bool{
	// Articles, including the two contractions.
	"el": true, "la": true, "los": true, "las": true,
	"un": true, "una": true, "unos": true, "unas": true,
	// "uno" is the lemma the feminine and plural forms resolve to, and
	// membership is tested against the lemma, not the surface. Without it
	// every "una"/"unos"/"unas" in a text was counted as vocabulary to
	// learn rather than as the article it is.
	"uno": true,
	"lo":  true, "al": true, "del": true,

	// Subject, object and reflexive pronouns.
	"yo": true, "tú": true, "él": true, "ella": true, "usted": true,
	"nosotros": true, "nosotras": true, "vosotros": true, "vosotras": true,
	"ellos": true, "ellas": true, "ustedes": true,
	"me": true, "te": true, "se": true, "nos": true, "os": true,
	"le": true, "les": true, "mí": true, "ti": true, "sí": true, "consigo": true,

	// Possessive determiners.
	"mi": true, "mis": true, "tu": true, "tus": true, "su": true, "sus": true,
	"nuestro": true, "nuestra": true, "nuestros": true, "nuestras": true,
	"vuestro": true, "vuestra": true, "vuestros": true, "vuestras": true,

	// Demonstratives.
	"este": true, "esta": true, "estos": true, "estas": true, "esto": true,
	"ese": true, "esa": true, "esos": true, "esas": true, "eso": true,
	"aquel": true, "aquella": true, "aquellos": true, "aquellas": true, "aquello": true,

	// Negation, the complementiser, and the handful of prepositions and
	// conjunctions so frequent that treating them as vocabulary would be
	// noise. Kept short on purpose: anything with real semantic content
	// ("durante", "mediante", "entre") stays a vocabulary item.
	"no": true, "que": true, "de": true, "a": true, "y": true, "e": true,
	"o": true, "u": true, "en": true, "ni": true,
}

// isFunctionWord reports whether a lemma is closed-class grammar rather
// than vocabulary. Used by the corpus tests.
func isFunctionWord(lemma string) bool { return closedClass[lemma] }
