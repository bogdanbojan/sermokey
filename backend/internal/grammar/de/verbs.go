package de

import "strings"

// Verbs is the reference dataset of 21 verbs spanning every German verb
// class (schwach, stark, gemischt, the three irregular auxiliaries, and one
// modal), each with its hand-authored simple-tense conjugation table. To
// add a verb, copy an existing entry of the same class and fill in every
// field — see the package doc comment for why irregular stems are hand-
// typed rather than generated.
//
// Three tables below ARE generated, though, because they are genuinely
// 100%-regular across this entire dataset (including every irregular verb):
// Konjunktiv I Präsens always follows (infinitive stem)+(-e/-est/-e/-en/-et/-en)
// — see konjunktivI — and modern German overwhelmingly prefers "würde +
// Infinitiv" over a synthetic Konjunktiv II Präsens for any verb besides
// sein/haben/werden/the modals (this is what every mainstream course —
// Netzwerk, Menschen, DaF-Kompakt, ... — actually teaches, and matches the
// instruction to favor common, everyday forms over technically-valid but
// archaic-sounding synthetic subjunctives like "hülfe" or "läse") — see
// wuerdeKonjunktivII. sein is irregular enough to need an explicit override
// for the first (Konjunktiv I).
//
// PastParticiple and Auxiliary are enough to derive every compound and
// nicht-finite tense (Perfekt, Plusquamperfekt, Futur I, Futur II,
// Konjunktiv II/I Perfekt, Infinitiv Perfekt, Partizip I) — see compound.go.
var Verbs = []Verb{
	// --- unregelmäßig (sein, haben, werden) ---
	{
		Infinitive: "sein", Group: GroupIrregular, TranslationEn: "to be",
		Auxiliary: "sein", PastParticiple: "gewesen", HasImperative: true,
		Contexts: []string{
			"müde nach der Arbeit", "froh über die gute Nachricht", "pünktlich zum Termin",
			"stolz auf ihre Tochter", "krank seit zwei Tagen", "bereit für die Prüfung",
			"neu in dieser Stadt",
		},
		Prasens:             Conjugation{"bin", "bist", "ist", "sind", "seid", "sind"},
		Prateritum:          Conjugation{"war", "warst", "war", "waren", "wart", "waren"},
		KonjunktivIIPrasens: Conjugation{"wäre", "wärst", "wäre", "wären", "wärt", "wären"},
		KonjunktivIPrasens:  Conjugation{"sei", "seiest", "sei", "seien", "seiet", "seien"},
		Imperativ:           Conjugation{Tu: "sei", Nous: "seien", Vous: "seid", Il: "seien"},
	},
	{
		Infinitive: "haben", Group: GroupIrregular, TranslationEn: "to have",
		Auxiliary: "haben", PastParticiple: "gehabt", HasImperative: true,
		Contexts: []string{
			"viel Geduld mit den Kindern", "Hunger nach dem Sport", "Angst vor Spinnen",
			"Glück bei der Prüfung", "Zeit für ein Gespräch", "Durst nach dem Spaziergang",
			"Lust auf ein Eis",
		},
		Prasens:             Conjugation{"habe", "hast", "hat", "haben", "habt", "haben"},
		Prateritum:          Conjugation{"hatte", "hattest", "hatte", "hatten", "hattet", "hatten"},
		KonjunktivIIPrasens: Conjugation{"hätte", "hättest", "hätte", "hätten", "hättet", "hätten"},
		KonjunktivIPrasens:  konjunktivI("haben"),
		Imperativ:           Conjugation{Tu: "hab", Nous: "haben", Vous: "habt", Il: "haben"},
	},
	{
		Infinitive: "werden", Group: GroupIrregular, TranslationEn: "to become",
		Auxiliary: "sein", PastParticiple: "geworden", HasImperative: true,
		Contexts: []string{
			"langsam müde am Abend", "jeden Tag älter", "schnell ungeduldig",
			"nervös vor der Prüfung", "rot vor Scham", "immer besser im Deutschen",
			"endlich erwachsen",
		},
		Prasens:             Conjugation{"werde", "wirst", "wird", "werden", "werdet", "werden"},
		Prateritum:          Conjugation{"wurde", "wurdest", "wurde", "wurden", "wurdet", "wurden"},
		KonjunktivIIPrasens: Conjugation{"würde", "würdest", "würde", "würden", "würdet", "würden"},
		KonjunktivIPrasens:  konjunktivI("werden"),
		Imperativ:           Conjugation{Tu: "werde", Nous: "werden", Vous: "werdet", Il: "werden"},
	},

	// --- Modalverb ---
	{
		Infinitive: "können", Group: GroupModal, TranslationEn: "can / to be able to",
		Auxiliary: "haben", PastParticiple: "gekonnt", HasImperative: false, // modals have no natural imperative
		Contexts: []string{
			"gut Deutsch sprechen", "morgen nicht kommen", "das Problem leicht lösen",
			"schon sehr gut schwimmen", "heute Abend nicht arbeiten", "die Frage nicht beantworten",
			"endlich wieder schlafen",
		},
		Prasens:             Conjugation{"kann", "kannst", "kann", "können", "könnt", "können"},
		Prateritum:          Conjugation{"konnte", "konntest", "konnte", "konnten", "konntet", "konnten"},
		KonjunktivIIPrasens: Conjugation{"könnte", "könntest", "könnte", "könnten", "könntet", "könnten"},
		KonjunktivIPrasens:  konjunktivI("können"),
	},

	// --- schwach (-en, regelmäßig) ---
	{
		Infinitive: "machen", Group: GroupWeak, TranslationEn: "to do / to make",
		Auxiliary: "haben", PastParticiple: "gemacht", HasImperative: true,
		Contexts: []string{
			"ihre Hausaufgaben jeden Abend", "eine Pause am Nachmittag", "das Bett jeden Morgen",
			"einen Spaziergang im Park", "das Essen für die Familie", "Fotos von der Reise",
			"Sport am Wochenende",
		},
		Prasens:             regularWeakPresent("machen"),
		Prateritum:          regularWeakPreterite("machen"),
		KonjunktivIIPrasens: wuerdeKonjunktivII("machen"),
		KonjunktivIPrasens:  konjunktivI("machen"),
		Imperativ:           Conjugation{Tu: "mach", Nous: "machen", Vous: "macht", Il: "machen"},
	},
	{
		Infinitive: "spielen", Group: GroupWeak, TranslationEn: "to play",
		Auxiliary: "haben", PastParticiple: "gespielt", HasImperative: true,
		Contexts: []string{
			"Fußball im Park", "Klavier jeden Abend", "mit ihrem kleinen Bruder",
			"Karten mit Freunden", "ein neues Videospiel", "Gitarre seit drei Jahren",
			"im Garten bei schönem Wetter",
		},
		Prasens:             regularWeakPresent("spielen"),
		Prateritum:          regularWeakPreterite("spielen"),
		KonjunktivIIPrasens: wuerdeKonjunktivII("spielen"),
		KonjunktivIPrasens:  konjunktivI("spielen"),
		Imperativ:           Conjugation{Tu: "spiel", Nous: "spielen", Vous: "spielt", Il: "spielen"},
	},
	{
		Infinitive: "wohnen", Group: GroupWeak, TranslationEn: "to live (reside)",
		Auxiliary: "haben", PastParticiple: "gewohnt", HasImperative: true,
		Contexts: []string{
			"seit zwei Jahren in Berlin", "in einer kleinen Wohnung", "bei ihren Eltern",
			"im dritten Stock", "ganz in der Nähe vom Bahnhof", "auf dem Land",
			"weit weg von der Stadt",
		},
		Prasens:             regularWeakPresent("wohnen"),
		Prateritum:          regularWeakPreterite("wohnen"),
		KonjunktivIIPrasens: wuerdeKonjunktivII("wohnen"),
		KonjunktivIPrasens:  konjunktivI("wohnen"),
		Imperativ:           Conjugation{Tu: "wohn", Nous: "wohnen", Vous: "wohnt", Il: "wohnen"},
	},
	{
		Infinitive: "suchen", Group: GroupWeak, TranslationEn: "to look for",
		Auxiliary: "haben", PastParticiple: "gesucht", HasImperative: true,
		Contexts: []string{
			"seit einer Stunde ihre Schlüssel", "eine neue Wohnung", "Arbeit in München",
			"eine Lösung für das Problem", "den Weg zum Bahnhof", "ein Geschenk für ihre Mutter",
			"nach dem passenden Wort",
		},
		Prasens:             regularWeakPresent("suchen"),
		Prateritum:          regularWeakPreterite("suchen"),
		KonjunktivIIPrasens: wuerdeKonjunktivII("suchen"),
		KonjunktivIPrasens:  konjunktivI("suchen"),
		Imperativ:           Conjugation{Tu: "such", Nous: "suchen", Vous: "sucht", Il: "suchen"},
	},
	{
		Infinitive: "kaufen", Group: GroupWeak, TranslationEn: "to buy",
		Auxiliary: "haben", PastParticiple: "gekauft", HasImperative: true,
		Contexts: []string{
			"frisches Brot beim Bäcker", "ein neues Fahrrad", "Blumen für seine Frau",
			"Gemüse auf dem Markt", "ein Geschenk zum Geburtstag", "ein neues Auto",
			"Milch und Eier im Supermarkt",
		},
		Prasens:             regularWeakPresent("kaufen"),
		Prateritum:          regularWeakPreterite("kaufen"),
		KonjunktivIIPrasens: wuerdeKonjunktivII("kaufen"),
		KonjunktivIPrasens:  konjunktivI("kaufen"),
		Imperativ:           Conjugation{Tu: "kauf", Nous: "kaufen", Vous: "kauft", Il: "kaufen"},
	},
	{
		Infinitive: "lernen", Group: GroupWeak, TranslationEn: "to learn",
		Auxiliary: "haben", PastParticiple: "gelernt", HasImperative: true,
		Contexts: []string{
			"jeden Tag Deutsch", "für die Prüfung am Freitag", "ein neues Instrument",
			"schnell neue Wörter", "mit großem Interesse", "gern etwas Neues",
			"fleißig für die Schule",
		},
		Prasens:             regularWeakPresent("lernen"),
		Prateritum:          regularWeakPreterite("lernen"),
		KonjunktivIIPrasens: wuerdeKonjunktivII("lernen"),
		KonjunktivIPrasens:  konjunktivI("lernen"),
		Imperativ:           Conjugation{Tu: "lern", Nous: "lernen", Vous: "lernt", Il: "lernen"},
	},

	// --- stark, haben (Ablaut, no change of state/location) ---
	{
		Infinitive: "sprechen", Group: GroupStrong, TranslationEn: "to speak",
		Auxiliary: "haben", PastParticiple: "gesprochen", HasImperative: true,
		Contexts: []string{
			"mit ihrer Schwester am Telefon", "sehr gut Englisch", "leise, um das Baby nicht zu wecken",
			"offen über ihre Gefühle", "mit dem Chef über das Projekt", "zu schnell, um verstanden zu werden",
			"mit Begeisterung über die Reise",
		},
		Prasens:             Conjugation{"spreche", "sprichst", "spricht", "sprechen", "sprecht", "sprechen"},
		Prateritum:          Conjugation{"sprach", "sprachst", "sprach", "sprachen", "spracht", "sprachen"},
		KonjunktivIIPrasens: wuerdeKonjunktivII("sprechen"),
		KonjunktivIPrasens:  konjunktivI("sprechen"),
		Imperativ:           Conjugation{Tu: "sprich", Nous: "sprechen", Vous: "sprecht", Il: "sprechen"},
	},
	{
		Infinitive: "essen", Group: GroupStrong, TranslationEn: "to eat",
		Auxiliary: "haben", PastParticiple: "gegessen", HasImperative: true,
		Contexts: []string{
			"jeden Morgen einen Apfel", "heute Abend im Restaurant", "zu viel Schokolade an Weihnachten",
			"immer Gemüse zum Mittagessen", "mit den Fingern beim Picknick", "ein Brötchen zum Frühstück",
			"gern italienisch",
		},
		Prasens:             Conjugation{"esse", "isst", "isst", "essen", "esst", "essen"},
		Prateritum:          Conjugation{"aß", "aßt", "aß", "aßen", "aßt", "aßen"},
		KonjunktivIIPrasens: wuerdeKonjunktivII("essen"),
		KonjunktivIPrasens:  konjunktivI("essen"),
		Imperativ:           Conjugation{Tu: "iss", Nous: "essen", Vous: "esst", Il: "essen"},
	},
	{
		Infinitive: "lesen", Group: GroupStrong, TranslationEn: "to read",
		Auxiliary: "haben", PastParticiple: "gelesen", HasImperative: true,
		Contexts: []string{
			"jeden Abend ein Buch", "die Zeitung beim Frühstück", "ein spannendes Buch im Urlaub",
			"laut vor der Klasse", "die Nachrichten auf dem Handy", "ein Gedicht von Goethe",
			"gern Krimis",
		},
		Prasens:             Conjugation{"lese", "liest", "liest", "lesen", "lest", "lesen"},
		Prateritum:          Conjugation{"las", "last", "las", "lasen", "last", "lasen"},
		KonjunktivIIPrasens: wuerdeKonjunktivII("lesen"),
		KonjunktivIPrasens:  konjunktivI("lesen"),
		Imperativ:           Conjugation{Tu: "lies", Nous: "lesen", Vous: "lest", Il: "lesen"},
	},
	{
		Infinitive: "nehmen", Group: GroupStrong, TranslationEn: "to take",
		Auxiliary: "haben", PastParticiple: "genommen", HasImperative: true,
		Contexts: []string{
			"den Zug um acht Uhr", "eine wichtige Entscheidung", "einen Kaffee vor der Arbeit",
			"Fotos während der Reise", "sich Zeit für die Antwort", "das Flugzeug nach Rom",
			"die Treppe statt des Aufzugs",
		},
		Prasens:             Conjugation{"nehme", "nimmst", "nimmt", "nehmen", "nehmt", "nehmen"},
		Prateritum:          Conjugation{"nahm", "nahmst", "nahm", "nahmen", "nahmt", "nahmen"},
		KonjunktivIIPrasens: wuerdeKonjunktivII("nehmen"),
		KonjunktivIPrasens:  konjunktivI("nehmen"),
		Imperativ:           Conjugation{Tu: "nimm", Nous: "nehmen", Vous: "nehmt", Il: "nehmen"},
	},
	{
		Infinitive: "sehen", Group: GroupStrong, TranslationEn: "to see",
		Auxiliary: "haben", PastParticiple: "gesehen", HasImperative: true,
		Contexts: []string{
			"einen guten Film im Kino", "ihre Freunde am Wochenende", "die Berge in der Ferne",
			"fern bis spät in die Nacht", "ihre Großeltern einmal im Monat", "einen Regenbogen nach dem Regen",
			"die Ergebnisse der Prüfung",
		},
		Prasens:             Conjugation{"sehe", "siehst", "sieht", "sehen", "seht", "sehen"},
		Prateritum:          Conjugation{"sah", "sahst", "sah", "sahen", "saht", "sahen"},
		KonjunktivIIPrasens: wuerdeKonjunktivII("sehen"),
		KonjunktivIPrasens:  konjunktivI("sehen"),
		Imperativ:           Conjugation{Tu: "sieh", Nous: "sehen", Vous: "seht", Il: "sehen"},
	},
	{
		Infinitive: "helfen", Group: GroupStrong, TranslationEn: "to help",
		Auxiliary: "haben", PastParticiple: "geholfen", HasImperative: true,
		Contexts: []string{
			"ihrer Mutter in der Küche", "seinem Freund beim Umzug", "den Kindern bei den Hausaufgaben",
			"gern in schwierigen Zeiten", "ihrem Nachbarn im Garten", "bei der Vorbereitung der Party",
			"immer, wenn es nötig ist",
		},
		Prasens:             Conjugation{"helfe", "hilfst", "hilft", "helfen", "helft", "helfen"},
		Prateritum:          Conjugation{"half", "halfst", "half", "halfen", "halft", "halfen"},
		KonjunktivIIPrasens: wuerdeKonjunktivII("helfen"),
		KonjunktivIPrasens:  konjunktivI("helfen"),
		Imperativ:           Conjugation{Tu: "hilf", Nous: "helfen", Vous: "helft", Il: "helfen"},
	},

	// --- stark, sein (Bewegung/Zustandswechsel: movement or change of state) ---
	{
		Infinitive: "gehen", Group: GroupStrong, TranslationEn: "to go",
		Auxiliary: "sein", PastParticiple: "gegangen", HasImperative: true,
		Contexts: []string{
			"jeden Morgen zur Schule", "am Samstag auf den Markt", "morgen zum Arzt",
			"im Sommer gern schwimmen", "heute Abend ins Kino", "jeden Sonntag spazieren",
			"zu Fuß zur Arbeit",
		},
		Prasens:             Conjugation{"gehe", "gehst", "geht", "gehen", "geht", "gehen"},
		Prateritum:          Conjugation{"ging", "gingst", "ging", "gingen", "gingt", "gingen"},
		KonjunktivIIPrasens: wuerdeKonjunktivII("gehen"),
		KonjunktivIPrasens:  konjunktivI("gehen"),
		Imperativ:           Conjugation{Tu: "geh", Nous: "gehen", Vous: "geht", Il: "gehen"},
	},
	{
		Infinitive: "kommen", Group: GroupStrong, TranslationEn: "to come",
		Auxiliary: "sein", PastParticiple: "gekommen", HasImperative: true,
		Contexts: []string{
			"am Sonntag zu Besuch", "immer pünktlich zur Arbeit", "aus einer kleinen Stadt",
			"später nach Hause", "mit dem Zug aus Hamburg", "gern zur Party",
			"vor der Dunkelheit nach Hause",
		},
		Prasens:             Conjugation{"komme", "kommst", "kommt", "kommen", "kommt", "kommen"},
		Prateritum:          Conjugation{"kam", "kamst", "kam", "kamen", "kamt", "kamen"},
		KonjunktivIIPrasens: wuerdeKonjunktivII("kommen"),
		KonjunktivIPrasens:  konjunktivI("kommen"),
		Imperativ:           Conjugation{Tu: "komm", Nous: "kommen", Vous: "kommt", Il: "kommen"},
	},
	{
		Infinitive: "fahren", Group: GroupStrong, TranslationEn: "to drive / to go (by vehicle)",
		Auxiliary: "sein", PastParticiple: "gefahren", HasImperative: true,
		Contexts: []string{
			"jeden Tag mit dem Fahrrad zur Arbeit", "im Sommer ans Meer", "mit dem Zug nach München",
			"zu schnell auf der Autobahn", "am Wochenende aufs Land", "mit dem Bus in die Stadt",
			"vorsichtig bei Regen",
		},
		Prasens:             Conjugation{"fahre", "fährst", "fährt", "fahren", "fahrt", "fahren"},
		Prateritum:          Conjugation{"fuhr", "fuhrst", "fuhr", "fuhren", "fuhrt", "fuhren"},
		KonjunktivIIPrasens: wuerdeKonjunktivII("fahren"),
		KonjunktivIPrasens:  konjunktivI("fahren"),
		Imperativ:           Conjugation{Tu: "fahr", Nous: "fahren", Vous: "fahrt", Il: "fahren"},
	},
	{
		Infinitive: "bleiben", Group: GroupStrong, TranslationEn: "to stay",
		Auxiliary: "sein", PastParticiple: "geblieben", HasImperative: true,
		Contexts: []string{
			"heute Abend lieber zu Hause", "ruhig trotz des Stresses", "noch eine Woche in Berlin",
			"bei schlechtem Wetter drinnen", "lange wach wegen der Prüfung", "gesund trotz der Kälte",
			"optimistisch trotz der Schwierigkeiten",
		},
		Prasens:             Conjugation{"bleibe", "bleibst", "bleibt", "bleiben", "bleibt", "bleiben"},
		Prateritum:          Conjugation{"blieb", "bliebst", "blieb", "blieben", "bliebt", "blieben"},
		KonjunktivIIPrasens: wuerdeKonjunktivII("bleiben"),
		KonjunktivIPrasens:  konjunktivI("bleiben"),
		Imperativ:           Conjugation{Tu: "bleib", Nous: "bleiben", Vous: "bleibt", Il: "bleiben"},
	},

	// --- gemischt (Ablaut im Präteritum + schwache Endungen) ---
	{
		Infinitive: "denken", Group: GroupMixed, TranslationEn: "to think",
		Auxiliary: "haben", PastParticiple: "gedacht", HasImperative: true,
		Contexts: []string{
			"oft an ihre Familie", "positiv über die Zukunft", "an den letzten Urlaub",
			"genau wie ihr Bruder", "an die gute alte Zeit", "logisch bei jedem Problem",
			"an ihre Kindheit in Bayern",
		},
		Prasens:             regularWeakPresent("denken"),
		Prateritum:          Conjugation{"dachte", "dachtest", "dachte", "dachten", "dachtet", "dachten"},
		KonjunktivIIPrasens: wuerdeKonjunktivII("denken"),
		KonjunktivIPrasens:  konjunktivI("denken"),
		Imperativ:           Conjugation{Tu: "denk", Nous: "denken", Vous: "denkt", Il: "denken"},
	},
}

