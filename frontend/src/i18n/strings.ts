// UI chrome strings (buttons, headings, feedback messages) for each
// language. This app follows an "immersive" convention already established
// for French: interface text is in the language being studied, not fixed to
// English — so switching to Spanish content also switches "Vérifier" to
// "Comprobar", not just the exercises. Content strings with a value to
// interpolate use {placeholders}; see format() in language.ts.
import type { LanguageCode } from '../types/api'

export interface UiStrings {
  brand: { title: string; tagline: string }
  home: {
    title: string
    intro: string
    loading: string
    error: string
    mixedTitle: string
    mixedUsage: string
    grammarTab: string
    vocabTab: string
    searchPlaceholder: string
    noResults: string
    continueTitle: string
    clusters: {
      basics: string
      people: string
      daily: string
      world: string
      schoolWork: string
      wordsConcepts: string
    }
  }
  tenseCard: { theory: string; practice: string; mixedPill: string }
  theory: {
    back: string
    loading: string
    error: string
    formation: string
    toRemember: string
    example: string
    practiceThisTense: string
    personLabels: string[] // je/tu/il/nous/vous/ils, in that order — table headers
  }
  quizSetup: {
    back: string
    loading: string
    error: string
    practiceTitlePrefix: string
    exerciseType: string
    mixed: string
    mcqOnly: string
    textOnly: string
    questionCount: string
    start: string
  }
  quiz: {
    back: string
    loadingExercises: string
    error: string
    noExercises: string
    questionLabel: string
    scoreLabel: string
    subjectLabel: string
    check: string
    next: string
    seeResults: string
    correct: string
    correctWithAccentWarning: string // "{answer}"
    incorrect: string // "{answer}"
    youWrote: string // "{text}"
    results: string
    toReview: string
    missedItem: string // "{infinitive} ({tense}, {person}) — {correctLabel}: {answer} · {wroteLabel}: {userAnswer}"
    correctLabel: string
    wroteLabel: string
    noMistakes: string
    backHome: string
    tryAgain: string
  }
  vocab: {
    sectionTitle: string
    sectionIntro: string
    mixedTitle: string
    mixedUsage: string
    wordsLabel: string // "{count} mots" fragment prefix
    back: string
    loading: string
    error: string
    newLabel: string
    learningLabel: string
    reviewLabel: string
    masteredLabel: string
    dueLabel: string
    startSession: string
    reveal: string
    again: string
    hard: string
    good: string
    easy: string
    sessionDone: string
    sessionDoneDetail: string // "{count}"
    nothingDue: string
    nothingDueDetail: string
    studyAnyway: string
    cardProgress: string // "{index} / {total}"
    directionLabel: string
    directionRecognition: string
    directionRecognitionHint: string
    directionProduction: string
    directionProductionHint: string
    directionMixed: string
    directionMixedHint: string
  }
  reading: {
    tab: string
    sectionIntro: string
    loading: string
    error: string
    unavailable: string
    levelLabel: string
    tokensLabel: string // "{count}"
    sectionsLabel: string // "{count}"
    newWordsLabel: string // "{count}"
    coverageLabel: string // "{percent}"
    bandComfortable: string
    bandChallenging: string
    bandHard: string
    bandHint: string
    read: string
    continueReading: string
    back: string
    prevSection: string
    nextSection: string
    sectionProgress: string // "{index} / {total}"
    savedCount: string // "{count}"
    finishTitle: string
    finishDetail: string // "{count}"
    reviewSaved: string
    noneSaved: string
    // Word panel
    noDefinition: string
    uncertainNote: string
    otherReadings: string
    chooseSense: string
    save: string
    saved: string
    unsave: string
    markKnown: string
    knownNote: string
    ignore: string
    ignoredNote: string
    undo: string
    nudge: string // "{count}"
    lookupHint: string
    attribution: string
    // Display controls
    displayLabel: string
    smaller: string
    larger: string
    serif: string
    sans: string
    highlightOn: string
    highlightOff: string
    statusUnknown: string
    statusLearning: string
    statusKnown: string
    // Selection
    selectionLookup: string
    selectionSave: string
    selectionTooLong: string // "{max}"
    // Onboarding
    coachTitle: string
    coachHover: string
    coachClick: string
    coachSelect: string
    coachDismiss: string
    // Saved-words deck
    savedDeckTitle: string
    savedDeckBlurb: string
    textDeckTitle: string // "{title}"
    contextLabel: string
  }
}


