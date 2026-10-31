package reading

import (
	"sort"
	"strings"
	"unicode"

	"github.com/bogdanbojan/sermokey/backend/internal/grammar"
)

// StandardAnalyzer resolves surface forms against three sources, in
// decreasing order of authority: a language's curated vocabulary, its
// generated lexicon, and its inflected-form index.
//
// Unlike a plain fallback chain it merges rather than short-circuits,
// because homographs routinely straddle sources. French "est" is a curated
// noun ("east") and also the third-person singular of être; returning only
// the first hit would hide the most common verb form in the language behind
// a compass point. Candidates are collected from everywhere and ranked.
//
// Everything here is language-independent. What differs between languages
// is supplied through LexicalConfig: the closed-class set, the elision
// table (French has one; Spanish and German do not), and how a surface is
// normalized for lookup.
type StandardAnalyzer struct {
	lex *Lexicon
	cfg LexicalConfig

	curated map[string][]grammar.VocabWord
}

// LexicalConfig is the per-language part of StandardAnalyzer.
type LexicalConfig struct {
	// Curated is the hand-built vocabulary list. Entries here outrank the
	// generated lexicon and carry the SRS card id the flashcard deck
	// already schedules.
	Curated []grammar.VocabWord

	// ClosedClass holds grammatical words — articles, clitic pronouns,
	// negation particles — which the reader treats as known rather than as
	// vocabulary to learn. Closed classes do not grow, so enumerating one
	// is complete by construction.
	ClosedClass map[string]bool

	// Elisions maps an elided clitic to the word it stands for ("l'" ->
	// "le"). French needs it so the clitic token is glossable at all;
	// Spanish and German leave it nil.
	Elisions map[string]string

	// Normalize prepares a surface for lookup: case folding plus whatever
	// punctuation folding the orthography needs.
	Normalize func(string) string

	// ReadingCategory is the synthetic vocabulary category that words
	// saved while reading fall into when they are not curated.
	ReadingCategory string

	// CapitalSignalsProperNoun says whether an initial capital is evidence
	// of a proper noun. True for French and Spanish, where only names and
	// sentence openings are capitalised. False for German, which
	// capitalises every noun — there the signal carries no information,
	// and acting on it would promote half the words on the page to names
	// and drop them from the vocabulary count entirely.
	CapitalSignalsProperNoun bool

	// DerivedSuffixes is the last-resort deinflection table. Defaults to
	// the Romance rules when nil; German supplies its own, since stripping
	// a final "-s" from a German word is almost always wrong.
	DerivedSuffixes []SuffixRule

	// PriorityLemmas breaks ties between readings the curriculum ranks
	// equally.
	//
	// Some homographs are genuinely ambiguous out of context: French
	// "suis" is the first person of both être and suivre, and both are
	// tier-1 verbs, so tier alone leaves the choice to table order. In
	// running text it is almost always être.
	//
	// The set is the auxiliaries and copulas — a deliberately tiny,
	// defensible frequency proxy rather than an attempt at real
	// disambiguation, which would need a context-sensitive tagger. The
	// losing reading stays available in the panel.
	PriorityLemmas map[string]bool

	// CuratedAliases returns extra lookup keys for a curated word, beyond
	// its lemma.
	//
	// German needs this. Its vocabulary list stores nouns with their
	// article ("der Hund") because the gender is half of what a learner
	// has to memorise, but running text contains the bare noun. Without an
	// alias every German noun would miss the curated layer, fall through
	// to the scraped lexicon, and be saved under a different card id than
	// the one the vocabulary tab schedules — silently splitting each
	// word's review history in two.
	CuratedAliases func(grammar.VocabWord) []string
}

// SuffixRule is one rule-based deinflection: strip Suffix, append
// Replacement, and report the result as Note.
type SuffixRule struct {
	Suffix      string
	Replacement string
	Note        string
}

const (
	// rankUncurated is the sort key for a lemma that has a dictionary
	// entry but no place in the curriculum. It sits below every curated
	// tier (1–3) and above nothing else.
	rankUncurated = 50

	// participleDemotion is added to an inflected reading's rank when the
	// surface is also a headword and the inflection is a participle.
	participleDemotion = 10

	// rankProperNoun is given to a proper-noun reading when the surface is
	// capitalised in the text. It outranks every curated tier because the
	// capitalisation is direct evidence.
	rankProperNoun = 0

	// ProperNounPOS is the part of speech cmd/lexgen uses for proper
	// nouns, kept distinct from "noun" so they can be excluded from
	// vocabulary counts.
	ProperNounPOS = "name"

	// maxPhraseCombinations bounds the lemma search for multi-word
	// surfaces.
	maxPhraseCombinations = 32
)

