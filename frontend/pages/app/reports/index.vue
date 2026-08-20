<template><DashboardReportWorkspace :show-financial="false" report-mode /></template>

<script setup lang="ts">
import DashboardReportWorkspace from '~/components/reports/DashboardReportWorkspace.vue'

definePageMeta({ layout: 'app', middleware: 'manager' })
useHead({ title: 'Reportes operativos' })
</script>
