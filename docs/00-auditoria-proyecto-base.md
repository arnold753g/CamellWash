# Auditoría del proyecto base

Fecha de corte: 2026-08-13

Fuente funcional: [`GUIA_UNICA_SISTEMA_LAVADERO.md`](../GUIA_UNICA_SISTEMA_LAVADERO.md)

Raíz inspeccionada: `D:\Users\arnol\Documents\Nueva Version del proyecto\Proyecto Base`

## Veredicto

El starter es **apto para evolucionar al Sistema de Gestión de Lavadero con adaptaciones controladas**. La arquitectura, la autenticación por sesión opaca, la seguridad transversal y el lenguaje visual se pueden reutilizar. El Módulo 1 debe adaptar el modelo binario `admin/user` a tres roles, retirar el registro público y completar la creación administrativa de usuarios.

La Fase 0 queda técnicamente documentada, pero requiere aprobación del usuario antes de implementar el Módulo 1. Git no está inicializado por decisión expresa, por lo que no existen rama, historial, `status` ni `diff` Git disponibles.

## Inventario técnico

| Área | Evidencia real |
| --- | --- |
| Backend | Go `1.25.0`, toolchain `1.25.12`, Gorilla Mux `1.8.1`, GORM `1.31.1`, validator `10.28.0`, bcrypt mediante `x/crypto`; módulos en `backend/cmd` e `backend/internal`. |
| Frontend | Nuxt instalado `4.5.2`, Vue `3.5.41`, Pinia `3`, PrimeVue `4.5`, Zod `3`, TypeScript `5.9`; aplicación SSR con guards cliente. |
| Datos | PostgreSQL 16, migraciones SQL embebidas, versionadas y transaccionales; GORM para acceso parametrizado. |
| Autenticación | Sesión opaca aleatoria en cookie HttpOnly; solo se persiste SHA-256 del token; expiración absoluta e inactiva; revocación en cambio de contraseña, rol o desactivación. |
| Seguridad HTTP | CSRF HMAC para métodos inseguros, lista CORS exacta, rate limiting por IP, headers defensivos, body JSON limitado y campos desconocidos rechazados. |
| API | REST JSON bajo `/api/v1`; envelope común `success/data/message/error/request_id/timestamp`; errores internos saneados. |
| UI | PrimeVue Aura, PrimeIcons, tema oscuro propio en `frontend/assets/theme.css`, layout con sidebar/header y estados reutilizables. |
| Infraestructura | Dockerfiles multi-stage, Compose de desarrollo y producción, Nginx; PostgreSQL no publicado en producción. |
| Pruebas | Tests unitarios Go de configuración, middleware, seguridad y router; tres tests Node de sesión/redirección. No hay todavía pruebas de integración con PostgreSQL para los handlers. |
| CI/CD | No se encontraron pipelines. Las verificaciones son locales y están documentadas en README. |
| Instrucciones | No existen archivos `AGENTS.md`. `AUDIT.md` documenta la construcción previa del starter y se conserva como antecedente técnico, no como especificación funcional. |

## Tabla de reutilización

