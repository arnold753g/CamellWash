<template><DashboardReportWorkspace show-financial report-mode /></template>

<script setup lang="ts">
import DashboardReportWorkspace from '~/components/reports/DashboardReportWorkspace.vue'

definePageMeta({ layout: 'admin', middleware: 'admin' })
useHead({ title: 'Reportes' })
</script>
