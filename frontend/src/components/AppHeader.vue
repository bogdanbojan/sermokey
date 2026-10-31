<script setup lang="ts">
import { navigate } from '../router/router'
import { useLanguage, setLanguage } from '../i18n/language'
import logo from '../assets/sermokey-logo.png'

const { lang, languages } = useLanguage()
</script>

<template>
  <header class="app-header">
    <div class="header-inner">
      <button class="brand" aria-label="Sermokey" @click="navigate({ name: 'home' })">
        <img class="brand-mark" :src="logo" alt="Sermokey" />
      </button>

      <div class="lang-switch" role="group" aria-label="Choix de la langue / Selección de idioma">
        <button
          v-for="l in languages"
          :key="l.code"
          type="button"
          class="lang-btn"
          :class="{ active: lang === l.code }"
          :aria-pressed="lang === l.code"
          @click="setLanguage(l.code)"
        >
          {{ l.name }}
        </button>
      </div>
    </div>
  </header>
</template>

<style scoped>
.app-header {
  border-bottom: 1px solid var(--color-border);
  background: var(--color-surface);
}
.header-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  /* Wrapping is the backstop. The wordmark and the switcher both refuse to
     shrink — the logo because a squashed wordmark looks broken, the
     switcher because its labels are words rather than icons — so on a
     screen too narrow for both they stack instead of pushing the page into
     a horizontal scroll, which would drag every view sideways with it. */
  flex-wrap: wrap;
  gap: var(--space-3);
  max-width: 880px;
  margin: 0 auto;
  padding: var(--space-3) var(--space-5);
}
.brand {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  background: none;
  border: none;
  cursor: pointer;
  text-align: left;
  padding: 0;
}
.brand-mark {
  display: block;
  height: 3.5rem;
  width: auto;
  flex-shrink: 0;
}

/* Segmented-control style switcher: native language names (not flags —
   flags represent countries, not languages, and several countries speak
   either language), both options always visible so the current choice and
   the alternative are equally discoverable, persisted via localStorage. */
.lang-switch {
  display: inline-flex;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: 2px;
  background: var(--color-bg);
  flex-shrink: 0;
}
.lang-btn {
  border: none;
  background: none;
  padding: 0.4rem 0.8rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-text-muted);
  border-radius: calc(var(--radius-sm) - 2px);
  cursor: pointer;
}
.lang-btn:hover {
  color: var(--color-text);
}
.lang-btn.active {
  background: var(--color-primary);
  color: var(--color-primary-contrast);
}

/* On a phone the wordmark and three language names do not fit side by side
   at desktop sizes. Scaling both down a step keeps them on one row, which
   is worth more than the extra few pixels of logo: a header that wraps
   costs vertical space on the screen where it is scarcest. */
@media (max-width: 640px) {
  .header-inner {
    padding: var(--space-2) var(--space-3);
    gap: var(--space-2);
  }
  .brand-mark {
    height: 2.5rem;
  }
  .lang-btn {
    padding: 0.35rem 0.5rem;
    font-size: 0.78rem;
  }
}
</style>
