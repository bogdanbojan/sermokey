import { describe, expect, it } from 'vitest'
import { renderTokens } from './tokens'
import { NO_ENTRY, type ReadingParagraph } from '../types/api'

/** "Le café" — 'é' is two bytes, so byte and UTF-16 offsets disagree. */
const para = (): ReadingParagraph => ({
  text: 'Le café.',
  //     Le | ' ' | café | .
  spans: [0, 2, 0, 2, 3, NO_ENTRY, 3, 8, 1, 8, 9, NO_ENTRY],
})

describe('renderTokens', () => {
  it('slices on byte offsets, not UTF-16 offsets', () => {
    const tokens = renderTokens(para())
    expect(tokens.map((t) => t.text)).toEqual(['Le', ' ', 'café', '.'])
  })

  it('marks words and inter-word material apart', () => {
    const tokens = renderTokens(para())
    expect(tokens.map((t) => t.word)).toEqual([true, false, true, false])
  })

  it('caches per paragraph object rather than per text', () => {
    const p = para()
    expect(renderTokens(p)).toBe(renderTokens(p))
  })

  it('does not thrash above the old 256-paragraph cache ceiling', () => {
    // The previous cache cleared itself wholesale past 256 entries, so a
    // long chapter re-encoded every paragraph on every access.
    const paras = Array.from({ length: 400 }, para)
    const first = paras.map(renderTokens)
    const second = paras.map(renderTokens)
    for (let i = 0; i < paras.length; i++) expect(first[i]).toBe(second[i])
  })
})