// NewStandardAnalyzer indexes the curated vocabulary and returns a ready
// analyzer.
func NewStandardAnalyzer(lex *Lexicon, cfg LexicalConfig) *StandardAnalyzer {
	if cfg.Normalize == nil {
		cfg.Normalize = func(s string) string { return strings.ToLower(strings.TrimSpace(s)) }
	}
	if cfg.DerivedSuffixes == nil {
		cfg.DerivedSuffixes = romanceSuffixes
	}
	a := &StandardAnalyzer{
		lex:     lex,
		cfg:     cfg,
		curated: make(map[string][]grammar.VocabWord, len(cfg.Curated)),
	}
	for _, w := range cfg.Curated {
		keys := []string{strings.ToLower(w.Lemma)}
		if cfg.CuratedAliases != nil {
			for _, alias := range cfg.CuratedAliases(w) {
				if alias = strings.ToLower(strings.TrimSpace(alias)); alias != "" && alias != keys[0] {
					keys = append(keys, alias)
				}
			}
		}
		for _, k := range keys {
			a.curated[k] = append(a.curated[k], w)
		}
	}
	return a
}

var _ Analyzer = (*StandardAnalyzer)(nil)

type candidate struct {
	analysis Analysis
	rank     int
	// direct records that the surface was already the headword. Breaks
	// ties between equally uncurated readings, where the simpler
	// explanation — this is the word itself, not an inflection of
	// something else — is the better default.
	direct bool
	// common records that the headword is spelled in lower case, i.e. an
	// ordinary word rather than a proper noun or an acronym.
	common bool
	// priority marks an auxiliary or copula; see PriorityLemmas.
	priority bool
}

// Analyze implements Analyzer.
func (a *StandardAnalyzer) Analyze(surface string) []Analysis {
	return a.analyze(surface, false)
}

// AnalyzeProperNoun implements ProperNounAnalyzer: the caller has seen this
// surface capitalised mid-sentence somewhere in the document, which settles
// the question Analyze has to stay agnostic about.
func (a *StandardAnalyzer) AnalyzeProperNoun(surface string) []Analysis {
	return a.analyze(surface, true)
}

func (a *StandardAnalyzer) analyze(surface string, knownProperNoun bool) []Analysis {
	trimmed := strings.ToLower(strings.TrimSpace(surface))
	norm := a.cfg.Normalize(surface)
	if norm == "" {
		return nil
	}
	capitalised := isCapitalised(surface)

	var cands []candidate
	seen := map[string]bool{}

	if full, ok := a.cfg.Elisions[trimmed]; ok {
		// The clitic stands for a whole word; gloss that word, and say so.
		a.collect(&cands, seen, full, "elided form of "+full, true)
	} else {
		a.collect(&cands, seen, norm, "", true)
		directCount := len(cands)

		// Every analysis of the surface as an inflection, not just the
		// first: a form routinely belongs to two different lemmas, and
		// which one is right is a ranking question, not a question the
		// generator can answer.
		forms := a.lex.Forms(norm)
		for _, f := range forms {
			lemma := strings.ToLower(f.Lemma)
			a.collect(&cands, seen, lemma, f.Note, false)

			// One further hop, when the first landed on a headword that is
			// itself an inflection. Participle-derived adjectives are
			// productive, so French "oubliée" resolves to the adjective
			// "oublié" — a real headword, where the generated chain
			// correctly stops — and a reader would be shown "forgotten"
			// instead of being connected to oublier, the verb already on a
			// flashcard in their deck.
			for _, f2 := range a.lex.Forms(lemma) {
				if strings.ToLower(f2.Lemma) != norm {
					a.collect(&cands, seen, strings.ToLower(f2.Lemma), f2.Note, false)
				}
			}
		}

		// Past participles are lexicalised as nouns and adjectives on a
		// large scale: French marché, entrée, employé. For those the
		// headword reading is normally right, even though the verb behind
		// it is usually the more essential word and would win on tier
		// alone — "Le marché du samedi" is a market, not something walked.
		// Finite forms behave the opposite way, so they keep the tier
		// ranking.
		if directCount > 0 && len(forms) > 0 && isParticiple(forms[0].Note) {
			for i := directCount; i < len(cands); i++ {
				cands[i].rank += participleDemotion
			}
		}

		// Promote a proper-noun reading when the text capitalised the word
		// — but only when the surface has no ordinary reading *of its
		// own*. The test is over the direct candidates alone, not over
		// everything collected: "Claire" has no headword reading besides
		// the given name, and the fact that it is also the feminine of
		// "clair" is exactly the competing analysis capitalisation is
		// meant to settle.
		//
		// Every sentence-initial word is capitalised, and Wiktionary
		// lists surnames for a great many common words, so "Maison" must
		// stay the house. This errs toward false negatives, which is the
		// right way to be wrong: the name reading is still in the panel.
		// Promote a proper-noun reading when the evidence supports it.
		//
		// Capitalisation alone is weak: every sentence-initial word has it,
		// and Wiktionary lists surnames for a great many common words, so
		// on its own it is only trusted when the surface has no ordinary
		// reading of its own. That caution is what keeps French "Maison"
		// the house — and what read Spanish "Clara" as the adjective claro.
		//
		// knownProperNoun lifts it. The caller has seen this surface
		// capitalised mid-sentence, where capitalisation is not positional
		// and so means what it appears to mean.
		if a.cfg.CapitalSignalsProperNoun && !a.cfg.ClosedClass[norm] &&
			(knownProperNoun || (capitalised && !hasOrdinaryReading(cands[:directCount]))) {
			for i := range cands {
				if cands[i].analysis.POS == ProperNounPOS {
					cands[i].rank = rankProperNoun
				}
			}
		}
	}

	if len(cands) == 0 {
		if as := a.analyzePhrase(norm); len(as) > 0 {
			return as
		}
		return a.derive(norm)
	}

	sort.SliceStable(cands, func(i, j int) bool {
		x, y := cands[i], cands[j]
		if x.rank != y.rank {
			return x.rank < y.rank
		}
		// Wiktionary indexes acronyms and surnames alongside ordinary
		// words, so French "le" collides with the abbreviation "LE". A
		// capitalised headword sorts before a lower-case one byte-wise,
		// which without this rule would make an acronym the primary
		// reading of the commonest article in the language.
		if x.common != y.common {
			return x.common
		}
		if x.priority != y.priority {
			return x.priority
		}
		return x.direct && !y.direct
	})

	out := make([]Analysis, 0, len(cands))
	for _, c := range cands {
		out = append(out, c.analysis)
	}
	return out
}

