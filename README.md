# Camell Wash

Plantilla full-stack reutilizable para iniciar aplicaciones con autenticación, perfiles y administración genérica de usuarios. Conserva el stack del proyecto de referencia —Go, Nuxt/Vue, PostgreSQL y Docker— y elimina por completo su dominio comercial.

## Arquitectura

```text
Browser
  └─ Nuxt 4 / Vue 3 / Pinia / PrimeVue
       └─ API JSON Go 1.25 / Gorilla Mux / GORM
            └─ PostgreSQL 16
```

- `frontend/`: interfaz SSR, layouts `/app` y `/admin`, guards cliente y tema centralizado.
- `backend/cmd/api`: servidor HTTP.
- `backend/cmd/migrate`: migraciones SQL versionadas e idempotentes.
- `backend/cmd/seed`: dos usuarios configurables, solo para desarrollo/pruebas.
- `backend/internal/httpapi`: composición explícita de rutas y middleware.
- `backend/internal/handlers`: autenticación, perfil y administración.
- `backend/internal/middleware`: sesión, CSRF, RBAC, rate limiting, logging y cabeceras.
- `backend/internal/database/migrations`: fuente de verdad del esquema.
- `nginx/`: reverse proxy de producción con servicios internos no publicados.

## Requisitos

- Go 1.25.12 o compatible con el `go.mod`.
- Node.js 22 y npm.
- PostgreSQL 16 o 17.
- Docker Compose, opcional pero recomendado.

## Instalación rápida con Docker

1. Crea la configuración local:

   ```powershell
   Copy-Item .env.example .env
   ```

2. Sustituye en `.env` al menos `DB_PASSWORD` y `SESSION_SECRET`. Usa valores aleatorios; `SESSION_SECRET` debe tener 32 caracteres o más.

3. Inicia los servicios:

   ```powershell
   docker compose up --build
   ```

La base se publica únicamente en `127.0.0.1`, igual que frontend y backend. En desarrollo, el backend aplica las migraciones al arrancar porque `DB_AUTO_MIGRATE=true` en Compose.

## Desarrollo local

Primero crea la base PostgreSQL indicada por `DB_NAME` y copia `.env.example` como `.env` en la raíz. El backend carga `../.env` cuando se ejecuta desde su carpeta.

Backend:

```powershell
Set-Location backend
go mod download
go run ./cmd/migrate
go run ./cmd/api
```

Frontend, en otra terminal:

```powershell
Set-Location frontend
npm ci
npm run dev
```

Direcciones predeterminadas:

- Frontend: `http://localhost:3000`
- API: `http://localhost:8080/api/v1`
- Salud: `http://localhost:8080/health`

## Variables de entorno

La lista completa y segura está en [`.env.example`](./.env.example). No se debe versionar `.env`.

| Grupo | Variables principales | Propósito |
| --- | --- | --- |
| App | `APP_ENV`, `APP_NAME`, `APP_SHORT_NAME`, `LOG_LEVEL` | Entorno, nombre y nivel de logs |
| Frontend | `NUXT_PUBLIC_API_BASE`, `FRONTEND_PORT` | URL pública de API y puerto local |
| Backend | `SERVER_HOST`, `SERVER_PORT`, `BACKEND_PORT`, `ALLOWED_ORIGINS`, `TRUST_PROXY` | Escucha HTTP, CORS y proxies confiables |
| Base de datos | `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`, `DB_AUTO_MIGRATE` | Conexión y estrategia de migración |
| Sesiones | `SESSION_SECRET`, `SESSION_TTL`, `SESSION_IDLE_TIMEOUT` | Firma CSRF y vencimientos absoluto/inactivo |
| Cookies | `AUTH_COOKIE_SECURE`, `AUTH_COOKIE_SAME_SITE` | Seguridad de la cookie HttpOnly |
| Contraseñas | `PASSWORD_BCRYPT_COST`, `LOGIN_MAX_ATTEMPTS`, `LOGIN_LOCKOUT` | Hashing y bloqueo contra fuerza bruta |
| Límites | `AUTH_RATE_LIMIT_PER_MINUTE`, `API_RATE_LIMIT_PER_MINUTE` | Límites por IP y proceso |
| Seed | `SEED_DEVELOPMENT`, `DEV_ADMIN_EMAIL`, `DEV_ADMIN_PASSWORD` | Administrador inicial no productivo |

