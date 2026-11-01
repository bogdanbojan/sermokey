import { beforeEach, describe, expect, it, vi } from 'vitest'

type Prefs = typeof import('./prefs')

async function fresh(): Promise<Prefs> {
  vi.resetModules()
  return import('./prefs')
}

beforeEach(() => localStorage.clear())

describe('reading preferences', () => {
  it('starts at the shipped defaults', async () => {
    const m = await fresh()
    expect(m.isDefaultReadingPrefs(m.readingPrefs.value)).toBe(true)
  })

  it('does not alias the defaults constant', async () => {
    // The regression this guards: `loadPrefs` returned DEFAULT_PREFS by
    // reference, so `readingPrefs.value` *was* the constant. Changing the
    // font size rewrote the defaults, and every later comparison against
    // them compared the object to itself.
    const m = await fresh()
    m.readingPrefs.value.fontSize = 1.5

    expect(m.DEFAULT_READING_PREFS.fontSize).toBe(1.125)
    expect(m.isDefaultReadingPrefs(m.readingPrefs.value)).toBe(false)
  })

  it('restores every field on reset', async () => {
    const m = await fresh()
    m.readingPrefs.value.fontSize = m.MAX_FONT_SIZE
    m.readingPrefs.value.lineHeight = m.MAX_LINE_HEIGHT
    m.readingPrefs.value.serif = false
    m.readingPrefs.value.highlighting = false

    m.resetReadingPrefs()
    expect(m.isDefaultReadingPrefs(m.readingPrefs.value)).toBe(true)
    expect(m.readingPrefs.value.serif).toBe(true)
  })

  it('clamps an out-of-range stored value on read', async () => {
    localStorage.setItem(
      'sermokey:reading-prefs',
      JSON.stringify({ v: 2, d: { fontSize: 99, lineHeight: 99, serif: true, highlighting: true } }),
    )
    const m = await fresh()
    expect(m.readingPrefs.value.fontSize).toBe(m.MAX_FONT_SIZE)
    expect(m.readingPrefs.value.lineHeight).toBe(m.MAX_LINE_HEIGHT)
  })
})