| Componente | Evidencia | Decisión | Justificación | Riesgo | Acción propuesta |
| --- | --- | --- | --- | --- | --- |
| Stack Go/Nuxt/PostgreSQL | `go.mod`, `package-lock.json`, Compose | REUTILIZAR | Es coherente, probado y suficiente para todo el alcance V1. | Dependencias con rangos semver en frontend. | Mantener lockfile y verificar en cada módulo. |
| Organización backend | `cmd/*`, `internal/{handlers,middleware,models,database}` | ADAPTAR | La separación actual funciona para un dominio pequeño. | Handlers crecerán al incorporar reglas transaccionales complejas. | Introducir servicios/casos de uso solo donde aparezca lógica de negocio real. |
| Sesiones opacas | `security/session.go`, `middleware/auth.go`, `auth_sessions` | REUTILIZAR | Evita JWT innecesario y ya soporta expiración/revocación. | La revocación se valida por solicitud, con coste de DB. | Mantener y ampliar pruebas de expiración/revocación. |
| Cookies y CSRF | `handlers/auth.go`, `middleware/auth.go` | REUTILIZAR | HttpOnly, Secure configurable, SameSite validado y token CSRF HMAC. | La protección requiere que frontend y API compartan configuración coherente. | Probar cookies y CSRF en integración. |
| RBAC backend | `RequireRoles` y subrouter `/admin` | ADAPTAR | El patrón central de middleware es correcto. | Solo reconoce `admin/user`. | Añadir los tres roles y pruebas negativas por endpoint. |
| Modelo de usuario | `users`, `models/user.go` | ADAPTAR | Ya protege hash, estado y datos públicos. | Usa email como login y divide nombre/apellido; faltan roles del negocio. | Resolver identificador de login y migrar sin exponer credenciales. |
| Registro público | `/auth/register`, `register.vue` | NO USAR | Contradice “solo Administrador crea usuarios”. | Permitiría altas fuera del control administrativo. | Retirar ruta/pantalla/configuración en Módulo 1. |
| Seed | `internal/seed/seed.go` | SUSTITUIR | Está bloqueado en producción y exige bandera, pero crea `admin` y `user`. | No implementa el bootstrap exacto del primer Administrador ni los tres roles. | Crear mecanismo explícito, idempotente y solo de bootstrap/desarrollo aprobado. |
| Migrador SQL | `database/migrate.go` | REUTILIZAR | Versionado, ordenado, transaccional e idempotente. | Migración inicial ya aplicada exige migraciones aditivas. | Añadir `000002_*` y probar desde DB vacía y actualizada. |
| Envelope/errores | `respond/respond.go`, helpers | REUTILIZAR | Forma uniforme y sin detalle interno. | Mensajes backend están en inglés mientras la UI está en español. | Conservar contrato; decidir localización sin cambiar forma de API. |
| Rate limiter | `middleware/rate_limit.go` | ADAPTAR | Protección razonable para una sola instancia. | Estado en memoria; no coordina varias réplicas y el contador de intentos puede sufrir carreras. | Mantener para V1 monolítica y probar concurrencia; documentar límite de escalado. |
| Store HTTP/Pinia | `stores/auth.ts` | REUTILIZAR | Credenciales y CSRF quedan en memoria; cookie viaja con `credentials`. | Guards son solo cliente y la configuración de registro puede divergir. | Mantener como UX, nunca como autorización; eliminar flag duplicado de registro. |
| PrimeVue y tema | plugin y `theme.css` | REUTILIZAR | Ya cubre formularios, tablas, modales, feedback y accesibilidad básica. | No existe prueba visual automatizada. | Reutilizar componentes y añadir verificación por breakpoint en cada módulo. |
| Layout `AppShell` | sidebar/header/nav responsive | ADAPTAR | Es la referencia visual y de navegación. | Navegación está codificada para `admin/user`. | Extender items y etiquetas por los tres roles sin rediseñar. |
| Dashboard genérico | `/admin` backend/frontend | RETIRADO EN MÓDULO 7 | Fue reemplazado por métricas reales de órdenes, transiciones y pagos. | Resuelta. | No reintroducir conteos genéricos de usuarios como dashboard del negocio. |
| Docker/Nginx | Dockerfiles y Compose | REUTILIZAR | Separación clara de servicios y configuración productiva segura. | Docker daemon no estuvo disponible durante esta auditoría. | Repetir prueba con daemon disponible antes de aprobar el Módulo 1. |

## Riesgos de seguridad y compatibilidad

### Bloqueantes para el Módulo 1

1. La base acepta solo `admin/user`; debe migrarse a `admin/manager/wash_staff` (nombres técnicos propuestos) con etiquetas visibles Administrador, Encargado y Personal de lavado.
2. El registro público contradice el control de altas y debe eliminarse, no solo ocultarse.
3. Falta un endpoint/caso de uso para que el Administrador cree usuarios con contraseña inicial segura.
4. “login” es ambiguo frente al email actual. Se recomienda conservar **correo normalizado como identificador de acceso único**, porque minimiza migración y reutiliza validación; requiere aprobación.
5. No hay tests reales con PostgreSQL para duplicados, revocación, administración o migración desde base vacía.

### Importantes, no bloqueantes para diseñar

