package fr

import (
	"testing"

	"github.com/bogdanbojan/sermokey/backend/internal/grammar"
)

// categoryBudget is the target size of each vocabulary category, chosen by
// communicative load rather than split evenly — see VocabCategories' doc
// comment in vocab.go for why uniform category sizes were the root cause of
// the dataset's earlier quality problems.
var categoryBudget = map[string]int{
	"verbes":     240,
	"fonction":   160,
	"adjectifs":  155,
	"abstrait":   130,
	"nourriture": 115,
	"maison":     105,
	"travail":    95,
	"temps":      95,
	"societe":    80,
	"nature":     75,
	"lieux":      75,
	"ecole":      75,
	"emotions":   75,
	"animaux":    65,
	"argent":     65,
	"sante":      65,
	"loisirs":    70,
	"ville":      65,
	"transport":  65,
	"medias":     65,
	"metiers":    60,
	"vetements":  55,
	"corps":      50,
	"famille":    50,
	"nombres":    48,
	"mesures":    45,
	"couleurs":   20,
}

// mustHaveLemmas is the closed-class guarantee for French: every item here
// sits in roughly the top 300 lemmas by frequency, so a list that omits any
// of them is mis-targeted no matter how many rare words it contains. An
// earlier version of this dataset was missing "rien", "personne",
// "pourquoi" and "falloir" while carrying "anthracite" and "rétine".
var mustHaveLemmas = []string{
	// Interrogatives and relatives.
	"qui", "quoi", "quel", "où", "comment", "pourquoi", "combien", "quand", "dont",
	// Negative and indefinite pronouns.
	"rien", "personne", "quelque chose", "quelqu'un", "aucun", "chacun",
	// Core determiners and quantifiers.
	"tout", "tous", "autre", "même", "chaque", "plusieurs", "quelques",
	// Subject and stressed pronouns.
	"moi", "toi", "lui", "elle", "eux", "celui", "celle", "ceux", "cela", "ça",
	// Conjunctions and discourse connectors.
	"et", "ou", "mais", "donc", "car", "parce que", "si", "comme", "puisque",
	"alors", "pourtant", "cependant", "ainsi", "enfin", "ensuite", "puis",
	// High-frequency adverbs.
	"aussi", "encore", "déjà", "toujours", "jamais", "souvent", "parfois",
	"beaucoup", "peu", "très", "trop", "assez", "plus", "moins", "presque",
	"plutôt", "partout", "vraiment", "peut-être", "surtout", "seulement",
	// Core prepositions.
	"de", "à", "dans", "sur", "sous", "avec", "sans", "pour", "par", "entre",
	"chez", "vers", "pendant", "avant", "après", "depuis", "contre", "parmi",
	// The top verbs — the single highest-value closed set in the language.
	"être", "avoir", "faire", "dire", "pouvoir", "aller", "voir", "vouloir",
	"venir", "devoir", "prendre", "trouver", "donner", "falloir", "parler",
	"mettre", "savoir", "passer", "regarder", "aimer", "croire", "demander",
	"rester", "répondre", "entendre", "penser", "arriver", "connaître",
	"devenir", "sentir", "sembler", "tenir", "comprendre", "rendre",
	"attendre", "sortir", "vivre", "chercher", "écrire", "lire", "suivre",
	"montrer", "utiliser", "expliquer", "rencontrer", "garder", "apporter",
	"commencer", "finir", "continuer", "perdre", "gagner", "partir", "jouer",
	"manger", "boire", "dormir", "ouvrir", "fermer", "acheter", "payer",
	"travailler", "apprendre", "oublier", "changer", "choisir", "aider",
	"quitter", "obtenir", "éviter", "créer", "recevoir",
}

func TestVocabDataIsWellFormed(t *testing.T) {
	for _, err := range grammar.ValidateVocab(grammar.VocabCheck{
		Lang:            "fr",
		Words:           Vocabulary,
		Categories:      VocabCategories,
		Budget:          categoryBudget,
		BudgetTolerance: 0.30,
		MustHave:        mustHaveLemmas,
	}) {
		t.Error(err)
	}
}

// TestVocabSizeInB2Range guards the overall target. Receptive vocabulary at
// B2 is roughly 3,250–3,750 word families (Milton 2010); a curated deck of
// ~2,400 high-frequency lemmas is the practical core of that, with the
// remainder picked up incidentally and through cognates.
func TestVocabSizeInB2Range(t *testing.T) {
	if n := len(Vocabulary); n < 2000 || n > 2800 {
		t.Errorf("fr: %d words, want 2000..2800", n)
	}
}