// collect appends every reading of one headword — curated and dictionary —
// deduplicated by lemma and part of speech.
func (a *StandardAnalyzer) collect(cands *[]candidate, seen map[string]bool, lemma, note string, direct bool) {
	curated := a.curated[lemma]
	entries := a.lex.Entries(lemma)

	// A curated word always produces an analysis, even when the scraped
	// lexicon has nothing for it: the curriculum is the authority on what
	// the learner is supposed to know.
	for _, w := range curated {
		key := lemma + "\x00" + w.POS
		if seen[key] {
			continue
		}
		seen[key] = true
		*cands = append(*cands, candidate{
			analysis: Analysis{
				Lemma:    w.Lemma,
				POS:      w.POS,
				Gender:   w.Gender,
				FormNote: note,
				Senses:   a.mergeSenses(w.TranslationEn, entries, w.POS),
				WordID:   w.ID,
				Function: a.cfg.ClosedClass[lemma],
				Proper:   w.POS == ProperNounPOS,
				Source:   SourceCurated,
			},
			rank:     w.Tier,
			direct:   direct,
			common:   isCommonSpelling(w.Lemma),
			priority: a.cfg.PriorityLemmas[lemma],
		})
	}

	for _, e := range entries {
		key := lemma + "\x00" + e.POS
		if seen[key] {
			continue
		}
		seen[key] = true
		*cands = append(*cands, candidate{
			analysis: Analysis{
				Lemma:    e.Lemma,
				POS:      e.POS,
				Gender:   e.Gender,
				FormNote: note,
				Senses:   toSenses(SplitGlosses(e.Glosses)),
				WordID:   e.Lemma + "|" + a.cfg.ReadingCategory,
				Function: a.cfg.ClosedClass[lemma],
				Proper:   e.POS == ProperNounPOS,
				Source:   SourceLexicon,
			},
			rank:     rankUncurated,
			direct:   direct,
			common:   isCommonSpelling(e.Lemma),
			priority: a.cfg.PriorityLemmas[lemma],
		})
	}
}

// mergeSenses puts the curated translation first — it is the gloss the
// learner already sees on that word's flashcard, and the two surfaces
// should agree — then appends additional dictionary senses for the same
// part of speech, so the panel can still show the fuller picture.
func (a *StandardAnalyzer) mergeSenses(translation string, entries []LexEntry, pos string) []Sense {
	out := []Sense{{Gloss: translation}}
	for _, e := range entries {
		if e.POS != pos {
			continue
		}
		for _, g := range SplitGlosses(e.Glosses) {
			if !strings.EqualFold(g, translation) {
				out = append(out, Sense{Gloss: g})
			}
		}
	}
	return out
}