La aplicación falla al arrancar si falta una variable crítica o si producción intenta usar una cookie no segura, un secreto corto/placeholder, CORS comodín o una combinación `SameSite=None` sin `Secure`.

## Base de datos y migraciones

El esquema se crea únicamente desde SQL versionado. No depende de cambios manuales ni de `AutoMigrate` de GORM.

```powershell
Set-Location backend
go run ./cmd/migrate
```

El comando puede ejecutarse repetidamente: registra versiones en `schema_migrations` y aplica cada archivo una sola vez dentro de una transacción. Para agregar una migración, crea el siguiente archivo secuencial en `backend/internal/database/migrations`, por ejemplo `000002_feature.sql`.

En producción, el servicio `migrate` de `docker-compose.prod.yml` debe terminar correctamente antes de que el backend inicie.

## Administrador inicial de desarrollo

El seed crea únicamente el Administrador definido por `DEV_ADMIN_EMAIL` y `DEV_ADMIN_PASSWORD`. Si la cuenta ya existe y es Administrador, no la duplica ni sustituye su contraseña:

```powershell
$env:SEED_DEVELOPMENT='true'
Set-Location backend
go run ./cmd/seed
```

Con Docker:

```powershell
docker compose run --rm backend /app/seed
```

El comando rechaza `APP_ENV=production`, exige confirmación mediante `SEED_DEVELOPMENT=true`, valida la política de contraseña y nunca contiene contraseñas codificadas en el repositorio. Las demás cuentas se crean exclusivamente desde la administración web.

## Autenticación

No se utilizan JWT ni tokens en `localStorage`/`sessionStorage`.

1. Login genera 256 bits aleatorios.
2. El valor crudo vive solo en una cookie `HttpOnly`, limitada a `/api/v1`.
3. PostgreSQL almacena únicamente SHA-256 del token.
4. El frontend conserva en memoria un token CSRF HMAC; toda mutación autenticada debe enviarlo en `X-CSRF-Token`.
5. La sesión tiene expiración absoluta e inactividad deslizante.
6. Logout, cambio de contraseña, cambio de rol y desactivación revocan sesiones en servidor.

La cookie usa `Secure=true` obligatoriamente en producción. `SameSite=lax` es el valor recomendado para un frontend y API del mismo sitio.

## Usuarios y roles

- `admin` — Administrador: crea y edita usuarios, asigna roles, activa/desactiva cuentas y accede a `/admin`.
- `manager` — Encargado: accede a `/app` y a su perfil; no puede consultar ni mutar administración de usuarios.
- `wash_staff` — Personal de lavado: accede a `/app` y a su perfil; no puede consultar ni mutar administración de usuarios.

El correo recortado y convertido a minúsculas es el login único. No existe registro público: solo un Administrador crea cuentas. Cada usuario puede cambiar su propia contraseña; todas las rutas `/admin/*` pasan por autenticación y `RequireRoles("admin")`. Un Administrador tampoco puede cambiar su propio rol o estado desde la tabla administrativa.

La migración `000002_lavadero_roles.sql` clasifica explícitamente `user@example.test` como Encargado (`manager`), según la decisión aprobada. Se detiene si encuentra cualquier otra cuenta con el rol legado `user`; nunca aplica una clasificación general silenciosa.

## Rutas base

Frontend:

```text
/login
/app
/app/clients
/app/orders
/app/profile
/admin
/admin/users
/admin/catalogs
/admin/prices
/admin/clients
/admin/orders
/403
/cualquier-ruta-invalida  (404)
```

API:

