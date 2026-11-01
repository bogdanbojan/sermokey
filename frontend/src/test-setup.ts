// Node 26 exposes an experimental `localStorage` global that is inert
// unless the process was started with `--localstorage-file`, and it takes
// precedence over the one jsdom would otherwise install. The result is a
// `window` with no Storage API at all, which makes every persistence test
// fail for a reason that has nothing to do with the code under test.
//
// Installing a spec-faithful in-memory Storage removes the dependency on
// which Node and jsdom versions happen to be resolved, and gives the tests
// a seam for simulating a full disk.
class MemoryStorage implements Storage {
  private map = new Map<string, string>()

  get length(): number {
    return this.map.size
  }

  key(index: number): string | null {
    return [...this.map.keys()][index] ?? null
  }

  getItem(key: string): string | null {
    return this.map.has(key) ? this.map.get(key)! : null
  }

  setItem(key: string, value: string): void {
    this.map.set(String(key), String(value))
  }

  removeItem(key: string): void {
    this.map.delete(key)
  }

  clear(): void {
    this.map.clear()
  }
}

function install(name: 'localStorage' | 'sessionStorage') {
  const storage = new MemoryStorage()
  for (const target of [globalThis, window] as unknown as Record<string, unknown>[]) {
    Object.defineProperty(target, name, {
      value: storage,
      configurable: true,
      writable: true,
    })
  }
  // `vi.spyOn(Storage.prototype, 'setItem')` is how the tests simulate a
  // quota error, so the prototype has to be the one the instance uses.
  Object.defineProperty(globalThis, 'Storage', {
    value: MemoryStorage,
    configurable: true,
    writable: true,
  })
}

install('localStorage')
install('sessionStorage')