// regularWeakPresent builds the fully regular Präsens table for a weak verb
// with no spelling-change issues (no stem ending in -t/-d, -s/-ß/-z, or
// -eln/-ern): stem + -e/-st/-t/-en/-t/-en.
func regularWeakPresent(infinitive string) Conjugation {
	stem := strings.TrimSuffix(infinitive, "en")
	return Conjugation{
		Je: stem + "e", Tu: stem + "st", Il: stem + "t",
		Nous: infinitive, Vous: stem + "t", Ils: infinitive,
	}
}

// regularWeakPreterite builds the fully regular Präteritum table for a weak
// verb: stem + -te/-test/-te/-ten/-tet/-ten.
func regularWeakPreterite(infinitive string) Conjugation {
	stem := strings.TrimSuffix(infinitive, "en")
	return Conjugation{
		Je: stem + "te", Tu: stem + "test", Il: stem + "te",
		Nous: stem + "ten", Vous: stem + "tet", Ils: stem + "ten",
	}
}

// konjunktivIOverrides holds the one verb in this dataset whose Konjunktiv I
// Präsens isn't the regular "(infinitive minus final -n) + -e/-est/-e/-en/-et/-en".
var konjunktivIOverrides = map[string]Conjugation{
	"sein": {"sei", "seiest", "sei", "seien", "seiet", "seien"},
}

