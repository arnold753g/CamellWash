<template>
  <NuxtRouteAnnouncer />
  <Toast position="top-right" />
  <ConfirmDialog />
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
</template>

<script setup lang="ts">
import ConfirmDialog from 'primevue/confirmdialog'
import Toast from 'primevue/toast'

const branding = useBranding()
useHead({ titleTemplate: (title) => title ? `${title} · ${branding.name}` : branding.name })
</script>
