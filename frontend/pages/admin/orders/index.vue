<template><OrderWorkspace /></template>

<script setup lang="ts">
import OrderWorkspace from '~/components/orders/OrderWorkspace.vue'

definePageMeta({ layout: 'admin', middleware: 'admin' })
useHead({ title: 'Órdenes de lavado' })
</script>
