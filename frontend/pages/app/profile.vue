<template>
  <section>
    <div class="page-heading"><h1>Mi perfil</h1><p>Actualiza tus datos permitidos o cambia tu contraseña.</p></div>
    <div class="detail-grid">
      <Card>
        <template #title>Datos personales</template>
        <template #content>
          <Message v-if="profileError" severity="error" :closable="false">{{ profileError }}</Message>
          <form class="stack" novalidate @submit.prevent="saveProfile">
            <div class="form-field"><label for="first-name">Nombre</label><InputText id="first-name" v-model.trim="profile.first_name" autocomplete="given-name" /></div>
            <div class="form-field"><label for="last-name">Apellido</label><InputText id="last-name" v-model.trim="profile.last_name" autocomplete="family-name" /></div>
            <div class="form-field"><label for="email">Correo</label><InputText id="email" :model-value="auth.user?.email" disabled /></div>
            <Button type="submit" label="Guardar perfil" icon="pi pi-save" :loading="savingProfile" />
          </form>
        </template>
      </Card>
      <Card>
        <template #title>Cambiar contraseña</template>
        <template #content>
          <Message v-if="passwordError" severity="error" :closable="false">{{ passwordError }}</Message>
          <form class="stack" novalidate @submit.prevent="changePassword">
            <div class="form-field"><label for="current-password">Contraseña actual</label><Password input-id="current-password" v-model="password.current" autocomplete="current-password" :feedback="false" toggle-mask /></div>
            <div class="form-field"><label for="new-password">Nueva contraseña</label><Password input-id="new-password" v-model="password.next" autocomplete="new-password" toggle-mask /><small class="form-help">Mínimo 12 caracteres. Al cambiarla se cerrarán todas tus sesiones.</small></div>
            <Button type="submit" label="Cambiar contraseña" icon="pi pi-key" :loading="savingPassword" />
          </form>
        </template>
      </Card>
    </div>
  </section>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Password from 'primevue/password'
import { useToast } from 'primevue/usetoast'
import type { User } from '~/types/auth'
import { apiError } from '~/utils/api-error'

definePageMeta({ layout: 'app', middleware: 'auth' })
useHead({ title: 'Mi perfil' })
const auth = useAuthStore()
const toast = useToast()
const profile = reactive({ first_name: auth.user?.first_name || '', last_name: auth.user?.last_name || '' })
const password = reactive({ current: '', next: '' })
const profileError = ref('')
const passwordError = ref('')
const savingProfile = ref(false)
const savingPassword = ref(false)

const saveProfile = async () => {
  profileError.value = ''
  if (profile.first_name.trim().length < 2 || profile.last_name.trim().length < 2) { profileError.value = 'Nombre y apellido deben tener al menos 2 caracteres.'; return }
  savingProfile.value = true
  try {
    const response = await auth.request<User>('/me', { method: 'PATCH', body: profile })
    if (response.data) auth.user = response.data
    toast.add({ severity: 'success', summary: 'Perfil actualizado', life: 3500 })
  } catch (error) { profileError.value = apiError(error).message }
  finally { savingProfile.value = false }
}

const changePassword = async () => {
  passwordError.value = ''
  if (password.next.length < 12) { passwordError.value = 'La nueva contraseña debe tener al menos 12 caracteres.'; return }
  savingPassword.value = true
  try {
    await auth.request('/me/password', { method: 'POST', body: { current_password: password.current, new_password: password.next } })
    auth.clear()
    await navigateTo('/login?password_changed=true')
  } catch (error) { passwordError.value = apiError(error).message }
  finally { savingPassword.value = false }
}
</script>
