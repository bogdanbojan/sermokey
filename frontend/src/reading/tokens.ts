// Turning paragraph span data into something renderable, and turning a
// browser text selection back into token indices.
//
// The second direction is the subtle one. Rendering one <span> per word
// over the verbatim paragraph text means the browser's own selection
// machinery does the hard part — hit testing, drag handling, touch
// long-press, accessibility — and all this module has to do is read the
// token index back off the DOM. Reimplementing selection over a canvas or
// a single opaque text node would have meant rebuilding all of that badly.
import { NO_ENTRY, SPAN_STRIDE, type ReadingParagraph } from '../types/api'

/** One rendered piece of a paragraph. */
export interface RenderToken {
  /** Index within the paragraph's token list; also the DOM data attribute. */
  index: number
  text: string
  /** Index into the section lexicon, or NO_ENTRY. */
  entry: number
  /** Whether this is a word (hoverable, selectable) or inter-word material. */
  word: boolean
}

/** Expands a paragraph's flat span triples into render tokens. */
export function renderTokens(p: ReadingParagraph): RenderToken[] {
  const cached = tokenCache.get(p)
  if (cached) return cached

  const bytes = bytesOf(p)
  const out: RenderToken[] = []
  for (let i = 0, n = 0; i < p.spans.length; i += SPAN_STRIDE, n++) {
    const start = p.spans[i]!
    const end = p.spans[i + 1]!
    const entry = p.spans[i + 2]!
    // Decoded once. `word` used to call sliceBytes a second time, and
    // because `||` only short-circuits when the token has a lexicon entry,
    // every space and punctuation mark — the majority of the span array —
    // was decoded twice.
    const text = decoder.decode(bytes.subarray(start, end))
    out.push({
      index: n,
      text,
      entry,
      word: entry !== NO_ENTRY || isWordish(text),
    })
  }
  tokenCache.set(p, out)
  return out
}

// The backend works in byte offsets (Go strings are bytes); JavaScript
// strings are UTF-16. French is full of accented characters, so the two
// disagree constantly and slicing with the raw offsets would cut words in
// half. Converting once per paragraph via a byte view keeps the wire format
// natural for Go without making the client guess.
const encoder = new TextEncoder()
const decoder = new TextDecoder()

// Keyed on the paragraph object, which is already a stable identity for
// the lifetime of a loaded section.
//
// The previous version keyed a WeakMap through a `Map<string, object>` of
// paragraph *text*, which held both the string and the key strongly and so
// defeated the WeakMap entirely — then cleared the whole cache whenever it
// passed 256 entries. Any EPUB chapter longer than 256 paragraphs
// therefore thrashed: each new paragraph wiped all the others, and the
// next access re-encoded from scratch.
const byteCache = new WeakMap<ReadingParagraph, Uint8Array>()
const tokenCache = new WeakMap<ReadingParagraph, RenderToken[]>()

function bytesOf(p: ReadingParagraph): Uint8Array {
  let bytes = byteCache.get(p)
  if (!bytes) {
    bytes = encoder.encode(p.text)
    byteCache.set(p, bytes)
  }
  return bytes
}

/** Whether a span with no dictionary entry should still be selectable —
 *  proper nouns and numerals have no gloss but are part of the sentence. */
function isWordish(s: string): boolean {
  return /[\p{L}\p{N}]/u.test(s)
}

/** The token index a DOM node belongs to, or undefined if it is not in a
 *  token span. */
export function tokenIndexOf(node: Node | null): number | undefined {
  let el = node instanceof Element ? node : node?.parentElement ?? null
  while (el) {
    const raw = (el as HTMLElement).dataset?.token
    if (raw !== undefined) {
      const n = Number(raw)
      return Number.isInteger(n) ? n : undefined
    }
    el = el.parentElement
  }
  return undefined
}

/** A contiguous run of tokens the reader has selected. */
export interface TokenRange {
  paragraph: number
  from: number
  to: number // inclusive
}

/**
 * Multi-word selections are capped. Formulaic sequences — the thing worth
 * capturing as a unit (Wray 2002) — are short, and an uncapped selection
 * would let a stray triple-click turn a whole paragraph into a flashcard.
 */
export const MAX_PHRASE_TOKENS = 8

/**
 * Reads the current selection back as a token range, snapped to whole
 * tokens. Returns undefined when the selection is empty, collapsed, spans
 * more than one paragraph, or is too long to be a phrase.
 */
export function selectionToRange(root: HTMLElement | null): TokenRange | undefined {
  const sel = window.getSelection()
  if (!sel || sel.isCollapsed || sel.rangeCount === 0) return undefined

  const range = sel.getRangeAt(0)
  if (root && !root.contains(range.commonAncestorContainer)) return undefined

  const startPara = paragraphIndexOf(range.startContainer)
  const endPara = paragraphIndexOf(range.endContainer)
  if (startPara === undefined || startPara !== endPara) return undefined

  const a = tokenIndexOf(range.startContainer)
  const b = tokenIndexOf(range.endContainer)
  if (a === undefined || b === undefined) return undefined

  const from = Math.min(a, b)
  const to = Math.max(a, b)
  if (to - from + 1 > MAX_PHRASE_TOKENS * 2) return undefined
  return { paragraph: startPara, from, to }
}

function paragraphIndexOf(node: Node | null): number | undefined {
  let el = node instanceof Element ? node : node?.parentElement ?? null
  while (el) {
    const raw = (el as HTMLElement).dataset?.paragraph
    if (raw !== undefined) {
      const n = Number(raw)
      return Number.isInteger(n) ? n : undefined
    }
    el = el.parentElement
  }
  return undefined
}

/** The text of a token range, trimmed — what gets looked up as a phrase. */
export function rangeText(tokens: RenderToken[], range: TokenRange): string {
  return tokens
    .slice(range.from, range.to + 1)
    .map((t) => t.text)
    .join('')
    .trim()
}

/**
 * The sentence containing a token, as plain text plus the token's offsets
 * within it. Stored with a saved word so its flashcard can show the word in
 * the context it was met in.
 *
 * Sentence splitting is done on terminal punctuation followed by a space.
 * It is not perfect — "M. Rahmani" will split early — but the result is
 * only ever shown to a human as context, never parsed, so an occasional
 * short fragment costs nothing, whereas an abbreviation list per language
 * would be real complexity for no real gain.
 */
export function sentenceAround(
  tokens: RenderToken[],
  index: number,
): { sentence: string; start: number; end: number } {
  let from = 0
  for (let i = index - 1; i >= 0; i--) {
    if (isSentenceEnd(tokens[i]!.text)) {
      from = i + 1
      break
    }
  }
  let to = tokens.length - 1
  for (let i = index; i < tokens.length; i++) {
    if (isSentenceEnd(tokens[i]!.text)) {
      to = i
      break
    }
  }

  const before = tokens.slice(from, index).map((t) => t.text).join('')
  const word = tokens[index]!.text
  const sentence = before + word + tokens.slice(index + 1, to + 1).map((t) => t.text).join('')
  const leading = sentence.length - sentence.trimStart().length
  return {
    sentence: sentence.trim(),
    start: before.length - leading,
    end: before.length - leading + word.length,
  }
}

function isSentenceEnd(text: string): boolean {
  return /[.!?…][\s»"']*$/.test(text) && text.trim() !== ''
}
