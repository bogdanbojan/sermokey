package de

// closedClass is the set of German grammatical words the reader treats as
// known rather than as vocabulary to learn.
//
// German declines its articles and pronouns for case, so the list is
// longer than its French or Spanish equivalents — "der", "die", "das",
// "den", "dem" and "des" are six surface forms of one determiner. Every
// one has to be here, because the entry that matters is whichever form the
// text happens to use.
//
// The rationale is the same as for the other languages: these are closed
// classes, they are excluded from the curated deck on purpose, and they
// are a large share of running tokens, so counting them as unknown
// vocabulary would make every text score as too hard.
var closedClass = map[string]bool{
	// Definite article, all cases and genders.
	"der": true, "die": true, "das": true, "den": true, "dem": true, "des": true,
	// Indefinite article and its negation.
	"ein": true, "eine": true, "einen": true, "einem": true, "einer": true, "eines": true,
	"kein": true, "keine": true, "keinen": true, "keinem": true, "keiner": true, "keines": true,

	// Personal pronouns.
	"ich": true, "du": true, "er": true, "sie": true, "es": true,
	"wir": true, "ihr": true, "man": true,
	"mich": true, "dich": true, "sich": true, "uns": true, "euch": true,
	"mir": true, "dir": true, "ihm": true, "ihn": true, "ihnen": true,

	// Possessives.
	"mein": true, "meine": true, "meinen": true, "meinem": true, "meiner": true,
	"dein": true, "deine": true, "deinen": true, "deinem": true, "deiner": true,
	"sein": true, "seine": true, "seinen": true, "seinem": true, "seiner": true,
	"unser": true, "unsere": true, "euer": true, "eure": true,

	// Demonstratives.
	"dieser": true, "diese": true, "dieses": true, "diesen": true, "diesem": true,
	"jener": true, "jene": true, "jenes": true,

	// Preposition+article contractions. German fuses these constantly
	// ("in dem" -> "im"), and they are the single most common class of
	// token after the bare articles themselves.
	"im": true, "am": true, "beim": true, "zum": true, "zur": true,
	"vom": true, "ins": true, "ans": true, "aufs": true, "fürs": true,
	"durchs": true, "ums": true, "übers": true, "unters": true,

	// Negation, complementisers and the highest-frequency function words.
	// Deliberately short beyond this: anything carrying real semantic
	// content ("während", "trotz", "wegen") stays a vocabulary item.
	"nicht": true, "dass": true, "und": true, "oder": true, "aber": true,
	"zu": true, "in": true, "von": true, "mit": true, "auf": true,
	"als": true, "wie": true, "so": true, "an": true,
	"bei": true, "nach": true, "aus": true, "für": true, "um": true,
}

// isFunctionWord reports whether a lemma is closed-class grammar rather
// than vocabulary. Used by the corpus tests.
func isFunctionWord(lemma string) bool { return closedClass[lemma] }
