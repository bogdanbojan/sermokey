// Groups the 24 flat vocabulary categories into 6 semantically coherent
// clusters for the Home page accordions — browsing 24 cards at once has no
// information architecture; 6 labeled, collapsible groups does. Mapping is
// by category ID per language (not by array position) so it stays correct
// even if a language's category list is reordered later.
import type { LanguageCode } from './types/api'

export type ClusterKey = 'basics' | 'people' | 'daily' | 'world' | 'schoolWork' | 'wordsConcepts'

const CLUSTER_BY_CATEGORY: Record<LanguageCode, Record<string, ClusterKey>> = {
  fr: {
    nombres: 'basics',
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
    ecole: 'schoolWork',
    travail: 'schoolWork',
    metiers: 'schoolWork',
    argent: 'schoolWork',
    adjectifs: 'wordsConcepts',
    verbes: 'wordsConcepts',
    fonction: 'wordsConcepts',
    abstrait: 'wordsConcepts',
    loisirs: 'wordsConcepts',
  },
  es: {
    numeros: 'basics',
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
    escuela: 'schoolWork',
    trabajo: 'schoolWork',
    profesiones: 'schoolWork',
    dinero: 'schoolWork',
    adjetivos: 'wordsConcepts',
    verbos: 'wordsConcepts',
    funcion: 'wordsConcepts',
    abstracto: 'wordsConcepts',
    ocio: 'wordsConcepts',
  },
}

/** Display order for the 6 clusters. */
export const CLUSTER_ORDER: ClusterKey[] = ['basics', 'people', 'daily', 'world', 'schoolWork', 'wordsConcepts']

export function clusterFor(lang: LanguageCode, categoryId: string): ClusterKey {
  return CLUSTER_BY_CATEGORY[lang][categoryId] ?? 'wordsConcepts'
}
