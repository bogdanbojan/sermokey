import { beforeEach, describe, expect, it, vi } from 'vitest'

const LANG = 'fr' as const

type DecksMod = typeof import('./decks')
type SavedMod = typeof import('./saved')

async function fresh(): Promise<{ decks: DecksMod; saved: SavedMod }> {
  vi.resetModules()
  return { decks: await import('./decks'), saved: await import('./saved') }
}

beforeEach(() => localStorage.clear())

describe('text decks', () => {
  it('derives a stable id from the text id', async () => {
    const { decks } = await fresh()
    expect(decks.textDeckID('la-parure')).toBe('text:la-parure')
    expect(decks.isTextDeck('text:la-parure')).toBe(true)
    expect(decks.isTextDeck('custom:abc')).toBe(false)
    expect(decks.textIDOf('text:la-parure')).toBe('la-parure')
    expect(decks.textIDOf('custom:abc')).toBeUndefined()
  })

  it('creates the deck only when a word is actually saved', async () => {
    const { decks } = await fresh()
    const d = decks.useDecks(LANG)
    // Merely resolving a destination must not create anything — opening a
    // book should not litter the library with empty decks.
    expect(d.resolveTarget({ id: 'la-parure', title: 'La Parure' })?.name).toBe('La Parure')
    expect(d.decks.value).toHaveLength(0)

    d.ensureTextDeck('la-parure', 'La Parure')
    expect(d.decks.value).toHaveLength(1)
  })

  it('does not create a second deck for the same text', async () => {
    const { decks } = await fresh()
    const d = decks.useDecks(LANG)
    d.ensureTextDeck('la-parure', 'La Parure')
    d.ensureTextDeck('la-parure', 'La Parure')
    expect(d.decks.value).toHaveLength(1)
  })

  it('follows a retitled text', async () => {
    const { decks } = await fresh()
    const d = decks.useDecks(LANG)
    d.ensureTextDeck('book', 'Untitled')
    d.ensureTextDeck('book', 'A Better Title')
    expect(d.decks.value).toHaveLength(1)
    expect(d.get('text:book')?.name).toBe('A Better Title')
  })
})

describe('custom decks', () => {
  it('creates and lists', async () => {
    const { decks } = await fresh()
    const d = decks.useDecks(LANG)
    const deck = d.create('Food words')
    expect(deck.name).toBe('Food words')
    expect(d.customDecks.value).toHaveLength(1)
    expect(d.textDecks.value).toHaveLength(0)
  })

  it('does not fork a deck when the same name is typed again', async () => {
    const { decks } = await fresh()
    const d = decks.useDecks(LANG)
    const a = d.create('Food')
    const b = d.create('  food  ')
    expect(b.id).toBe(a.id)
    expect(d.customDecks.value).toHaveLength(1)
  })

  it('survives a reload', async () => {
    const first = await fresh()
    first.decks.useDecks(LANG).create('Food')

    const second = await fresh()
    expect(second.decks.useDecks(LANG).customDecks.value.map((x) => x.name)).toEqual(['Food'])
  })
})

describe('destination', () => {
  it("defaults to the text being read, so saving stays one tap", async () => {
    const { decks } = await fresh()
    const d = decks.useDecks(LANG)
    expect(d.target.value).toBe('auto')
    expect(d.resolveTarget({ id: 'la-parure', title: 'La Parure' })?.id).toBe('text:la-parure')
  })

  it('falls back to no deck outside a text', async () => {
    const { decks } = await fresh()
    expect(decks.useDecks(LANG).resolveTarget(undefined)).toBeUndefined()
  })

  it('sticks once chosen, so it is set once rather than per word', async () => {
    const { decks } = await fresh()
    const d = decks.useDecks(LANG)
    const food = d.create('Food')
    d.setTarget(food.id)
    expect(d.resolveTarget({ id: 'la-parure', title: 'La Parure' })?.id).toBe(food.id)
  })

  it('returns to auto when the chosen deck is forgotten', async () => {
    const { decks } = await fresh()
    const d = decks.useDecks(LANG)
    const food = d.create('Food')
    d.setTarget(food.id)
    d.remove(food.id)
    expect(d.target.value).toBe('auto')
  })
})

