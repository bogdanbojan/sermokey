import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  exportAll,
  importAll,
  isBackup,
  isRecord,
  ownedKeys,
  read,
  storageHealth,
  write,
} from './storage'

const parseObject = (u: unknown) => (isRecord(u) ? u : null)

beforeEach(() => {
  localStorage.clear()
  storageHealth.value = 'ok'
  vi.restoreAllMocks()
})

describe('read', () => {
  it('returns the fallback for a missing key', () => {
    expect(read('sermokey:absent', parseObject, { a: 1 })).toEqual({ a: 1 })
  })

  it('round-trips a written value', () => {
    write('sermokey:x', { hello: 'world' })
    expect(read('sermokey:x', parseObject, {})).toEqual({ hello: 'world' })
  })

  it('survives the literal string "null" instead of crashing later', () => {
    // The regression this guards. `JSON.parse('null')` returns null and
    // does *not* throw, so the old `catch` never ran and the `as T` cast
    // handed null to callers as a non-nullable object. The first property
    // access threw, on every page load, permanently.
    localStorage.setItem('sermokey:x', 'null')
    expect(read('sermokey:x', parseObject, { safe: true })).toEqual({ safe: true })
  })

  it('rejects an array where an object is expected', () => {
    localStorage.setItem('sermokey:x', '[1,2,3]')
    expect(read('sermokey:x', parseObject, { safe: true })).toEqual({ safe: true })
  })

  it('survives truncated JSON', () => {
    localStorage.setItem('sermokey:x', '{"a":')
    expect(read('sermokey:x', parseObject, { safe: true })).toEqual({ safe: true })
  })

  it('survives a parser that throws', () => {
    localStorage.setItem('sermokey:x', '{"a":1}')
    const boom = () => {
      throw new Error('bad shape')
    }
    expect(read('sermokey:x', boom, { safe: true })).toEqual({ safe: true })
  })

  it('reads a pre-envelope value as version 1', () => {
    // Installs that predate versioning stored the bare value.
    localStorage.setItem('sermokey:legacy', JSON.stringify({ old: true }))
    const seen: number[] = []
    const out = read(
      'sermokey:legacy',
      parseObject,
      {},
      (data, version) => {
        seen.push(version)
        return data
      },
    )
    expect(seen).toEqual([1])
    expect(out).toEqual({ old: true })
  })

  it('does not call migrate for a current-version envelope', () => {
    write('sermokey:x', { a: 1 })
    const migrate = vi.fn((d: unknown) => d)
    read('sermokey:x', parseObject, {}, migrate)
    expect(migrate).not.toHaveBeenCalled()
  })

  it('falls back when migrate discards the payload', () => {
    localStorage.setItem('sermokey:x', JSON.stringify({ old: true }))
    expect(read('sermokey:x', parseObject, { safe: true }, () => null)).toEqual({ safe: true })
  })
})

describe('write', () => {
  it('reports ok and leaves health alone on success', () => {
    expect(write('sermokey:x', { a: 1 })).toBe('ok')
    expect(storageHealth.value).toBe('ok')
  })

  it('reports quota and flags health when the disk is full', () => {
    // The regression this guards: this used to be a console.warn, so the
    // UI advanced to the next card as though the review had been saved and
    // the learner lost the whole session on reload with no warning.
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new DOMException('full', 'QuotaExceededError')
    })
    expect(write('sermokey:x', { a: 1 })).toBe('quota')
    expect(storageHealth.value).toBe('quota')
  })

  it('distinguishes an unavailable store from a full one', () => {
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('denied')
    })
    expect(write('sermokey:x', { a: 1 })).toBe('unavailable')
    expect(storageHealth.value).toBe('unavailable')
  })

  it('recovers health once a write succeeds again', () => {
    storageHealth.value = 'quota'
    write('sermokey:x', { a: 1 })
    expect(storageHealth.value).toBe('ok')
  })
})

describe('backup', () => {
  it('captures only sermokey keys', () => {
    write('sermokey:srs:fr', { card: 1 })
    write('sermokey:saved:es', { word: 2 })
    localStorage.setItem('unrelated', 'x')

    const backup = exportAll()
    expect(Object.keys(backup.data).sort()).toEqual(['sermokey:saved:es', 'sermokey:srs:fr'])
    expect(isBackup(backup)).toBe(true)
  })

  it('round-trips through export and import', () => {
    write('sermokey:srs:fr', { a: 1 })
    const backup = exportAll()
    localStorage.clear()
    expect(ownedKeys()).toEqual([])

    importAll(backup)
    expect(read('sermokey:srs:fr', parseObject, {})).toEqual({ a: 1 })
  })

  it('preserves corrupt values verbatim so a partial deck is still recoverable', () => {
    localStorage.setItem('sermokey:srs:fr', 'null')
    expect(exportAll().data['sermokey:srs:fr']).toBe('null')
  })

  it('refuses to import keys outside the sermokey namespace', () => {
    importAll({
      app: 'sermokey',
      version: 2,
      exportedAt: '',
      data: { 'evil:key': 'x', 'sermokey:ok': '1' },
    })
    expect(localStorage.getItem('evil:key')).toBeNull()
    expect(localStorage.getItem('sermokey:ok')).toBe('1')
  })

  it('rejects a file that is not a sermokey backup', () => {
    expect(isBackup({ app: 'other', data: {} })).toBe(false)
    expect(isBackup(null)).toBe(false)
    expect(isBackup({ app: 'sermokey' })).toBe(false)
  })
})

describe('isRecord', () => {
  it('rejects the values that caused the original crashes', () => {
    expect(isRecord(null)).toBe(false)
    expect(isRecord([])).toBe(false)
    expect(isRecord('s')).toBe(false)
    expect(isRecord(undefined)).toBe(false)
    expect(isRecord({})).toBe(true)
  })
})
