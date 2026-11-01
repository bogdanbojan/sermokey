import { describe, expect, it } from 'vitest'
import { MAX_QUIZ_COUNT, MIN_QUIZ_COUNT, __parseHashForTests as parseHash } from './router'

describe('parseHash', () => {
  it('falls back to home for anything unrecognised', () => {
    expect(parseHash('').name).toBe('home')
    expect(parseHash('#/nonsense').name).toBe('home')
  })

  it('does not throw on a malformed percent escape', () => {
    // A truncated shared link. This used to throw a URIError at module
    // scope, before Vue mounted — a blank page with no way to recover.
    expect(() => parseHash('#/theory/%E0%A4%A')).not.toThrow()
    expect(parseHash('#/theory/%E0%A4%A').name).toBe('theory')
  })

  it('clamps the quiz count instead of trusting the URL', () => {
    const big = parseHash('#/quiz/x/mcq/100000')
    expect(big.name === 'quiz' && big.count).toBe(MAX_QUIZ_COUNT)

    // Number('') is 0 and Number.isFinite(0) is true, so this used to ask
    // for a zero-question quiz and render "no exercises".
    const empty = parseHash('#/quiz/x/mcq/')
    expect(empty.name).toBe('home')

    const zero = parseHash('#/quiz/x/mcq/0')
    expect(zero.name === 'quiz' && zero.count).toBe(MIN_QUIZ_COUNT)
  })

  it('round-trips the deck variants', () => {
    expect(parseHash('#/flashcards/saved')).toEqual({ name: 'flashcards', deck: { kind: 'saved' } })
    expect(parseHash('#/flashcards/text/abc')).toEqual({
      name: 'flashcards',
      deck: { kind: 'text', id: 'abc' },
    })
  })

  it('rejects a negative section', () => {
    const r = parseHash('#/read/book/-3')
    expect(r.name === 'read' && r.section).toBe(0)
  })
})
