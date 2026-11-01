// UI chrome strings (buttons, headings, feedback messages) for each
// language. This app follows an "immersive" convention already established
// for French: interface text is in the language being studied, not fixed to
// English — so switching to Spanish content also switches "Vérifier" to
// "Comprobar", not just the exercises. Content strings with a value to
// interpolate use {placeholders}; see format() in language.ts.
import type { LanguageCode } from '../types/api'

/**
 * A count-bearing string in its grammatical forms.
 *
 * The rest of this file hardcodes plurals ("{count} mots"), which renders
 * "1 mots". Fixing that everywhere means reshaping a dozen keys across
 * four languages; new strings at least start correct.
 */
export interface PluralForms {
  one: string
  other: string
}

export interface UiStrings {
  brand: { title: string; tagline: string }
  /** Shared across screens; previously these were retyped per view or
   *  simply absent, which is why no error state had a retry control. */
  common: {
    retry: string
    dismiss: string
    cancel: string
    close: string
  }
  /** Persistence is the one failure the learner must never be left to
   *  discover on their own: everything they have done lives in the browser
   *  and nowhere else. */
  storage: {
    quotaTitle: string
    quotaDetail: string
    unavailableTitle: string
    unavailableDetail: string
    backup: string
    restore: string
    restoreConfirm: string
    crashTitle: string
    crashDetail: string
    reload: string
    reset: string
    resetConfirm: string
  }
  settings: {
    interfaceLanguage: string
    followStudyLanguage: string
  }
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
    /** Classifier above the mixed-review title. Distinct from mixedTitle:
     *  the two were the same string, so the card read "RÉVISION MIXTE /
     *  Révision mixte" while every other card pairs a classifier with a
     *  different name. */
    mixedPill: string
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
    undo: string
    undone: string
    saveFailed: string
    leechLabel: string
    leechHint: string
    unsuspend: string
    intervalToday: string
    intervalDays: string
    intervalMonths: string
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
    statusLegendTitle: string
    lineHeightLabel: string
    tighter: string
    looser: string
    sectionJump: string
    textSizeLabel: string
    typefaceLabel: string
    highlightLabel: string
    resetDisplay: string
    displayMenu: string
    // Decks
    deckDestination: string
    deckChange: string
    deckMyWords: string
    deckNew: string
    deckNewPlaceholder: string
    deckCreate: string
    deckAddedTo: string
    deckManage: string
    /** Kind shown on the all-saved-words card. */
    deckAll: string
    deckForget: string
    deckForgetConfirm: string
    /** Kind shown on a deck whose source text has been deleted. It is also
     *  the only text deck that can be forgotten, so the pill explains why
     *  this card has a remove control when its siblings do not. */
    deckOrphaned: string
    /** Reassurance in the delete-a-text confirmation. */
    removeKeepsWords: PluralForms
    deckEmpty: string
    /** Plural forms, selected by Intl.PluralRules. See plural(). */
    deckCountLabel: PluralForms
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
    // EPUB import
    import: string
    importing: string // "{name}"
    importHint: string
    importFailed: string
    importWrongType: string
    importedPill: string
    remove: string
    removeConfirm: string // "{title}"
  }
}


/**
 * Languages the *interface* can be rendered in: every study language, plus
 * English. Keeping this wider than LanguageCode is what lets the interface
 * language be chosen independently of the content language.
 */
export type UiLanguageCode = LanguageCode | 'en'

