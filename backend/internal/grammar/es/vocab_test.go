package es

import (
	"testing"

	"github.com/bogdanbojan/sermokey/backend/internal/grammar"
)

// categoryBudget is the target size of each vocabulary category, chosen by
// communicative load rather than split evenly — see VocabCategories' doc
// comment in vocab.go.
var categoryBudget = map[string]int{
	"verbos":      230,
	"funcion":     150,
	"adjetivos":   150,
	"abstracto":   130,
	"comida":      115,
	"casa":        100,
	"tiempo":      90,
	"trabajo":     85,
	"sociedad":    78,
	"naturaleza":  75,
	"lugares":     73,
	"escuela":     73,
	"animales":    70,
	"ocio":        70,
	"emociones":   70,
	"medios":      65,
	"transporte":  63,
	"dinero":      62,
	"salud":       60,
	"ciudad":      60,
	"profesiones": 60,
	"ropa":        55,
	"numeros":     52,
	"familia":     50,
	"cuerpo":      49,
	"medidas":     42,
	"colores":     21,
}

// mustHaveLemmas is the closed-class guarantee for Spanish. Every item sits
// in roughly the top 300 lemmas by frequency (cf. the Instituto Cervantes
// PCIC A1–B1 inventories), so a list omitting any of them is mis-targeted
// no matter how many rare words it carries.
var mustHaveLemmas = []string{
	// Interrogatives.
	"qué", "quién", "cuál", "cómo", "por qué", "cuánto", "dónde", "cuándo",
	// Negative and indefinite pronouns.
	"nada", "nadie", "algo", "alguien", "ninguno", "cualquiera",
	// Core determiners and quantifiers.
	"todo", "todos", "otro", "mismo", "cada", "varios", "algunos",
	// Personal pronouns, possessives, demonstratives.
	"yo", "tú", "él", "ella", "nosotros", "vosotros", "ellos", "usted",
	"mi", "tu", "su", "nuestro", "este", "ese", "esto", "eso", "se", "lo", "le",
	// Conjunctions and discourse connectors.
	"y", "o", "pero", "porque", "si", "como", "aunque", "mientras",
	"entonces", "sin embargo", "además", "es decir", "de hecho",
	// High-frequency adverbs.
	"también", "ya", "todavía", "siempre", "nunca", "a menudo", "a veces",
	"mucho", "poco", "muy", "demasiado", "bastante", "más", "menos", "casi",
	"quizás", "solo", "sobre todo",
	// Core prepositions.
	"de", "a", "en", "sobre", "bajo", "con", "sin", "para", "por", "entre",
	"hacia", "hasta", "desde", "durante", "antes", "después", "contra",
	// The top verbs.
	"ser", "estar", "haber", "tener", "hacer", "poder", "decir", "ir", "ver",
	"querer", "venir", "deber", "saber", "dar", "poner", "parecer", "quedar",
	"creer", "hablar", "llevar", "dejar", "seguir", "encontrar", "llamar",
	"llegar", "pasar", "pensar", "volver", "conocer", "salir", "sentir",
	"mirar", "contar", "empezar", "esperar", "buscar", "existir", "entrar",
	"trabajar", "escribir", "perder", "producir", "ocurrir", "pedir",
	"recordar", "terminar", "permitir", "aparecer", "conseguir", "comenzar",
	"servir", "sacar", "necesitar", "mantener", "resultar", "leer", "caer",
	"cambiar", "presentar", "crear", "abrir", "considerar", "oír", "usar",
	"mostrar", "explicar", "comer", "beber", "vivir", "estudiar", "comprar",
}

func TestVocabDataIsWellFormed(t *testing.T) {
	for _, err := range grammar.ValidateVocab(grammar.VocabCheck{
		Lang:            "es",
		Words:           Vocabulary,
		Categories:      VocabCategories,
		Budget:          categoryBudget,
		BudgetTolerance: 0.30,
		MustHave:        mustHaveLemmas,
	}) {
		t.Error(err)
	}
}

func TestVocabSizeInB2Range(t *testing.T) {
	if n := len(Vocabulary); n < 2000 || n > 2800 {
		t.Errorf("es: %d words, want 2000..2800", n)
	}
}
