<script setup lang="ts">
// Generic collapsible section. Open/closed state is controlled by the
// parent (v-model) so it can be persisted (see useExpandedState in
// recent.ts) — accordions remembering what you left open/closed across
// visits is a small but real "respect the user's prior choices" touch.
const props = defineProps<{
  modelValue: boolean
  title: string
  badge?: string | number
}>()
const emit = defineEmits<{ 'update:modelValue': [boolean] }>()

function toggle() {
  emit('update:modelValue', !props.modelValue)
}
</script>

<template>
  <div class="accordion" :class="{ open: modelValue }">
    <button
      type="button"
      class="accordion-header"
      :aria-expanded="modelValue"
      @click="toggle"
    >
      <span class="accordion-title">{{ title }}</span>
      <span v-if="badge" class="accordion-badge">{{ badge }}</span>
      <span class="accordion-chevron" aria-hidden="true">›</span>
    </button>
    <div class="accordion-body">
      <div class="accordion-body-inner">
        <slot />
      </div>
    </div>
  </div>
</template>

<style scoped>
.accordion {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-surface);
  overflow: hidden;
}
.accordion + .accordion {
  margin-top: var(--space-3);
}
.accordion-header {
  width: 100%;
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-4) var(--space-5);
  background: none;
  border: none;
  cursor: pointer;
  text-align: left;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--color-text);
}
.accordion-title {
  flex: 1;
}
.accordion-badge {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.15rem 0.55rem;
  border-radius: 999px;
  background: var(--color-accent);
  color: var(--color-primary-contrast);
}
.accordion-chevron {
  font-size: 1.3rem;
  color: var(--color-text-muted);
  transition: transform 0.2s ease;
  line-height: 1;
}
.accordion.open .accordion-chevron {
  transform: rotate(90deg);
}

/* The grid-rows trick: animating from 0fr to 1fr smoothly animates height
   without ever measuring scrollHeight in JS, and it degrades gracefully
   (no animation, just appears) in any browser without grid-template-rows
   transition support. */
.accordion-body {
  display: grid;
  grid-template-rows: 0fr;
  transition: grid-template-rows 0.25s ease;
}
.accordion.open .accordion-body {
  grid-template-rows: 1fr;
}
.accordion-body-inner {
  overflow: hidden;
  padding: 0 var(--space-5);
}
.accordion.open .accordion-body-inner {
  padding: 0 var(--space-5) var(--space-5);
}
</style>
