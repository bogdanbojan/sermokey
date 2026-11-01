# Changes

Nothing here is committed. Backend: `go test ./...` green, `go vet` clean, `-race` clean.
Frontend: 75 tests across 5 files (there were none), `vue-tsc` clean, production build green.

Two rounds of work: the audit fixes, then the reader-controls and decks
round at the top of this document.

Screenshots are in `assets/changes/`.

---

## Contents

00. [EPUB truncation](#00-epub-truncation) — a regression I introduced, and the fix
0. [Reader controls and decks](#0-reader-controls-and-decks) — the second round of work
1. [SRS scheduling](#1-srs-scheduling) — three bugs that silently corrupted or crashed review history
2. [Persistence](#2-persistence) — three independent silent-data-loss paths
3. [Broken flows](#3-broken-flows) — blank screens, empty-as-error, page-wiping errors
4. [Keyboard and screen reader](#4-keyboard-and-screen-reader) — what the comments claimed vs. what shipped
5. [Mobile](#5-mobile)
6. [Backend hardening](#6-backend-hardening)
7. [Analyzer correctness](#7-analyzer-correctness)
8. [Performance](#8-performance)
9. [Interface language](#9-interface-language)
10. [Test infrastructure](#10-test-infrastructure)
11. [Things I changed my mind about](#11-things-i-changed-my-mind-about)
12. [Not done](#12-not-done)



---

## 00. EPUB truncation

**A regression I introduced in the zip-bomb work, reported as: "only part of
the book is imported, in 3 parts".**

### What went wrong

The decompression budget charged each read its **allowance** rather than
the bytes it actually read:

```go
allowance := bud.take(maxEntryBytes)   // spends 32 MB, every time
...
return io.ReadAll(io.LimitReader(rc, allowance))
```

`take` subtracts what it hands out, so every entry cost 32 MB of a 256 MB
budget no matter how small it was. That is **eight reads** in total — and
three of them go to `container.xml`, the package document and the table of
contents before a single chapter is touched. Every book truncated to about
five chapters, regardless of its size.

Reproduced before fixing, with a 40 KB synthetic novel:

```
$ go test -run TestFullBookImportsEveryChapter
    archive is 40 KB for 120 chapters
    imported 6 of 120 chapters          ← the bug
--- FAIL
```

### The fix

Reserving and charging are now separate operations: `reserve` says how many
bytes the next read *may* consume without spending anything, and `charge`
records what was genuinely decompressed.

```go
allowance := bud.reserve(maxEntryBytes)
...
data, err := io.ReadAll(io.LimitReader(rc, allowance))
// Charged even on error: the bytes were decompressed either way, which is
// the cost the budget exists to bound.
bud.charge(int64(len(data)))
```

### Verified

| | Before | After |
|---|---|---|
| 120-chapter book (40 KB) | 6 sections | **120** |
| 360-chapter book (283 KB, 4.5M runes) | ~6 | **360** |
| Zip bomb: 400 × 10 MB of markup yielding no paragraphs | rejected | **still rejected** |

End to end against the running server, with a 90-chapter EPUB:

```
POST /api/texts/import  ->  "sectionCount": 90
GET  .../?section=89    ->  "Chapitre 90", 40 paragraphs
GET  .../?section=45    ->  "Chapitre 46", 40 paragraphs
```

Three regression tests now cover it: a full book, a novel-sized book
(which also guards the rune and spine ceilings), and the bomb.

**Why it slipped through.** My original hardening test asserted only that a
*hostile* archive is rejected. It never asserted that a legitimate one
survives — so the budget could be arbitrarily too aggressive and still
pass. A limit needs a test on both sides of it.

---

## 0a. Deck organisation

Two things you noticed were the same problem: the mixed-review card's pill
duplicated its own title, and of three cards in a row only two carried the
accent border.

### The accent border was carrying meaning it could not carry

`.vocab-card.mixed` and `.saved-card` each set a 2px accent border;
`.deck-card` did not. Three cards side by side, two highlighted, no stated
rule — which is exactly how it read.

The underlying issue is information architecture, not CSS. The Vocabulaire
tab had come to hold two different kinds of thing in one undifferentiated
grid:

- **The curated curriculum** — fixed, authored, ~2,200 words
- **The reader's own collections** — grows with every page read

**What changed:** the grouping is now structural and labelled, so the
accent can carry a single rule — *the "everything" card of its group*.

```
Mes paquets
  [Mes mots ●]  [Marché ×]  [Le marché du samedi]

Vocabulaire
  [Révision mixte ●]
  ▸ Les bases (4)   ▸ Les gens (6)  …
```

Exactly one accented card per group, and the question "why is that one not
highlighted?" no longer arises.

### Should custom decks exist?

Yes — but they needed a visible *kind*, and ordering that reflects intent:

| Card | Kind pill | Forgettable | Order |
|---|---|---|---|
| Mes mots | Tous mes mots | no | first — the aggregate |
| A deck you made | Mes paquets | **yes** | before text decks |
| A text's deck | Textes | no | last |

Decks you made sort ahead of text decks because creating one was a
deliberate act, where a text's deck appears merely because a word happened
to be saved there. And only a deck you made can be forgotten — a text's
deck is a fact about where words came from, so "forget" on it would either
be a no-op or would have to discard the words' context.

### The pill duplication

`VocabCard` rendered `isMixed ? t.vocab.mixedTitle : nameEn` as the pill
and `nameNative` as the heading — and the mixed card is passed
`t.vocab.mixedTitle` for *both*. So it read "RÉVISION MIXTE / Révision
mixte" while every other card pairs a classifier with a different name
("NUMBERS / Nombres").

A new `vocab.mixedPill` key fixes it: `TOUT LE VOCABULAIRE / Révision
mixte`. That also removed the hardcoded English `name-en="All vocabulary"`
the audit had flagged.

### Deleting a book leaves its deck behind, on purpose

A deck named after a book outlived the book with no way to remove it, so it
sat there permanently. But the fix must not swing the other way: deleting a
book must never silently delete a grouping the reader can still see.

**The rule now:** a text deck becomes *forgettable* when its text is gone —
and nothing is ever removed without being asked for.

Making that true needed two changes that pull against each other:

1. **Derivation is scoped to texts still in the library.** A text deck is
   normally rebuilt from its words' `context.textId` on every render. That
   is what made it un-removable: forgetting it deleted the stored entry and
   the next render recreated it. Scoping derivation to live texts means a
   forgotten orphan stays forgotten.

2. **The deck is materialised when the book is deleted.** Step 1 alone
   introduces the opposite bug. A text deck that was only ever *derived* —
   because its words went to a custom deck instead — would vanish along
   with the book, which is exactly the silent deletion to avoid. So the
   delete handler stores the deck first, turning it into an orphan the
   reader keeps for as long as they like.

The confirmation says so plainly, rather than leaving them to find out:

> Retirer « Le Petit Prince » de votre bibliothèque ?
>
> Vos 3 mots enregistrés et leur paquet sont conservés.

And the orphan's pill explains why it alone carries a remove control:

| State | Pill | × |
|---|---|---|
| Text still in the library | `TEXTES` | no — removing it would be a visible no-op |
| Text deleted | `TEXTE SUPPRIMÉ` | **yes** |
| A deck you made | `MES PAQUETS` | **yes** |

Verified end to end — save three words (two to the text's deck, one to a
custom deck), delete the book, then forget the orphan deliberately:

```
1. while the book exists   Mes mots 3 · Vocabulaire du renard 1 · Le Petit Prince 3
2. after deleting the book Mes mots 3 · Vocabulaire du renard 1 · Le Petit Prince 3  (TEXTE SUPPRIMÉ, ×)
3. after forgetting it     Mes mots 3 · Vocabulaire du renard 1
   words still in Mes mots: 3
```

A custom deck is never affected by deleting a book, whatever it is named.
Four tests cover the lifecycle.

![An orphaned deck](changes/decks-orphaned.png)

### Incidental cleanup

The saved-words card and the deck card had grown into near-identical markup
with separately-maintained styles — which is how two things that should
look alike stop looking alike. Both are now one `DeckCard` component, so
the three kinds cannot drift. `margin-top: auto` on the stats row also
aligns the buttons across cards whether or not a card has a blurb.

---

## 0. Reader controls and decks

A second round, prompted by two questions: are six display controls all
needed, and how does a word get into a *named* deck?

### 0.1 Six toolbar buttons became one `Aa` menu

**Before** — `A− A+ ↕− ↕+ Serif H`, six 32px buttons permanently in the
toolbar of the one screen whose whole job is to get out of the way. Two of
them were mine, from the previous round, which is how the row got to six.
`H` was the keyboard shortcut masquerading as a label and meant nothing on
sight.

**After** — one `Aa` button opening a panel. That is the pattern every
long-form reader converges on (Kindle, Apple Books, Instapaper, Pocket),
for two reasons: these are set-once preferences that do not deserve
permanent chrome, and a row of six icon buttons is neither readable nor
tappable on a phone.

Inside, each control gets a real label — `Text size`, `Line spacing`,
`Typeface`, `Highlight words` — and the steppers *disable at their limits*,
so "this does nothing" is distinguishable from "this is broken". On a phone
the panel docks to the screen edges rather than hanging off a 44px trigger.

| | Before | After |
|---|---|---|
| Toolbar controls | 6 | 1 |
| Smallest target | ~32×26 | 44×40 |
| Highlight control | `H` | labelled switch, `H` shown as the shortcut |
| At a stepper's limit | silently clamps | disabled |

![Reader display menu](changes/reader-display-menu.png)

On a phone the panel docks to the screen edges, where a 16rem popover hung
off a 44px trigger would have overflowed:

![Reader display menu at 375px](changes/reader-display-menu-375.png)

### 0.2 Reset

You were right that it is marginal — you *can* tap back. But only if you
remember where you started, and after a few taps nobody does. As a seventh
toolbar button it was not worth the clutter; as one line at the bottom of a
panel it costs nothing and answers a question the steppers cannot. It is
disabled while the settings are already at their defaults, so it also
doubles as an indicator of whether you have changed anything.

### 0.3 A latent bug that Reset exposed

Building it surfaced a real aliasing bug that predates this work.
`loadPrefs()` returned the `DEFAULT_PREFS` **object itself** as its
fallback, so `readingPrefs.value` *was* that constant. The first
`readingPrefs.value.fontSize = …` rewrote the shipped defaults for the rest
of the session.

Nothing had ever compared against them, so it was invisible — until Reset
needed to know what "default" meant, and found the mutated value. The
symptom was a Reset button that stayed disabled after changing the font
size, caught by driving the real UI rather than by reading the code.

`DEFAULT_PREFS` is now frozen, so a recurrence throws rather than corrupting
state quietly, and `loadPrefs` spreads it. Four tests cover it.

### 0.4 Decks

The design question is not "how do we add a picker" — it is **where the
picker must not be**. Saving a word is the highest-frequency action in the
reader, and a dialog in front of it taxes every single one. So:

- **Save stays one tap.** The button is unchanged.
- **The destination is shown, not asked.** A line above the button reads
  `Add to  La Parure ▾`, so a word is never filed somewhere invisible.
- **It defaults to the text being read.** That is the "deck named after the
  text" — automatic, not something to set up first.
- **The choice sticks.** A reader working through a book wants that book's
  deck for every word in it, not once per word. Verified: after choosing
  "Maupassant", the next word's destination still reads "Maupassant".
- **New decks are created inline**, in the same list, with no trip to a
  settings screen.

![Deck picker](changes/deck-picker-open.png)

**The model matters more than the UI.** A deck is a *view over saved words*,
not a separate store:

- Every saved word still lives in "My words". A deck adds membership.
- Deleting a deck therefore **cannot lose words** — which is why forgetting
  one needs no scary warning, and why the confirm says so explicitly.
- A word can sit in several decks without being duplicated, so its SRS card
  id stays single-valued and its review history is never forked.
- A text deck is *derivable*: "I met this word in La Parure" is already
  recorded in the word's context. So words saved before decks existed show
  up in their text's deck with no migration, and a word filed into a custom
  deck still appears under the text it came from.

That last point is visible on Home — the same two words produce both cards:

```
Textes       La Parure (début)   2 mots   (no forget button)
Mes paquets  Maupassant          1 mot    (forgettable)
```

Only a deck *you made* can be forgotten. A text's deck is a fact about where
words came from, not a grouping anyone chose, so a "forget" on it would
either be a no-op or would have to discard the words' context.

![Decks on the home screen](changes/home-decks.png)

### 0.5 Plurals

The new count string is the one in the app that reads `1 mot` rather than
`1 mots`. `strings.ts` otherwise hardcodes plurals, which the audit flagged;
reshaping a dozen keys across four languages is its own job, but a string I
was adding should at least start correct. `plural()` uses `Intl.PluralRules`
rather than `count === 1`, so a language with more than two categories needs
only its forms added. Tested across all four.

### 0.6 What I did not add

- **Multi-deck selection at save time.** A word *can* be in several decks,
  but offering that in the picker would reintroduce the friction the whole
  design avoids. Single-select at save; add to more decks later.
- **Rename.** `rename()` exists in the store and nothing calls it yet. The
  create-time name is almost always right, and a rename affordance on every
  card would compete with the forget button for the same corner.
- **Drag between decks.** Not without a deck-detail screen to drag within.

---

## 1. SRS scheduling

`srs/sm2.ts` is 113 lines of pure functions holding the entire value of the product. It had no tests. Three of its bugs were reproducible in a few lines.

I extracted the original logic verbatim into a scratch script before touching anything, so the "before" column here is measured, not inferred:

```
$ TZ=America/New_York node oldsm2.mjs
1. card graded "good" on Jan 1 -> dueDate 2026-01-02
   isDue at 20:00 on Jan 1 (should be false): true
2. threw on easy #13: RangeError Invalid time value
3. card with 3 reviews then one lapse -> bucket: new
```

### 1.1 The day boundary was UTC; the arithmetic was local

**Before** — `sm2.ts:44-52`:

```ts
export function toISODate(d: Date): string {
  return d.toISOString().slice(0, 10)        // UTC day
}
function addDays(date: Date, days: number): Date {
  const result = new Date(date)
  result.setDate(result.getDate() + Math.round(days))   // local day
  return result
}
```

Mixing the two gives a day boundary belonging to neither calendar. In New York, at 20:00 on Jan 1 `toISODate` already reports Jan 2, so `isDue` returns true for everything scheduled for tomorrow. Every evening session pulled the next day's queue forward, compressing every interval in the deck by up to a day. At UTC+13 it fails the other way: the day doesn't roll over until 13:00 local.

**After** — the local calendar throughout, with DST absorbed by rounding between two local midnights:

```ts
export function toISODate(d: Date): string {
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function daysBetween(fromISO: string, to: Date): number {
  const from = fromISODate(fromISO)
  if (!from) return 0
  return Math.round((localMidnight(to).getTime() - from.getTime()) / 86_400_000)
}
```

Tests run in four timezones:

```
$ for tz in America/New_York Pacific/Auckland UTC Europe/Paris; do TZ=$tz npx vitest run; done
   30 passed  (×4)
```

**Migration:** existing `dueDate`s are left alone. They cannot be recomputed — the hour and zone each was written in are not recorded — so they drift by at most a day and self-correct on the next review. That is documented in the file rather than papered over with a fake migration.

### 1.2 Ease factor had a floor but no ceiling, interval had no cap

**Before** — `sm2.ts:86-87`:

```ts
else intervalDays = Math.round(intervalDays * easeFactor * 1.3)
easeFactor = easeFactor + 0.15        // unbounded
```

Each "Easy" multiplies by the factor *and* raises it. Growth is super-exponential, the interval leaves the range of a JS `Date`, `addDays` returns an Invalid Date, and `toISOString()` throws `RangeError` straight out of the grading handler — unhandled, with no error boundary. **13 consecutive "Easy" grades.**

**After** — `MAX_INTERVAL_DAYS = 36500`, `MAX_EASE_FACTOR = 5.0`, both clamped on every write, plus a 200-random-grade fuzz test asserting the state stays finite and the date stays valid.

### 1.3 Every lapsed card was reported as new

**Before** — `review()` resets `repetitions` to 0 on a lapse (`:64`), and `bucketOf` derived the bucket from exactly that (`:109`):

```ts
if (!state || state.repetitions === 0) return 'new'
```

So a card with forty reviews behind it joined the "new" pile the moment it was failed once. That number feeds the session-complete panel and every per-category "N new" hint on Home — all of them wrong on any deck with recent lapses.

**After** — an explicit `phase: 'new' | 'learning' | 'review' | 'relearning'` stored on the card, with `relearning` existing precisely so a lapse is distinguishable from never-seen. Cards written before the field existed are reconstructed from what v1 *did* record:

```ts
// Version 1 reset `repetitions` to 0 on a lapse but left the interval
// behind, so a card with no repetitions and a non-zero interval has
// certainly been failed at least once.
if (repetitions === 0 && intervalDays === 0) phase = 'new'
else if (repetitions === 0) { phase = 'relearning'; lapses = 1 }
```

### 1.4 Anki parity (requested)

| | Before | After |
|---|---|---|
| Lapse counter | none | `lapses`, incremented only when a *graduated* card fails |
| Leeches | a failed card cycled forever | suspended at 8 lapses, withheld from the queue, surfaced with an unsuspend action |
| Failed card in-session | deferred to tomorrow | re-queued; SM-2 as published repeats until recalled |
| Overdue credit | ignored; `lastReviewed` was written and read by nothing | Anki's `delay` term — a 6-day card answered 60 days late is scheduled from the retention actually demonstrated |
| Review queue size | unbounded | `reviewLimit` (120); two weeks away no longer produces a 400-card wall |
| New-card placement | all reviews, then all new | interleaved |
| Queue order | comparator never returned `0`, so tied due dates shuffled per call | stable, reproducible |
| Clock | two `new Date()` per review | one, threaded through |
| Interval shown on buttons | no — graded blind | yes (`1j / 1j / 1j / 4j` in the screenshot) |
| Undo | none | single-step, keyboard `Z` |
| Corrupt `dueDate` | `undefined <= "2026-01-01"` is `false` → card never due again, silently, forever | treated as due |

![Flashcards with grade intervals](changes/flashcards-grades-desktop.png)

---

## 2. Persistence

Everything the learner has ever done lives in `localStorage` and nowhere else. There were three independent ways to lose it silently, and no backup path at all.

All seven stores now go through one module, `src/storage.ts`.

### 2.1 `JSON.parse` was guarded against throwing, not against the wrong shape

**Before** — `srs/store.ts:59-66`, and the same shape in `recent.ts`, `saved.ts`, `prefs.ts`, `HomeView.vue`:

```ts
const raw = localStorage.getItem(storageKey(lang))
return raw ? (JSON.parse(raw) as Deck) : {}
```

If the value is the four characters `null`: `raw` is truthy, so the ternary takes the parse branch; `JSON.parse('null')` returns `null` and **does not throw**, so the `catch` never runs; the `as Deck` cast launders it into a non-nullable type. The first `deck.value[cardId]` throws `TypeError` — on every page load, permanently, because the bad value is re-read at boot.

**After** — `read()` takes a parser that must actively vouch for the value, and a malformed *entry* costs that entry rather than the whole deck:

```ts
function parseDeck(u: unknown): Deck | null {
  if (!isRecord(u)) return null
  const deck: Deck = {}
  for (const [id, raw] of Object.entries(u)) {
    if (!isRecord(raw)) continue
    if (typeof raw['easeFactor'] !== 'number' || typeof raw['intervalDays'] !== 'number') continue
    deck[id] = normalizeState(raw as unknown as CardState)
  }
  return deck
}
```

Covered by tests for `"null"`, arrays, truncated JSON, a throwing parser, and mixed valid/corrupt decks.

### 2.2 Quota failures were a `console.warn`

**Before** — `srs/store.ts:68-74`:

```ts
try { localStorage.setItem(storageKey(lang), JSON.stringify(deck)) }
catch (e) { console.warn('Could not persist flashcard progress to localStorage:', e) }
```

The UI advanced to the next card as though the review had landed. Study for an hour, reload, lose all of it, with no signal at any point.

**After** — `write()` returns `'ok' | 'quota' | 'unavailable'`, and the caller is required to care:

```ts
const { previous, saved } = deck.value.recordReview(cardId, g)
// A review that did not persist has not happened. Advancing anyway is
// how an hour of study disappears on the next reload with no warning.
if (saved !== 'ok') {
  saveWarning.value = t.value.vocab.saveFailed
  return
}
```

Plus a `storageHealth` ref driving a persistent banner in the shell.

### 2.3 Two tabs destroyed each other's work

**Before** — each store was a module-level ref loaded once, and every write replaced the whole object:

```ts
deck.value = { ...deck.value, [cardId]: next }
saveDeck(lang, deck.value)          // whole deck
```

No `storage` listener existed anywhere. Tab A open since morning, Tab B reviews 80 cards, Tab A grades one → Tab A writes its morning snapshot plus one card. **80 reviews gone.**

**After** — every store subscribes and reloads on an external write:

```ts
subscribe(storageKey(lang), () => {
  existing!.value = loadDeck(lang)
})
```

### 2.4 No backup, no versioning, no recovery

**Before:** no export, no import, no schema version on any key, and `main.ts` was five lines with no `errorHandler` — so §2.1's crash was a permanent white page.

**After:**
- A `{ v, d }` envelope on every value; pre-envelope values are read as version 1 and re-wrapped.
- Export/import of all `sermokey:*` keys as JSON. Corrupt values are copied **verbatim** so a partially-damaged deck still round-trips.
- An `onErrorCaptured` boundary in `App.vue` whose recovery screen leads with the backup, not the reset — whatever is still readable is worth more than a working app.

### 2.5 Everything else that touched storage

| File | Before | After |
|---|---|---|
| `i18n/language.ts:10` | unguarded `getItem` at **module scope** — throws in Safari private mode before Vue mounts | guarded |
| `srs/store.ts:32,38` | unguarded | guarded |
| `HomeView.vue:65,69,170` | unguarded; the quota throw was inside a click handler | guarded |
| `ReaderView.vue:314,317` | unguarded | guarded |
| `prefs.ts:105` | `ReturnType<typeof ref<T>>` resolved to `Ref<T \| undefined>`, hence `?? {}` and `?.` everywhere | `Ref<LookupCounts>` |

`grep -c localStorage` across the four store files: **0**.

---

## 3. Broken flows

### 3.1 "Study anyway" rendered a blank page

**Before** — `FlashcardsView.vue:112-120`:

```ts
const newCardLimit = includeNotDue ? allCards.length : 15
queue.value = deck.value.sessionQueue(allCards, newCardLimit)
if (queue.value.length === 0 && !includeNotDue) phase.value = 'empty'
else phase.value = 'studying'
```

But `sessionQueue` only ever emitted **due** cards plus **never-seen** cards — not-due cards were dropped at `store.ts:123`. `newCardLimit` widens only the new-card slice. So on a mature deck with nothing due: queue is `[]`, `includeNotDue` is true → `phase = 'studying'` → but the template guard is `phase === 'studying' && current`, and `current` is `undefined`. **No branch matched. Blank page.** That button is the one on the empty state, so it was the common path.

**After** — `sessionQueue` gained an explicit `includeNotDue` option that actually admits them, the phase is derived from the queue rather than the flag, and there is a catch-all `v-else` so no combination of conditions can render nothing:

```ts
phase.value = queue.value.length === 0 ? 'empty' : 'studying'
```

Test: `sessionQueue(cards)` → 0; `sessionQueue(cards, { includeNotDue: true })` → 2.

### 3.2 An empty deck was reported as a load failure

**Before** — `FlashcardsView.vue:106-110` set `phase = 'error'` with *"Could not load the vocabulary."* That is what a new user saw on `#/flashcards/saved` — the one deck that is supposed to start empty.

**After** — `phase = 'empty'`, with copy and an action appropriate to "you have not saved anything yet" rather than to "nothing is due".

![Empty saved deck](changes/flashcards-empty-saved-deck.png)

### 3.3 A failed delete wiped the entire Home page

**Before** — `HomeView.vue:87` wrote to the page-level `error` ref, and the template renders `v-else-if="error"` *before* `<template v-else>` containing everything. A failed `DELETE` replaced the tabs, the search box, the library and all grammar and vocabulary content with one line of red text, with no retry.

**After** — a separate `actionError`, rendered as a dismissible inline banner beside the library. And every one of the six error states in the app now has a retry button; `load()` was already idempotent in all of them.

### 3.4 Deleting a text left orphans

**Before** — the reading position and the "continue" chip survived. The chip then matched no text, fell through to the vocabulary branch of `goToRecent`, and navigated to a category that does not exist. Dead end, back button only.

**After** — `forgetPosition()` and `recents.forget(id)` on delete.

### 3.5 Out-of-order responses

**Before** — none of the six `load()` functions had an `AbortController`, a sequence guard, or `onCleanup`. Arrowing through sections started several fetches and the last to resolve won — section 2's body could render while the URL said section 5. Worse, `ReaderView.vue:72` then read `props.section` *after* the await and saved that as the reading position, recording a section the learner was not looking at.

**After** — a sequence token in every `load()`, and the request parameters captured *before* the await:

```ts
const requested = { lang: lang.value, textId: props.textId, section: props.section }
const fetched = await api.getTextSection(requested.lang, requested.textId, requested.section)
if (token !== loadToken) return
```

### 3.6 Language switch flashed a 404

**Before** — `setLanguage` set `lang.value` and *then* navigated. `hashchange` is delivered as a task, not a microtask, so the pre-flush watcher on `lang` refetched the current route under the new language first — a guaranteed 404 on `/api/theory/{french-id}?lang=es`, shown as a red error, before the route changed.

**After** — navigate first, then set the language.

### 3.7 Other

| Issue | Before | After |
|---|---|---|
| `WordPanel` key — `ReaderView.vue:427` | `entry.surface + token`, so `"est"` at token 7 of two different paragraphs collided *and* the surface watcher didn't fire; `chosenSense` stayed stale and Save became a silent no-op | keyed on `paragraph:token:wordId` |
| `lookupSelection` catch — `ReaderView.vue:243` | cleanup sat **outside** the `try`, so a network failure wiped the selection and showed nothing | cleanup inside; error surfaced |
| Quiz count — `router.ts:48` | `Number('')` is `0` and `isFinite(0)` is true → a zero-question quiz; no upper bound either | clamped to `[1, 50]` |
| `parseHash` — `router.ts:94` | `decodeURIComponent` throws `URIError` on a malformed escape, at **module scope**, before Vue mounts | `safeDecode` |
| `/api/languages` — `language.ts:34` | an unknown code set `lang`, made `t` undefined, threw on the next render, and **persisted the bad code** — blank on every subsequent load | filtered against the strings this build actually has |
| `api/client.ts:28-49` | error path tolerated a non-JSON body, success path did not — a 200 with an empty body threw a bare `SyntaxError` every caller then misreported | both tolerate it |
| Enter-to-advance — `QuizView.vue:108` | the branch was unreachable: the input unmounts on answer, so Enter reached no handler | moved to the window |
| MCQ keyboard | none at all, while flashcards had 1-4 | 1-4 |
| Progress bars | `index / total` — 0% on the first card, never 100% | `(index + 1) / total` |

### 3.8 A misleading error, caught live

While taking screenshots I ran Vite on :5199 against a backend allowing :5173 and got:

> Could not reach the backend at http://localhost:8080. Is it running? (cd backend && go run ./cmd/server)

The backend was running. `fetch` rejects identically for connection-refused and a CORS rejection, and `catch {}` discarded the cause. The message sent me to restart a server that was already up — which is the single most likely cause in development.

**After:**

```ts
`Could not reach the backend at ${BASE_URL}. It may not be running, ` +
  `or it may be refusing this origin (${window.location.origin}) via CORS — ` +
  `set FRONTEND_ORIGIN on the backend to match.`
```

with `err.cause` preserved. `getJSON` was also a near-duplicate of `request` that had already drifted (it never grew the 204 handling); it is now the same code path.

---

## 4. Keyboard and screen reader

The sharpest problem in the codebase was not a bug class — it was comments describing behaviour the code did not have.

### 4.1 The reader was pointer-only

`ReaderView.vue:262` said:

> It also makes the feature operable without a pointer, rather than a mouse-only toy.

Tokens were inert `<span>`s with no `tabindex` and no key handler. A keyboard or screen-reader user could not look up a single word — the central interaction. The hover popup is `aria-hidden="true"` on the stated grounds that the pinned panel is the accessible route to the same information, but there was no keyboard route to the pinned panel either.

**After** — tokens that will actually respond are focusable and activatable:

```html
<span
  v-if="tk.word"
  class="tok"
  :tabindex="isActionable(tk.entry) ? 0 : undefined"
  :role="isActionable(tk.entry) ? 'button' : undefined"
>
```

with Enter/Space handled on the article, so the handler count doesn't scale with chapter length. Tokens with no entry (proper nouns, numerals) are no longer focusable *and* no longer get `cursor: pointer` — they never did anything.

![Keyboard-opened word panel](changes/reader-panel-focused.png)

That screenshot was produced by `Tab` then `Enter` — no pointer involved.

### 4.2 Escape was unreachable from inside the dialog

**Before** — `ReaderView.vue:267`:

```ts
const typing = (e.target as HTMLElement)?.tagName === 'INPUT'
if (typing) return                  // returns BEFORE the Escape case
```

The panel's sense selector is radio inputs. Focus one, press Escape, nothing happens — a `role="dialog"` you could not dismiss from the keyboard.

**After** — Escape handled before the typing guard, and the guard widened to `TEXTAREA` and `contenteditable`.

### 4.3 Shift+Arrow navigated away instead of extending the selection

**Before** — the handler checked `metaKey`/`ctrlKey`/`altKey` but not `shiftKey`, so Shift+Arrow paged to the next section. `SelectionToolbar.vue:12` claims phrase capture "works … with the keyboard for free" — Shift+Arrow is *the* keyboard route to a native selection, and the shortcut handler destroyed it.

**After** — `if (e.shiftKey) return`.

### 4.4 The dialog never received focus

**Before** — `WordPanel.vue:102` had `role="dialog"` and `tabindex="-1"`, and nothing ever called `.focus()`. No trap, no `aria-modal`, no focus restoration — and `.panel { outline: none }` removed the ring from the element that was supposed to be focused.

**After** — focus on mount, a Tab trap, restoration to the originating element on unmount, `aria-modal="true"`, and the focus ring kept.

### 4.5 Flashcard shortcuts were dead until you clicked the background

**Before** — `FlashcardsView.vue:200`: `<section tabindex="0" @keydown="onKeydown">`, which nothing focused, with `section { outline: none }` at `:318` hiding whether it had worked. The UI advertised `1`–`4` on the buttons. And grading unmounts the focused button, so even after clicking, focus fell to `<body>` and the shortcuts stopped after the first card.

**After** — bound to the window, as `ReaderView` already did. The screenshot in §1 was produced by pressing Space on a freshly loaded page with no prior click.

### 4.6 Focus indicators

**Before** — `grep focus-visible src/` → **zero matches** in the whole app. The only two `:focus` rules *removed* the outline (`HomeView.vue:459`, `QuizView.vue:280`), replacing it with a border-colour change — a colour-only indicator, which fails WCAG 2.4.7.

**After** — one global rule plus all three `outline: none` sites converted:

```css
:focus-visible {
  outline: 2px solid var(--color-accent);
  outline-offset: 2px;
  border-radius: var(--radius-sm);
}
```

### 4.7 Motion

**Before** — `prefers-reduced-motion` honoured in exactly one component (the popover's entrance). The accordion's `grid-template-rows` transition is the largest moving element in the app and was unguarded, as were the progress bars, coverage bars, every button-press `translateY`, and the per-token background fade.

**After** — a global reduce block in `style.css`.

### 4.8 Contrast

`--color-success: #2f8f5b` measured **4.04:1** on white and **3.60:1** on its own tinted background — below the 4.5:1 text floor. It is used as *text* for the "Correct!" quiz feedback, the Easy grade button, the comfortable difficulty band and the neuter gender badge. Darkened to `#1f7049`.

### 4.9 The status legend that was translated but never built

`reading.statusUnknown` / `statusLearning` / `statusKnown` existed, translated into all three languages, since the feature shipped. Nothing rendered them — so the three underline styles on the page meant nothing to a first-time reader. Now rendered:

![Reader legend and controls](changes/reader-legend-and-controls.png)

The underline styles themselves were already colourblind-safe (solid + tint / dotted / none), which was good work; it just had no key.

### 4.10 Other

`role="status"` on every loading state and `role="alert"` on every error (there were **no** live regions anywhere); `aria-pressed` on the direction and display toggles; `scope="col"` on theory table headers; `<html lang>` driven from the interface language instead of hardcoded `fr` — which also means German prose is no longer hyphenated by French rules via `hyphens: auto`.

---

## 5. Mobile

Measured on a 320px-wide iPhone SE viewport.

### 5.1 Grade buttons

**Before** — `repeat(4, 1fr)` with no breakpoint. At 375px that is ~70px per button, of which 35px is padding, leaving **~35px of label** for "À revoir", "Difficile", "Nochmal", "De nuevo". The two `repeat(4, 1fr)` grids strongly suggest the flashcard screen was never opened on a phone.

**After** — `repeat(auto-fit, minmax(7.5rem, 1fr))`, plus `min-height: 2.75rem` for WCAG 2.5.5:

![Grade buttons at 320px](changes/flashcards-grades-375.png)

Same for the session stats grid ("In Bearbeitung" in a quarter of 375px).

### 5.2 Theory tables forced the whole page sideways

**Before** — a 6-column table with `width: 100%` and no overflow wrapper. Person labels are `ellos / ellas / ustedes`; at 351px of content that is ~34px of text per column, so the table widened the **document**, dragging every view horizontally — the exact outcome `AppHeader.vue:44` and `HomeView.vue:411` both identify as the thing to avoid.

**After** — a `.table-scroll` wrapper. Measured:

```
375px: { "pageOverflows": false, "docScrollWidth": 320, "innerWidth": 320,
         "tableScrollsItself": true }
```

The table scrolls; the page does not.

![Theory at 375px](changes/theory-table-375.png)

### 5.3 Other

| Issue | Before | After |
|---|---|---|
| First-run coachmark | told **every** phone user to "hover a word" | branched on `(hover: hover)` |
| Bottom sheet | the Save/Known/Ignore row sat under the iOS home indicator | `env(safe-area-inset-bottom)` + `viewport-fit=cover` |
| `100vh` | mobile Safari's *largest* viewport — page extends under the collapsing URL bar | `100dvh` (and `dvh` on the sheet) |
| `SelectionToolbar` | anchored at `y - 6` and translated fully above it, so selecting on the first visible line put it off-screen; `WordPopover` clamps and flips, this didn't | clamped and flipped |
| Touch targets | close `×` ~28×22, reader controls ~32×26, toolbar button ~26px, language buttons made *smaller* by the mobile media query | ≥ 2.25–2.75rem |

---

## 6. Backend hardening

### 6.1 No request body limit on import

**Before** — `reading_import.go:62`:

```go
if err := r.ParseMultipartForm(8 << 20); err != nil {
```

`ParseMultipartForm`'s argument is *maxMemory*, **not** a body cap. Everything above 8 MB is spooled to `$TMPDIR` with no limit, and the `header.Size > MaxImportBytes` check at `:73` runs only after that has finished. Unauthenticated disk exhaustion. `grep MaxBytesReader` → no matches anywhere.

**After** — one line, plus 413 handling and temp-file cleanup:

```go
r.Body = http.MaxBytesReader(w, r.Body, MaxImportBytes+(1<<20))
```

Test: a 36 MB upload now returns 413.

I also corrected the comment at `:79`, which claimed `header.Size` is "client-supplied". It is computed by `ReadForm` from the bytes actually read. The `LimitReader` is fine as defence in depth; the stated reason was wrong and would stop someone looking at the real gap.

### 6.2 Zip bomb

**Before** — `MaxTotalRunes` only advances for entries that *yield paragraphs*:

```go
data, err := readEntry(files, full)
heading, paragraphs := extractText(data)
if len(paragraphs) == 0 { continue }        // totalRunes NOT incremented
```

2000 spine items of `<style>` padding — which `skipElements` discards — decompress **64 GB** while the counter sits at zero and the `break` never fires. There was also no cap on entry count, and `zip.NewReader` reads the central directory eagerly: a 32 MB archive can declare ~700k entries.

**After** — a shared decompressed-byte budget threaded through `readEntry`, checked in the spine loop alongside the rune count, plus `MaxZipEntries = 20_000`. Skipped chapters are now counted and logged rather than silently producing a two-chapter stub.

Also `resolve()`: `strings.TrimPrefix(joined, "../")` strips **one** `../`, so `../../../x` still escaped. Replaced with `fs.ValidPath`. Nothing is written to disk from the archive today, which is the only reason this wasn't exploitable.

### 6.3 No panic recovery

**Before** — `grep "recover()"` → **no matches**, in a codebase that panics deliberately in `registry.go`, `catalog.go`, `parse.go` and `lexdata.go`. `net/http` recovers per connection so the process survives, but the client gets a truncated response and `withLogging` never records the request, because the handler never returns.

**After** — `withRecovery` returning a JSON 500, re-panicking on `http.ErrAbortHandler`, with the stack logged.

### 6.4 403/404 chosen by substring match

**Before** — `reading_import.go:122`:

```go
if strings.Contains(err.Error(), "bundled") {
    status = http.StatusForbidden
}
```

`library.go:90` returns `"unknown text: %s"` with the caller's id interpolated. So `DELETE /api/texts/bundled-anything` returned **403 instead of 404**.

**After** — `ErrBundledText` / `ErrNoSuchText` sentinels and `errors.Is`. Verified live:

```
DELETE bundled-nonsense    -> 404
DELETE la-parure (bundled) -> 403
```

### 6.5 Timeouts killed large imports *after* they committed

**Before** — `WriteTimeout: 10 * time.Second` starts when headers are read, so it covers handler execution. A large import blows through it — but `Library.Add` and `libraryStore.Save` have *already run*. The book is imported and persisted, the learner sees a failure, retries, and pays for the work twice. `ReadTimeout: 10s` also required a sustained 3.2 MB/s for a 32 MB upload.

**After** — generous server-wide timeouts, with read-only routes bounded individually by `http.TimeoutHandler(30s)`. Also `MaxHeaderBytes`, `ErrorLog`, `errors.Is(err, http.ErrServerClosed)`, and the bind-failure path no longer calls `log.Fatalf` from inside a goroutine (which skipped shutdown entirely).

### 6.6 No caching, no compression

**Before** — `grep -i "etag|cache-control"` → no matches. The section payload is the largest in the API, fully determined by document id and index, and was re-downloaded in full on every navigation.

**After** — measured live:

```
identity: 58,536 bytes | gzip: 11,672 bytes        (5.0x)
If-None-Match -> 304, 0 bytes
```

Marshalling into a buffer before touching the header also fixed a latent bug: `writeJSON` committed the status *then* streamed, so a mid-stream encode failure left a `200` with truncated JSON.

### 6.7 A bug I introduced, then caught

Adding `Vary` broke `Vary: Origin`. `http.TimeoutHandler` copies the inner handler's headers over the outer ones with `dst[k] = vv` — a whole-slice **assignment**, not an append. Any header written by both layers loses the outer value.

Caught by checking the live response rather than trusting the code. `Vary` is now set in one place, with the reason documented, and there is a regression test.

### 6.8 Other

Import concurrency semaphore (each import holds the upload plus a built `Document` and spends minutes of CPU); `r.Context()` honoured; `%q` on the logged path (it was attacker-controlled and unescaped, so `/api/%0a%0afake` injected newlines into the log); status and byte count in the access log; `Access-Control-Max-Age` so preflights aren't paid per request.

---

## 7. Analyzer correctness

### 7.1 The tiebreak that demoted every German noun

**Before** — `analyzer.go:294`:

```go
if x.common != y.common { return x.common }
// common = isCommonSpelling(lemma) = (lemma == strings.ToLower(lemma))
```

The rule exists to stop Wiktionary's acronyms outranking ordinary words — French `LE` must not beat the article `le`. Correct for French and Spanish.

For German it inverts. **Every German noun headword is capitalised**, so `common` is false for all of them, and at equal rank any lowercase homograph wins. Curated nouns fared worse still, because the vocabulary stores them with their article (`"der Hund"`).

My first attempt was a language flag disabling the tiebreak for German. **An existing test caught it**: `"die" is not marked as a function word`. Disabling the rule let the acronym `CIA` win. So the tiebreak *was* doing real work for German too — the problem was never "lowercase", it was that lowercase is a bad proxy for "ordinary word".

**After** — test for what the rule is actually trying to exclude. Language-neutral, no config flag, and strictly better for all three:

```go
func isOrdinaryHeadword(lemma, pos string) bool {
	return !isAcronym(lemma) && pos != ProperNounPOS
}
```

Verified live against the running server:

| surface | before | after |
|---|---|---|
| `die` | `der` (article) | `der` (article) ✓ |
| `Arm` | `arm` — *"poor"* | `der Arm` — the body part |
| `Essen` | `essen` — *"to eat"* | `essen`, with the **name** reading demoted below both real readings |
| `Sie` | `sie` | `sie` ✓ |
| `Hund` | — | `der Hund` |

The French side is visible in the reader screenshot: the alternatives for `c'` list `ce · this, that` first, with `CE · EC (European Community)` and `CE · a part of the primary school` after it.

### 7.2 `entryPool` cached under a key the analyzer didn't see

**Before** — `reading_handlers.go:178-182` cached under `strings.ToLower(surface)` but analyzed the **raw** `surface`. The analyzer branches on capitalisation, so `"Claire"` and `"claire"` shared one slot and whichever occurred first *in document order* fixed the reading for both — the same word glossed as a name in one paragraph and a common noun in another.

**After** — the key is analyzed. Folding is safe because the reliable proper-noun signal is the document's name set, decided once from mid-sentence capitalisation and looked up case-insensitively. Sentence-initial capitals — all this discards — are precisely the noise that pass exists to ignore.

My first attempt here keyed on the raw surface instead, which an existing pooling test rejected as payload duplication. The test was right.

### 7.3 `table.line()` returned the trailing newline on the last row

**Before** — `lexdata.go:130-140` inferred the row end from the *next* row's offset, which assumes exactly one newline between rows. The final row has no next offset, so it kept its `\n`. That leaks into `FormNote`, which is not trimmed downstream. A blank line anywhere would also have folded two rows together silently.

**After** — scan for the newline. Two regression tests.

### 7.4 `handleTextSection`'s comment was wrong

`reading_handlers.go:71` claimed *"Everything here is precomputed at startup; the request path only pools and marshals."* Line 102 builds a fresh `entryPool` per request and runs the full analyzer on every distinct surface. The comment would have stopped anyone looking there while profiling. Corrected, and the cost is now largely offset by the ETag and gzip work in §6.6.

---

## 8. Performance

### 8.1 The token cache thrashed on any real chapter

**Before** — `tokens.ts:47-65`:

```ts
const byteCache = new WeakMap<object, Uint8Array>()
const keyFor = new Map<string, object>()        // STRONG, keyed by paragraph text

if (keyFor.size > 256) keyFor.clear()           // wipes everything
```

The `Map` held both the paragraph string and the key object strongly, defeating the `WeakMap` entirely — then cleared the whole cache past 256 entries. Any EPUB chapter longer than 256 paragraphs degenerates: paragraph 257 wipes all 256, paragraph 1 re-encodes on next access, repeat. Worst case a ~0% hit rate while doing full `TextEncoder.encode()` of every paragraph repeatedly.

**After** — a real `WeakMap<ReadingParagraph, …>`. The paragraph object was already a stable identity. Test asserts 400 paragraphs stay cached.

### 8.2 Every non-word token was decoded twice

`tokens.ts:30-35` called `sliceBytes` once for `text` and again for `word`. `||` short-circuits only when the token *has* a lexicon entry — and whitespace and punctuation, the majority of the span array, never do.

### 8.3 54 full deck scans per render, on every keystroke

**Before** — `HomeView.vue:348-349` called `statsFor(...)` twice per category inline in the template. 27 categories × 2 = 54 full passes over the deck, each allocating a throwaway object per card. It re-ran on **every keystroke** in the search box, because `searchQuery` is read inside the same computed.

**After** — one `computed` keyed by category.

### 8.4 `selectionchange` on every character of a drag

`ReaderView.vue:190` re-tokenised the paragraph, allocated two arrays, and called `getBoundingClientRect()` — a forced layout flush — on every tick, with no throttle. Now coalesced to one run per frame with `requestAnimationFrame`, cancelled on unmount.

### 8.5 Two arrays per token per render

`classesFor` returned `string[]`; Vue's `normalizeClass` then joined it back into a string. There are six possible values. Returns a string now.

---

## 9. Interface language

The immersive convention is a defensible product position, but it was undermined by its own content: `ReadingSense.gloss` is English, `formNote` is English ("first-person plural present of parler"), `translationEn` is the answer side of every card. The app had already conceded the L1 — it just refused to show it in the chrome.

Concretely, a day-one German learner could not read the four grade buttons they were being asked to press (`Nochmal / Schwer / Gut / Leicht`), or the only destructive confirmation in the app, fired through `window.confirm`. The one escape hatch was the language switcher, which is self-defeating: it also switches away the content you came for.

**After** — a second preference, defaulting to `'auto'`, which reproduces the original behaviour exactly:

```ts
export type UiLangPreference = UiLanguageCode | 'auto'

export const uiLang = computed<UiLanguageCode>(() =>
  uiLangPreference.value === 'auto' ? lang.value : uiLangPreference.value,
)
```

plus a complete `en: UiStrings` block (~130 keys) and an `EN` toggle in the header, separate from the content-language control.

![Interface in English, content still French](changes/home-ui-english.png)

### Dead strings now rendered

Seven groups were translated into all three languages and rendered nowhere:

| Key | What it was for |
|---|---|
| `reading.statusUnknown/Learning/Known` | the reader status legend (§4.9) |
| `reading.unavailable` | explaining a missing Texts tab instead of silently hiding it |
| `quiz.missedItem` | `QuizView.vue:209` bypassed it and assembled the line from fragments with hardcoded punctuation, making the em-dash, colon and clause order untranslatable |

### Hardcoded English removed

`FlashcardsView.vue:257` rendered a literal `'en'` pill on every production card — in the German UI, a pill reading "EN" with no key for it. The part-of-speech pill describes the target-language word, so on a production card (where the prompt is the English gloss) it was describing the wrong language; it is now hidden there. `AppHeader.vue:16`'s `aria-label` was hardcoded, bilingual, and missing German.

---

## 10. Test infrastructure

**Before** — `package.json` had no test runner. Zero frontend tests.

**After** — Vitest + jsdom, 75 tests across 5 files.

Two environment problems worth recording:

1. jsdom exposes `localStorage` only for a real origin; the default `about:blank` is an opaque origin where the Storage API does not exist. Fixed with `environmentOptions.jsdom.url`.
2. Node 26 ships an experimental `localStorage` global that is inert without `--localstorage-file` and **takes precedence over jsdom's**. The symptom is a `window` with `document` but no Storage API:
   ```
   window: object | localStorage: undefined | document: object
   (node:22229) ExperimentalWarning: localStorage is not available because
   --localstorage-file was not provided.
   ```
   `src/test-setup.ts` installs a spec-faithful in-memory `Storage`, which also gives the quota tests a clean seam.

Vitest bundles its own Vite, which conflicts with the project's Vite 8 (rolldown) plugin types and broke `vue-tsc -b`. Test config lives in a separate `vitest.config.ts` with the reason documented.

**Backend:** 7 new tests for the hardening work — panic→JSON 500, oversized upload→413, too many zip entries, 403-vs-404, sentinel errors, conditional 304, gzip, and the `Vary`/`TimeoutHandler` regression.

---

## 11. Things I changed my mind about

Worth recording, because in each case a test or a measurement overruled the plan:

1. **German tiebreak.** Planned a language flag. An existing test proved the rule was load-bearing for German too; the real fix was narrower, language-neutral, and needed no config (§7.1).
2. **`entryPool` key.** Planned to key on the raw surface. An existing pooling test showed that doubles the payload for every sentence-initial word. Keying on the folded form and *analyzing* it is both correct and free (§7.2).
3. **Due-date migration.** Planned to recompute stored `dueDate`s. They can't be — the hour and zone aren't recorded. Left alone with a comment, rather than a migration that would invent data (§1.1).
4. **`Vary`.** My own change silently broke a header I had just added, because of `TimeoutHandler`'s copy semantics. Only visible by curling the running server (§6.7).

---

## 12. Not done

- **Dark mode.** `style.css:98` is `color-scheme: light`, which opts out of even the browser's automatic darkening — on a surface people read on for 30+ minutes, often at night. The token layer is already structured so ~12 variable overrides would do it.
- **Streaks / daily stats.** `lastReviewed` is now load-bearing (§1.4), so a streak counter is a `reduce` over existing state.
- **NFC normalisation.** Both tokenizers explicitly handle combining marks, so the code knows NFD input exists — but lexicon keys are NFC and `NormalizeSurface` doesn't normalise. NFD text (common from macOS and some EPUB producers) would tokenize perfectly and then fail *every* accented lookup, silently dropping coverage to near zero. I did not build the fix because I have not confirmed it occurs in practice, and the honest first step is a test proving it does. Flagging it as the highest-value remaining robustness item.
- **`/api/texts` payload.** `WordFrequency` is the largest field on a DTO documented as carrying no content, and it grows per book. Splitting it into `/api/texts/{id}/frequency` would honour the stated design.
- **Reader paragraph position.** `savePosition` still hardcodes `paragraph: 0`. I added the section jump control instead, which addresses the navigation complaint more directly; true scroll-position restore needs an IntersectionObserver and is a separate piece of work.
