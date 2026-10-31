<p align="center">
  <img src="assets/sermokey-logo.png" alt="Sermokey" width="420" />
</p>

<p align="center"><strong>Unlock new languages.</strong></p>

Sermokey is a language learning platform focused on grammar: interactive
theory pages, conjugation drills, vocabulary practice, and spaced-repetition
flashcards, with immersive UI text in the language you're studying.

Currently supported: **French**, **Spanish**, and **German**.

<p align="center">
  <img src="assets/screenshots/homepage.png" alt="Sermokey homepage" width="720" />
</p>

## Features

- **Theory** — reference pages per tense/category, grouped and searchable.
- **Exercises** — generated conjugation and translation drills per tense,
  with multiple exercise types.
- **Vocabulary** — ~2,000–2,250 curated words per language across 27
  categories, with a reversible quiz direction.
- **Flashcards / SRS** — spaced-repetition review persisted locally in the
  browser.
- **Immersive i18n** — switching the study language also switches the
  interface copy (e.g. "Vérifier" vs "Comprobar" vs "Prüfen"), not just the
  exercises.

## Vocabulary selection

Word lists are built from the intersection of frequency and pedagogy rather
than either alone — pure frequency produces an unteachable list, pure
thematic grouping produces categories padded with rare words:

1. **Frequency floor** — lemmas sit inside roughly the top 5,000 of a
   reference frequency source, with the bulk inside the top 3,000.
2. **Thematic budget** — category sizes are proportional to communicative
   load, not split evenly. `verbes` carries ~270 words; `couleurs` ~20.
3. **Closed-class guarantee** — every pronoun, interrogative, determiner,
   preposition and conjunction, plus the top ~150 verbs, must be present.
   These are finite sets, so completeness is a checkable property, and
   `grammar.ValidateVocab` checks it on every test run.

Reference sources: French — Lonsdale & Le Bras, *A Frequency Dictionary of
French*, plus CEFRLex/FLELex. Spanish — the Instituto Cervantes *Plan
Curricular* CEFR inventory, plus Davies' frequency dictionary. German — the
Goethe-Institut A1/A2/B1 Wortlisten, plus Tschirner & Möhring.

Target size is ~2,000–2,250 lemmas per language. B2 receptive vocabulary is
roughly 3,250–3,750 word families (Milton 2010); a curated deck of this size
is the practical core of that, with the remainder acquired incidentally and
through cognates.

## Tech stack

- **Backend**: Go (`backend/`) — a small JSON API (`net/http`), no external
  framework.
- **Frontend**: Vue 3 + TypeScript + Vite (`frontend/`).

## Getting started

```sh
make install   # npm install for the frontend
make run       # starts backend (:8080) and frontend (:5173) in the background
make logs      # tail both logs
make stop      # stop whatever `make run` started
```

Run `make help` for the full list of targets (`build`, `test`, `clean`, ...).

## Project layout

```
backend/   Go JSON API (cmd/server, internal/grammar, internal/httpapi)
frontend/  Vue 3 + TypeScript + Vite single-page app
assets/    Project branding (logo, screenshots)
```

## Roadmap

See open issues for upcoming work.

## License

[MIT](LICENSE)
