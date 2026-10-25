<script setup lang="ts">
import { useRoute } from './router/router'
import AppHeader from './components/AppHeader.vue'
import HomeView from './views/HomeView.vue'
import TheoryView from './views/TheoryView.vue'
import QuizSetupView from './views/QuizSetupView.vue'
import QuizView from './views/QuizView.vue'
import FlashcardsView from './views/FlashcardsView.vue'

const route = useRoute()
</script>

<template>
  <AppHeader />
  <main class="page">
    <HomeView v-if="route.name === 'home'" />
    <TheoryView v-else-if="route.name === 'theory'" :tense="route.tense" />
    <QuizSetupView v-else-if="route.name === 'quiz-setup'" :tense="route.tense" />
    <QuizView v-else-if="route.name === 'quiz'" :tense="route.tense" :type="route.type" :count="route.count" />
    <FlashcardsView v-else-if="route.name === 'flashcards'" :category="route.category" />
  </main>
</template>