describe('membership', () => {
  const word = (id: string, over: Record<string, unknown> = {}) => ({
    id,
    lemma: id,
    pos: 'noun',
    gloss: 'g',
    savedAt: '2026-01-01T00:00:00.000Z',
    ...over,
  })

  function seedSaved(words: Record<string, unknown>) {
    localStorage.setItem('sermokey:saved:fr', JSON.stringify({ v: 2, d: words }))
  }

  it('lists words explicitly in a deck', async () => {
    seedSaved({ a: word('a', { decks: ['custom:1'] }), b: word('b') })
    const { saved } = await fresh()
    expect(saved.useSavedWords(LANG).inDeck('custom:1').map((w) => w.id)).toEqual(['a'])
  })

  it('finds words saved before decks existed via their text context', async () => {
    // No `decks` field, but the source text is recorded in the context —
    // so the text's deck must still find it, without a migration.
    seedSaved({
      old: word('old', { context: { textId: 'la-parure', textTitle: 'La Parure', section: 0, sentence: '', start: 0, end: 0 } }),
    })
    const { saved } = await fresh()
    expect(saved.useSavedWords(LANG).inDeck('text:la-parure').map((w) => w.id)).toEqual(['old'])
  })

  it('adds and removes membership without duplicating the word', async () => {
    seedSaved({ a: word('a') })
    const { saved } = await fresh()
    const s = saved.useSavedWords(LANG)

    s.addToDeck('a', 'custom:1')
    s.addToDeck('a', 'custom:1')
    expect(s.decksOf('a')).toEqual(['custom:1'])
    expect(s.all()).toHaveLength(1)

    s.removeFromDeck('a', 'custom:1')
    expect(s.decksOf('a')).toEqual([])
  })

  it('keeps a word in several decks at once', async () => {
    seedSaved({ a: word('a', { decks: ['custom:1', 'custom:2'] }) })
    const { saved } = await fresh()
    const s = saved.useSavedWords(LANG)
    expect(s.inDeck('custom:1')).toHaveLength(1)
    expect(s.inDeck('custom:2')).toHaveLength(1)
  })

  it('keeps the words when a deck is forgotten', async () => {
    // A deck is a grouping. Losing it must not lose the vocabulary or the
    // review history attached to it.
    seedSaved({ a: word('a', { decks: ['custom:1'] }), b: word('b', { decks: ['custom:1'] }) })
    const { saved } = await fresh()
    const s = saved.useSavedWords(LANG)

    s.removeFromAllDecks('custom:1')
    expect(s.all()).toHaveLength(2)
    expect(s.inDeck('custom:1')).toHaveLength(0)
  })
})

describe('plural count label', () => {
  it('agrees with the number in every shipped language', async () => {
    const { plural, strings } = await import('../i18n/strings')
    const cases: Array<[string, number, string]> = [
      ['fr', 1, '1 mot'],
      ['fr', 2, '2 mots'],
      ['es', 1, '1 palabra'],
      ['es', 2, '2 palabras'],
      ['de', 1, '1 Wort'],
      ['de', 2, '2 Wörter'],
      ['en', 1, '1 word'],
      ['en', 2, '2 words'],
    ]
    for (const [lang, n, want] of cases) {
      const forms = strings[lang as 'fr'].reading.deckCountLabel
      expect(plural(forms, n, lang)).toBe(want)
    }
  })
})