- El limitador por IP vive en memoria y no es distribuido; es adecuado solo mientras haya una instancia backend.
- El incremento de intentos fallidos es lectura-modificación-escritura y debe endurecerse frente a solicitudes concurrentes.
- `.env.example` contiene `DB_PASSWORD=1234`; aunque no es un secreto real, induce una configuración débil y debe cambiarse por placeholder obligatorio.
- Hay dos flags de registro (`REGISTRATION_ENABLED` y `NUXT_PUBLIC_REGISTRATION_ENABLED`) que pueden divergir.
- La fuente Outfit se descarga de Google Fonts; si el despliegue debe operar sin terceros, deberá autoalojarse sin cambiar tipografía.
- El build muestra avisos de Node por patrones `exports` obsoletos en dependencias transitivas; no bloquean el build, pero deben vigilarse.
- No existe CI/CD ni lint frontend separado. Typecheck, tests y build pasan, pero la cobertura funcional es todavía baja.
- La política de borrado debe mantenerse como baja lógica; no se debe añadir endpoint de borrado de usuarios.

## Decisiones funcionales y recomendaciones

| Decisión | Recomendación razonada | Estado |
| --- | --- | --- |
| CI obligatorio y único | CI opcional persistido como `NULL`; `00000000` reservado para presentación; CI real normalizado y único. | RESUELTA en el Módulo 4. |
| Vehículos sin placa | Permitir `placa = NULL` y usar el ID interno del vehículo como referencia temporal; aplicar índice único parcial solo a placas reales normalizadas. | RESUELTA en el Módulo 4. |
| Escritura “Raibanizada” | No corregir ni crear seeds hasta que el negocio confirme la etiqueta comercial exacta. La interfaz debe mostrar la marca/nombre aprobado, aunque internamente se use un slug estable. | PENDIENTE antes del Módulo 2. |
| Precio cero | Bloqueado (`precio > 0`). La matriz comienza sin tarifas y el Administrador carga cada importe; una ausencia se representa como hueco, no como `0.00`. | RESUELTA en el Módulo 3. |
| Transiciones del Personal | Solo personal activo y asignado: `PENDIENTE → EN_PROCESO` y `EN_PROCESO → CONTROL_CALIDAD`; ninguna otra transición. | RECOMENDADA; aprobar antes del Módulo 5. |
| Aprobación de calidad | Administrador o Encargado ejecutan `CONTROL_CALIDAD → TERMINADO`; el Personal no aprueba su propio trabajo. | RECOMENDADA; aprobar antes del Módulo 5. |
| Momento del pago | Cobrar dentro de la operación atómica `TERMINADO → ENTREGADO`; sin pago válido no se entrega. | RECOMENDADA; aprobar antes del Módulo 6. |
| Métodos de pago | Confirmar con el negocio una lista cerrada inicial; candidatos a validar, no aprobados: Efectivo, QR y Transferencia. No crear seeds ni enum hasta confirmación. | PENDIENTE antes del Módulo 6. |
| Corrección/anulación | Nunca borrar: conservar un pago único por orden y un registro de auditoría con motivo, actor y fecha para anular/corregir; una orden no queda entregada con pago anulado. El detalle de re-registro debe cerrarse al diseñar el módulo. | PENDIENTE antes del Módulo 6. |
| Reportes financieros del Encargado | Solo reportes operativos; puede registrar/ver el pago de la orden que opera, pero no recibe agregados financieros generales. | RESUELTA en el Módulo 7. |
| Zona horaria | `America/La_Paz` para presentación y límites del día; persistencia en UTC. | RESUELTA por los datos del trabajo. |

Decisiones adicionales que bloquean el Módulo 1:

1. **Identificador de acceso — RESUELTA:** correo electrónico normalizado (`trim` + minúsculas) como login único visible.
2. **Migración de `user` — RESUELTA EN POLÍTICA:** la migración se detiene si encuentra el rol `user`; su clasificación como Encargado o Personal de lavado debe ser explícita. La base configurada contiene una fila pendiente de esa decisión.
3. **Primer Administrador — RESUELTA:** conservar el Administrador ya definido en `.env`; el seed idempotente queda limitado a esa cuenta, requiere confirmación y está bloqueado en producción. Todas las demás altas son administrativas.

## Contrato visual

### Referencias obligatorias

