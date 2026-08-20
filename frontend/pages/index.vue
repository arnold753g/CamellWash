<template><div /></template>

<script setup lang="ts">
definePageMeta({ middleware: 'auth' })
const auth = useAuthStore()
onMounted(() => { void navigateTo(auth.isAdmin ? '/admin' : '/app', { replace: true }) })
</script>
