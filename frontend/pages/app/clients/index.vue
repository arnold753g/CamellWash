<template><CustomerVehicleWorkspace /></template>

<script setup lang="ts">
import CustomerVehicleWorkspace from '~/components/customers/CustomerVehicleWorkspace.vue'

definePageMeta({ layout: 'app', middleware: 'manager' })
useHead({ title: 'Clientes y vehículos' })
</script>
