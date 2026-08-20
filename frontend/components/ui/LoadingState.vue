<template>
  <div class="loading-state" role="status" aria-live="polite">
    <ProgressSpinner stroke-width="4" />
    <span class="sr-only">Cargando</span>
  </div>
</template>

<script setup lang="ts">
import ProgressSpinner from 'primevue/progressspinner'
</script>