// Deleting a book must never delete the grouping built from it. The deck
// outlives the text, keeps its words, and is removable only on purpose.
describe('a deck whose text has been deleted', () => {
  const ctx = (textId: string, title: string) => ({
    textId,
    textTitle: title,
    section: 0,
    sentence: '',
    start: 0,
    end: 0,
  })

  function seedSaved(words: Record<string, unknown>) {
    localStorage.setItem('sermokey:saved:fr', JSON.stringify({ v: 2, d: words }))
  }

  const word = (id: string, over: Record<string, unknown> = {}) => ({
    id,
    lemma: id,
    pos: 'noun',
    gloss: 'g',
    savedAt: '2026-01-01T00:00:00.000Z',
    ...over,
  })

  it('survives as a stored deck even when it was only ever derived', async () => {
    // The word went to a custom deck, so ensureTextDeck was never called
    // while reading — the text's deck existed only by derivation from the
    // word's context. Deleting the book is the moment it has to be made
    // real, or it vanishes with no one having asked for that.
    seedSaved({
      a: word('a', { decks: ['custom:1'], context: ctx('petit-prince', 'Le Petit Prince') }),
    })
    const { decks, saved } = await fresh()
    const d = decks.useDecks(LANG)
    expect(d.get('text:petit-prince')).toBeUndefined()

    // What HomeView.onRemoveText does on a successful delete.
    expect(saved.useSavedWords(LANG).fromText('petit-prince')).toHaveLength(1)
    d.ensureTextDeck('petit-prince', 'Le Petit Prince')

    expect(d.get('text:petit-prince')?.name).toBe('Le Petit Prince')
  })

  it('keeps its words, and keeps them after it is forgotten', async () => {
    seedSaved({
      a: word('a', { decks: ['text:petit-prince'], context: ctx('petit-prince', 'Le Petit Prince') }),
      b: word('b', { decks: ['text:petit-prince'], context: ctx('petit-prince', 'Le Petit Prince') }),
    })
    const { decks, saved } = await fresh()
    const d = decks.useDecks(LANG)
    const s = saved.useSavedWords(LANG)
    d.ensureTextDeck('petit-prince', 'Le Petit Prince')

    expect(s.inDeck('text:petit-prince')).toHaveLength(2)

    // Forgetting is the reader's decision, and it costs them nothing but
    // the grouping.
    s.removeFromAllDecks('text:petit-prince')
    d.remove('text:petit-prince')

    expect(d.get('text:petit-prince')).toBeUndefined()
    expect(s.all()).toHaveLength(2)
  })

  it('does not come back once forgotten', async () => {
    // The words still carry context.textId, so only scoping derivation to
    // texts still in the library keeps a forgotten orphan gone. This is
    // the rule HomeView.deckCards applies; asserted here on its inputs.
    seedSaved({ a: word('a', { context: ctx('petit-prince', 'Le Petit Prince') }) })
    const { decks, saved } = await fresh()
    const d = decks.useDecks(LANG)
    d.ensureTextDeck('petit-prince', 'Le Petit Prince')
    d.remove('text:petit-prince')

    const libraryTextIDs = new Set<string>() // the book was deleted
    const derivable = saved
      .useSavedWords(LANG)
      .all()
      .filter((w) => w.context && libraryTextIDs.has(w.context.textId))

    expect(d.get('text:petit-prince')).toBeUndefined()
    expect(derivable).toHaveLength(0)
  })

  it('a custom deck is never touched by deleting a book', async () => {
    seedSaved({ a: word('a', { decks: ['custom:1'], context: ctx('petit-prince', 'Le Petit Prince') }) })
    const { decks, saved } = await fresh()
    const d = decks.useDecks(LANG)
    const custom = d.create('Le Petit Prince')
    saved.useSavedWords(LANG).addToDeck('a', custom.id)

    d.ensureTextDeck('petit-prince', 'Le Petit Prince')
    d.remove('text:petit-prince')

    expect(d.get(custom.id)?.name).toBe('Le Petit Prince')
    expect(saved.useSavedWords(LANG).inDeck(custom.id)).toHaveLength(1)
  })
})
