package es

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
	{ID: "numeros", NameNative: "Números", NameEn: "Numbers"},
	{ID: "tiempo", NameNative: "El tiempo", NameEn: "Time & calendar"},
	{ID: "familia", NameNative: "La familia", NameEn: "Family"},
	{ID: "cuerpo", NameNative: "El cuerpo", NameEn: "Body"},
	{ID: "ropa", NameNative: "La ropa", NameEn: "Clothing"},
	{ID: "comida", NameNative: "La comida", NameEn: "Food & drink"},
	{ID: "casa", NameNative: "La casa", NameEn: "House & furniture"},
	{ID: "colores", NameNative: "Los colores", NameEn: "Colors"},
	{ID: "adjetivos", NameNative: "Adjetivos comunes", NameEn: "Common adjectives"},
	{ID: "verbos", NameNative: "Verbos comunes", NameEn: "Common verbs"},
	{ID: "naturaleza", NameNative: "La naturaleza", NameEn: "Nature & weather"},
	{ID: "animales", NameNative: "Los animales", NameEn: "Animals"},
	{ID: "profesiones", NameNative: "Las profesiones", NameEn: "Professions"},
	{ID: "lugares", NameNative: "Los lugares", NameEn: "Places"},
	{ID: "transporte", NameNative: "El transporte", NameEn: "Transportation"},
	{ID: "escuela", NameNative: "La escuela", NameEn: "School"},
	{ID: "trabajo", NameNative: "El trabajo", NameEn: "Work & technology"},
	{ID: "emociones", NameNative: "Las emociones", NameEn: "Emotions & personality"},
	{ID: "funcion", NameNative: "Palabras funcionales", NameEn: "Function words"},
	{ID: "abstracto", NameNative: "Conceptos abstractos", NameEn: "Abstract concepts"},
	{ID: "dinero", NameNative: "El dinero", NameEn: "Money & shopping"},
	{ID: "salud", NameNative: "La salud", NameEn: "Health"},
	{ID: "ocio", NameNative: "El ocio", NameEn: "Sports & leisure"},
	{ID: "ciudad", NameNative: "En la ciudad", NameEn: "City & directions"},
}

// mustParseVocab parses the embedded CSV (header: lemma,pos,gender,translation,category,tier).
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
			panic("es: vocab_data.csv line " + strconv.Itoa(lineNum) + ": expected 6 fields, got " + strconv.Itoa(len(fields)) + ": " + line)
		}
		tier, err := strconv.Atoi(strings.TrimSpace(fields[5]))
		if err != nil {
			panic("es: vocab_data.csv line " + strconv.Itoa(lineNum) + ": invalid tier: " + fields[5])
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
		panic("es: failed to read vocab_data.csv: " + err.Error())
	}
	return words
}
