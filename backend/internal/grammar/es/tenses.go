package es

import "github.com/bogdanbojan/sermokey/backend/internal/grammar"

type TenseTheory = grammar.TenseTheory
type GroupEndings = grammar.GroupEndings

// Theories holds the theory content for every tense, keyed by TenseID.
var Theories = map[TenseID]TenseTheory{
	Present: {
		ID: Present, NameNative: "Presente", NameEn: "Present",
		Usage:     "Actions happening right now, habitual actions, and general truths.",
		Formation: "Drop the infinitive ending (-ar, -er, -ir) and add the presente endings for the verb's class.",
		Groups: []GroupEndings{
			{GroupName: string(GroupAR), Explanation: "-ar verbs (e.g. hablar).",
				Endings:     [6]string{"-o", "-as", "-a", "-amos", "-áis", "-an"},
				ExampleVerb: "hablar", ExampleForms: [6]string{"hablo", "hablas", "habla", "hablamos", "habláis", "hablan"}},
			{GroupName: string(GroupER), Explanation: "-er verbs (e.g. comer).",
				Endings:     [6]string{"-o", "-es", "-e", "-emos", "-éis", "-en"},
				ExampleVerb: "comer", ExampleForms: [6]string{"como", "comes", "come", "comemos", "coméis", "comen"}},
			{GroupName: string(GroupIR), Explanation: "-ir verbs (e.g. vivir) — identical to -er except nosotros/vosotros.",
				Endings:     [6]string{"-o", "-es", "-e", "-imos", "-ís", "-en"},
				ExampleVerb: "vivir", ExampleForms: [6]string{"vivo", "vives", "vive", "vivimos", "vivís", "viven"}},
		},
		IrregularNotes: "Many verbs change their stem vowel when it's stressed (e→ie: pensar→pienso; o→ue: volver→vuelve; e→i: pedir→pide) — nosotros/vosotros keep the unstressed, unchanged stem (pensamos, volvemos, pedimos). Others are fully irregular (ser, estar, ir, haber) and must be memorized.",
	},
	Imperfect: {
		ID: Imperfect, NameNative: "Pretérito imperfecto", NameEn: "Imperfect",
		Usage:     "Descriptions in the past, habitual or repeated past actions, and ongoing states ('was doing', 'used to do'). Never has a stem change, even for verbs that are stem-changing in the presente.",
		Formation: "-ar verbs: stem + -aba/-abas/-aba/-ábamos/-abais/-aban. -er/-ir verbs: stem + -ía/-ías/-ía/-íamos/-íais/-ían.",
		Groups: []GroupEndings{
			{GroupName: string(GroupAR), Explanation: "-ar verbs.",
				Endings:     [6]string{"-aba", "-abas", "-aba", "-ábamos", "-abais", "-aban"},
				ExampleVerb: "hablar", ExampleForms: [6]string{"hablaba", "hablabas", "hablaba", "hablábamos", "hablabais", "hablaban"}},
			{GroupName: "-er / -ir", Explanation: "-er and -ir verbs share the same endings.",
				Endings:     [6]string{"-ía", "-ías", "-ía", "-íamos", "-íais", "-ían"},
				ExampleVerb: "comer / vivir", ExampleForms: [6]string{"comía / vivía", "comías / vivías", "comía / vivía", "comíamos / vivíamos", "comíais / vivíais", "comían / vivían"}},
		},
		IrregularNotes: "Only three verbs are irregular here: ser (era, eras, era, éramos, erais, eran), ir (iba, ibas, iba, íbamos, ibais, iban), and ver (not covered in this app). Every other verb — even otherwise very irregular ones like tener or hacer — is fully regular in the imperfecto.",
	},
	Preterite: {
		ID: Preterite, NameNative: "Pretérito indefinido", NameEn: "Preterite (simple past)",
		Usage:     "A completed action at a specific point in the past ('spoke', 'ate') — the everyday simple past, much more central in spoken Spanish than French's passé composé-style distinction between completed and ongoing past actions.",
		Formation: "-ar verbs: stem + -é/-aste/-ó/-amos/-asteis/-aron. -er/-ir verbs: stem + -í/-iste/-ió/-imos/-isteis/-ieron.",
		Groups: []GroupEndings{
			{GroupName: string(GroupAR), Explanation: "-ar verbs.",
				Endings:     [6]string{"-é", "-aste", "-ó", "-amos", "-asteis", "-aron"},
				ExampleVerb: "hablar", ExampleForms: [6]string{"hablé", "hablaste", "habló", "hablamos", "hablasteis", "hablaron"}},
			{GroupName: "-er / -ir", Explanation: "-er and -ir verbs share the same endings.",
				Endings:     [6]string{"-í", "-iste", "-ió", "-imos", "-isteis", "-ieron"},
				ExampleVerb: "comer / vivir", ExampleForms: [6]string{"comí / viví", "comiste / viviste", "comió / vivió", "comimos / vivimos", "comisteis / vivisteis", "comieron / vivieron"}},
		},
		IrregularNotes: "-ar/-er verbs that spelling-change in the presente (buscar→busqué, llegar→llegué) only change in the yo form here, to keep the same consonant sound before é. -ir stem-changers (pedir→pidió/pidieron) change in the 3rd persons only. Many common verbs have a fully irregular preterite stem (tener→tuve, hacer→hice, poder→pude, querer→quise, estar→estuve) with a shared, different set of endings (-e, -iste, -o, -imos, -isteis, -ieron — no accents). Ser and ir are identical in this tense: fui, fuiste, fue, fuimos, fuisteis, fueron.",
	},
	FutureSimple: {
		ID: FutureSimple, NameNative: "Futuro simple", NameEn: "Simple future",
		Usage:     "Actions that will happen in the future ('will do').",
		Formation: "Add the futuro endings directly to the infinitive — unlike the other simple tenses, there's no difference between -ar/-er/-ir here.",
		Groups: []GroupEndings{
			{GroupName: "Todos los verbos", Explanation: "One universal set of endings applies to every verb.",
				Endings:     [6]string{"-é", "-ás", "-á", "-emos", "-éis", "-án"},
				ExampleVerb: "hablar / comer / vivir", ExampleForms: [6]string{"hablaré / comeré / viviré", "hablarás / comerás / vivirás", "hablará / comerá / vivirá", "hablaremos / comeremos / viviremos", "hablaréis / comeréis / viviréis", "hablarán / comerán / vivirán"}},
		},
		IrregularNotes: "A dozen common verbs have an irregular stem (the infinitive gets shortened or altered) but keep the same endings: tener→tendr-, haber→habr-, poder→podr-, querer→querr-, hacer→har-, decir→dir- (not covered in this app). E.g. tener→tendré, tendrás, tendrá...",
	},
	ConditionalPresent: {
		ID: ConditionalPresent, NameNative: "Condicional simple", NameEn: "Simple conditional",
		Usage:     "Hypothetical actions ('would do'), polite requests, and suggestions.",
		Formation: "Same stem as the futuro simple (including any irregular stem), plus a universal set of endings identical to the imperfecto's -er/-ir endings.",
		Groups: []GroupEndings{
			{GroupName: "Todos los verbos", Explanation: "Futuro stem + -ía/-ías/-ía/-íamos/-íais/-ían.",
				Endings:     [6]string{"-ía", "-ías", "-ía", "-íamos", "-íais", "-ían"},
				ExampleVerb: "hablar / tener", ExampleForms: [6]string{"hablaría / tendría", "hablarías / tendrías", "hablaría / tendría", "hablaríamos / tendríamos", "hablaríais / tendríais", "hablarían / tendrían"}},
		},
		IrregularNotes: "Because it reuses the futuro stem, any verb with an irregular future (tendr-, podr-, har-, ...) carries that same irregularity into the condicional.",
	},
	SubjunctivePresent: {
		ID: SubjunctivePresent, NameNative: "Subjuntivo presente", NameEn: "Present subjunctive",
		Usage:     "Used after expressions of wish, doubt, emotion, or necessity — almost always introduced by 'que' (e.g. 'espero que...', 'quiero que...').",
		Formation: "Take the yo form of the presente, drop the -o, and add the 'opposite vowel' endings: -ar verbs get e-endings, -er/-ir verbs get a-endings.",
		Groups: []GroupEndings{
			{GroupName: string(GroupAR), Explanation: "-ar verbs take e-endings.",
				Endings:     [6]string{"-e", "-es", "-e", "-emos", "-éis", "-en"},
				ExampleVerb: "hablar", ExampleForms: [6]string{"que hable", "que hables", "que hable", "que hablemos", "que habléis", "que hablen"}},
			{GroupName: "-er / -ir", Explanation: "-er and -ir verbs take a-endings.",
				Endings:     [6]string{"-a", "-as", "-a", "-amos", "-áis", "-an"},
				ExampleVerb: "comer / vivir", ExampleForms: [6]string{"que coma / viva", "que comas / vivas", "que coma / viva", "que comamos / vivamos", "que comáis / viváis", "que coman / vivan"}},
		},
		IrregularNotes: "-ir stem-changers (e→i, e.g. pedir) keep the change in every person, including nosotros/vosotros: pidamos, pidáis. -ar/-er stem-changers (e→ie, o→ue) behave like the presente: unchanged in nosotros/vosotros (pensemos, volvamos). Since the subjuntivo stem comes from the irregular yo form, verbs irregular there stay irregular here too: tener→tenga, hacer→haga, poder→pueda.",
	},
	ImperativePresent: {
		ID: ImperativePresent, NameNative: "Imperativo afirmativo", NameEn: "Affirmative imperative",
		Usage:     "Commands, instructions, and suggestions. Five forms exist: tú, usted, nosotros, vosotros, ustedes — unlike the indicative, 'usted'/'ustedes' commands borrow their form from the subjuntivo presente.",
		Formation: "tú: like the presente's él/ella form (with some irregular short forms: ten, ve, haz, sé...). usted/ustedes/nosotros: identical to the subjuntivo presente. vosotros: infinitive with the final -r replaced by -d.",
		Groups: []GroupEndings{
			{GroupName: string(GroupAR), Explanation: "tú drops -as from the presente form; vosotros replaces -r with -d.",
				Endings:     [6]string{"", "-a", "(subj.)", "(subj.)", "-d", "(subj.)"},
				ExampleVerb: "hablar", ExampleForms: [6]string{"", "habla", "hable", "hablemos", "hablad", "hablen"}},
			{GroupName: "-er / -ir", Explanation: "Same pattern as -ar, with -e instead of -a for tú.",
				Endings:     [6]string{"", "-e", "(subj.)", "(subj.)", "-d", "(subj.)"},
				ExampleVerb: "comer / vivir", ExampleForms: [6]string{"", "come / vive", "coma / viva", "comamos / vivamos", "comed / vivid", "coman / vivan"}},
		},
		IrregularNotes: "A handful of very common verbs have a short, irregular tú form: ten (tener), ve (ir), haz (hacer), sé (ser), di (decir, not covered). A few verbs (like haber, poder) have no natural imperative at all. Negative commands ('don't speak') are not covered here: they always use the subjuntivo presente for every person, including tú.",
	},
	PresentPerfect: {
		ID: PresentPerfect, NameNative: "Pretérito perfecto compuesto", NameEn: "Present perfect",
		Usage:     "A completed action with some connection to the present ('have done') — used more in Spain than in much of Latin America, where the pretérito indefinido often covers this meaning too.",
		Formation: "haber conjugated in the presente + the past participle. Unlike French, there is only ever one auxiliary (haber), and the participle never agrees with the subject.",
		Groups: []GroupEndings{
			{GroupName: "Todos los verbos", Explanation: "haber (presente) + participio — the participle is always invariable.",
				ExampleVerb: "hablar", ExampleForms: [6]string{"he hablado", "has hablado", "ha hablado", "hemos hablado", "habéis hablado", "han hablado"}},
		},
		IrregularNotes: "Participle irregulars to memorize: hacer→hecho, volver→vuelto (and, outside this app's verb list, decir→dicho, poner→puesto, ver→visto, escribir→escrito, romper→roto, abrir→abierto).",
	},
	Pluperfect: {
		ID: Pluperfect, NameNative: "Pluscuamperfecto", NameEn: "Pluperfect",
		Usage:     "An action that had already happened before another past action ('had done').",
		Formation: "Same participle rules as the pretérito perfecto compuesto, but haber is conjugated in the imperfecto instead of the presente.",
		Groups: []GroupEndings{
			{GroupName: "Todos los verbos", Explanation: "haber (imperfecto) + participio.",
				ExampleVerb: "hablar", ExampleForms: [6]string{"había hablado", "habías hablado", "había hablado", "habíamos hablado", "habíais hablado", "habían hablado"}},
		},
		IrregularNotes: "Same participle irregulars as the pretérito perfecto compuesto apply here.",
	},
	FuturePerfect: {
		ID: FuturePerfect, NameNative: "Futuro perfecto", NameEn: "Future perfect",
		Usage:     "An action that will already be finished before a future point in time ('will have done').",
		Formation: "Same participle rules as the pretérito perfecto compuesto, but haber is conjugated in the futuro simple instead of the presente.",
		Groups: []GroupEndings{
			{GroupName: "Todos los verbos", Explanation: "haber (futuro simple) + participio.",
				ExampleVerb: "hablar", ExampleForms: [6]string{"habré hablado", "habrás hablado", "habrá hablado", "habremos hablado", "habréis hablado", "habrán hablado"}},
		},
		IrregularNotes: "Example usage: « Cuando llegues, ya habré terminado. » Same participle irregulars as the pretérito perfecto compuesto.",
	},
	ConditionalPerfect: {
		ID: ConditionalPerfect, NameNative: "Condicional compuesto", NameEn: "Conditional perfect",
		Usage:     "A hypothetical action that didn't happen ('would have done') — commonly the result clause of a 'si + pluscuamperfecto' sentence, and used for regret.",
		Formation: "Same participle rules as the pretérito perfecto compuesto, but haber is conjugated in the condicional simple instead of the presente.",
		Groups: []GroupEndings{
			{GroupName: "Todos los verbos", Explanation: "haber (condicional simple) + participio.",
				ExampleVerb: "hablar", ExampleForms: [6]string{"habría hablado", "habrías hablado", "habría hablado", "habríamos hablado", "habríais hablado", "habrían hablado"}},
		},
		IrregularNotes: "Example usage: « Si lo hubiera sabido, habría venido antes. » Same participle irregulars as the pretérito perfecto compuesto.",
	},
	SubjunctivePerfect: {
		ID: SubjunctivePerfect, NameNative: "Pretérito perfecto de subjuntivo", NameEn: "Present perfect subjunctive",
		Usage:     "A completed action in a subjunctive-triggering clause ('that ... have done') — the subordinate action happened before the main clause's action. Also introduced by 'que'.",
		Formation: "Same participle rules as the pretérito perfecto compuesto, but haber is conjugated in the subjuntivo presente instead of the presente.",
		Groups: []GroupEndings{
			{GroupName: "Todos los verbos", Explanation: "haber (subjuntivo presente) + participio.",
				ExampleVerb: "hablar", ExampleForms: [6]string{"que haya hablado", "que hayas hablado", "que haya hablado", "que hayamos hablado", "que hayáis hablado", "que hayan hablado"}},
		},
		IrregularNotes: "Example usage: « Me alegro de que hayas llegado bien. » Same participle irregulars as the pretérito perfecto compuesto.",
	},
	PresentProgressive: {
		ID: PresentProgressive, NameNative: "Presente continuo", NameEn: "Present progressive",
		Usage:     "Emphasizes that an action is happening right now ('is doing') — more vivid/emphatic than the simple presente, and much more commonly used than French's présent continu.",
		Formation: "estar conjugated in the presente + the gerundio of the verb being practiced (NOT the infinitive — this is the key structural difference from French's 'en train de').",
		Groups: []GroupEndings{
			{GroupName: "Todos los verbos", Explanation: "estar (presente) + gerundio.",
				ExampleVerb: "hablar", ExampleForms: [6]string{"estoy hablando", "estás hablando", "está hablando", "estamos hablando", "estáis hablando", "están hablando"}},
		},
		IrregularNotes: "-ir stem-changers keep their change in the gerundio: pedir→pidiendo (not 'pediendo'). A few verbs have a special gerundio: poder→pudiendo, ir→yendo.",
	},
	RecentPast: {
		ID: RecentPast, NameNative: "Pretérito reciente", NameEn: "Recent past",
		Usage:     "An action that just happened a moment ago ('have just done').",
		Formation: "acabar conjugated in the presente + 'de' + the infinitive of the verb being practiced. Unlike French's passé récent, 'de' never contracts — Spanish has no elision.",
		Groups: []GroupEndings{
			{GroupName: "Todos los verbos", Explanation: "acabar (presente) + de + infinitivo.",
				ExampleVerb: "hablar", ExampleForms: [6]string{"acabo de hablar", "acabas de hablar", "acaba de hablar", "acabamos de hablar", "acabáis de hablar", "acaban de hablar"}},
		},
		IrregularNotes: "Only acabar's own conjugation varies (it's fully regular); the infinitive of the practiced verb never changes.",
	},
	NearFuture: {
		ID: NearFuture, NameNative: "Futuro próximo (ir a + infinitivo)", NameEn: "Near future",
		Usage:     "An action about to happen soon ('going to do') — the everyday, spoken-Spanish alternative to the futuro simple, extremely common.",
		Formation: "ir conjugated in the presente + 'a' + the infinitive of the verb being practiced.",
		Groups: []GroupEndings{
			{GroupName: "Todos los verbos", Explanation: "ir (presente) + a + infinitivo.",
				ExampleVerb: "hablar", ExampleForms: [6]string{"voy a hablar", "vas a hablar", "va a hablar", "vamos a hablar", "vais a hablar", "van a hablar"}},
		},
		IrregularNotes: "The 'a' is always kept — it never contracts or disappears, even before a vowel-initial infinitive (e.g. 'voy a estudiar').",
	},
	PastInfinitive: {
		ID: PastInfinitive, NameNative: "Infinitivo compuesto", NameEn: "Past infinitive",
		Usage:     "A completed action, used after certain conjunctions — most commonly 'después de' ('after having done'). Invariable: exactly one form per verb, no subject agreement.",
		Formation: "haber (the infinitive itself) + the past participle.",
		Groups: []GroupEndings{
			{GroupName: "Todos los verbos", Explanation: "haber + participio — a single invariable form.",
				ExampleVerb: "hablar / volver", ExampleForms: [6]string{"haber hablado / haber vuelto", "", "", "", "", ""}},
		},
		IrregularNotes: "Contrast with the plain infinitive (hablar), which is the dictionary form already shown as the hint on every exercise in this app.",
	},
	Gerund: {
		ID: Gerund, NameNative: "Gerundio", NameEn: "Gerund",
		Usage:     "Describes an action happening at the same time as the main verb, or how/why something happens ('by doing', 'while doing'). Also the building block of the presente continuo. Invariable: one form per verb, no subject agreement, and — unlike French's gérondif — never prefixed with anything like 'en'.",
		Formation: "-ar verbs: stem + -ando. -er/-ir verbs: stem + -iendo.",
		Groups: []GroupEndings{
			{GroupName: string(GroupAR), Explanation: "stem + -ando.",
				ExampleVerb: "hablar", ExampleForms: [6]string{"hablando", "", "", "", "", ""}},
			{GroupName: "-er / -ir", Explanation: "stem + -iendo.",
				ExampleVerb: "comer / vivir", ExampleForms: [6]string{"comiendo / viviendo", "", "", "", "", ""}},
		},
		IrregularNotes: "-ir stem-changers (e→i, o→u) keep the change here: pedir→pidiendo, (dormir→durmiendo, not covered). poder is specially irregular: pudiendo. ir is fully suppletive: yendo.",
	},
	PerfectGerund: {
		ID: PerfectGerund, NameNative: "Gerundio compuesto", NameEn: "Perfect gerund",
		Usage:     "A formal, mostly written form expressing a completed action before the main clause ('having done X, ...'). Invariable: one form per verb, no subject agreement.",
		Formation: "habiendo (the gerundio of haber, always the same) + the past participle.",
		Groups: []GroupEndings{
			{GroupName: "Todos los verbos", Explanation: "habiendo + participio — a single invariable form.",
				ExampleVerb: "hablar / volver", ExampleForms: [6]string{"habiendo hablado / habiendo vuelto", "", "", "", "", ""}},
		},
		IrregularNotes: "Mostly seen in formal writing, e.g. « Habiendo terminado su trabajo, se fue a casa. » Rare in everyday speech.",
	},
}