// konjunktivI builds the Konjunktiv I Präsens table for a verb's infinitive.
// This single formula is correct for every regular AND irregular verb in
// this dataset except sein (see konjunktivIOverrides): Konjunktiv I always
// attaches weak personal endings to the plain infinitive stem, completely
// bypassing any Präsens stem-vowel change (so sprechen → spreche/sprechest/
// spreche/..., never spriche/sprichest/...) — which is exactly why it stays
// regular even for verbs that are wildly irregular in the Präsens.
func konjunktivI(infinitive string) Conjugation {
	if override, ok := konjunktivIOverrides[infinitive]; ok {
		return override
	}
	stem := strings.TrimSuffix(infinitive, "n") // "sprechen" -> "spreche"
	return Conjugation{
		Je: stem, Tu: stem + "st", Il: stem,
		Nous: infinitive, Vous: stem + "t", Ils: infinitive,
	}
}

// wuerdeKonjunktivII builds the "würde + Infinitiv" periphrastic Konjunktiv
// II Präsens used by every verb in this dataset except sein/haben/werden/
// können — see the package-level var doc comment for why.
func wuerdeKonjunktivII(infinitive string) Conjugation {
	return Conjugation{
		Je: "würde " + infinitive, Tu: "würdest " + infinitive, Il: "würde " + infinitive,
		Nous: "würden " + infinitive, Vous: "würdet " + infinitive, Ils: "würden " + infinitive,
	}
}
