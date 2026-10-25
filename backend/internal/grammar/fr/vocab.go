package fr

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
var VocabCategories = []grammar.VocabCategoryInfo{
	{ID: "nombres", NameNative: "Nombres", NameEn: "Numbers"},
	{ID: "temps", NameNative: "Le temps", NameEn: "Time & calendar"},
	{ID: "famille", NameNative: "La famille", NameEn: "Family"},
	{ID: "corps", NameNative: "Le corps", NameEn: "Body"},
	{ID: "vetements", NameNative: "Les vêtements", NameEn: "Clothing"},
	{ID: "nourriture", NameNative: "La nourriture", NameEn: "Food & drink"},
	{ID: "maison", NameNative: "La maison", NameEn: "House & furniture"},
	{ID: "couleurs", NameNative: "Les couleurs", NameEn: "Colors"},
	{ID: "adjectifs", NameNative: "Adjectifs courants", NameEn: "Common adjectives"},
	{ID: "verbes", NameNative: "Verbes courants", NameEn: "Common verbs"},
	{ID: "nature", NameNative: "La nature", NameEn: "Nature & weather"},
	{ID: "animaux", NameNative: "Les animaux", NameEn: "Animals"},
	{ID: "metiers", NameNative: "Les métiers", NameEn: "Professions"},
	{ID: "lieux", NameNative: "Les lieux", NameEn: "Places"},
	{ID: "transport", NameNative: "Les transports", NameEn: "Transportation"},
	{ID: "ecole", NameNative: "L'école", NameEn: "School"},
	{ID: "travail", NameNative: "Le travail", NameEn: "Work & technology"},
	{ID: "emotions", NameNative: "Les émotions", NameEn: "Emotions & personality"},
	{ID: "fonction", NameNative: "Mots-outils", NameEn: "Function words"},
	{ID: "abstrait", NameNative: "Mots abstraits", NameEn: "Abstract concepts"},
	{ID: "argent", NameNative: "L'argent", NameEn: "Money & shopping"},
	{ID: "sante", NameNative: "La santé", NameEn: "Health"},
	{ID: "loisirs", NameNative: "Les loisirs", NameEn: "Sports & leisure"},
	{ID: "ville", NameNative: "En ville", NameEn: "City & directions"},
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
			panic("fr: vocab_data.csv line " + strconv.Itoa(lineNum) + ": expected 6 fields, got " + strconv.Itoa(len(fields)) + ": " + line)
		}
		tier, err := strconv.Atoi(strings.TrimSpace(fields[5]))
		if err != nil {
			panic("fr: vocab_data.csv line " + strconv.Itoa(lineNum) + ": invalid tier: " + fields[5])
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
		panic("fr: failed to read vocab_data.csv: " + err.Error())
	}
	return words
}