```text
GET    /health
POST   /api/v1/auth/login
GET    /api/v1/auth/session
POST   /api/v1/auth/logout
GET    /api/v1/me
PATCH  /api/v1/me
POST   /api/v1/me/password
GET    /api/v1/admin/dashboard
GET    /api/v1/reports/operations?date_from=YYYY-MM-DD&date_to=YYYY-MM-DD
GET    /api/v1/admin/reports/financial?date_from=YYYY-MM-DD&date_to=YYYY-MM-DD
POST   /api/v1/admin/users
GET    /api/v1/admin/users
GET    /api/v1/admin/users/:id
PATCH  /api/v1/admin/users/:id
PATCH  /api/v1/admin/users/:id/role
PATCH  /api/v1/admin/users/:id/status
GET    /api/v1/admin/vehicle-types
POST   /api/v1/admin/vehicle-types
GET    /api/v1/admin/vehicle-types/:id
PATCH  /api/v1/admin/vehicle-types/:id
PATCH  /api/v1/admin/vehicle-types/:id/status
GET    /api/v1/admin/service-categories
POST   /api/v1/admin/service-categories
GET    /api/v1/admin/service-categories/:id
PATCH  /api/v1/admin/service-categories/:id
PATCH  /api/v1/admin/service-categories/:id/status
GET    /api/v1/admin/services
POST   /api/v1/admin/services
GET    /api/v1/admin/services/:id
PATCH  /api/v1/admin/services/:id
PATCH  /api/v1/admin/services/:id/status
GET    /api/v1/catalogs/vehicle-types
GET    /api/v1/catalogs/service-categories
GET    /api/v1/catalogs/services
GET    /api/v1/admin/service-prices/matrix
POST   /api/v1/admin/service-prices
PATCH  /api/v1/admin/service-prices/{service_id}/{vehicle_type_id}
GET    /api/v1/pricing/current
GET    /api/v1/clients
POST   /api/v1/clients
GET    /api/v1/clients/{id}
PATCH  /api/v1/clients/{id}
GET    /api/v1/clients/{id}/vehicles
POST   /api/v1/clients/{id}/vehicles
GET    /api/v1/vehicles/{id}
PATCH  /api/v1/vehicles/{id}
GET    /api/v1/clients/{id}/orders
GET    /api/v1/vehicles/{id}/orders
GET    /api/v1/orders
POST   /api/v1/orders
GET    /api/v1/orders/{id}
PATCH  /api/v1/orders/{id}
POST   /api/v1/orders/{id}/transitions
POST   /api/v1/orders/{id}/cancel
GET    /api/v1/orders/{id}/payments
POST   /api/v1/orders/{id}/payments
POST   /api/v1/payments/{id}/void
POST   /api/v1/payments/{id}/correct
GET    /api/v1/admin/payment-reviews
POST   /api/v1/admin/payment-reviews/{id}
GET    /api/v1/operations/wash-staff
GET    /api/v1/operations/service-options
```

## Catálogos

Tipos de vehículo, categorías y servicios son extensibles y usan baja lógica. Los nombres se recortan, colapsan espacios y se comparan en minúsculas; tipos/categorías son únicos globalmente y servicios dentro de cada categoría. Los endpoints `/admin/*` son exclusivos del Administrador. Los tres roles pueden consultar `/catalogs/*`, que devuelve únicamente opciones activas aptas para formularios operativos.

No existen endpoints de borrado físico. La migración `000003_catalogs.sql` usa `ON DELETE RESTRICT` y constraints únicos. `000004_catalog_initial_values.sql` carga los valores iniciales aprobados con `ON CONFLICT DO NOTHING`. Son registros ordinarios: pueden agregarse otros, renombrarse, cambiarse de categoría o activarse/desactivarse desde la interfaz.

## Precios

La matriz `/admin/prices` cruza servicios y tipos de vehículo activos. Comienza sin tarifas precargadas y muestra cada hueco como `Sin precio`; solo el Administrador crea o actualiza importes. `service_prices.price` usa `NUMERIC(12,2)`, exige un valor mayor que cero y la API transporta el decimal como texto para no perder precisión.

La combinación servicio/tipo es única. Una tarifa de catálogo inactivo se conserva, deja de ser vigente y vuelve a estar disponible al reactivar la combinación. Administrador y Encargado pueden consultar `/pricing/current`; Personal de lavado no accede a información financiera.

## Clientes y vehículos