export const strings: Record<LanguageCode, UiStrings> = {
  fr: {
    brand: { title: 'Grammaire', tagline: 'conjugaison française' },
    home: {
      title: 'Grammaire française',
      intro: 'Choisissez un temps à étudier et à pratiquer, ou lancez une révision mixte de tous les temps.',
      loading: 'Chargement…',
      error: 'Une erreur est survenue.',
      mixedTitle: 'Révision mixte',
      mixedUsage: 'Un mélange de tous les temps pour un entraînement complet.',
      grammarTab: 'Grammaire',
      vocabTab: 'Vocabulaire',
      searchPlaceholder: 'Rechercher…',
      noResults: 'Aucun résultat.',
      continueTitle: 'Reprendre',
      clusters: {
        basics: 'Les bases',
        people: 'Les gens',
        daily: 'La vie quotidienne',
        world: 'Le monde',
        schoolWork: 'École, travail & médias',
        wordsConcepts: 'Mots & concepts',
      },
    },
    tenseCard: { theory: 'Théorie', practice: 'Pratiquer', mixedPill: 'Révision' },
    theory: {
      back: '← Retour',
      loading: 'Chargement…',
      error: 'Impossible de charger cette page.',
      formation: 'Formation',
      toRemember: 'À retenir',
      example: 'Exemple',
      practiceThisTense: 'Pratiquer ce temps →',
      personLabels: ['je', 'tu', 'il / elle / on', 'nous', 'vous', 'ils / elles'],
    },
    quizSetup: {
      back: '← Retour',
      loading: 'Chargement…',
      error: 'Impossible de charger ce temps.',
      practiceTitlePrefix: 'Pratiquer :',
      exerciseType: "Type d'exercice",
      mixed: 'Mélange (choix multiple + rédaction)',
      mcqOnly: 'Choix multiple uniquement',
      textOnly: 'Rédaction uniquement',
      questionCount: 'Nombre de questions',
      start: 'Commencer →',
    },
    quiz: {
      back: '← Retour',
      loadingExercises: 'Chargement des exercices…',
      error: 'Impossible de charger les exercices.',
      noExercises: "Aucun exercice n'a pu être généré.",
      questionLabel: 'Question',
      scoreLabel: 'Score :',
      subjectLabel: 'sujet :',
      check: 'Vérifier',
      next: 'Suivant →',
      seeResults: 'Voir les résultats →',
      correct: '✓ Correct !',
      correctWithAccentWarning:
        "✓ Correct ! Attention cependant aux accents : la forme correcte s'écrit «\u00a0{answer}\u00a0».",
      incorrect: '✗ Pas tout à fait — la bonne réponse était «\u00a0{answer}\u00a0».',
      youWrote: 'Vous avez écrit : «\u00a0{text}\u00a0»',
      results: 'Résultats',
      toReview: 'À revoir',
      missedItem: '{infinitive} ({tense}, {person}) — bonne réponse : {answer} · vous aviez écrit : {userAnswer}',
      correctLabel: 'bonne réponse',
      wroteLabel: 'vous aviez écrit',
      noMistakes: 'Aucune erreur — excellent travail !',
      backHome: "Retour à l'accueil",
      tryAgain: 'Recommencer ↻',
    },
    vocab: {
      sectionTitle: 'Vocabulaire',
      sectionIntro: 'Apprenez du vocabulaire essentiel avec des fiches à répétition espacée.',
      mixedTitle: 'Révision mixte',
      mixedUsage: 'Un mélange de tout le vocabulaire pour une révision complète.',
      wordsLabel: 'mots',
      back: '← Retour',
      loading: 'Chargement…',
      error: 'Impossible de charger le vocabulaire.',
      newLabel: 'Nouveaux',
      learningLabel: 'En cours',
      reviewLabel: 'À réviser',
      masteredLabel: 'Maîtrisés',
      dueLabel: "à réviser aujourd'hui",
      startSession: 'Commencer →',
      reveal: 'Révéler',
      again: 'À revoir',
      hard: 'Difficile',
      good: 'Bien',
      easy: 'Facile',
      sessionDone: 'Séance terminée !',
      sessionDoneDetail: 'Vous avez révisé {count} cartes.',
      nothingDue: "Rien à réviser aujourd'hui",
      nothingDueDetail: 'Revenez plus tard, ou étudiez quelques cartes en avance.',
      studyAnyway: 'Étudier quand même',
      cardProgress: '{index} / {total}',
      directionLabel: 'Sens de révision',
      directionRecognition: 'Reconnaissance',
      directionRecognitionHint: 'Voir le mot, deviner le sens',
      directionProduction: 'Production',
      directionProductionHint: 'Voir le sens, deviner le mot',
      directionMixed: 'Mélange',
      directionMixedHint: 'Les deux sens, mélangés',
    },
    reading: {
      tab: 'Textes',
      sectionIntro: 'Lisez, survolez un mot pour sa définition, et ajoutez-le à vos cartes.',
      loading: 'Chargement…',
      error: 'Impossible de charger le texte.',
      unavailable: 'La lecture n’est pas encore disponible dans cette langue.',
      levelLabel: 'Niveau',
      tokensLabel: '{count} mots',
      sectionsLabel: '{count} parties',
      newWordsLabel: '{count} mots nouveaux',
      coverageLabel: '{percent} % connus',
      bandComfortable: 'Confortable',
      bandChallenging: 'Exigeant',
      bandHard: 'Difficile',
      bandHint: 'Calculé à partir de vos cartes : au-dessus de 98 % de mots connus, on lit sans effort.',
      read: 'Lire',
      continueReading: 'Reprendre',
      back: '← Bibliothèque',
      prevSection: '← Précédent',
      nextSection: 'Suivant →',
      sectionProgress: 'Partie {index} / {total}',
      savedCount: '{count} enregistrés',
      finishTitle: 'Texte terminé',
      finishDetail: 'Vous avez enregistré {count} mots dans ce texte.',
      reviewSaved: 'Réviser ces mots',
      noneSaved: 'Aucun mot enregistré pour l’instant.',
      noDefinition: 'Aucune définition trouvée.',
      uncertainNote: 'Forme devinée — à vérifier.',
      otherReadings: 'Autres lectures',
      chooseSense: 'Choisissez le sens à retenir',
      save: 'Ajouter aux cartes',
      saved: 'Enregistré',
      unsave: 'Retirer',
      markKnown: 'Je le sais déjà',
      knownNote: 'Marqué comme connu',
      ignore: 'Ignorer',
      ignoredNote: 'Ignoré (nom propre, chiffre…)',
      undo: 'Annuler',
      nudge: 'Vous avez cherché ce mot {count} fois. À ajouter ?',
      lookupHint: 'Cliquez pour épingler',
      attribution: 'Définitions : Wiktionnaire (CC BY-SA)',
      displayLabel: 'Affichage',
      smaller: 'Réduire le texte',
      larger: 'Agrandir le texte',
      serif: 'Serif',
      sans: 'Sans',
      highlightOn: 'Surlignage activé',
      highlightOff: 'Surlignage désactivé',
      statusUnknown: 'nouveau',
      statusLearning: 'en cours',
      statusKnown: 'connu',
      selectionLookup: 'Chercher',
      selectionSave: 'Ajouter l’expression',
      selectionTooLong: 'Sélection trop longue ({max} mots maximum).',
      coachTitle: 'Comment lire ici',
      coachHover: 'Survolez un mot pour un aperçu rapide.',
      coachClick: 'Cliquez pour l’épingler et l’ajouter à vos cartes.',
      coachSelect: 'Glissez sur plusieurs mots pour une expression.',
      coachDismiss: 'Compris',
      savedDeckTitle: 'Mes mots',
      savedDeckBlurb: 'Les mots enregistrés pendant la lecture',
      textDeckTitle: 'Mots de « {title} »',
      contextLabel: 'Rencontré dans',
    },
  },
  es: {
    brand: { title: 'Gramática', tagline: 'conjugación española' },
    home: {
      title: 'Gramática española',
      intro: 'Elige un tiempo para estudiar y practicar, o inicia un repaso mixto de todos los tiempos.',
      loading: 'Cargando…',
      error: 'Ha ocurrido un error.',
      mixedTitle: 'Repaso mixto',
      mixedUsage: 'Una mezcla de todos los tiempos para un entrenamiento completo.',
      grammarTab: 'Gramática',
      vocabTab: 'Vocabulario',
      searchPlaceholder: 'Buscar…',
      noResults: 'Sin resultados.',
      continueTitle: 'Continuar',
      clusters: {
        basics: 'Lo esencial',
        people: 'Las personas',
        daily: 'La vida cotidiana',
        world: 'El mundo',
        schoolWork: 'Escuela, trabajo y medios',
        wordsConcepts: 'Palabras y conceptos',
      },
    },
    tenseCard: { theory: 'Teoría', practice: 'Practicar', mixedPill: 'Repaso' },
    theory: {
      back: '← Volver',
      loading: 'Cargando…',
      error: 'No se pudo cargar esta página.',
      formation: 'Formación',
      toRemember: 'Para recordar',
      example: 'Ejemplo',
      practiceThisTense: 'Practicar este tiempo →',
      personLabels: ['yo', 'tú', 'él / ella / usted', 'nosotros/as', 'vosotros/as', 'ellos / ellas / ustedes'],
    },
    quizSetup: {
      back: '← Volver',
      loading: 'Cargando…',
      error: 'No se pudo cargar este tiempo.',
      practiceTitlePrefix: 'Practicar:',
      exerciseType: 'Tipo de ejercicio',
      mixed: 'Mezcla (opción múltiple + redacción)',
      mcqOnly: 'Solo opción múltiple',
      textOnly: 'Solo redacción',
      questionCount: 'Número de preguntas',
      start: 'Empezar →',
    },
    quiz: {
      back: '← Volver',
      loadingExercises: 'Cargando ejercicios…',
      error: 'No se pudieron cargar los ejercicios.',
      noExercises: 'No se pudo generar ningún ejercicio.',
      questionLabel: 'Pregunta',
      scoreLabel: 'Puntos:',
      subjectLabel: 'sujeto:',
      check: 'Comprobar',
      next: 'Siguiente →',
      seeResults: 'Ver resultados →',
      correct: '✓ ¡Correcto!',
      correctWithAccentWarning:
        '✓ ¡Correcto! Pero cuidado con los acentos: la forma correcta se escribe «\u00a0{answer}\u00a0».',
      incorrect: '✗ No del todo — la respuesta correcta era «\u00a0{answer}\u00a0».',
      youWrote: 'Escribiste: «\u00a0{text}\u00a0»',
      results: 'Resultados',
      toReview: 'Para repasar',
      missedItem: '{infinitive} ({tense}, {person}) — respuesta correcta: {answer} · escribiste: {userAnswer}',
      correctLabel: 'respuesta correcta',
      wroteLabel: 'escribiste',
      noMistakes: '¡Sin errores — excelente trabajo!',
      backHome: 'Volver al inicio',
      tryAgain: 'Repetir ↻',
    },
    vocab: {
      sectionTitle: 'Vocabulario',
      sectionIntro: 'Aprende vocabulario esencial con tarjetas de repetición espaciada.',
      mixedTitle: 'Repaso mixto',
      mixedUsage: 'Una mezcla de todo el vocabulario para un repaso completo.',
      wordsLabel: 'palabras',
      back: '← Volver',
      loading: 'Cargando…',
      error: 'No se pudo cargar el vocabulario.',
      newLabel: 'Nuevas',
      learningLabel: 'Aprendiendo',
      reviewLabel: 'Para repasar',
      masteredLabel: 'Dominadas',
      dueLabel: 'para repasar hoy',
      startSession: 'Empezar →',
      reveal: 'Revelar',
      again: 'De nuevo',
      hard: 'Difícil',
      good: 'Bien',
      easy: 'Fácil',
      sessionDone: '¡Sesión terminada!',
      sessionDoneDetail: 'Repasaste {count} tarjetas.',
      nothingDue: 'Nada que repasar hoy',
      nothingDueDetail: 'Vuelve más tarde, o estudia algunas tarjetas por adelantado.',
      studyAnyway: 'Estudiar de todos modos',
      cardProgress: '{index} / {total}',
      directionLabel: 'Sentido de repaso',
      directionRecognition: 'Reconocimiento',
      directionRecognitionHint: 'Ver la palabra, adivinar el significado',
      directionProduction: 'Producción',
      directionProductionHint: 'Ver el significado, adivinar la palabra',
      directionMixed: 'Mezcla',
      directionMixedHint: 'Ambos sentidos, mezclados',
    },
    reading: {
      tab: 'Textos',
      sectionIntro: 'Lee, pasa el ratón sobre una palabra para ver su definición y añádela a tus tarjetas.',
      loading: 'Cargando…',
      error: 'No se ha podido cargar el texto.',
      unavailable: 'La lectura aún no está disponible en este idioma.',
      levelLabel: 'Nivel',
      tokensLabel: '{count} palabras',
      sectionsLabel: '{count} partes',
      newWordsLabel: '{count} palabras nuevas',
      coverageLabel: '{percent} % conocidas',
      bandComfortable: 'Cómodo',
      bandChallenging: 'Exigente',
      bandHard: 'Difícil',
      bandHint: 'Calculado con tus tarjetas: por encima del 98 % de palabras conocidas se lee sin esfuerzo.',
      read: 'Leer',
      continueReading: 'Continuar',
      back: '← Biblioteca',
      prevSection: '← Anterior',
      nextSection: 'Siguiente →',
      sectionProgress: 'Parte {index} / {total}',
      savedCount: '{count} guardadas',
      finishTitle: 'Texto terminado',
      finishDetail: 'Has guardado {count} palabras de este texto.',
      reviewSaved: 'Repasar estas palabras',
      noneSaved: 'Todavía no has guardado ninguna palabra.',
      noDefinition: 'No se ha encontrado definición.',
      uncertainNote: 'Forma deducida: conviene comprobarla.',
      otherReadings: 'Otras lecturas',
      chooseSense: 'Elige el sentido que quieres recordar',
      save: 'Añadir a las tarjetas',
      saved: 'Guardada',
      unsave: 'Quitar',
      markKnown: 'Ya la sé',
      knownNote: 'Marcada como conocida',
      ignore: 'Ignorar',
      ignoredNote: 'Ignorada (nombre propio, cifra…)',
      undo: 'Deshacer',
      nudge: 'Has buscado esta palabra {count} veces. ¿La añades?',
      lookupHint: 'Haz clic para fijar',
      attribution: 'Definiciones: Wikcionario (CC BY-SA)',
      displayLabel: 'Visualización',
      smaller: 'Reducir el texto',
      larger: 'Ampliar el texto',
      serif: 'Serif',
      sans: 'Sans',
      highlightOn: 'Resaltado activado',
      highlightOff: 'Resaltado desactivado',
      statusUnknown: 'nueva',
      statusLearning: 'en curso',
      statusKnown: 'conocida',
      selectionLookup: 'Buscar',
      selectionSave: 'Añadir la expresión',
      selectionTooLong: 'Selección demasiado larga ({max} palabras como máximo).',
      coachTitle: 'Cómo leer aquí',
      coachHover: 'Pasa el ratón sobre una palabra para una vista rápida.',
      coachClick: 'Haz clic para fijarla y añadirla a tus tarjetas.',
      coachSelect: 'Arrastra sobre varias palabras para una expresión.',
      coachDismiss: 'Entendido',
      savedDeckTitle: 'Mis palabras',
      savedDeckBlurb: 'Las palabras guardadas durante la lectura',
      textDeckTitle: 'Palabras de «{title}»',
      contextLabel: 'Encontrada en',
    },
  },
  de: {
    brand: { title: 'Grammatik', tagline: 'deutsche Konjugation' },
    home: {
      title: 'Deutsche Grammatik',
      intro: 'Wähle eine Zeitform zum Lernen und Üben, oder starte eine gemischte Wiederholung aller Zeitformen.',
      loading: 'Wird geladen…',
      error: 'Ein Fehler ist aufgetreten.',
      mixedTitle: 'Gemischte Wiederholung',
      mixedUsage: 'Eine Mischung aller Zeitformen für ein vollständiges Training.',
      grammarTab: 'Grammatik',
      vocabTab: 'Wortschatz',
      searchPlaceholder: 'Suchen…',
      noResults: 'Keine Ergebnisse.',
      continueTitle: 'Weitermachen',
      clusters: {
        basics: 'Grundlagen',
        people: 'Menschen',
        daily: 'Alltag',
        world: 'Die Welt',
        schoolWork: 'Schule, Arbeit & Medien',
        wordsConcepts: 'Wörter & Konzepte',
      },
    },
    tenseCard: { theory: 'Theorie', practice: 'Üben', mixedPill: 'Wiederholung' },
    theory: {
      back: '← Zurück',
      loading: 'Wird geladen…',
      error: 'Diese Seite konnte nicht geladen werden.',
      formation: 'Bildung',
      toRemember: 'Zu merken',
      example: 'Beispiel',
      practiceThisTense: 'Diese Zeitform üben →',
      personLabels: ['ich', 'du', 'er / sie / es', 'wir', 'ihr', 'sie / Sie'],
    },
    quizSetup: {
      back: '← Zurück',
      loading: 'Wird geladen…',
      error: 'Diese Zeitform konnte nicht geladen werden.',
      practiceTitlePrefix: 'Üben:',
      exerciseType: 'Übungstyp',
      mixed: 'Gemischt (Multiple-Choice + Texteingabe)',
      mcqOnly: 'Nur Multiple-Choice',
      textOnly: 'Nur Texteingabe',
      questionCount: 'Anzahl der Fragen',
      start: 'Starten →',
    },
    quiz: {
      back: '← Zurück',
      loadingExercises: 'Übungen werden geladen…',
      error: 'Die Übungen konnten nicht geladen werden.',
      noExercises: 'Es konnten keine Übungen erstellt werden.',
      questionLabel: 'Frage',
      scoreLabel: 'Punkte:',
      subjectLabel: 'Subjekt:',
      check: 'Prüfen',
      next: 'Weiter →',
      seeResults: 'Ergebnisse ansehen →',
      correct: '✓ Richtig!',
      correctWithAccentWarning:
        '✓ Richtig! Achte aber auf Umlaute und ß: Die richtige Form lautet „{answer}“.',
      incorrect: '✗ Nicht ganz — die richtige Antwort war „{answer}“.',
      youWrote: 'Du hast geschrieben: „{text}“',
      results: 'Ergebnisse',
      toReview: 'Zu wiederholen',
      missedItem: '{infinitive} ({tense}, {person}) — richtige Antwort: {answer} · du hast geschrieben: {userAnswer}',
      correctLabel: 'richtige Antwort',
      wroteLabel: 'du hast geschrieben',
      noMistakes: 'Keine Fehler — ausgezeichnete Arbeit!',
      backHome: 'Zurück zur Startseite',
      tryAgain: 'Nochmal ↻',
    },
    vocab: {
      sectionTitle: 'Wortschatz',
      sectionIntro: 'Lerne wichtigen Wortschatz mit Karteikarten nach dem Prinzip der verteilten Wiederholung.',
      mixedTitle: 'Gemischte Wiederholung',
      mixedUsage: 'Eine Mischung aus dem gesamten Wortschatz für eine vollständige Wiederholung.',
      wordsLabel: 'Wörter',
      back: '← Zurück',
      loading: 'Wird geladen…',
      error: 'Der Wortschatz konnte nicht geladen werden.',
      newLabel: 'Neu',
      learningLabel: 'In Bearbeitung',
      reviewLabel: 'Zu wiederholen',
      masteredLabel: 'Gemeistert',
      dueLabel: 'heute fällig',
      startSession: 'Starten →',
      reveal: 'Aufdecken',
      again: 'Nochmal',
      hard: 'Schwer',
      good: 'Gut',
      easy: 'Leicht',
      sessionDone: 'Sitzung beendet!',
      sessionDoneDetail: 'Du hast {count} Karten wiederholt.',
      nothingDue: 'Heute nichts zu wiederholen',
      nothingDueDetail: 'Komm später wieder, oder lerne ein paar Karten im Voraus.',
      studyAnyway: 'Trotzdem lernen',
      cardProgress: '{index} / {total}',
      directionLabel: 'Wiederholungsrichtung',
      directionRecognition: 'Erkennen',
      directionRecognitionHint: 'Wort sehen, Bedeutung erraten',
      directionProduction: 'Produktion',
      directionProductionHint: 'Bedeutung sehen, Wort erraten',
      directionMixed: 'Gemischt',
      directionMixedHint: 'Beide Richtungen gemischt',
    },
    reading: {
      tab: 'Texte',
      sectionIntro: 'Lies, fahre über ein Wort für die Bedeutung und nimm es in deine Karten auf.',
      loading: 'Wird geladen…',
      error: 'Der Text konnte nicht geladen werden.',
      unavailable: 'Lesen ist in dieser Sprache noch nicht verfügbar.',
      levelLabel: 'Niveau',
      tokensLabel: '{count} Wörter',
      sectionsLabel: '{count} Teile',
      newWordsLabel: '{count} neue Wörter',
      coverageLabel: '{percent} % bekannt',
      bandComfortable: 'Angenehm',
      bandChallenging: 'Anspruchsvoll',
      bandHard: 'Schwer',
      bandHint: 'Berechnet aus deinen Karten: ab 98 % bekannter Wörter liest man mühelos.',
      read: 'Lesen',
      continueReading: 'Fortsetzen',
      back: '← Bibliothek',
      prevSection: '← Zurück',
      nextSection: 'Weiter →',
      sectionProgress: 'Teil {index} / {total}',
      savedCount: '{count} gespeichert',
      finishTitle: 'Text beendet',
      finishDetail: 'Du hast {count} Wörter aus diesem Text gespeichert.',
      reviewSaved: 'Diese Wörter wiederholen',
      noneSaved: 'Noch keine Wörter gespeichert.',
      noDefinition: 'Keine Bedeutung gefunden.',
      uncertainNote: 'Geratene Form — bitte prüfen.',
      otherReadings: 'Andere Lesarten',
      chooseSense: 'Wähle die Bedeutung, die du lernen willst',
      save: 'Zu den Karten hinzufügen',
      saved: 'Gespeichert',
      unsave: 'Entfernen',
      markKnown: 'Kenne ich schon',
      knownNote: 'Als bekannt markiert',
      ignore: 'Ignorieren',
      ignoredNote: 'Ignoriert (Eigenname, Zahl…)',
      undo: 'Rückgängig',
      nudge: 'Du hast dieses Wort {count}-mal nachgeschlagen. Hinzufügen?',
      lookupHint: 'Klicken zum Anheften',
      attribution: 'Bedeutungen: Wiktionary (CC BY-SA)',
      displayLabel: 'Darstellung',
      smaller: 'Text verkleinern',
      larger: 'Text vergrößern',
      serif: 'Serif',
      sans: 'Sans',
      highlightOn: 'Hervorhebung an',
      highlightOff: 'Hervorhebung aus',
      statusUnknown: 'neu',
      statusLearning: 'im Lernen',
      statusKnown: 'bekannt',
      selectionLookup: 'Nachschlagen',
      selectionSave: 'Wendung hinzufügen',
      selectionTooLong: 'Auswahl zu lang (höchstens {max} Wörter).',
      coachTitle: 'So liest du hier',
      coachHover: 'Fahre über ein Wort für eine kurze Vorschau.',
      coachClick: 'Klicke, um es anzuheften und zu den Karten hinzuzufügen.',
      coachSelect: 'Ziehe über mehrere Wörter für eine Wendung.',
      coachDismiss: 'Verstanden',
      savedDeckTitle: 'Meine Wörter',
      savedDeckBlurb: 'Beim Lesen gespeicherte Wörter',
      textDeckTitle: 'Wörter aus „{title}“',
      contextLabel: 'Gefunden in',
    },
  },
}

/** Interpolates {placeholders} in a template string, e.g. format(t.quiz.incorrect, { answer: 'parle' }). */
export function format(template: string, vars: Record<string, string | number>): string {
  return template.replace(/\{(\w+)\}/g, (_, key: string) => String(vars[key] ?? ''))
}
