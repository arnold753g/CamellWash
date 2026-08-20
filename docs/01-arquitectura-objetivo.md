# Arquitectura objetivo

Fuente funcional: [`GUIA_UNICA_SISTEMA_LAVADERO.md`](../GUIA_UNICA_SISTEMA_LAVADERO.md).

## Decisión arquitectónica

Se mantiene el monolito modular existente: frontend Nuxt/Vue, API REST Go y PostgreSQL. No se cambia framework, ORM, transporte de autenticación ni librería visual. El sistema crecerá por capacidades dentro de las carpetas actuales y solo incorporará una capa de servicio cuando una operación tenga reglas transaccionales, autorización contextual o concurrencia que no deban residir en el handler.

```text
Nuxt/Vue + Pinia + PrimeVue
  └─ API REST JSON con cookie de sesión + CSRF
       ├─ middleware: autenticación, RBAC, rate limit, CORS, logs
       ├─ handlers: HTTP, DTO y traducción de errores
       ├─ services: reglas de negocio y transacciones complejas
       ├─ models/persistencia: GORM y consultas parametrizadas
       └─ PostgreSQL: SQL versionado, constraints e índices
```

## Contratos que se conservan

- API bajo `/api/v1` con envelope común y errores saneados.
- Sesión opaca en cookie HttpOnly y token CSRF en memoria frontend.
- Timestamps persistidos en UTC; presentación con `America/La_Paz`.
- Migraciones SQL secuenciales, embebidas, transaccionales e idempotentes.
- Autorización backend obligatoria; guards frontend solo mejoran navegación.
- PrimeVue y tokens de `theme.css` como único lenguaje visual.
- Configuración por entorno, `.env` ignorado y `.env.example` sin secretos reales.

## Evolución por módulo

| Módulo | Backend | Frontend | Datos |
| --- | --- | --- | --- |
| 1 | Adaptar auth/RBAC y administración de usuarios. | Login y administración por tres roles. | Migrar roles/login; sesiones existentes coherentes. |
| 2 | CRUD y consultas activas de catálogos. | Listados/formularios administrativos. | Tipos, categorías y servicios con baja lógica. |
| 3 | Caso de uso de tarifa vigente. | Matriz editable. | Precio decimal único por combinación. |
| 4 | Casos de uso de cliente/vehículo. | Listado, detalle y relación uno-a-muchos. | Normalización y unicidad aprobadas. |
| 5 | Servicios transaccionales de orden y máquina de estados. | Flujo operativo responsive. | Snapshots, asignaciones e historial append-only. |
| 6 | Servicio idempotente de cobro/entrega. | Modal de cobro y estado. | Un pago válido por orden. |
| 7 | Consultas agregadas y permisos de reporte. | Dashboard con componentes existentes. | Índices basados en consultas reales. |

## Límites y criterios de introducción de capas

- Los handlers no deben calcular importes, decidir transiciones ni confiar en campos protegidos del cliente.
- Las operaciones multitabla de órdenes, estados y pagos viven en servicios y usan `db.Transaction`.
- Los repositorios dedicados se crean únicamente si reducen duplicación o permiten probar reglas sin acoplarlas a HTTP.
- No se añade bus de eventos, microservicios, GraphQL, JWT, Redis ni nueva librería de estado/UI en V1 sin evidencia y autorización.
- La concurrencia se resuelve primero con constraints, updates condicionales, bloqueos de fila o claves idempotentes de PostgreSQL.

## Estrategia de pruebas

- Unitarias: normalización, validadores, permisos, máquina de estados, totales y fórmulas.
- Integración PostgreSQL: migraciones, constraints, transacciones, revocación, concurrencia e idempotencia.
- HTTP: matrices de rol/asignación, errores uniformes, CSRF e IDOR.
- Frontend: utilidades/stores y flujos críticos; typecheck y build siempre.
- Visual manual: breakpoints y referencias definidos en la auditoría hasta incorporar una herramienta automatizada existente o autorizada.

## Despliegue y operación

- Desarrollo: Compose actual o procesos locales con PostgreSQL.
- Producción: migración obligatoria antes del backend, cookies Secure, registro público deshabilitado, CORS exacto y servicios internos detrás de Nginx.
- El rate limit en memoria limita el despliegue inicial a una instancia backend; escalar horizontalmente requerirá un almacén compartido aprobado.

