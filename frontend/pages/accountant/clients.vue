<template>
  <CustomerVehicleWorkspace read-only />
</template>

<script setup lang="ts">
import CustomerVehicleWorkspace from '~/components/customers/CustomerVehicleWorkspace.vue'
definePageMeta({ layout: 'app', middleware: 'accountant' })
useHead({ title: 'Consulta de clientes' })
</script>
