// Groups the flat vocabulary categories into 6 semantically coherent
// clusters for the Home page accordions — browsing 27 cards at once has no
// information architecture; 6 labeled, collapsible groups does. Mapping is
// by category ID per language (not by array position) so it stays correct
// even if a language's category list is reordered later.
//
// Every category ID declared by the backend must appear here. clusterFor
// falls back to 'wordsConcepts' rather than throwing, so a forgotten entry
// silently misfiles a whole category instead of failing loudly — hence
// assertEveryCategoryMapped below, which the Home view calls once in dev.
import type { LanguageCode } from './types/api'

export type ClusterKey = 'basics' | 'people' | 'daily' | 'world' | 'schoolWork' | 'wordsConcepts'

const CLUSTER_BY_CATEGORY: Record<LanguageCode, Record<string, ClusterKey>> = {
  fr: {
    nombres: 'basics',
    mesures: 'basics',
    temps: 'basics',
    couleurs: 'basics',
    famille: 'people',
    corps: 'people',
    emotions: 'people',
    maison: 'daily',
    nourriture: 'daily',
    vetements: 'daily',
    sante: 'daily',
    nature: 'world',
    animaux: 'world',
    lieux: 'world',
    ville: 'world',
    transport: 'world',
    societe: 'world',
    ecole: 'schoolWork',
    travail: 'schoolWork',
    metiers: 'schoolWork',
    argent: 'schoolWork',
    medias: 'schoolWork',
    adjectifs: 'wordsConcepts',
    verbes: 'wordsConcepts',
    fonction: 'wordsConcepts',
    abstrait: 'wordsConcepts',
    loisirs: 'wordsConcepts',
  },
  es: {
    numeros: 'basics',
    medidas: 'basics',
    tiempo: 'basics',
    colores: 'basics',
    familia: 'people',
    cuerpo: 'people',
    emociones: 'people',
    casa: 'daily',
    comida: 'daily',
    ropa: 'daily',
    salud: 'daily',
    naturaleza: 'world',
    animales: 'world',
    lugares: 'world',
    ciudad: 'world',
    transporte: 'world',
    sociedad: 'world',
    escuela: 'schoolWork',
    trabajo: 'schoolWork',
    profesiones: 'schoolWork',
    dinero: 'schoolWork',
    medios: 'schoolWork',
    adjetivos: 'wordsConcepts',
    verbos: 'wordsConcepts',
    funcion: 'wordsConcepts',
    abstracto: 'wordsConcepts',
    ocio: 'wordsConcepts',
  },
  de: {
    zahlen: 'basics',
    masse: 'basics',
    zeit: 'basics',
    farben: 'basics',
    familie: 'people',
    koerper: 'people',
    gefuehle: 'people',
    haus: 'daily',
    essen: 'daily',
    kleidung: 'daily',
    gesundheit: 'daily',
    natur: 'world',
    tiere: 'world',
    orte: 'world',
    stadt: 'world',
    verkehr: 'world',
    gesellschaft: 'world',
    schule: 'schoolWork',
    arbeit: 'schoolWork',
    berufe: 'schoolWork',
    geld: 'schoolWork',
    medien: 'schoolWork',
    adjektive: 'wordsConcepts',
    verben: 'wordsConcepts',
    funktion: 'wordsConcepts',
    abstrakt: 'wordsConcepts',
    freizeit: 'wordsConcepts',
  },
}

/** Display order for the 6 clusters. */
export const CLUSTER_ORDER: ClusterKey[] = ['basics', 'people', 'daily', 'world', 'schoolWork', 'wordsConcepts']

export function clusterFor(lang: LanguageCode, categoryId: string): ClusterKey {
  return CLUSTER_BY_CATEGORY[lang][categoryId] ?? 'wordsConcepts'
}

/**
 * Warns (in dev only) about any backend category with no cluster mapping.
 * Without this the fallback in clusterFor hides the mistake completely:
 * the category still renders, just dumped into "Words & concepts".
 */
export function assertEveryCategoryMapped(lang: LanguageCode, categoryIds: string[]) {
  if (!import.meta.env.DEV) return
  const map = CLUSTER_BY_CATEGORY[lang]
  const unmapped = categoryIds.filter((id) => !(id in map))
  if (unmapped.length) {
    console.warn(
      `vocabClusters: ${lang} has no cluster mapping for ${unmapped.join(', ')} — ` +
        `they will all be filed under "wordsConcepts".`,
    )
  }
}