- Login/formulario: `frontend/pages/login.vue`.
- Listado administrativo, búsqueda, tabla, paginación, modal y confirmación: `frontend/pages/admin/users/index.vue`.
- Tarjetas/resumen: `frontend/pages/admin/index.vue`.
- Estructura global: `frontend/components/app/AppShell.vue`.
- Carga y vacío: `frontend/components/ui/LoadingState.vue` y `EmptyState.vue`.
- Permiso denegado: `frontend/pages/403.vue`.

### Tokens reales

| Familia | Valores |
| --- | --- |
| Tipografía | Outfit 400/500/600/700/800; fallback Inter/system sans. Base 16 px; títulos con `line-height: 1.08` y tracking `-0.025em`. |
| Marca | Primario `#bafe06`, hover/secundario `#aaca02`, activo/acento `#7d9505`, acento oscuro `#333d04`. |
| Superficies | Fondo `#080c0f`, fondo profundo `#010100`, superficie `#111517`, elevada `#181c1d`. |
| Texto/borde | Texto `#f8f9f2`, muted `#b9baba`, disabled `#757b7a`, borde `#303432`. |
| Estados | Success `#8ee55d`, warning `#f5c451`, danger `#ff7474`, info `#78c8ff`, cada uno con fondo semitransparente. |
| Espaciado | Escala `0.25, 0.5, 0.75, 1, 1.25, 1.5, 2, 2.5, 3rem`. |
| Radios | `0.625, 0.875, 1.25, 1.625rem`; pills a `999px`. |
| Sombras | `--shadow-sm`, `--shadow-md`, `--shadow-glow`; no crear sombras ad hoc repetidas. |
| Layout | Contenido máximo `80rem`, sidebar `17rem`, header `4.5rem`; ancho mínimo 20rem. |
| Breakpoints | Escritorio desde `64rem`; adaptación del auth a `56rem`; móvil hasta `40rem`. |
| Movimiento | 160/240 ms y anulación global con `prefers-reduced-motion`. |

### Reglas de reutilización

- Mantener tema oscuro, paleta lima, Outfit, logo configurable, sidebar y header.
- Usar PrimeVue existente para Button, InputText, Password, Select, DataTable, Card, Dialog, ConfirmDialog, Message, Tag, Toast, ProgressSpinner y Paginator.
- Usar `.page-heading`, `.toolbar`, `.table-wrap`, `.form-field`, `.detail-grid`, `.stack` y `.cluster` antes de crear nuevas abstracciones.
- Conservar objetivos táctiles de al menos 44 px, foco visible, labels explícitos, `skip-link` y texto/icono junto al color de estado.
- Tablas deben vivir en `.table-wrap` para scroll horizontal; en móvil, barras y formularios se apilan.
- Cada pantalla debe incluir carga, vacío, error y permiso denegado; acciones sensibles requieren confirmación y bloqueo de doble envío.
- No introducir otra librería UI, otra paleta o estilos globales paralelos sin autorización.

### Método de verificación por módulo

1. Comparar la pantalla nueva con la referencia equivalente anterior.
2. Ejecutar test, typecheck y build frontend.
3. Inspeccionar a 320/360 px, 640 px, 896 px y 1024 px o más; comprobar overflow, foco, teclado, modales y textos largos.
4. Registrar capturas o evidencia de preview cuando haya navegador disponible.

En esta fase la inspección interactiva fue **NO EJECUTADA**: el entorno no expuso ningún navegador controlable. La inspección estática de Vue/CSS sí fue completada y el build pasó.

## Verificaciones de la Fase 0

| Comando/comprobación | Resultado |
| --- | --- |
| `go test ./...` | APROBADO. |
| `go vet ./...` | APROBADO. |
| `gofmt -l .` | APROBADO, sin archivos reportados. |
| `npm test` | APROBADO, 3/3. |
| `npm run typecheck` | APROBADO. |
| `npm run build` | APROBADO con dos avisos deprecados de dependencias transitivas. |
| `npm audit --omit=dev --audit-level=high` | APROBADO, 0 vulnerabilidades reportadas. |
| `docker compose config --quiet` (dev/prod) | APROBADO; advertencias de acceso al config local de Docker. |
| Docker/DB limpia | NO EJECUTADA: daemon Docker no disponible. |
| `govulncheck ./...` | NO EJECUTADA: herramienta no instalada. |
| Preview visual interactivo | NO EJECUTADA: no había navegador controlable. |
