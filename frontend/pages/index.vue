<template><div /></template>

<script setup lang="ts">
import { roleHomePath } from '~/utils/role-access'

definePageMeta({ middleware: 'auth' })
const auth = useAuthStore()
onMounted(() => { void navigateTo(roleHomePath(auth.user?.role), { replace: true }) })
</script>
