<template><main class="error-page"><div><div class="error-code">403</div><h1>Acceso denegado</h1><p class="muted">Tu cuenta no tiene permiso para acceder a esta ruta.</p><Button label="Volver a mi inicio" icon="pi pi-home" @click="goHome" /></div></main></template>
<script setup lang="ts">
import Button from 'primevue/button'
import { roleHomePath } from '~/utils/role-access'

useHead({ title: 'Acceso denegado' })
const auth = useAuthStore()
const goHome = () => navigateTo(auth.user ? roleHomePath(auth.user.role) : '/')
</script>
