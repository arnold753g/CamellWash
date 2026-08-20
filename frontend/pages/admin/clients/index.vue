<template><CustomerVehicleWorkspace /></template>

<script setup lang="ts">
import CustomerVehicleWorkspace from '~/components/customers/CustomerVehicleWorkspace.vue'

definePageMeta({ layout: 'admin', middleware: 'admin' })
useHead({ title: 'Clientes y vehículos' })
</script>