export const strings: Record<UiLanguageCode, UiStrings> = {
  fr: {
    brand: { title: 'Grammaire', tagline: 'conjugaison française' },
    common: { retry: 'Réessayer', dismiss: 'Fermer', cancel: 'Annuler', close: 'Fermer' },
    storage: {
      quotaTitle: 'Stockage saturé',
      quotaDetail: 'Vos dernières réponses n’ont pas pu être enregistrées. Exportez une sauvegarde, puis libérez de l’espace.',
      unavailableTitle: 'Stockage indisponible',
      unavailableDetail: 'Votre navigateur bloque le stockage local. Vos progrès ne seront pas conservés.',
      backup: 'Exporter une sauvegarde',
      restore: 'Restaurer',
      restoreConfirm: 'Remplacer toutes vos données actuelles par cette sauvegarde ?',
      crashTitle: 'Quelque chose s’est mal passé',
      crashDetail: 'Exportez d’abord une sauvegarde — elle fonctionne même si l’application ne s’affiche plus.',
      reload: 'Recharger',
      reset: 'Effacer les données locales',
      resetConfirm: 'Effacer définitivement toutes vos données ? Exportez une sauvegarde d’abord.',
    },
    settings: { interfaceLanguage: 'Langue de l’interface', followStudyLanguage: 'Suivre la langue étudiée' },
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
      mixedPill: 'Tout le vocabulaire',
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
      undo: 'Annuler',
      undone: 'Réponse annulée.',
      saveFailed: 'Réponse non enregistrée — vérifiez le stockage.',
      leechLabel: 'Carte difficile',
      leechHint: 'Oubliée {count} fois. Reformulez-la plutôt que de la répéter.',
      unsuspend: 'Remettre en révision',
      intervalToday: 'aujourd’hui',
      intervalDays: '{count} j',
      intervalMonths: '{count} mois',
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
      statusLegendTitle: 'Légende',
      lineHeightLabel: 'Interligne',
      tighter: 'Resserrer l’interligne',
      looser: 'Élargir l’interligne',
      sectionJump: 'Aller à la partie',
      textSizeLabel: 'Taille du texte',
      typefaceLabel: 'Police',
      highlightLabel: 'Surligner les mots',
      resetDisplay: 'Rétablir les réglages',
      displayMenu: 'Réglages de lecture',
      deckDestination: 'Ajouter à',
      deckChange: 'Changer',
      deckMyWords: 'Mes mots',
      deckNew: 'Nouveau paquet',
      deckNewPlaceholder: 'Nom du paquet',
      deckCreate: 'Créer',
      deckAddedTo: 'Ajouté à « {name} »',
      deckManage: 'Mes paquets',
      deckAll: 'Tous mes mots',
      deckForget: 'Oublier ce paquet',
      deckForgetConfirm: 'Oublier « {name} » ? Les mots sont conservés dans Mes mots.',
      deckOrphaned: 'Texte supprimé',
      removeKeepsWords: { one: 'Le mot enregistré et son paquet sont conservés.', other: 'Vos {count} mots enregistrés et leur paquet sont conservés.' },
      deckEmpty: 'Aucun paquet pour l’instant.',
      deckCountLabel: { one: '{count} mot', other: '{count} mots' },
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
      import: 'Importer un EPUB',
      importing: 'Lecture de {name}…',
      importHint: 'ou glissez un fichier .epub ici',
      importFailed: 'Impossible d’importer ce fichier.',
      importWrongType: 'Choisissez un fichier .epub.',
      importedPill: 'Importé',
      remove: 'Retirer',
      removeConfirm: 'Retirer « {title} » de votre bibliothèque ?',
    },
  },
  es: {
    brand: { title: 'Gramática', tagline: 'conjugación española' },
    common: { retry: 'Reintentar', dismiss: 'Cerrar', cancel: 'Cancelar', close: 'Cerrar' },
    storage: {
      quotaTitle: 'Almacenamiento lleno',
      quotaDetail: 'No se pudieron guardar tus últimas respuestas. Exporta una copia y libera espacio.',
      unavailableTitle: 'Almacenamiento no disponible',
      unavailableDetail: 'Tu navegador bloquea el almacenamiento local. Tu progreso no se guardará.',
      backup: 'Exportar una copia',
      restore: 'Restaurar',
      restoreConfirm: '¿Reemplazar todos tus datos actuales por esta copia?',
      crashTitle: 'Algo ha salido mal',
      crashDetail: 'Exporta primero una copia — funciona aunque la aplicación no se muestre.',
      reload: 'Recargar',
      reset: 'Borrar los datos locales',
      resetConfirm: '¿Borrar definitivamente todos tus datos? Exporta una copia antes.',
    },
    settings: { interfaceLanguage: 'Idioma de la interfaz', followStudyLanguage: 'Seguir el idioma estudiado' },
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
      mixedPill: 'Todo el vocabulario',
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
      undo: 'Deshacer',
      undone: 'Respuesta deshecha.',
      saveFailed: 'Respuesta no guardada — revisa el almacenamiento.',
      leechLabel: 'Tarjeta difícil',
      leechHint: 'Olvidada {count} veces. Reescríbela en vez de repetirla.',
      unsuspend: 'Volver a repasar',
      intervalToday: 'hoy',
      intervalDays: '{count} d',
      intervalMonths: '{count} m',
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
      statusLegendTitle: 'Leyenda',
      lineHeightLabel: 'Interlineado',
      tighter: 'Reducir el interlineado',
      looser: 'Aumentar el interlineado',
      sectionJump: 'Ir a la parte',
      textSizeLabel: 'Tamaño del texto',
      typefaceLabel: 'Tipografía',
      highlightLabel: 'Resaltar palabras',
      resetDisplay: 'Restablecer los ajustes',
      displayMenu: 'Ajustes de lectura',
      deckDestination: 'Añadir a',
      deckChange: 'Cambiar',
      deckMyWords: 'Mis palabras',
      deckNew: 'Nuevo mazo',
      deckNewPlaceholder: 'Nombre del mazo',
      deckCreate: 'Crear',
      deckAddedTo: 'Añadido a «{name}»',
      deckManage: 'Mis mazos',
      deckAll: 'Todas mis palabras',
      deckForget: 'Olvidar este mazo',
      deckForgetConfirm: '¿Olvidar «{name}»? Las palabras se conservan en Mis palabras.',
      deckOrphaned: 'Texto eliminado',
      removeKeepsWords: { one: 'La palabra guardada y su mazo se conservan.', other: 'Tus {count} palabras guardadas y su mazo se conservan.' },
      deckEmpty: 'Aún no hay mazos.',
      deckCountLabel: { one: '{count} palabra', other: '{count} palabras' },
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
      import: 'Importar un EPUB',
      importing: 'Leyendo {name}…',
      importHint: 'o arrastra aquí un archivo .epub',
      importFailed: 'No se ha podido importar este archivo.',
      importWrongType: 'Elige un archivo .epub.',
      importedPill: 'Importado',
      remove: 'Quitar',
      removeConfirm: '¿Quitar «{title}» de tu biblioteca?',
    },
  },
  de: {
    brand: { title: 'Grammatik', tagline: 'deutsche Konjugation' },
    common: { retry: 'Erneut versuchen', dismiss: 'Schließen', cancel: 'Abbrechen', close: 'Schließen' },
    storage: {
      quotaTitle: 'Speicher voll',
      quotaDetail: 'Deine letzten Antworten konnten nicht gespeichert werden. Exportiere eine Sicherung und schaffe Platz.',
      unavailableTitle: 'Speicher nicht verfügbar',
      unavailableDetail: 'Dein Browser blockiert den lokalen Speicher. Dein Fortschritt wird nicht gesichert.',
      backup: 'Sicherung exportieren',
      restore: 'Wiederherstellen',
      restoreConfirm: 'Alle aktuellen Daten durch diese Sicherung ersetzen?',
      crashTitle: 'Etwas ist schiefgelaufen',
      crashDetail: 'Exportiere zuerst eine Sicherung — das funktioniert auch, wenn die App nicht mehr lädt.',
      reload: 'Neu laden',
      reset: 'Lokale Daten löschen',
      resetConfirm: 'Alle Daten endgültig löschen? Exportiere vorher eine Sicherung.',
    },
    settings: { interfaceLanguage: 'Sprache der Oberfläche', followStudyLanguage: 'Der Lernsprache folgen' },
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
      mixedPill: 'Gesamter Wortschatz',
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
      undo: 'Rückgängig',
      undone: 'Antwort zurückgenommen.',
      saveFailed: 'Antwort nicht gespeichert — Speicher prüfen.',
      leechLabel: 'Schwierige Karte',
      leechHint: '{count}-mal vergessen. Schreibe sie neu, statt sie zu wiederholen.',
      unsuspend: 'Wieder wiederholen',
      intervalToday: 'heute',
      intervalDays: '{count} T',
      intervalMonths: '{count} Mon.',
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
      statusLegendTitle: 'Legende',
      lineHeightLabel: 'Zeilenabstand',
      tighter: 'Zeilenabstand verringern',
      looser: 'Zeilenabstand vergrößern',
      sectionJump: 'Zu Teil springen',
      textSizeLabel: 'Schriftgröße',
      typefaceLabel: 'Schriftart',
      highlightLabel: 'Wörter hervorheben',
      resetDisplay: 'Einstellungen zurücksetzen',
      displayMenu: 'Leseeinstellungen',
      deckDestination: 'Hinzufügen zu',
      deckChange: 'Ändern',
      deckMyWords: 'Meine Wörter',
      deckNew: 'Neuer Stapel',
      deckNewPlaceholder: 'Name des Stapels',
      deckCreate: 'Erstellen',
      deckAddedTo: 'Zu „{name}“ hinzugefügt',
      deckManage: 'Meine Stapel',
      deckAll: 'Alle meine Wörter',
      deckForget: 'Diesen Stapel vergessen',
      deckForgetConfirm: '„{name}“ vergessen? Die Wörter bleiben in Meine Wörter.',
      deckOrphaned: 'Text gelöscht',
      removeKeepsWords: { one: 'Das gespeicherte Wort und sein Stapel bleiben erhalten.', other: 'Deine {count} gespeicherten Wörter und ihr Stapel bleiben erhalten.' },
      deckEmpty: 'Noch keine Stapel.',
      deckCountLabel: { one: '{count} Wort', other: '{count} Wörter' },
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
      import: 'EPUB importieren',
      importing: '{name} wird gelesen…',
      importHint: 'oder eine .epub-Datei hierher ziehen',
      importFailed: 'Diese Datei konnte nicht importiert werden.',
      importWrongType: 'Bitte eine .epub-Datei wählen.',
      importedPill: 'Importiert',
      remove: 'Entfernen',
      removeConfirm: '„{title}“ aus deiner Bibliothek entfernen?',
    },
  },
  // English is the one interface language that is never a study language.
  // It exists as an escape hatch: the immersive convention above is right
  // for an intermediate learner and wrong for a day-one beginner, who
  // cannot read the grade buttons they are being asked to press, or the
  // confirmation dialog in front of a destructive action. Content — glosses
  // and translations — was already English, so this only makes the chrome
  // agree with it for learners who need that.
  en: {
    brand: { title: 'Grammar', tagline: 'conjugation practice' },
    common: { retry: 'Try again', dismiss: 'Dismiss', cancel: 'Cancel', close: 'Close' },
    storage: {
      quotaTitle: 'Storage is full',
      quotaDetail: 'Your last answers could not be saved. Export a backup, then free up some space.',
      unavailableTitle: 'Storage unavailable',
      unavailableDetail: 'Your browser is blocking local storage. Your progress will not be kept.',
      backup: 'Export a backup',
      restore: 'Restore',
      restoreConfirm: 'Replace all of your current data with this backup?',
      crashTitle: 'Something went wrong',
      crashDetail: 'Export a backup first — it works even when the app will not load.',
      reload: 'Reload',
      reset: 'Clear local data',
      resetConfirm: 'Permanently erase all of your data? Export a backup first.',
    },
    settings: { interfaceLanguage: 'Interface language', followStudyLanguage: 'Follow the study language' },
    home: {
      title: 'Grammar',
      intro: 'Pick a tense to study and practise, or start a mixed review of every tense.',
      loading: 'Loading…',
      error: 'Something went wrong.',
      mixedTitle: 'Mixed review',
      mixedUsage: 'A mix of every tense, for all-round practice.',
      grammarTab: 'Grammar',
      vocabTab: 'Vocabulary',
      searchPlaceholder: 'Search…',
      noResults: 'No results.',
      continueTitle: 'Pick up where you left off',
      clusters: {
        basics: 'The basics',
        people: 'People',
        daily: 'Everyday life',
        world: 'The world',
        schoolWork: 'School, work & media',
        wordsConcepts: 'Words & concepts',
      },
    },
    tenseCard: { theory: 'Theory', practice: 'Practise', mixedPill: 'Review' },
    theory: {
      back: '← Back',
      loading: 'Loading…',
      error: 'Could not load this page.',
      formation: 'Formation',
      toRemember: 'Worth remembering',
      example: 'Example',
      practiceThisTense: 'Practise this tense →',
      personLabels: ['I', 'you', 'he / she / one', 'we', 'you (pl.)', 'they'],
    },
    quizSetup: {
      back: '← Back',
      loading: 'Loading…',
      error: 'Could not load this tense.',
      practiceTitlePrefix: 'Practise:',
      exerciseType: 'Exercise type',
      mixed: 'Mixed (multiple choice + typing)',
      mcqOnly: 'Multiple choice only',
      textOnly: 'Typing only',
      questionCount: 'Number of questions',
      start: 'Start →',
    },
    quiz: {
      back: '← Back',
      loadingExercises: 'Loading exercises…',
      error: 'Could not load the exercises.',
      noExercises: 'No exercises could be generated.',
      questionLabel: 'Question',
      scoreLabel: 'Score:',
      subjectLabel: 'subject:',
      check: 'Check',
      next: 'Next →',
      seeResults: 'See results →',
      correct: '✓ Correct!',
      correctWithAccentWarning:
        '✓ Correct! Watch the accents though: the written form is “{answer}”.',
      incorrect: '✗ Not quite — the answer was “{answer}”.',
      youWrote: 'You wrote: “{text}”',
      results: 'Results',
      toReview: 'Worth reviewing',
      missedItem:
        '{infinitive} ({tense}, {person}) — correct answer: {answer} · you wrote: {userAnswer}',
      correctLabel: 'correct answer',
      wroteLabel: 'you wrote',
      noMistakes: 'No mistakes — nicely done!',
      backHome: 'Back to home',
      tryAgain: 'Start again ↻',
    },
    vocab: {
      sectionTitle: 'Vocabulary',
      sectionIntro: 'Learn core vocabulary with spaced-repetition flashcards.',
      mixedTitle: 'Mixed review',
      mixedPill: 'All vocabulary',
      mixedUsage: 'A mix of all the vocabulary, for a full review.',
      wordsLabel: 'words',
      back: '← Back',
      loading: 'Loading…',
      error: 'Could not load the vocabulary.',
      newLabel: 'New',
      learningLabel: 'Learning',
      reviewLabel: 'To review',
      masteredLabel: 'Mastered',
      dueLabel: 'due today',
      startSession: 'Start →',
      reveal: 'Reveal',
      again: 'Again',
      hard: 'Hard',
      good: 'Good',
      easy: 'Easy',
      sessionDone: 'Session complete!',
      sessionDoneDetail: 'You reviewed {count} cards.',
      nothingDue: 'Nothing due today',
      nothingDueDetail: 'Come back later, or study a few cards ahead of schedule.',
      studyAnyway: 'Study anyway',
      cardProgress: '{index} / {total}',
      directionLabel: 'Review direction',
      directionRecognition: 'Recognition',
      directionRecognitionHint: 'See the word, recall the meaning',
      directionProduction: 'Production',
      directionProductionHint: 'See the meaning, recall the word',
      directionMixed: 'Mixed',
      directionMixedHint: 'Both directions, shuffled',
      undo: 'Undo',
      undone: 'Answer undone.',
      saveFailed: 'Answer not saved — check your storage.',
      leechLabel: 'Difficult card',
      leechHint: 'Forgotten {count} times. Rewrite it rather than drilling it again.',
      unsuspend: 'Put back into review',
      intervalToday: 'today',
      intervalDays: '{count}d',
      intervalMonths: '{count}mo',
    },
    reading: {
      tab: 'Texts',
      sectionIntro: 'Read, hover a word for its meaning, and add it to your cards.',
      loading: 'Loading…',
      error: 'Could not load the text.',
      unavailable: 'Reading is not available in this language yet.',
      levelLabel: 'Level',
      tokensLabel: '{count} words',
      sectionsLabel: '{count} parts',
      newWordsLabel: '{count} new words',
      coverageLabel: '{percent}% known',
      bandComfortable: 'Comfortable',
      bandChallenging: 'Challenging',
      bandHard: 'Hard',
      bandHint: 'Based on your cards: above 98% known words, reading is effortless.',
      read: 'Read',
      continueReading: 'Continue',
      back: '← Library',
      prevSection: '← Previous',
      nextSection: 'Next →',
      sectionProgress: 'Part {index} / {total}',
      savedCount: '{count} saved',
      finishTitle: 'Text finished',
      finishDetail: 'You saved {count} words from this text.',
      reviewSaved: 'Review these words',
      noneSaved: 'No words saved yet.',
      noDefinition: 'No definition found.',
      uncertainNote: 'Guessed form — worth checking.',
      otherReadings: 'Other readings',
      chooseSense: 'Choose the meaning to learn',
      save: 'Add to cards',
      saved: 'Saved',
      unsave: 'Remove',
      markKnown: 'I already know this',
      knownNote: 'Marked as known',
      ignore: 'Ignore',
      ignoredNote: 'Ignored (name, numeral…)',
      undo: 'Undo',
      nudge: 'You have looked this word up {count} times. Add it?',
      lookupHint: 'Click to pin',
      attribution: 'Definitions: Wiktionary (CC BY-SA)',
      displayLabel: 'Display',
      smaller: 'Smaller text',
      larger: 'Larger text',
      serif: 'Serif',
      sans: 'Sans',
      highlightOn: 'Highlighting on',
      highlightOff: 'Highlighting off',
      statusUnknown: 'new',
      statusLearning: 'learning',
      statusKnown: 'known',
      statusLegendTitle: 'Key',
      lineHeightLabel: 'Line spacing',
      tighter: 'Tighter line spacing',
      looser: 'Looser line spacing',
      sectionJump: 'Jump to part',
      textSizeLabel: 'Text size',
      typefaceLabel: 'Typeface',
      highlightLabel: 'Highlight words',
      resetDisplay: 'Reset to defaults',
      displayMenu: 'Reading settings',
      deckDestination: 'Add to',
      deckChange: 'Change',
      deckMyWords: 'My words',
      deckNew: 'New deck',
      deckNewPlaceholder: 'Deck name',
      deckCreate: 'Create',
      deckAddedTo: 'Added to “{name}”',
      deckManage: 'My decks',
      deckAll: 'All my words',
      deckForget: 'Forget this deck',
      deckForgetConfirm: 'Forget “{name}”? The words stay in My words.',
      deckOrphaned: 'Deleted text',
      removeKeepsWords: { one: 'The saved word and its deck are kept.', other: 'Your {count} saved words and their deck are kept.' },
      deckEmpty: 'No decks yet.',
      deckCountLabel: { one: '{count} word', other: '{count} words' },
      selectionLookup: 'Look up',
      selectionSave: 'Add the phrase',
      selectionTooLong: 'Selection too long ({max} words maximum).',
      coachTitle: 'How to read here',
      coachHover: 'Hover a word for a quick preview.',
      coachClick: 'Click to pin it and add it to your cards.',
      coachSelect: 'Drag across several words for a phrase.',
      coachDismiss: 'Got it',
      savedDeckTitle: 'My words',
      savedDeckBlurb: 'Words saved while reading',
      textDeckTitle: 'Words from “{title}”',
      contextLabel: 'Met in',
      import: 'Import an EPUB',
      importing: 'Reading {name}…',
      importHint: 'or drop a .epub file here',
      importFailed: 'Could not import this file.',
      importWrongType: 'Please choose a .epub file.',
      importedPill: 'Imported',
      remove: 'Remove',
      removeConfirm: 'Remove “{title}” from your library?',
    },
  },
}

/**
 * Picks the right plural form for a count, then interpolates it.
 *
 * Intl.PluralRules rather than `count === 1`, so a language with more than
 * two categories only needs its forms added to PluralForms, not a new
 * branch at every call site.
 */
export function plural(forms: PluralForms, count: number, locale: string): string {
  const category = new Intl.PluralRules(locale).select(count)
  const template = category === 'one' ? forms.one : forms.other
  return format(template, { count: new Intl.NumberFormat(locale).format(count) })
}

/** Interpolates {placeholders} in a template string, e.g. format(t.quiz.incorrect, { answer: 'parle' }). */
export function format(template: string, vars: Record<string, string | number>): string {
  return template.replace(/\{(\w+)\}/g, (_, key: string) => String(vars[key] ?? ''))
}
