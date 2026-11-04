<template>
  <DashboardReportWorkspace show-financial report-mode />
</template>

<script setup lang="ts">
import DashboardReportWorkspace from '~/components/reports/DashboardReportWorkspace.vue'

definePageMeta({ layout: 'app', middleware: 'accountant' })
useHead({ title: 'Reportes contables' })
</script>
