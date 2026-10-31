package de

import "github.com/bogdanbojan/sermokey/backend/internal/grammar"

// TenseTheory is a re-export of the shared theory shape.
type TenseTheory = grammar.TenseTheory
type GroupEndings = grammar.GroupEndings

// Theories holds the theory content for every tense, keyed by TenseID.
var Theories = map[TenseID]TenseTheory{
	Prasens: {
		ID: Prasens, NameNative: "Präsens", NameEn: "Present",
		Usage:     "Actions happening right now, habitual actions, general truths — and, very commonly in German, future events paired with a time word ('Ich fahre morgen nach Berlin').",
		Formation: "Drop the infinitive ending -en and add the Präsens endings. Many strong verbs additionally change their stem vowel (e→i, e→ie, or a→ä) in the du and er/sie/es forms only.",
		Groups: []GroupEndings{
			{GroupName: string(GroupWeak), Explanation: "No stem-vowel change (e.g. machen).",
				Endings:     [6]string{"-e", "-st", "-t", "-en", "-t", "-en"},
				ExampleVerb: "machen", ExampleForms: [6]string{"mache", "machst", "macht", "machen", "macht", "machen"}},
			{GroupName: string(GroupStrong), Explanation: "Many strong verbs change e→i or e→ie, only in du/er.",
				ExampleVerb: "sprechen / sehen", ExampleForms: [6]string{"spreche / sehe", "sprichst / siehst", "spricht / sieht", "sprechen / sehen", "sprecht / seht", "sprechen / sehen"}},
			{GroupName: string(GroupIrregular), Explanation: "sein, haben, and werden are fully irregular and must be memorized.",
				ExampleVerb: "sein / haben / werden", ExampleForms: [6]string{"bin / habe / werde", "bist / hast / wirst", "ist / hat / wird", "sind / haben / werden", "seid / habt / werdet", "sind / haben / werden"}},
		},
		IrregularNotes: "Modal verbs like können drop the usual -e/-t endings in ich/er entirely: ich kann, er kann (never 'kanne'/'kannt') — a missing ending is the hallmark of every modal verb's Präsens.",
	},
	Prateritum: {
		ID: Prateritum, NameNative: "Präteritum", NameEn: "Simple past (narrative past)",
		Usage:     "The written, narrative past ('spoke', 'went') — used in stories, news, and formal writing. In everyday speech the Perfekt is strongly preferred for most verbs, except sein, haben, werden, and the modal verbs, which commonly use the Präteritum even when speaking ('ich war', 'ich hatte', 'ich musste').",
		Formation: "Weak verbs: stem + -te/-test/-te/-ten/-tet/-ten, no vowel change. Strong verbs change their stem vowel (Ablaut) and take NO ending at all in ich/er. Mixed verbs do both at once: an Ablaut stem plus the weak -te endings.",
		Groups: []GroupEndings{
			{GroupName: string(GroupWeak), Explanation: "stem + -te, no vowel change.",
				Endings:     [6]string{"-te", "-test", "-te", "-ten", "-tet", "-ten"},
				ExampleVerb: "machen", ExampleForms: [6]string{"machte", "machtest", "machte", "machten", "machtet", "machten"}},
			{GroupName: string(GroupStrong), Explanation: "Stem-vowel change (Ablaut); ich/er take no ending.",
				ExampleVerb: "sprechen / gehen", ExampleForms: [6]string{"sprach / ging", "sprachst / gingst", "sprach / ging", "sprachen / gingen", "spracht / gingt", "sprachen / gingen"}},
			{GroupName: string(GroupMixed), Explanation: "Ablaut stem + weak endings, both at once.",
				ExampleVerb: "denken", ExampleForms: [6]string{"dachte", "dachtest", "dachte", "dachten", "dachtet", "dachten"}},
		},
		IrregularNotes: "There's no reliable rule for which vowel a strong verb changes to — this is why strong-verb principal parts (Infinitiv – Präteritum – Partizip II) are traditionally memorized as a fixed triple, e.g. sprechen – sprach – gesprochen.",
	},
	Perfekt: {
		ID: Perfekt, NameNative: "Perfekt", NameEn: "Present perfect (conversational past)",
		Usage:     "The everyday SPOKEN past tense for most verbs — this is the default way to talk about the past out loud in German, even where English would reach for a simple past ('Ich habe gestern gearbeitet' = both 'I worked' and 'I have worked').",
		Formation: "haben or sein, conjugated in the Präsens, + the Partizip II — pushed all the way to the end of the clause. sein is used for intransitive verbs of motion or change of state (gehen, kommen, fahren, bleiben, werden, sein itself); haben for everything else.",
		Groups: []GroupEndings{
			{GroupName: "haben-Verben (schwach)", Explanation: "Weak Partizip II: ge- + stem + -t.",
				ExampleVerb: "machen", ExampleForms: [6]string{"habe gemacht", "hast gemacht", "hat gemacht", "haben gemacht", "habt gemacht", "haben gemacht"}},
			{GroupName: "haben-Verben (stark)", Explanation: "Strong Partizip II: ge- + stem + -en, often with a vowel change.",
				ExampleVerb: "sprechen", ExampleForms: [6]string{"habe gesprochen", "hast gesprochen", "hat gesprochen", "haben gesprochen", "habt gesprochen", "haben gesprochen"}},
			{GroupName: "sein-Verben", Explanation: "Verbs of motion/change of state take sein instead of haben.",
				ExampleVerb: "gehen", ExampleForms: [6]string{"bin gegangen", "bist gegangen", "ist gegangen", "sind gegangen", "seid gegangen", "sind gegangen"}},
		},
		IrregularNotes: "Many strong participles are irregular and must be memorized (nehmen→genommen, sehen→gesehen, essen→gegessen). Verbs with an inseparable prefix (be-, ver-, ge-, ent-, er-, ...) or ending in -ieren never take a ge- prefix at all — none of those appear in this app's verb list, since every verb here is a plain, unprefixed one.",
	},
	Plusquamperfekt: {
		ID: Plusquamperfekt, NameNative: "Plusquamperfekt", NameEn: "Past perfect",
		Usage:     "An action that had already happened before another past action ('had done') — typically paired with a Präteritum clause: 'Nachdem ich gegessen hatte, ging ich ins Bett.'",
		Formation: "Same haben/sein choice and Partizip II as the Perfekt, but the auxiliary is conjugated in the Präteritum instead of the Präsens.",
		Groups: []GroupEndings{
			{GroupName: "haben-Verben", Explanation: "haben (Präteritum) + Partizip II.",
				ExampleVerb: "machen", ExampleForms: [6]string{"hatte gemacht", "hattest gemacht", "hatte gemacht", "hatten gemacht", "hattet gemacht", "hatten gemacht"}},
			{GroupName: "sein-Verben", Explanation: "sein (Präteritum) + Partizip II, for sein-Verben.",
				ExampleVerb: "gehen", ExampleForms: [6]string{"war gegangen", "warst gegangen", "war gegangen", "waren gegangen", "wart gegangen", "waren gegangen"}},
		},
		IrregularNotes: "Same haben/sein choice and participle irregulars as the Perfekt apply here.",
	},
	FuturI: {
		ID: FuturI, NameNative: "Futur I", NameEn: "Future",
		Usage:     "Actions that will happen in the future ('will do') — though German very often just uses the plain Präsens with a future time word instead ('ich fahre morgen'). Futur I is mainly reserved for predictions, promises, and confident assumptions about the present ('Er wird wohl zu Hause sein' = 'He's probably at home').",
		Formation: "werden, conjugated in the Präsens, + the plain infinitive of the verb being practiced — pushed to the end of the clause.",
		Groups: []GroupEndings{
			{GroupName: "Alle Verben", Explanation: "werden (Präsens) + Infinitiv.",
				ExampleVerb: "machen", ExampleForms: [6]string{"werde machen", "wirst machen", "wird machen", "werden machen", "werdet machen", "werden machen"}},
		},
		IrregularNotes: "The infinitive never changes; only werden agrees with the subject.",
	},
	FuturII: {
		ID: FuturII, NameNative: "Futur II", NameEn: "Future perfect",
		Usage:     "An action that will already be completed by some future point ('will have done') — or, just as commonly, a confident guess about something in the PAST ('Er wird wohl schon angekommen sein' = 'He's probably already arrived').",
		Formation: "werden (Präsens) + Partizip II + the verb's own auxiliary infinitive (haben or sein) — both pushed to the end of the clause, participle before the auxiliary infinitive.",
		Groups: []GroupEndings{
			{GroupName: "haben-Verben", Explanation: "werden (Präsens) + Partizip II + haben.",
				ExampleVerb: "machen", ExampleForms: [6]string{"werde gemacht haben", "wirst gemacht haben", "wird gemacht haben", "werden gemacht haben", "werdet gemacht haben", "werden gemacht haben"}},
			{GroupName: "sein-Verben", Explanation: "werden (Präsens) + Partizip II + sein, for sein-Verben.",
				ExampleVerb: "gehen", ExampleForms: [6]string{"werde gegangen sein", "wirst gegangen sein", "wird gegangen sein", "werden gegangen sein", "werdet gegangen sein", "werden gegangen sein"}},
		},
		IrregularNotes: "Rare in everyday speech; mostly used for a confident present-tense guess about the past, not literally about the future.",
	},
	KonjunktivIIPrasens: {
		ID: KonjunktivIIPrasens, NameNative: "Konjunktiv II Präsens", NameEn: "Present subjunctive II (hypothetical)",
		Usage:     "Hypothetical or unreal actions ('would do'), polite requests, and wishes — e.g. 'Könntest du mir helfen?', 'Ich hätte gern einen Kaffee.'",
		Formation: "sein, haben, werden, and the modal verbs use a short, synthetic form (wäre, hätte, würde, könnte, ...). Every other verb instead uses werden's own würde + the plain infinitive: the old synthetic subjunctive of an ordinary verb (spräche, läse, gäbe) still exists and is grammatical, but sounds literary/old-fashioned in everyday German — this app always generates 'würde + Infinitiv' for those, matching how it's actually taught and used.",
		Groups: []GroupEndings{
			{GroupName: "sein / haben / werden / Modalverben", Explanation: "Short, synthetic forms — must be memorized.",
				ExampleVerb: "sein / haben / werden / können", ExampleForms: [6]string{"wäre / hätte / würde / könnte", "wärst / hättest / würdest / könntest", "wäre / hätte / würde / könnte", "wären / hätten / würden / könnten", "wärt / hättet / würdet / könntet", "wären / hätten / würden / könnten"}},
			{GroupName: "Alle anderen Verben", Explanation: "würde (conjugated) + Infinitiv — the modern, everyday form.",
				ExampleVerb: "machen", ExampleForms: [6]string{"würde machen", "würdest machen", "würde machen", "würden machen", "würdet machen", "würden machen"}},
		},
		IrregularNotes: "The synthetic forms of ordinary verbs (spräche, gäbe, läse) are still grammatically correct and appear in formal writing, but a learner who always reaches for 'würde + Infinitiv' instead will sound perfectly natural.",
	},
	KonjunktivIPrasens: {
		ID: KonjunktivIPrasens, NameNative: "Konjunktiv I Präsens", NameEn: "Present subjunctive I (reported speech)",
		Usage:     "Reports what someone else said, without the speaker vouching for it ('indirekte Rede') — a distinction French and Spanish don't grammaticalize this way. Very common in journalism: 'Der Minister sagte, er werde zurücktreten.'",
		Formation: "Infinitive stem (infinitive minus the final -n) + -e/-est/-e/-en/-et/-en — completely regular, and critically, it never takes the Präsens' stem-vowel change (so 'er spreche', never 'er spriche').",
		Groups: []GroupEndings{
			{GroupName: "Alle Verben", Explanation: "Infinitivstamm + -e/-est/-e/-en/-et/-en.",
				Endings:     [6]string{"-e", "-est", "-e", "-en", "-et", "-en"},
				ExampleVerb: "sprechen", ExampleForms: [6]string{"spreche", "sprechest", "spreche", "sprechen", "sprechet", "sprechen"}},
		},
		IrregularNotes: "Whenever a verb's Konjunktiv I happens to look identical to its ordinary Präsens indicative (true for ich/wir/sie with most verbs), writers fall back to Konjunktiv II instead, to keep the 'this is reported, not asserted' signal audible. sein is the one true exception to the regular formation: sei, seiest, sei, seien, seiet, seien.",
	},
	KonjunktivIIPerfekt: {
		ID: KonjunktivIIPerfekt, NameNative: "Konjunktiv II Perfekt", NameEn: "Past subjunctive II (unreal past)",
		Usage:     "A hypothetical action that didn't happen ('would have done') — the standard way to express regret or an unreal past condition: 'Wenn ich das gewusst hätte, wäre ich nicht gekommen.'",
		Formation: "hätte or wäre (haben/sein's own Konjunktiv II Präsens) + Partizip II — pushed to the end of the clause. Same haben/sein choice as the Perfekt.",
		Groups: []GroupEndings{
			{GroupName: "haben-Verben", Explanation: "hätte + Partizip II.",
				ExampleVerb: "machen", ExampleForms: [6]string{"hätte gemacht", "hättest gemacht", "hätte gemacht", "hätten gemacht", "hättet gemacht", "hätten gemacht"}},
			{GroupName: "sein-Verben", Explanation: "wäre + Partizip II, for sein-Verben.",
				ExampleVerb: "gehen", ExampleForms: [6]string{"wäre gegangen", "wärst gegangen", "wäre gegangen", "wären gegangen", "wärt gegangen", "wären gegangen"}},
		},
		IrregularNotes: "Unlike the Konjunktiv II Präsens, there's no 'würde haben'/'würde sein' periphrastic alternative here — hätte/wäre are themselves already short and universally used, so this tense has no archaic-sounding-synthetic-form problem to avoid.",
	},
	KonjunktivIPerfekt: {
		ID: KonjunktivIPerfekt, NameNative: "Konjunktiv I Perfekt", NameEn: "Past subjunctive I (reported past speech)",
		Usage:     "Reports what someone else said about a completed action ('that ... had done') — the past-tense counterpart of Konjunktiv I Präsens: 'Sie sagte, sie habe den Film schon gesehen.'",
		Formation: "habe or sei (haben/sein's own Konjunktiv I Präsens) + Partizip II — pushed to the end of the clause. Same haben/sein choice as the Perfekt.",
		Groups: []GroupEndings{
			{GroupName: "haben-Verben", Explanation: "habe + Partizip II.",
				ExampleVerb: "machen", ExampleForms: [6]string{"habe gemacht", "habest gemacht", "habe gemacht", "haben gemacht", "habet gemacht", "haben gemacht"}},
			{GroupName: "sein-Verben", Explanation: "sei + Partizip II, for sein-Verben.",
				ExampleVerb: "gehen", ExampleForms: [6]string{"sei gegangen", "seiest gegangen", "sei gegangen", "seien gegangen", "seiet gegangen", "seien gegangen"}},
		},
		IrregularNotes: "Exactly as common in journalism as Konjunktiv I Präsens, and follows the same 'switch to Konjunktiv II when the form would be ambiguous' convention.",
	},
	Imperativ: {
		ID: Imperativ, NameNative: "Imperativ", NameEn: "Imperative",
		Usage:     "Commands, instructions, and suggestions. Four forms exist: du, wir ('let's ...'), ihr, and the formal Sie — unlike du/wir/ihr, the Sie-form keeps its pronoun, placed right after the verb: 'Sprechen Sie bitte langsamer!'",
		Formation: "du: the Präsens stem with no ending at all (strong e→i/e→ie verbs keep their stem change here: sprich, sieh — but -st is never added). wir/Sie: identical to the plain infinitive, with the pronoun stated after the verb. ihr: identical to the ihr-form of the Präsens.",
		Groups: []GroupEndings{
			{GroupName: string(GroupWeak), Explanation: "du drops the Präsens ending entirely; wir/Sie equal the infinitive.",
				ExampleVerb: "machen", ExampleForms: [6]string{"", "mach", "machen", "machen", "macht", ""}},
			{GroupName: string(GroupStrong), Explanation: "e→i/e→ie verbs keep their Präsens stem change in du, with no added -st.",
				ExampleVerb: "sprechen / sehen", ExampleForms: [6]string{"", "sprich / sieh", "sprechen / sehen", "sprechen / sehen", "sprecht / seht", ""}},
		},
		IrregularNotes: "sein's du-form borrows its stem from the Konjunktiv ('sei', not a 'bin'-based form); haben's du-form is the short 'hab'. Modal verbs like können have no natural imperative at all.",
	},
	InfinitivPerfekt: {
		ID: InfinitivPerfekt, NameNative: "Infinitiv Perfekt", NameEn: "Perfect infinitive",
		Usage:     "A completed action with no subject agreement at all — most often after 'ohne ... zu' ('without having done'), 'um ... zu', or verbs like 'scheinen' ('Sie scheint das Buch gelesen zu haben' = 'She seems to have read the book').",
		Formation: "Partizip II + zu + the verb's own auxiliary infinitive (haben or sein) — note the word order: unlike French's 'avoir parlé', the participle comes FIRST and the auxiliary infinitive LAST.",
		Groups: []GroupEndings{
			{GroupName: "Alle Verben", Explanation: "Partizip II + zu + haben/sein — a single invariable form.",
				ExampleVerb: "machen / gehen", ExampleForms: [6]string{"gemacht zu haben / gegangen zu sein", "", "", "", "", ""}},
		},
		IrregularNotes: "Contrast with the plain infinitive (machen), which is the dictionary form already shown as the hint on every exercise in this app.",
	},
	PartizipI: {
		ID: PartizipI, NameNative: "Partizip I", NameEn: "Present participle",
		Usage:     "Describes an action happening at the same time as the main clause, most often as a sentence-initial adverbial ('Lachend verließ sie den Raum' = 'Laughing, she left the room') or as an adjective ('das schlafende Kind'). Invariable here: one form per verb, no subject agreement — as an adjective it would additionally take case/gender endings, which this app doesn't cover.",
		Formation: "Infinitiv + d — almost perfectly regular, including for strong verbs.",
		Groups: []GroupEndings{
			{GroupName: "Alle Verben", Explanation: "Infinitiv + d — a single invariable form.",
				ExampleVerb: "machen / sprechen", ExampleForms: [6]string{"machend / sprechend", "", "", "", "", ""}},
		},
		IrregularNotes: "sein is the one exception: 'seiend', not the expected 'seind'.",
	},
}
