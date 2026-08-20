<template><AppShell :mode="auth.isAdmin ? 'admin' : 'staff'"><slot /></AppShell></template>
<script setup lang="ts">const auth = useAuthStore()</script>