Administrador y Encargado gestionan clientes y varios vehículos desde `/admin/clients` y `/app/clients`. El CI y la placa pueden quedar ausentes: PostgreSQL conserva `NULL`, mientras el CI se representa como `00000000` en pantalla/reporte y un vehículo sin placa usa temporalmente su ID interno. `00000000` no puede registrarse como CI real.

Los CI y placas reales son únicos por su forma canónica. El listado de clientes enmascara CI/teléfono y la ficha autorizada muestra el dato completo. En esta versión el propietario de un vehículo no puede cambiarse y no existen rutas de borrado físico.

## Órdenes de lavado

Administrador y Encargado crean y gestionan órdenes desde `/admin/orders` y `/app/orders`; Personal de lavado consulta únicamente las órdenes donde está asignado. Cada orden congela el nombre y precio de sus servicios, calcula el total en servidor y registra cada transición en un historial append-only.

El flujo implementado llega hasta `DELIVERED`: la entrega exige un pago válido y puede realizarse junto con el cobro al retirar el vehículo o después de un pago anticipado. La edición comercial y la cancelación quedan bloqueadas cuando existe un pago válido, y `version` evita que dos solicitudes concurrentes apliquen la misma transición.

## Pagos y entrega

Administrador y Encargado pueden registrar el pago total al dejar el vehículo o al retirarlo. El backend toma el monto desde la orden; no existen importes editables, anticipos ni pagos parciales. Los métodos admitidos son efectivo, QR, transferencia y tarjeta, con referencia externa opcional.

El pago temprano no cambia el estado operativo. Al retirar, el cobro y `FINISHED → DELIVERED` pueden ejecutarse en una sola transacción; una orden ya pagada también puede entregarse posteriormente. La idempotencia y el índice de pago válido evitan duplicados por doble clic o concurrencia.

Una anulación del Encargado queda pendiente en `/admin/payment-reviews`. El Administrador la revisa después sin detener el trabajo. Los pagos entregados solo se corrigen mediante reemplazo atómico; no se implementaron devoluciones reales.

## Dashboard y reportes

El dashboard diario y los reportes por periodo usan fechas inclusivas de `America/La_Paz`. Órdenes recibidas se cuentan por recepción; entregas y cancelaciones por su transición; el ciclo mide recepción a entrega. Los ingresos suman únicamente pagos `VALID` por fecha de pago y las anulaciones se presentan aparte por fecha de anulación.

Administrador y Encargado consultan operación y exportan CSV. Solo el Administrador recibe ingresos, ticket promedio, distribución por método y anulaciones agregadas. Personal de lavado no accede a agregados. Las consultas usan índices de la migración 10 y no persisten contadores duplicados.

## Tema visual

La fuente de verdad es [`frontend/assets/theme.css`](./frontend/assets/theme.css).

Para cambiar colores, edita una vez:

```css
--color-primary          /* #FF5500: CTA, enlaces e indicadores */
--color-secondary        /* #E64D00: degradados */
--color-accent           /* #B83D00: brillo y profundidad */
--color-accent-dark      /* #3B1C0F: superficies teñidas */
--color-background       /* #080C0F: fondo principal */
--color-background-deep  /* #010100: secciones profundas */
--color-surface          /* #111517: cards */
--color-surface-elevated /* #181C1D: inputs y elevación */
--color-text             /* #F8F9F2: contenido principal */
--color-text-muted       /* #B9BABA: contenido secundario */
--color-text-disabled    /* #757B7A: contenido desactivado */
--color-border           /* #303432: bordes y separadores */
--color-success
--color-warning
--color-danger
```

La plantilla utiliza una identidad oscura única y la clase raíz `.app-dark` permite que PrimeVue consuma su variante oscura. Los componentes consumen tokens semánticos y no contienen paletas de cliente dispersas.

Para cambiar fuentes, edita:

```css
--font-primary
--font-secondary
--font-size-base
--font-weight-regular
--font-weight-medium
--font-weight-semibold
--font-weight-bold
--font-weight-extrabold
```

La fuente configurada es Outfit y se carga en `frontend/nuxt.config.ts` únicamente en los pesos 400, 500, 600, 700 y 800. Si agregas fuentes propias, define `@font-face` en el mismo archivo y mantén los archivos bajo `frontend/public/fonts/`.

## Branding

