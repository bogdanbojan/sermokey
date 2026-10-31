package de

import (
	"bufio"
	_ "embed"
	"strconv"
	"strings"

	"github.com/bogdanbojan/sermokey/backend/internal/grammar"
)

//go:embed vocab_data.csv
var vocabCSV string

// Vocabulary is the full flashcard word list, parsed once at package init
// from vocab_data.csv. To extend it, just append rows to that file — no
// code changes needed, it's parsed generically.
var Vocabulary = mustParseVocab(vocabCSV)

// VocabCategories describes every vocabulary category in display order
// (the order categories first appear in vocab_data.csv).
//
// Category sizes are budgeted by communicative load rather than split
// evenly — see fr.VocabCategories for the rationale.
//
// Note the German-specific convention: noun lemmas carry their article
// ("der Hund", not "Hund"), because German gender is far less predictable
// from word form than in French or Spanish and is conventionally memorised
// with the noun. The Gender column is still populated so the flashcard UI
// can render its ♂/♀/N badge, which makes it partly redundant here — that
// asymmetry with fr/es is deliberate.
var VocabCategories = []grammar.VocabCategoryInfo{
	{ID: "zahlen", NameNative: "Zahlen", NameEn: "Numbers"},
	{ID: "masse", NameNative: "Maße & Mengen", NameEn: "Measurement & quantity"},
	{ID: "zeit", NameNative: "Die Zeit", NameEn: "Time & calendar"},
	{ID: "familie", NameNative: "Die Familie", NameEn: "Family"},
	{ID: "koerper", NameNative: "Der Körper", NameEn: "Body"},
	{ID: "kleidung", NameNative: "Die Kleidung", NameEn: "Clothing"},
	{ID: "essen", NameNative: "Essen & Trinken", NameEn: "Food & drink"},
	{ID: "haus", NameNative: "Das Haus", NameEn: "House & furniture"},
	{ID: "farben", NameNative: "Die Farben", NameEn: "Colors"},
	{ID: "adjektive", NameNative: "Häufige Adjektive", NameEn: "Common adjectives"},
	{ID: "verben", NameNative: "Häufige Verben", NameEn: "Common verbs"},
	{ID: "natur", NameNative: "Die Natur", NameEn: "Nature & weather"},
	{ID: "tiere", NameNative: "Die Tiere", NameEn: "Animals"},
	{ID: "berufe", NameNative: "Die Berufe", NameEn: "Professions"},
	{ID: "orte", NameNative: "Orte", NameEn: "Places"},
	{ID: "verkehr", NameNative: "Der Verkehr", NameEn: "Transportation"},
	{ID: "schule", NameNative: "Die Schule", NameEn: "School"},
	{ID: "arbeit", NameNative: "Die Arbeit", NameEn: "Work"},
	{ID: "medien", NameNative: "Medien & Kommunikation", NameEn: "Media & communication"},
	{ID: "gefuehle", NameNative: "Gefühle & Charakter", NameEn: "Emotions & personality"},
	{ID: "funktion", NameNative: "Funktionswörter", NameEn: "Function words"},
	{ID: "abstrakt", NameNative: "Abstrakte Begriffe", NameEn: "Abstract concepts"},
	{ID: "geld", NameNative: "Geld & Einkaufen", NameEn: "Money & shopping"},
	{ID: "gesundheit", NameNative: "Die Gesundheit", NameEn: "Health"},
	{ID: "freizeit", NameNative: "Sport & Freizeit", NameEn: "Sports & leisure"},
	{ID: "stadt", NameNative: "In der Stadt", NameEn: "City & directions"},
	{ID: "gesellschaft", NameNative: "Gesellschaft & Umwelt", NameEn: "Society & environment"},
}

// mustParseVocab parses the embedded CSV (header: lemma,pos,gender,translation,category,tier).
// Panics on malformed data since this runs once at package init time against
// our own authored file — a parse failure here is a build-time bug, not a
// runtime condition to recover from.
func mustParseVocab(csv string) []grammar.VocabWord {
	scanner := bufio.NewScanner(strings.NewReader(csv))
	var words []grammar.VocabWord
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if lineNum == 1 || strings.TrimSpace(line) == "" {
			continue // header or blank line
		}
		fields := strings.Split(line, ",")
		if len(fields) != 6 {
			panic("de: vocab_data.csv line " + strconv.Itoa(lineNum) + ": expected 6 fields, got " + strconv.Itoa(len(fields)) + ": " + line)
		}
		tier, err := strconv.Atoi(strings.TrimSpace(fields[5]))
		if err != nil {
			panic("de: vocab_data.csv line " + strconv.Itoa(lineNum) + ": invalid tier: " + fields[5])
		}
		lemma := strings.TrimSpace(fields[0])
		category := strings.TrimSpace(fields[4])
		words = append(words, grammar.VocabWord{
			ID:            lemma + "|" + category,
			Lemma:         lemma,
			POS:           strings.TrimSpace(fields[1]),
			Gender:        strings.TrimSpace(fields[2]),
			TranslationEn: strings.TrimSpace(fields[3]),
			Category:      category,
			Tier:          tier,
		})
	}
	if err := scanner.Err(); err != nil {
		panic("de: failed to read vocab_data.csv: " + err.Error())
	}
	return words
}