// analyzePhrase resolves a multi-word surface by lemmatising each word and
// looking up the joined result.
//
// It is the inverse of the merging the tokenizer does, and needed for the
// same reason: "se lève" is one lexical item in running text, but the
// headword is "se lever". Wiktionary records plurals of multi-word nouns,
// so "pommes de terre" resolves directly, but it does not conjugate
// reflexive verbs — without this a merged phrase would become a token with
// no definition at all, visibly worse than not merging it.
//
// Every reading of every word is tried, not just the best one: French
// "couche" is primarily a noun, and only its secondary reading as a form of
// coucher leads anywhere when preceded by "se".
func (a *StandardAnalyzer) analyzePhrase(norm string) []Analysis {
	words := strings.Fields(norm)
	if len(words) < 2 {
		return nil
	}
	options := make([][]string, 0, len(words))
	for _, w := range words {
		seen := map[string]bool{}
		var lemmas []string
		for _, an := range a.Analyze(w) {
			if !seen[an.Lemma] {
				seen[an.Lemma] = true
				lemmas = append(lemmas, an.Lemma)
			}
		}
		if len(lemmas) == 0 {
			return nil
		}
		options = append(options, lemmas)
	}

	for _, joined := range joinCombinations(options, maxPhraseCombinations) {
		if joined == norm {
			continue // already tried as the direct surface
		}
		var cands []candidate
		a.collect(&cands, map[string]bool{}, joined, "", true)
		if len(cands) == 0 {
			continue
		}
		sort.SliceStable(cands, func(i, j int) bool { return cands[i].rank < cands[j].rank })
		out := make([]Analysis, 0, len(cands))
		for _, c := range cands {
			an := c.analysis
			if an.FormNote == "" {
				an.FormNote = "form of " + an.Lemma
			}
			out = append(out, an)
		}
		return out
	}
	return nil
}

// romanceSuffixes is the default last-resort deinflection table, tried
// only when neither the dictionary nor the form index recognises a
// surface. With a 350k-entry form index the realistic callers are typos,
// OCR artefacts and proper nouns, so the rules stay few and conservative:
// regular plurals and feminines, the endings most likely to be productive
// on a word the dictionary has never seen.
var romanceSuffixes = []SuffixRule{
	{"aux", "al", "plural"},
	{"eaux", "eau", "plural"},
	{"es", "e", "plural"},
	{"s", "", "plural"},
	{"x", "", "plural"},
	{"e", "", "feminine"},
}

// derive attempts rule-based deinflection and marks whatever it finds as a
// guess, so the UI can present it as one rather than as a dictionary fact.
func (a *StandardAnalyzer) derive(norm string) []Analysis {
	for _, r := range a.cfg.DerivedSuffixes {
		if !strings.HasSuffix(norm, r.Suffix) {
			continue
		}
		stem := strings.TrimSuffix(norm, r.Suffix) + r.Replacement
		if len([]rune(stem)) < 2 {
			continue
		}
		var cands []candidate
		a.collect(&cands, map[string]bool{}, stem, r.Note+" of "+stem, false)
		if len(cands) == 0 {
			continue
		}
		out := cands[0].analysis
		out.Source = SourceDerived
		return []Analysis{out}
	}
	return nil
}

// joinCombinations enumerates the cartesian product of per-word lemma
// options, best-first, stopping at limit.
func joinCombinations(options [][]string, limit int) []string {
	out := []string{""}
	for _, lemmas := range options {
		next := make([]string, 0, len(out)*len(lemmas))
		for _, prefix := range out {
			for _, l := range lemmas {
				if prefix == "" {
					next = append(next, l)
				} else {
					next = append(next, prefix+" "+l)
				}
				if len(next) >= limit {
					break
				}
			}
			if len(next) >= limit {
				break
			}
		}
		out = next
	}
	return out
}

func hasOrdinaryReading(cands []candidate) bool {
	for _, c := range cands {
		if c.analysis.POS != ProperNounPOS {
			return true
		}
	}
	return false
}

// isParticiple reports whether an inflection note describes a participle.
// Notes come from Wiktionary as either editor prose ("past participle of
// marcher") or assembled tags ("participle past of aller"), so both
// spellings have to match.
func isParticiple(note string) bool {
	return strings.Contains(strings.ToLower(note), "participle")
}

// isCommonSpelling reports whether a headword is written as an ordinary
// lower-case word, as opposed to a proper noun or an acronym.
func isCommonSpelling(lemma string) bool { return lemma == strings.ToLower(lemma) }

func isCapitalised(surface string) bool {
	for _, r := range strings.TrimSpace(surface) {
		return unicode.IsUpper(r)
	}
	return false
}

func toSenses(glosses []string) []Sense {
	out := make([]Sense, 0, len(glosses))
	for _, g := range glosses {
		out = append(out, Sense{Gloss: g})
	}
	return out
}
