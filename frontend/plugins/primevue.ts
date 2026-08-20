import PrimeVue from 'primevue/config'
import Aura from '@primevue/themes/aura'
import ConfirmationService from 'primevue/confirmationservice'
import ToastService from 'primevue/toastservice'

export default defineNuxtPlugin((nuxtApp) => {
  nuxtApp.vueApp.use(PrimeVue, {
    theme: { preset: Aura, options: { darkModeSelector: '.app-dark', cssLayer: false } },
    ripple: true,
    locale: {
      startsWith: 'Empieza con', contains: 'Contiene', notContains: 'No contiene', endsWith: 'Termina con',
      equals: 'Igual', notEquals: 'Distinto', noFilter: 'Sin filtro', lt: 'Menor que', lte: 'Menor o igual',
      gt: 'Mayor que', gte: 'Mayor o igual', dateIs: 'Fecha igual', dateIsNot: 'Fecha distinta',
      dateBefore: 'Fecha anterior', dateAfter: 'Fecha posterior', clear: 'Limpiar', apply: 'Aplicar',
      matchAll: 'Coincidir todo', matchAny: 'Coincidir cualquiera', addRule: 'Agregar regla', removeRule: 'Quitar regla',
      accept: 'Sí', reject: 'No', choose: 'Elegir', upload: 'Subir', cancel: 'Cancelar', emptyMessage: 'Sin resultados'
    }
  })
  nuxtApp.vueApp.use(ToastService)
  nuxtApp.vueApp.use(ConfirmationService)
})
