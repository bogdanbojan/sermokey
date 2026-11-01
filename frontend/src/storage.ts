// Every byte of user state in Sermokey lives in localStorage: review
// history, saved words, known/ignored sets, reading positions. There is no
// account and no server copy, so this module is the only thing standing
// between a learner and the loss of months of work. It exists because the
// per-store `JSON.parse` + `as T` + `catch {}` pattern it replaces had
// three independent silent-failure modes:
//
//  1. `JSON.parse('null')` does not throw. It returns `null`, the `as T`
//     cast launders it into a non-nullable type, and the first property
//     access crashes the app — permanently, since the bad value is re-read
//     on every boot. Shape validation, not a cast, is the fix: `read` takes
//     a parser that must actively vouch for the value.
//
//  2. A `QuotaExceededError` was caught and written to `console.warn`. The
//     UI carried on as though the write had landed, so a learner could
//     review for an hour and lose all of it on reload with no signal. Now
//     `write` reports the outcome and `storageHealth` makes it visible.
//
//  3. Two tabs each held their own in-memory snapshot and wrote the whole
//     object back, so the stale tab silently destroyed the other's work.
//     `subscribe` listens for the `storage` event and lets stores reload.
//
// Values are wrapped in a `{ v, d }` envelope so the shape can be migrated
// later. Pre-envelope values are read as version 1 and re-wrapped on the
// next write, which keeps existing installs working.
import { ref } from 'vue'

/** Current envelope version. Bump when a stored shape changes incompatibly. */
export const STORAGE_VERSION = 2

interface Envelope<T> {
  v: number
  d: T
}

export type WriteResult = 'ok' | 'quota' | 'unavailable'

/**
 * Whether persistence is working. Watched by the app shell so a learner
 * finds out *before* losing a session rather than after.
 */
export const storageHealth = ref<WriteResult>('ok')

/** Narrowing helper for parsers: a plain JSON object, not null or an array. */
export function isRecord(u: unknown): u is Record<string, unknown> {
  return typeof u === 'object' && u !== null && !Array.isArray(u)
}

function isEnvelope(u: unknown): u is Envelope<unknown> {
  return isRecord(u) && typeof u['v'] === 'number' && 'd' in u
}

/**
 * Reads and validates one key.
 *
 * `parse` receives the raw decoded value and must return a valid `T` or
 * `null`. Returning `null` — or anything throwing, or the key being absent
 * — yields `fallback`. Corrupt data therefore degrades to a usable default
 * instead of crashing the app.
 *
 * `migrate` is given the payload of an older envelope and may upgrade it;
 * returning `null` discards it. Legacy values written before envelopes
 * existed arrive here as version 1.
 */
export function read<T>(
  key: string,
  parse: (u: unknown) => T | null,
  fallback: T,
  migrate?: (data: unknown, version: number) => unknown,
): T {
  let raw: string | null
  try {
    raw = localStorage.getItem(key)
  } catch {
    // Safari private mode and blocked-storage settings throw on access,
    // not just on write.
    storageHealth.value = 'unavailable'
    return fallback
  }
  if (raw === null) return fallback

  let decoded: unknown
  try {
    decoded = JSON.parse(raw)
  } catch {
    return fallback
  }

  let version = 1
  let payload: unknown = decoded
  if (isEnvelope(decoded)) {
    version = decoded.v
    payload = decoded.d
  }

  if (version !== STORAGE_VERSION && migrate) {
    payload = migrate(payload, version)
    if (payload === null) return fallback
  }

  try {
    const parsed = parse(payload)
    return parsed === null ? fallback : parsed
  } catch {
    return fallback
  }
}

/**
 * Writes one key, reporting whether it actually persisted.
 *
 * Callers that are about to discard the in-memory original — advancing to
 * the next flashcard, for instance — must check the result. `storageHealth`
 * is updated as a side effect so the shell can warn even when an individual
 * caller does not care.
 */
export function write(key: string, value: unknown): WriteResult {
  const envelope: Envelope<unknown> = { v: STORAGE_VERSION, d: value }
  try {
    localStorage.setItem(key, JSON.stringify(envelope))
    if (storageHealth.value !== 'ok') storageHealth.value = 'ok'
    return 'ok'
  } catch (e) {
    // DOMException name is the only reliable discriminator; the code is
    // 22 in most browsers but 1014 in Firefox.
    const quota =
      e instanceof DOMException &&
      (e.name === 'QuotaExceededError' || e.name === 'NS_ERROR_DOM_QUOTA_REACHED')
    const result: WriteResult = quota ? 'quota' : 'unavailable'
    storageHealth.value = result
    console.error(`sermokey: could not persist "${key}" (${result})`, e)
    return result
  }
}

export function remove(key: string): void {
  try {
    localStorage.removeItem(key)
  } catch {
    /* nothing useful to do; the value simply stays */
  }
}

/**
 * Calls `onExternalChange` when another tab writes `key`.
 *
 * The `storage` event fires only in *other* documents, so this cannot loop
 * back on our own writes. Returns an unsubscribe function.
 */
export function subscribe(key: string, onExternalChange: () => void): () => void {
  const handler = (e: StorageEvent) => {
    // A null key means the whole store was cleared.
    if (e.key === null || e.key === key) onExternalChange()
  }
  window.addEventListener('storage', handler)
  return () => window.removeEventListener('storage', handler)
}

// --- Backup --------------------------------------------------------------

/** Every key this app owns. Used for export and for the reset path. */
export function ownedKeys(): string[] {
  const keys: string[] = []
  try {
    for (let i = 0; i < localStorage.length; i++) {
      const k = localStorage.key(i)
      if (k && k.startsWith('sermokey:')) keys.push(k)
    }
  } catch {
    /* unreadable storage exports as empty */
  }
  return keys.sort()
}

export interface Backup {
  app: 'sermokey'
  version: number
  exportedAt: string
  data: Record<string, string>
}

/**
 * Snapshots all Sermokey state as a plain object.
 *
 * Raw strings are copied verbatim rather than decoded and re-encoded, so a
 * backup of partially-corrupt storage still round-trips and a learner can
 * recover what survived.
 */
export function exportAll(): Backup {
  const data: Record<string, string> = {}
  for (const k of ownedKeys()) {
    try {
      const v = localStorage.getItem(k)
      if (v !== null) data[k] = v
    } catch {
      /* skip unreadable key */
    }
  }
  return { app: 'sermokey', version: STORAGE_VERSION, exportedAt: new Date().toISOString(), data }
}

export function isBackup(u: unknown): u is Backup {
  return isRecord(u) && u['app'] === 'sermokey' && isRecord(u['data'])
}

/**
 * Restores a backup, replacing all current Sermokey state.
 *
 * Returns the number of keys written. The caller is expected to reload the
 * page afterwards: the module-level store refs were populated at import
 * time and have no way to know their backing keys just changed.
 */
export function importAll(backup: Backup): number {
  for (const k of ownedKeys()) remove(k)
  let written = 0
  for (const [k, v] of Object.entries(backup.data)) {
    if (!k.startsWith('sermokey:')) continue
    try {
      localStorage.setItem(k, v)
      written++
    } catch {
      storageHealth.value = 'quota'
    }
  }
  return written
}

/** Deletes all Sermokey state. The recovery screen's last resort. */
export function clearAll(): void {
  for (const k of ownedKeys()) remove(k)
}
