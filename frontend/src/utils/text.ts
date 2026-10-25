// Grading for free-text exercises. We deliberately treat an accent-stripped
// match as *correct* (not merely "close") per product decision: French/
// Spanish keyboards/IMEs make accents annoying to type, and the pedagogical
// value is in getting the right word/ending, not in keyboard dexterity. We
// still want to nudge the learner to notice they dropped an accent, so we
// surface a warning instead of silently accepting it.
export type GradeResult =
  | { correct: true; accentWarning: false }
  | { correct: true; accentWarning: true }
  | { correct: false }

// Unicode NFD-decomposes "ñ" into "n" + a combining tilde (U+0303) — the
// *same* combining mark used for e.g. "ã". Stripping combining marks
// naively would therefore silently turn "año" into "ano", which is wrong:
// in Spanish ñ is a distinct letter from n, not an accented n (it changes
// the word's meaning/pronunciation entirely). We protect ñ/Ñ from
// decomposition so only true diacritics (á, é, í, ó, ú, ü, and French's
// à/â/î/ï/ô/ù/û/ë/ê/ç...) get stripped.
const NY_LOWER_PLACEHOLDER = '\u0000'
const NY_UPPER_PLACEHOLDER = '\u0001'

function stripAccents(s: string): string {
  const protected_ = s.replace(/ñ/g, NY_LOWER_PLACEHOLDER).replace(/Ñ/g, NY_UPPER_PLACEHOLDER)
  const stripped = protected_.normalize('NFD').replace(/[\u0300-\u036f]/g, '')
  return stripped
    .replace(new RegExp(NY_LOWER_PLACEHOLDER, 'g'), 'ñ')
    .replace(new RegExp(NY_UPPER_PLACEHOLDER, 'g'), 'Ñ')
}

function normalize(s: string): string {
  return stripAccents(s.trim().toLowerCase())
}

export function gradeTextAnswer(userInput: string, correctAnswer: string): GradeResult {
  const trimmed = userInput.trim()
  if (trimmed.toLowerCase() === correctAnswer.toLowerCase()) {
    return { correct: true, accentWarning: false }
  }
  if (normalize(trimmed) === normalize(correctAnswer)) {
    return { correct: true, accentWarning: true }
  }
  return { correct: false }
}