- Nombre completo: `APP_NAME` / `NUXT_PUBLIC_APP_NAME`.
- Nombre corto: `APP_SHORT_NAME` / `NUXT_PUBLIC_APP_SHORT_NAME`.
- Logo/favicon principal: `frontend/public/branding/camell-wash-logo.png` (archivo oficial proporcionado por el propietario).
- Colores y fuentes: `frontend/assets/theme.css`.

El logo oficial de Camell Wash se conserva sin modificaciones de contenido. Como ya incluye el nombre, la interfaz no agrega un rótulo duplicado junto a la imagen.

## Componentes reutilizables

PrimeVue cubre Button, Input, Password, Select, Dialog, ConfirmDialog, Card, DataTable, Tag, Message, Toast, Spinner y Paginator. Se añadieron `AppShell`, `EmptyState` y `LoadingState` solo donde existe valor común real. Esto evita mantener duplicados de la librería UI.

## Calidad y pruebas

```powershell
# Backend
Set-Location backend
go test ./...
go vet ./...
go build ./cmd/api
go build ./cmd/migrate
go build ./cmd/seed

# Frontend
Set-Location ../frontend
npm test
npm run typecheck
npm run build
npm audit --omit=dev

# Integración real del Módulo 1 en una base temporal aislada
Set-Location ..
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-module1.ps1

# Integración real del Módulo 2: catálogos, permisos y baja lógica
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-module2.ps1

# Integración real del Módulo 3: precisión, permisos, faltantes y concurrencia
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-module3.ps1

# Integración real del Módulo 4: privacidad, CI/placa nulos, permisos y propietario inmutable
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-module4.ps1

# Integración real del Módulo 5: órdenes, snapshots, asignaciones, estados y concurrencia
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-module5.ps1

# Integración real del Módulo 6: pagos, entrega, idempotencia, anulación y revisión
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-module6.ps1

# Integración real del Módulo 7: reportes, fórmulas, fechas, permisos e índices
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-module7.ps1
```

No hay script de lint separado; `go vet`, `gofmt` y el typecheck estricto cubren las comprobaciones estáticas disponibles en este starter.

## Producción

1. Usa secretos aleatorios externos al repositorio.
2. Configura `APP_ENV=production`, orígenes HTTPS exactos y `NUXT_PUBLIC_API_BASE` público.
3. Mantén `AUTH_COOKIE_SECURE=true`; el Compose de producción lo fuerza.
4. Termina TLS en un balanceador/reverse proxy o amplía Nginx con certificados administrados. No publiques HTTP directo a Internet.
5. Usa `DB_SSLMODE=require` cuando PostgreSQL esté fuera de la red privada.
6. No expongas un endpoint de registro público; las altas son administrativas.
7. Cambia el rate limiter en memoria por Redis si despliegas varias réplicas.
8. Ejecuta backups, monitorización y rotación de secretos según tu infraestructura.

Comando base:

```powershell
docker compose -f docker-compose.prod.yml up --build -d
```

PostgreSQL, backend y frontend no publican puertos en el Compose de producción; solo Nginx expone el puerto de entrada.

## Crear un proyecto nuevo desde esta base

1. Copia esta carpeta sin `.git`, `.env`, `node_modules`, `.nuxt`, `.output` ni binarios.
2. Crea un `.env` nuevo desde `.env.example` y genera secretos nuevos.
3. Cambia nombre, logo, favicon, colores y fuentes en los puntos descritos arriba.
4. Conserva `users`, `auth_sessions`, `/me` y RBAC como infraestructura.
5. Agrega cada nuevo dominio en paquetes backend y páginas/componentes frontend propios; no mezcles reglas comerciales dentro de autenticación.
6. Añade migraciones secuenciales y tests de autorización para cada endpoint nuevo.
7. Inicializa un repositorio Git nuevo y ejecuta toda la batería de calidad.

## Alcance deliberadamente excluido

La plantilla no inventa correo transaccional, verificación de email, recuperación de contraseña, MFA, almacenamiento de archivos ni permisos granulares. Esas capacidades requieren decisiones de proveedor y producto. Añádelas cuando el nuevo sistema tenga requisitos concretos.
