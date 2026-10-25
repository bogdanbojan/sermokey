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
        schoolWork: 'École, travail & argent',
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
        schoolWork: 'Escuela, trabajo y dinero',
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
  },
}

/** Interpolates {placeholders} in a template string, e.g. format(t.quiz.incorrect, { answer: 'parle' }). */
export function format(template: string, vars: Record<string, string | number>): string {
  return template.replace(/\{(\w+)\}/g, (_, key: string) => String(vars[key] ?? ''))
}
