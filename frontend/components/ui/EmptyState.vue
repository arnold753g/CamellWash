<template>
  <div class="empty-state" role="status">
    <i :class="icon" aria-hidden="true" />
    <h2>{{ title }}</h2>
    <p>{{ description }}</p>
    <slot />
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ title: string, description: string, icon?: string }>(), { icon: 'pi pi-inbox' })
</script>

