package fr

// Verbs is the reference dataset of 16 verbs spanning all three French verb
// groups, each with its hand-authored simple-tense conjugation table. To add
// a verb, copy an existing entry of the same group and fill in every
// field — there is intentionally no rule-based generation for irregular
// stems, see the package doc comment for why.
//
// PastParticiple and Auxiliary are enough to derive every compound,
// periphrastic, and non-finite tense (passé composé, plus-que-parfait,
// futur antérieur, conditionnel passé, subjonctif passé, impératif passé,
// présent continu, passé récent, futur proche, infinitif passé, gérondif
// présent/passé) — see compound.go. For être-verbs (aller, venir), the
// derived compound tenses use the default masculine-singular/plural past
// participle agreement for simplicity; see the "passe-compose" theory entry
// for the real agreement rule.
var Verbs = []Verb{
	// --- 1er groupe (-er) ---
	{
		Infinitive: "parler", Group: GroupFirst, TranslationEn: "to speak",
		Auxiliary: "avoir", PastParticiple: "parlé", HasImperative: true,
		Contexts: []string{
			"au téléphone avec sa sœur", "français très bien", "pendant des heures de politique",
			"doucement pour ne pas réveiller le bébé", "à ses collègues du nouveau projet",
			"trop vite pour qu'on le comprenne", "avec passion de ses voyages",
		},
		Present:             Conjugation{"parle", "parles", "parle", "parlons", "parlez", "parlent"},
		Imparfait:           Conjugation{"parlais", "parlais", "parlait", "parlions", "parliez", "parlaient"},
		FutureSimple:        Conjugation{"parlerai", "parleras", "parlera", "parlerons", "parlerez", "parleront"},
		ConditionnelPresent: Conjugation{"parlerais", "parlerais", "parlerait", "parlerions", "parleriez", "parleraient"},
		SubjonctifPresent:   Conjugation{"parle", "parles", "parle", "parlions", "parliez", "parlent"},
		ImperatifPresent:    Conjugation{Tu: "parle", Nous: "parlons", Vous: "parlez"},
	},
	{
		Infinitive: "manger", Group: GroupFirst, TranslationEn: "to eat",
		Auxiliary: "avoir", PastParticiple: "mangé", HasImperative: true,
		Contexts: []string{
			"une pomme chaque matin", "au restaurant ce soir", "trop de chocolat à Noël",
			"des légumes à chaque repas", "avec ses doigts en pique-nique", "un sandwich sur le pouce",
			"au petit-déjeuner avant l'école",
		},
		Present:             Conjugation{"mange", "manges", "mange", "mangeons", "mangez", "mangent"},
		Imparfait:           Conjugation{"mangeais", "mangeais", "mangeait", "mangions", "mangiez", "mangeaient"},
		FutureSimple:        Conjugation{"mangerai", "mangeras", "mangera", "mangerons", "mangerez", "mangeront"},
		ConditionnelPresent: Conjugation{"mangerais", "mangerais", "mangerait", "mangerions", "mangeriez", "mangeraient"},
		SubjonctifPresent:   Conjugation{"mange", "manges", "mange", "mangions", "mangiez", "mangent"},
		ImperatifPresent:    Conjugation{Tu: "mange", Nous: "mangeons", Vous: "mangez"},
	},
	{
		Infinitive: "aimer", Group: GroupFirst, TranslationEn: "to like / to love",
		Auxiliary: "avoir", PastParticiple: "aimé", HasImperative: true,
		Contexts: []string{
			"la musique classique", "voyager en été", "les films d'horreur", "se promener en forêt",
			"les longues soirées d'été", "lire avant de dormir", "cuisiner pour ses amis",
		},
		Present:             Conjugation{"aime", "aimes", "aime", "aimons", "aimez", "aiment"},
		Imparfait:           Conjugation{"aimais", "aimais", "aimait", "aimions", "aimiez", "aimaient"},
		FutureSimple:        Conjugation{"aimerai", "aimeras", "aimera", "aimerons", "aimerez", "aimeront"},
		ConditionnelPresent: Conjugation{"aimerais", "aimerais", "aimerait", "aimerions", "aimeriez", "aimeraient"},
		SubjonctifPresent:   Conjugation{"aime", "aimes", "aime", "aimions", "aimiez", "aiment"},
		ImperatifPresent:    Conjugation{Tu: "aime", Nous: "aimons", Vous: "aimez"},
	},
	{
		Infinitive: "chercher", Group: GroupFirst, TranslationEn: "to look for",
		Auxiliary: "avoir", PastParticiple: "cherché", HasImperative: true,
		Contexts: []string{
			"ses clés depuis une heure", "un nouvel appartement", "du travail à Paris",
			"une solution au problème", "son chemin dans la vieille ville", "un cadeau pour sa mère",
			"à comprendre la leçon",
		},
		Present:             Conjugation{"cherche", "cherches", "cherche", "cherchons", "cherchez", "cherchent"},
		Imparfait:           Conjugation{"cherchais", "cherchais", "cherchait", "cherchions", "cherchiez", "cherchaient"},
		FutureSimple:        Conjugation{"chercherai", "chercheras", "cherchera", "chercherons", "chercherez", "chercheront"},
		ConditionnelPresent: Conjugation{"chercherais", "chercherais", "chercherait", "chercherions", "chercheriez", "chercheraient"},
		SubjonctifPresent:   Conjugation{"cherche", "cherches", "cherche", "cherchions", "cherchiez", "cherchent"},
		ImperatifPresent:    Conjugation{Tu: "cherche", Nous: "cherchons", Vous: "cherchez"},
	},
	{
		Infinitive: "habiter", Group: GroupFirst, TranslationEn: "to live (reside)",
		Auxiliary: "avoir", PastParticiple: "habité", HasImperative: true,
		Contexts: []string{
			"à Lyon depuis deux ans", "près de la gare", "dans un petit village", "avec ses parents",
			"au troisième étage", "dans une grande ville animée", "loin de sa famille",
		},
		Present:             Conjugation{"habite", "habites", "habite", "habitons", "habitez", "habitent"},
		Imparfait:           Conjugation{"habitais", "habitais", "habitait", "habitions", "habitiez", "habitaient"},
		FutureSimple:        Conjugation{"habiterai", "habiteras", "habitera", "habiterons", "habiterez", "habiteront"},
		ConditionnelPresent: Conjugation{"habiterais", "habiterais", "habiterait", "habiterions", "habiteriez", "habiteraient"},
		SubjonctifPresent:   Conjugation{"habite", "habites", "habite", "habitions", "habitiez", "habitent"},
		ImperatifPresent:    Conjugation{Tu: "habite", Nous: "habitons", Vous: "habitez"},
	},

	// --- 2e groupe (-ir, type finir) ---
	{
		Infinitive: "finir", Group: GroupSecond, TranslationEn: "to finish",
		Auxiliary: "avoir", PastParticiple: "fini", HasImperative: true,
		Contexts: []string{
			"ses devoirs avant le dîner", "le livre ce week-end", "le projet à temps", "la course en premier",
			"son assiette avant de sortir", "le marathon malgré la fatigue", "la réunion plus tôt que prévu",
		},
		Present:             Conjugation{"finis", "finis", "finit", "finissons", "finissez", "finissent"},
		Imparfait:           Conjugation{"finissais", "finissais", "finissait", "finissions", "finissiez", "finissaient"},
		FutureSimple:        Conjugation{"finirai", "finiras", "finira", "finirons", "finirez", "finiront"},
		ConditionnelPresent: Conjugation{"finirais", "finirais", "finirait", "finirions", "finiriez", "finiraient"},
		SubjonctifPresent:   Conjugation{"finisse", "finisses", "finisse", "finissions", "finissiez", "finissent"},
		ImperatifPresent:    Conjugation{Tu: "finis", Nous: "finissons", Vous: "finissez"},
	},
	{
		Infinitive: "choisir", Group: GroupSecond, TranslationEn: "to choose",
		Auxiliary: "avoir", PastParticiple: "choisi", HasImperative: true,
		Contexts: []string{
			"un dessert au menu", "la meilleure option", "une destination pour les vacances",
			"entre deux candidats", "ses mots avec soin", "un cadeau d'anniversaire", "la couleur des murs",
		},
		Present:             Conjugation{"choisis", "choisis", "choisit", "choisissons", "choisissez", "choisissent"},
		Imparfait:           Conjugation{"choisissais", "choisissais", "choisissait", "choisissions", "choisissiez", "choisissaient"},
		FutureSimple:        Conjugation{"choisirai", "choisiras", "choisira", "choisirons", "choisirez", "choisiront"},
		ConditionnelPresent: Conjugation{"choisirais", "choisirais", "choisirait", "choisirions", "choisiriez", "choisiraient"},
		SubjonctifPresent:   Conjugation{"choisisse", "choisisses", "choisisse", "choisissions", "choisissiez", "choisissent"},
		ImperatifPresent:    Conjugation{Tu: "choisis", Nous: "choisissons", Vous: "choisissez"},
	},
	{
		Infinitive: "réussir", Group: GroupSecond, TranslationEn: "to succeed",
		Auxiliary: "avoir", PastParticiple: "réussi", HasImperative: true,
		Contexts: []string{
			"l'examen de français", "son projet grâce à ses efforts", "à convaincre ses parents",
			"malgré les difficultés", "son entretien d'embauche", "à joindre ses amis",
			"grâce à beaucoup de travail",
		},
		Present:             Conjugation{"réussis", "réussis", "réussit", "réussissons", "réussissez", "réussissent"},
		Imparfait:           Conjugation{"réussissais", "réussissais", "réussissait", "réussissions", "réussissiez", "réussissaient"},
		FutureSimple:        Conjugation{"réussirai", "réussiras", "réussira", "réussirons", "réussirez", "réussiront"},
		ConditionnelPresent: Conjugation{"réussirais", "réussirais", "réussirait", "réussirions", "réussiriez", "réussiraient"},
		SubjonctifPresent:   Conjugation{"réussisse", "réussisses", "réussisse", "réussissions", "réussissiez", "réussissent"},
		ImperatifPresent:    Conjugation{Tu: "réussis", Nous: "réussissons", Vous: "réussissez"},
	},

	// --- 3e groupe (irrégulier) ---
	{
		Infinitive: "être", Group: GroupThird, TranslationEn: "to be",
		Auxiliary: "avoir", PastParticiple: "été", HasImperative: true,
		Contexts: []string{
			"très fatigué après le travail", "en retard ce matin", "content de ce résultat",
			"prêt pour le voyage", "surpris par la nouvelle", "calme avant l'examen", "heureux de les revoir",
		},
		Present:             Conjugation{"suis", "es", "est", "sommes", "êtes", "sont"},
		Imparfait:           Conjugation{"étais", "étais", "était", "étions", "étiez", "étaient"},
		FutureSimple:        Conjugation{"serai", "seras", "sera", "serons", "serez", "seront"},
		ConditionnelPresent: Conjugation{"serais", "serais", "serait", "serions", "seriez", "seraient"},
		SubjonctifPresent:   Conjugation{"sois", "sois", "soit", "soyons", "soyez", "soient"},
		ImperatifPresent:    Conjugation{Tu: "sois", Nous: "soyons", Vous: "soyez"},
	},
	{
		Infinitive: "avoir", Group: GroupThird, TranslationEn: "to have",
		Auxiliary: "avoir", PastParticiple: "eu", HasImperative: true,
		Contexts: []string{
			"beaucoup de patience", "faim après le sport", "besoin d'aide", "de la chance aujourd'hui",
			"peur des araignées", "le temps de discuter", "envie d'un café",
		},
		Present:             Conjugation{"ai", "as", "a", "avons", "avez", "ont"},
		Imparfait:           Conjugation{"avais", "avais", "avait", "avions", "aviez", "avaient"},
		FutureSimple:        Conjugation{"aurai", "auras", "aura", "aurons", "aurez", "auront"},
		ConditionnelPresent: Conjugation{"aurais", "aurais", "aurait", "aurions", "auriez", "auraient"},
		SubjonctifPresent:   Conjugation{"aie", "aies", "ait", "ayons", "ayez", "aient"},
		ImperatifPresent:    Conjugation{Tu: "aie", Nous: "ayons", Vous: "ayez"},
	},
	{
		Infinitive: "aller", Group: GroupThird, TranslationEn: "to go",
		Auxiliary: "être", PastParticiple: "allé", HasImperative: true,
		Contexts: []string{
			"à l'école à pied", "au marché le samedi", "chez le médecin demain", "en vacances cet été",
			"au cinéma ce week-end", "voir ses grands-parents", "à la piscine le dimanche",
		},
		Present:             Conjugation{"vais", "vas", "va", "allons", "allez", "vont"},
		Imparfait:           Conjugation{"allais", "allais", "allait", "allions", "alliez", "allaient"},
		FutureSimple:        Conjugation{"irai", "iras", "ira", "irons", "irez", "iront"},
		ConditionnelPresent: Conjugation{"irais", "irais", "irait", "irions", "iriez", "iraient"},
		SubjonctifPresent:   Conjugation{"aille", "ailles", "aille", "allions", "alliez", "aillent"},
		ImperatifPresent:    Conjugation{Tu: "va", Nous: "allons", Vous: "allez"},
	},
	{
		Infinitive: "faire", Group: GroupThird, TranslationEn: "to do / to make",
		Auxiliary: "avoir", PastParticiple: "fait", HasImperative: true,
		Contexts: []string{
			"la cuisine pour ses amis", "ses devoirs tous les soirs", "du sport le matin",
			"le ménage avant les invités", "la vaisselle après le dîner", "une promenade au parc",
			"de son mieux pour réussir",
		},
		Present:             Conjugation{"fais", "fais", "fait", "faisons", "faites", "font"},
		Imparfait:           Conjugation{"faisais", "faisais", "faisait", "faisions", "faisiez", "faisaient"},
		FutureSimple:        Conjugation{"ferai", "feras", "fera", "ferons", "ferez", "feront"},
		ConditionnelPresent: Conjugation{"ferais", "ferais", "ferait", "ferions", "feriez", "feraient"},
		SubjonctifPresent:   Conjugation{"fasse", "fasses", "fasse", "fassions", "fassiez", "fassent"},
		ImperatifPresent:    Conjugation{Tu: "fais", Nous: "faisons", Vous: "faites"},
	},
	{
		Infinitive: "prendre", Group: GroupThird, TranslationEn: "to take",
		Auxiliary: "avoir", PastParticiple: "pris", HasImperative: true,
		Contexts: []string{
			"le train de huit heures", "une décision importante", "un café avant de partir",
			"des photos pendant le voyage", "des notes pendant le cours", "son temps pour répondre",
			"l'avion pour Rome",
		},
		Present:             Conjugation{"prends", "prends", "prend", "prenons", "prenez", "prennent"},
		Imparfait:           Conjugation{"prenais", "prenais", "prenait", "prenions", "preniez", "prenaient"},
		FutureSimple:        Conjugation{"prendrai", "prendras", "prendra", "prendrons", "prendrez", "prendront"},
		ConditionnelPresent: Conjugation{"prendrais", "prendrais", "prendrait", "prendrions", "prendriez", "prendraient"},
		SubjonctifPresent:   Conjugation{"prenne", "prennes", "prenne", "prenions", "preniez", "prennent"},
		ImperatifPresent:    Conjugation{Tu: "prends", Nous: "prenons", Vous: "prenez"},
	},
	{
		Infinitive: "pouvoir", Group: GroupThird, TranslationEn: "to be able to (can)",
		Auxiliary: "avoir", PastParticiple: "pu", HasImperative: false, // no natural imperative form
		Contexts: []string{
			"venir à la fête demain", "finir le travail à temps", "partir plus tôt",
			"aider ses amis ce week-end", "répondre à la question", "comprendre le problème",
			"se reposer ce soir",
		},
		Present:             Conjugation{"peux", "peux", "peut", "pouvons", "pouvez", "peuvent"},
		Imparfait:           Conjugation{"pouvais", "pouvais", "pouvait", "pouvions", "pouviez", "pouvaient"},
		FutureSimple:        Conjugation{"pourrai", "pourras", "pourra", "pourrons", "pourrez", "pourront"},
		ConditionnelPresent: Conjugation{"pourrais", "pourrais", "pourrait", "pourrions", "pourriez", "pourraient"},
		SubjonctifPresent:   Conjugation{"puisse", "puisses", "puisse", "puissions", "puissiez", "puissent"},
	},
	{
		Infinitive: "vouloir", Group: GroupThird, TranslationEn: "to want",
		Auxiliary: "avoir", PastParticiple: "voulu", HasImperative: true, // mostly used in the fixed polite form "veuillez"
		Contexts: []string{
			"apprendre le français", "partir en vacances", "changer de travail", "acheter une nouvelle voiture",
			"revoir ce film", "essayer ce restaurant", "déménager l'année prochaine",
		},
		Present:             Conjugation{"veux", "veux", "veut", "voulons", "voulez", "veulent"},
		Imparfait:           Conjugation{"voulais", "voulais", "voulait", "voulions", "vouliez", "voulaient"},
		FutureSimple:        Conjugation{"voudrai", "voudras", "voudra", "voudrons", "voudrez", "voudront"},
		ConditionnelPresent: Conjugation{"voudrais", "voudrais", "voudrait", "voudrions", "voudriez", "voudraient"},
		SubjonctifPresent:   Conjugation{"veuille", "veuilles", "veuille", "voulions", "vouliez", "veuillent"},
		ImperatifPresent:    Conjugation{Tu: "veuille", Nous: "veuillons", Vous: "veuillez"},
	},
	{
		Infinitive: "venir", Group: GroupThird, TranslationEn: "to come",
		Auxiliary: "être", PastParticiple: "venu", HasImperative: true,
		Contexts: []string{
			"chez nous dimanche", "de Paris", "avec nous au cinéma", "plus tard dans la soirée",
			"nous voir après le travail", "pour l'anniversaire de sa fille", "avant la tombée de la nuit",
		},
		Present:             Conjugation{"viens", "viens", "vient", "venons", "venez", "viennent"},
		Imparfait:           Conjugation{"venais", "venais", "venait", "venions", "veniez", "venaient"},
		FutureSimple:        Conjugation{"viendrai", "viendras", "viendra", "viendrons", "viendrez", "viendront"},
		ConditionnelPresent: Conjugation{"viendrais", "viendrais", "viendrait", "viendrions", "viendriez", "viendraient"},
		SubjonctifPresent:   Conjugation{"vienne", "viennes", "vienne", "venions", "veniez", "viennent"},
		ImperatifPresent:    Conjugation{Tu: "viens", Nous: "venons", Vous: "venez"},
	},
}
