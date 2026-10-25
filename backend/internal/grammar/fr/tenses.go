package fr

import "github.com/bogdanbojan/sermokey/backend/internal/grammar"

// TenseTheory is a re-export of the shared theory shape.
type TenseTheory = grammar.TenseTheory
type GroupEndings = grammar.GroupEndings

// Theories holds the theory content for every tense, keyed by TenseID.
var Theories = map[TenseID]TenseTheory{
	Present: {
		ID: Present, NameNative: "Présent", NameEn: "Present",
		Usage:     "Actions happening right now, habitual actions, and general truths.",
		Formation: "Drop the infinitive ending (-er, -ir, -re) and add the présent endings for the verb's group.",
		Groups: []GroupEndings{
			{GroupName: string(GroupFirst), Explanation: "-er verbs (e.g. parler).",
				Endings:     [6]string{"-e", "-es", "-e", "-ons", "-ez", "-ent"},
				ExampleVerb: "parler", ExampleForms: [6]string{"parle", "parles", "parle", "parlons", "parlez", "parlent"}},
			{GroupName: string(GroupSecond), Explanation: "Regular -ir verbs (e.g. finir) insert -iss- in the plural.",
				Endings:     [6]string{"-is", "-is", "-it", "-issons", "-issez", "-issent"},
				ExampleVerb: "finir", ExampleForms: [6]string{"finis", "finis", "finit", "finissons", "finissez", "finissent"}},
			{GroupName: string(GroupThird), Explanation: "Irregular verbs — stems and endings vary and must be memorized.",
				ExampleVerb: "être / avoir", ExampleForms: [6]string{"suis / ai", "es / as", "est / a", "sommes / avons", "êtes / avez", "sont / ont"}},
		},
		IrregularNotes: "Many 3e groupe verbs still share a family pattern, e.g. verbs ending in -endre (prendre, attendre) go -ends, -ends, -end, -enons, -enez, -ennent.",
	},
	Imparfait: {
		ID: Imparfait, NameNative: "Imparfait", NameEn: "Imperfect",
		Usage:     "Descriptions in the past, habitual or repeated past actions, and ongoing states ('was doing', 'used to do').",
		Formation: "Take the nous form of the présent, drop -ons to get the stem, then add the imparfait endings. The endings are the same for every verb (only être has an irregular stem: ét-).",
		Groups: []GroupEndings{
			{GroupName: "Toutes les verbes", Explanation: "One universal set of endings applies to every verb, regular or irregular.",
				Endings:     [6]string{"-ais", "-ais", "-ait", "-ions", "-iez", "-aient"},
				ExampleVerb: "parler / finir / être", ExampleForms: [6]string{"parlais / finissais / étais", "parlais / finissais / étais", "parlait / finissait / était", "parlions / finissions / étions", "parliez / finissiez / étiez", "parlaient / finissaient / étaient"}},
		},
		IrregularNotes: "Spelling-change -er verbs like manger keep the soft 'g' sound by adding an e before endings starting with a: je mangeais, but nous mangions (no e needed before i).",
	},
	FutureSimple: {
		ID: FutureSimple, NameNative: "Futur simple", NameEn: "Simple future",
		Usage:     "Actions that will happen in the future ('will do').",
		Formation: "Add the futur endings directly to the infinitive (for -re verbs, drop the final e first). The endings are always the same — only the stem changes.",
		Groups: []GroupEndings{
			{GroupName: "Toutes les verbes", Explanation: "One universal set of endings applies to every verb.",
				Endings:     [6]string{"-ai", "-as", "-a", "-ons", "-ez", "-ont"},
				ExampleVerb: "parler / finir", ExampleForms: [6]string{"parlerai / finirai", "parleras / finiras", "parlera / finira", "parlerons / finirons", "parlerez / finirez", "parleront / finiront"}},
		},
		IrregularNotes: "A handful of very common verbs have an irregular stem that must be memorized, e.g. être→ser-, avoir→aur-, aller→ir-, faire→fer-, pouvoir→pourr-, vouloir→voudr-, venir→viendr-. The endings on top of those stems never change.",
	},
	ConditionnelPresent: {
		ID: ConditionnelPresent, NameNative: "Conditionnel présent", NameEn: "Present conditional",
		Usage:     "Hypothetical actions ('would do'), polite requests, and suggestions.",
		Formation: "Same stem as the futur simple (including any irregular stem), plus the imparfait endings.",
		Groups: []GroupEndings{
			{GroupName: "Toutes les verbes", Explanation: "Futur stem + imparfait endings.",
				Endings:     [6]string{"-ais", "-ais", "-ait", "-ions", "-iez", "-aient"},
				ExampleVerb: "parler / être", ExampleForms: [6]string{"parlerais / serais", "parlerais / serais", "parlerait / serait", "parlerions / serions", "parleriez / seriez", "parleraient / seraient"}},
		},
		IrregularNotes: "Because it reuses the futur stem, any verb with an irregular future (voudr-, viendr-, ser-, ...) carries that same irregularity into the conditionnel.",
	},
	SubjonctifPresent: {
		ID: SubjonctifPresent, NameNative: "Subjonctif présent", NameEn: "Present subjunctive",
		Usage:     "Used after certain expressions of necessity, emotion, doubt, or wish — almost always introduced by 'que' (e.g. 'il faut que...').",
		Formation: "Take the ils/elles form of the présent, drop -ent, add -e/-es/-e/-ent for je/tu/il/ils. The nous/vous forms often fall back to the imparfait stem instead.",
		Groups: []GroupEndings{
			{GroupName: "Toutes les verbes", Explanation: "General pattern; many irregular verbs have a second stem for nous/vous.",
				Endings:     [6]string{"-e", "-es", "-e", "-ions", "-iez", "-ent"},
				ExampleVerb: "parler / finir", ExampleForms: [6]string{"que je parle / finisse", "que tu parles / finisses", "qu'il parle / finisse", "que nous parlions / finissions", "que vous parliez / finissiez", "qu'ils parlent / finissent"}},
		},
		IrregularNotes: "Several common irregular verbs use two stems: one for je/tu/il/ils and a different one (often matching the imparfait) for nous/vous — e.g. venir: que je vienne but que nous venions. Être and avoir are fully irregular: que je sois, que j'aie.",
	},
	ImperatifPresent: {
		ID: ImperatifPresent, NameNative: "Impératif présent", NameEn: "Present imperative",
		Usage:     "Commands, instructions, and suggestions. No subject pronoun is used — only three forms exist: tu, nous, vous.",
		Formation: "Take the tu/nous/vous forms of the présent and drop the pronoun. For -er verbs, the tu form also drops the final -s.",
		Groups: []GroupEndings{
			{GroupName: string(GroupFirst), Explanation: "-er verbs drop the -s from the tu form.",
				Endings:     [6]string{"-e", "", "", "-ons", "-ez", ""},
				ExampleVerb: "parler", ExampleForms: [6]string{"parle", "", "", "parlons", "parlez", ""}},
			{GroupName: string(GroupSecond), Explanation: "Regular -ir verbs keep the présent forms as-is.",
				Endings:     [6]string{"-is", "", "", "-issons", "-issez", ""},
				ExampleVerb: "finir", ExampleForms: [6]string{"finis", "", "", "finissons", "finissez", ""}},
		},
		IrregularNotes: "Être, avoir, and savoir borrow their stem from the subjonctif instead of the présent: sois, soyons, soyez / aie, ayons, ayez. A few verbs (like pouvoir) have no natural imperative at all.",
	},
	PasseCompose: {
		ID: PasseCompose, NameNative: "Passé composé", NameEn: "Compound past",
		Usage:     "A completed action in the past — roughly the English simple past ('spoke') or present perfect ('have spoken').",
		Formation: "Auxiliary verb (avoir or être) conjugated in the présent, plus the past participle. Most verbs use avoir. A small set of verbs of movement/state (aller, venir, naître, mourir, partir, sortir, entrer, rester, retourner, arriver, descendre, monter, tomber — the 'Dr & Mrs Vandertramp' verbs) plus all reflexive/pronominal verbs use être instead.",
		Groups: []GroupEndings{
			{GroupName: string(GroupFirst), Explanation: "Past participle: -er → -é.",
				ExampleVerb: "parler", ExampleForms: [6]string{"ai parlé", "as parlé", "a parlé", "avons parlé", "avez parlé", "ont parlé"}},
			{GroupName: string(GroupSecond), Explanation: "Past participle: -ir → -i.",
				ExampleVerb: "finir", ExampleForms: [6]string{"ai fini", "as fini", "a fini", "avons fini", "avez fini", "ont fini"}},
			{GroupName: string(GroupThird), Explanation: "Past participle is irregular and must be memorized (e.g. faire→fait, prendre→pris).",
				ExampleVerb: "aller (être verb)", ExampleForms: [6]string{"suis allé", "es allé", "est allé", "sommes allés", "êtes allés", "sont allés"}},
		},
		IrregularNotes: "With être, the past participle agrees in gender and number with the subject (elle est allée, ils sont allés, elles sont allées). This app shows the default masculine form for simplicity.",
	},
	PlusQueParfait: {
		ID: PlusQueParfait, NameNative: "Plus-que-parfait", NameEn: "Pluperfect",
		Usage:     "An action that had already happened before another past action ('had done').",
		Formation: "Same auxiliary and agreement rules as the passé composé, but the auxiliary is conjugated in the imparfait instead of the présent.",
		Groups: []GroupEndings{
			{GroupName: string(GroupFirst), Explanation: "avoir (imparfait) + past participle.",
				ExampleVerb: "parler", ExampleForms: [6]string{"avais parlé", "avais parlé", "avait parlé", "avions parlé", "aviez parlé", "avaient parlé"}},
			{GroupName: string(GroupThird), Explanation: "être (imparfait) + past participle, for être-verbs.",
				ExampleVerb: "aller", ExampleForms: [6]string{"étais allé", "étais allé", "était allé", "étions allés", "étiez allés", "étaient allés"}},
		},
		IrregularNotes: "Same être/avoir choice and past-participle agreement rules as the passé composé apply here.",
	},
	FuturAnterieur: {
		ID: FuturAnterieur, NameNative: "Futur antérieur", NameEn: "Future perfect",
		Usage:     "An action that will already be finished before a future point in time ('will have done') — often in sentences with quand/dès que + futur antérieur, futur simple.",
		Formation: "Same auxiliary and agreement rules as the passé composé, but the auxiliary is conjugated in the futur simple instead of the présent.",
		Groups: []GroupEndings{
			{GroupName: string(GroupFirst), Explanation: "avoir (futur simple) + past participle.",
				ExampleVerb: "parler", ExampleForms: [6]string{"aurai parlé", "auras parlé", "aura parlé", "aurons parlé", "aurez parlé", "auront parlé"}},
			{GroupName: string(GroupThird), Explanation: "être (futur simple) + past participle, for être-verbs.",
				ExampleVerb: "aller", ExampleForms: [6]string{"serai allé", "seras allé", "sera allé", "serons allés", "serez allés", "seront allés"}},
		},
		IrregularNotes: "Same être/avoir choice and past-participle agreement rules as the passé composé apply here. Example usage: « Quand j'aurai fini, je t'appellerai. »",
	},
	ConditionnelPasse: {
		ID: ConditionnelPasse, NameNative: "Conditionnel passé", NameEn: "Past conditional",
		Usage:     "A hypothetical action that didn't happen ('would have done') — commonly the result clause of a 'si + plus-que-parfait' sentence, and used for regret.",
		Formation: "Same auxiliary and agreement rules as the passé composé, but the auxiliary is conjugated in the conditionnel présent instead of the présent.",
		Groups: []GroupEndings{
			{GroupName: string(GroupFirst), Explanation: "avoir (conditionnel présent) + past participle.",
				ExampleVerb: "parler", ExampleForms: [6]string{"aurais parlé", "aurais parlé", "aurait parlé", "aurions parlé", "auriez parlé", "auraient parlé"}},
			{GroupName: string(GroupThird), Explanation: "être (conditionnel présent) + past participle, for être-verbes.",
				ExampleVerb: "aller", ExampleForms: [6]string{"serais allé", "serais allé", "serait allé", "serions allés", "seriez allés", "seraient allés"}},
		},
		IrregularNotes: "Example usage: « Si j'avais su, je serais venu plus tôt. » Same être/avoir and agreement rules as the passé composé.",
	},
	SubjonctifPasse: {
		ID: SubjonctifPasse, NameNative: "Subjonctif passé", NameEn: "Past subjunctive",
		Usage:     "A completed action in a subjonctif-triggering clause ('that ... have done') — used when the subordinate action happened before the main clause's action.",
		Formation: "Same auxiliary and agreement rules as the passé composé, but the auxiliary is conjugated in the subjonctif présent instead of the présent. Also introduced by 'que', like the subjonctif présent.",
		Groups: []GroupEndings{
			{GroupName: string(GroupFirst), Explanation: "avoir (subjonctif présent) + past participle.",
				ExampleVerb: "parler", ExampleForms: [6]string{"que j'aie parlé", "que tu aies parlé", "qu'il ait parlé", "que nous ayons parlé", "que vous ayez parlé", "qu'ils aient parlé"}},
			{GroupName: string(GroupThird), Explanation: "être (subjonctif présent) + past participle, for être-verbes.",
				ExampleVerb: "aller", ExampleForms: [6]string{"que je sois allé", "que tu sois allé", "qu'il soit allé", "que nous soyons allés", "que vous soyez allés", "qu'ils soient allés"}},
		},
		IrregularNotes: "Example usage: « Je suis content qu'elle soit arrivée à temps. » Same être/avoir and agreement rules as the passé composé.",
	},
	ImperatifPasse: {
		ID: ImperatifPasse, NameNative: "Impératif passé", NameEn: "Past imperative",
		Usage:     "A rare, formal command about something that must be completed by a future point ('have done X by then') — only tu/nous/vous exist, like the impératif présent.",
		Formation: "Same auxiliary and agreement rules as the passé composé, but the auxiliary is conjugated in the impératif présent instead of the présent, and no subject pronoun is used.",
		Groups: []GroupEndings{
			{GroupName: string(GroupFirst), Explanation: "avoir (impératif) + past participle.",
				Endings:     [6]string{"", "", "", "", "", ""},
				ExampleVerb: "parler", ExampleForms: [6]string{"", "aie parlé", "", "ayons parlé", "ayez parlé", ""}},
		},
		IrregularNotes: "Rarely used in everyday speech; mostly seen in formal writing, e.g. « Ayez terminé ce rapport avant vendredi. » A few verbs (like pouvoir) have no imperative at all, present or passé.",
	},
	PresentContinu: {
		ID: PresentContinu, NameNative: "Présent continu", NameEn: "Present progressive",
		Usage:     "Emphasizes that an action is happening right at this moment ('is doing') — more emphatic than the simple présent, which can also cover this meaning.",
		Formation: "être conjugated in the présent + 'en train de' + the infinitive of the verb being practiced. The infinitive never changes; only être agrees with the subject.",
		Groups: []GroupEndings{
			{GroupName: "Toutes les verbes", Explanation: "être (présent) + en train de + infinitif.",
				ExampleVerb: "parler", ExampleForms: [6]string{"suis en train de parler", "es en train de parler", "est en train de parler", "sommes en train de parler", "êtes en train de parler", "sont en train de parler"}},
		},
		IrregularNotes: "Only être's conjugation varies; practicing this tense mostly tests whether you remember the fixed phrase 'en train de' plus être's présent.",
	},
	PasseRecent: {
		ID: PasseRecent, NameNative: "Passé récent", NameEn: "Recent past",
		Usage:     "An action that just happened a moment ago ('have just done').",
		Formation: "venir conjugated in the présent + 'de' (or 'd'' before a vowel sound) + the infinitive of the verb being practiced.",
		Groups: []GroupEndings{
			{GroupName: "Toutes les verbes", Explanation: "venir (présent) + de/d' + infinitif.",
				ExampleVerb: "parler / arriver", ExampleForms: [6]string{"viens de parler / viens d'arriver", "viens de parler / viens d'arriver", "vient de parler / vient d'arriver", "venons de parler / venons d'arriver", "venez de parler / venez d'arriver", "viennent de parler / viennent d'arriver"}},
		},
		IrregularNotes: "'De' elides to 'd'' before an infinitive starting with a vowel sound (arriver, avoir, étudier, habiter...), just like 'je'→'j'' and 'que'→'qu''.",
	},
	FuturProche: {
		ID: FuturProche, NameNative: "Futur proche", NameEn: "Near future",
		Usage:     "An action about to happen soon ('going to do') — the everyday, spoken-French alternative to the futur simple.",
		Formation: "aller conjugated in the présent + the infinitive of the verb being practiced, with no connecting word.",
		Groups: []GroupEndings{
			{GroupName: "Toutes les verbes", Explanation: "aller (présent) + infinitif.",
				ExampleVerb: "parler", ExampleForms: [6]string{"vais parler", "vas parler", "va parler", "allons parler", "allez parler", "vont parler"}},
		},
		IrregularNotes: "No elision or connecting word is needed here, unlike passé récent's 'de'.",
	},
	InfinitifPasse: {
		ID: InfinitifPasse, NameNative: "Infinitif passé", NameEn: "Past infinitive",
		Usage:     "A completed action, used after certain conjunctions — most commonly 'après' ('after having done'). Unlike every tense above, this form never agrees with a subject: there is exactly one form per verb.",
		Formation: "avoir or être at the infinitive (i.e. literally the word 'avoir'/'être') + the past participle. Agreement with être still applies if the understood subject requires it.",
		Groups: []GroupEndings{
			{GroupName: "Toutes les verbes", Explanation: "avoir/être (infinitif) + past participle — a single invariable form.",
				ExampleVerb: "parler / aller", ExampleForms: [6]string{"avoir parlé / être allé", "", "", "", "", ""}},
		},
		IrregularNotes: "Contrast with the infinitif présent, which is simply the dictionary form (parler) — already shown as the hint on every exercise in this app.",
	},
	GerondifPresent: {
		ID: GerondifPresent, NameNative: "Gérondif présent", NameEn: "Present gerund",
		Usage:     "Describes an action happening at the same time as the main verb, or how/why something happens ('by doing', 'while doing'). Invariable: one form per verb, no subject agreement.",
		Formation: "'en' + the participe présent (take the nous form of the présent, drop -ons, add -ant).",
		Groups: []GroupEndings{
			{GroupName: "Toutes les verbes", Explanation: "en + (nous-form stem) + ant — a single invariable form.",
				ExampleVerb: "parler / finir", ExampleForms: [6]string{"en parlant / en finissant", "", "", "", "", ""}},
		},
		IrregularNotes: "Only avoir (en ayant) and être (en étant) have an irregular participe présent; every other verb in this app follows the regular rule.",
	},
	GerondifPasse: {
		ID: GerondifPasse, NameNative: "Gérondif passé", NameEn: "Perfect participle",
		Usage:     "A formal, mostly written form expressing a completed action before the main clause ('having done X, ...'). Invariable: one form per verb, no subject agreement.",
		Formation: "ayant or étant (the participe présent of avoir/être) + the past participle — no 'en' this time, unlike the gérondif présent.",
		Groups: []GroupEndings{
			{GroupName: "Toutes les verbes", Explanation: "ayant/étant + past participle — a single invariable form.",
				ExampleVerb: "parler / aller", ExampleForms: [6]string{"ayant parlé / étant allé", "", "", "", "", ""}},
		},
		IrregularNotes: "Mostly seen in formal writing, e.g. « Ayant terminé son travail, elle est rentrée chez elle. » Rare in everyday speech.",
	},
}
