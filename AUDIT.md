# Auditoría del proyecto de referencia y construcción del starter

## Alcance y método

El proyecto de referencia se inspeccionó en modo de solo lectura: árbol de archivos, rutas, modelos, configuración, migración/bootstrap, seeds, Docker, Nginx, frontend, tests y patrones de seguridad. Se excluyeron de la huella de integridad `.git`, dependencias, builds, uploads, el archivo comprimido y binarios generados.

Huella SHA-256 agregada inicial del código/configuración relevante del origen:

```text
495432d218011e5c9e306f909e3ca3079961190591d3ba0066e4f89bbc9894f2
```

La misma huella se verificó al finalizar. El único elemento no versionado ya presente en el origen era su propio archivo `.rar`; esta tarea no lo creó ni modificó.

## Arquitectura detectada

- Backend Go 1.24.3, Gorilla Mux, GORM, PostgreSQL y validación `go-playground/validator`.
- Frontend Nuxt 4, Vue 3, Pinia, PrimeVue y Tailwind.
- Sesiones opacas persistidas en PostgreSQL; cookie HttpOnly y protección CSRF HMAC.
- Docker multi-stage para frontend/backend, PostgreSQL 16 y Nginx.
- WebSocket con PostgreSQL LISTEN/NOTIFY, correo SMTP/OTP, uploads y generación PDF.
- Cinco roles ligados al dominio: administrador y cuatro perfiles comerciales/operativos.
- Migración principal con GORM `AutoMigrate` y un bootstrap SQL comercial grande embebido en Go; el directorio de migraciones montado por Docker no contenía migraciones versionadas.

## Problemas encontrados

### Críticos/altos

1. **Autorización inconsistente.** Muchas rutas comerciales estaban bajo autenticación genérica, pero no bajo middleware de rol. Algunos handlers hacían comprobaciones internas y otros dependían de IDs del path. Esto aumenta el riesgo de IDOR y escalamiento horizontal/vertical.
2. **Detalles internos expuestos.** Numerosas respuestas enviaban `err.Error()` al cliente en errores de base de datos, validación, archivos y exportación; podían revelar SQL, constraints y estructura interna.
3. **Tokens WebSocket en query string.** El flujo documentaba autenticación por token en URL, con riesgo de filtración en logs, historiales y proxies.
4. **Bloqueo de login incoherente.** Tras cinco fallos se asignaba `status=suspended`; el login rechazaba ese estado antes de poder ejecutar la rama que lo reactivaba, por lo que el bloqueo temporal podía volverse permanente sin intervención administrativa.
5. **Migraciones toleradas al fallar.** El backend registraba fallos de migración/bootstrap/seed y continuaba iniciando, permitiendo una instancia parcialmente operativa con esquema incorrecto.

### Medios

1. **Defaults inseguros.** Configuración incluía contraseña PostgreSQL `1234`, nombre de base comercial y secreto de sesión placeholder como fallbacks de código.
2. **Política de contraseña débil.** Los DTO solo exigían ocho caracteres y el cambio de contraseña dejaba activas las sesiones de otros dispositivos.
3. **Rate limiting evadible.** Se confiaba siempre en `X-Forwarded-For` y, sin proxy, `RemoteAddr` incluía el puerto efímero. Un cliente podía falsear la clave o producir una clave nueva por conexión.
4. **Logging excesivo.** GORM estaba en nivel `Info` y existían múltiples logs de payload/estado en frontend y WebSocket. Aunque no se observó logging explícito de contraseñas, la práctica ampliaba la superficie de datos.
5. **Errores de arranque dispersos.** Parte de la validación vivía en `main`, parte en `config` y parte usaba defaults silenciosos.
6. **Uploads servidos ampliamente.** El backend exponía todo `uploads/` mediante `http.FileServer`; la separación público/privado dependía de disciplina de carpetas y Nginx solo bloqueaba algunas extensiones ejecutables.
7. **Entorno versionable.** `.gitignore` excluía `.env` pero permitía explícitamente `.env.production`, una ruta frecuente para filtrar secretos o direcciones reales.
8. **Desarrollo exponía PostgreSQL.** Compose publicaba la base sin limitar explícitamente la interfaz de escucha.

### Mantenibilidad y frontend

1. El dominio turístico cruzaba modelos, handlers, SQL, servicios, layouts, páginas, assets, rutas y navegación; no era extraíble por configuración.
2. Había términos, imágenes, logos, fuentes y textos comerciales en decenas de archivos.
3. Existían colores hex/RGBA y sombras ad hoc en componentes y páginas pese a contar con un `theme.css` parcial.
4. Los guards Nuxt eran solo cliente. Esto es válido para experiencia de navegación, pero no sustituye la autorización backend; en el origen esa segunda capa era desigual.
5. Había dependencias amplias para mapas, charts, PDF, Excel, carruseles, parallax, animación, fechas, WebSocket y persistencia que no pertenecen a un starter mínimo.
6. El árbol contenía `node_modules`, `.output`, un binario backend, uploads y un archivo comprimido grande; son artefactos que no deben formar parte de una plantilla.

## Aspectos positivos conservados

- Token de sesión aleatorio y opaco en vez de JWT innecesario.
- Hash del token persistido, no token crudo.
- Cookie HttpOnly con `Secure` y `SameSite` configurables.
- CSRF con HMAC y comparación constante.
- Hash de contraseñas con bcrypt.
- Respuestas JSON con forma común.
- CORS con lista exacta y credenciales.
- Docker multi-stage y usuarios no-root en imágenes de aplicación.
- Health check y pruebas unitarias iniciales.
- PrimeVue como base adecuada para componentes accesibles y reutilizables.

## Diseño de la nueva plantilla

Se conservó el stack sin migraciones de framework, lenguaje, ORM o base de datos. Se eligió una arquitectura pequeña por capas:

```text
HTTP router
  ├─ middleware transversal
  ├─ auth handlers
  ├─ profile handler
  └─ admin handler
        └─ GORM / PostgreSQL / SQL migrations
```

No se agregó repository/service/use-case por entidad porque, con solo dos tablas y operaciones simples, aumentaría la ceremonia sin reducir acoplamiento real. Los nuevos dominios pueden incorporar servicios cuando aparezca lógica propia.

## Cambios realizados

### Backend

- Reescritura mínima con `users` y `auth_sessions`.
- Roles cerrados `admin/user` y estados `active/inactive`, protegidos también con constraints SQL.
- Registro público forzado a `user` y deshabilitable.
- Endpoints `/me` sin ID controlado por cliente para eliminar IDOR del perfil.
- Grupo `/admin` protegido con middleware RBAC en backend.
- Sesiones opacas de 256 bits, cookie HttpOnly, hash SHA-256 en DB, vencimiento absoluto/inactivo y CSRF HMAC.
- Revocación de todas las sesiones en cambio de contraseña/rol y al desactivar.
- Bcrypt con coste configurable (10–14) y política de 12–72 bytes.
- Bloqueo temporal de intentos sin cambiar el estado permanente de la cuenta.
- Bodies JSON limitados a 1 MiB, campos desconocidos rechazados y validación backend.
- Respuestas internas sanitizadas; los detalles técnicos quedan solo en logs.
- Logs estructurados `slog`, sin bodies, cookies, tokens ni contraseñas.
- Rate limiting por IP con `TRUST_PROXY` explícito; no se acepta `X-Forwarded-For` por defecto.
- CORS exacto, security headers, recovery de pánicos, request ID y timeouts HTTP.
- Configuración central validada al iniciar y sin credenciales fallback.
- Migraciones SQL embebidas, transaccionales, versionadas e idempotentes.
- Seed explícito, configurable y bloqueado en producción.
- Comandos separados `api`, `migrate` y `seed`.

### Frontend

- Rutas genéricas `/login`, `/register`, `/app`, `/app/profile`, `/admin`, `/admin/users`, `/403` y 404 global.
- Store Pinia guarda usuario/CSRF solo en memoria; las credenciales no usan almacenamiento web.
- Guards separados `auth`, `user`, `admin` y `guest`.
- Layout común responsive con sidebar, header, menú, perfil y logout.
- Formularios con Zod, errores inline/general, carga y prevención de doble envío.
- Tabla administrativa, paginación, empty/loading states y confirmación para cambios sensibles.
- Reutilización directa de PrimeVue para no duplicar Button/Input/Select/Dialog/Card/Table/Tag/Message/Toast/Spinner/Paginator.
- Tokens semánticos centralizados para color, tipografía, estado, espacio, radio, sombra y movimiento.
- Tema oscuro premium con paleta verde centralizada, foco visible, objetivos táctiles de 44 px, skip link y `prefers-reduced-motion`.
- Branding central mediante runtime config y dos assets SVG genéricos.

### Infraestructura

- `.gitignore` y `.dockerignore` endurecidos para secretos, certificados, logs, uploads, backups y builds.
- Dockerfiles multi-stage y runtime no-root.
- Compose de desarrollo publica servicios solo en loopback.
- Compose de producción publica únicamente Nginx; DB/backend/frontend permanecen internos.
- Servicio de migración obligatorio antes del backend de producción.
- Nginx con límites, CSP, headers, tamaño de body y ocultación de versión.

## Código eliminado/no trasladado

- Agencias, atracciones, paquetes, salidas, compras, pagos, ventas, promociones y reportes.
- Roles y layouts comerciales/operativos del proyecto de referencia.
- WebSocket, notificaciones y PostgreSQL LISTEN/NOTIFY.
- SMTP, OTP, verificación de email y recuperación de contraseña.
- Uploads, comprobantes, galerías, mapas, PDF/Excel y archivos reales.
- Seeds geográficos/comerciales, funciones/triggers SQL de negocio y datos de Bolivia.
- Footer/páginas comerciales, logos, favicon, fuente e imágenes de la marca original.
- Dependencias Go de WebSocket, pgx directo y PDF.
- Dependencias frontend de charts, fechas, GSAP, PDF, mapas, Excel, carruseles, lazy/parallax, vee-validate y persisted state.
- Builds, dependencias instaladas, binarios, archivos subidos, backups y comprimidos del origen.

## Verificación realizada

### Proyecto original

- `go test ./...`: correcto.
- `go vet ./...`: correcto.
- tests frontend: 2/2 correctos.
- `npm audit --offline --omit=dev`: 0 vulnerabilidades conocidas.

### Starter

- `gofmt`: aplicado.
- `go test ./...`: correcto.
- `go vet ./...`: correcto.
- build `api`, `migrate`, `seed`: correcto.
- tests frontend: 3/3 correctos.
- `nuxt typecheck`: correcto.
- `nuxt build`: correcto.
- `npm audit --audit-level=high`: 0 vulnerabilidades conocidas.
- `govulncheck -show verbose ./...`: 0 vulnerabilidades en símbolos o paquetes importados. Se actualizaron `pgx`, `x/text`, `x/crypto` y `x/sys` a versiones corregidas; esto elevó Go del starter a 1.25. Solo queda el aviso de módulo para `x/crypto/openpgp`, paquete no importado, no alcanzable y sin versión corregida; `x/crypto` se conserva para bcrypt.
- `docker compose config` en desarrollo y producción: correcto.
- PostgreSQL real temporal: migración inicial + segunda ejecución idempotente + seed: correcto.
- Batería HTTP real: health 200; anónimo 401; perfil user 200; user→admin 403; mutación sin CSRF 403; mutación válida 200; admin dashboard/listado y búsquedas escapadas 200; cambio de rol 200; sesión afectada revocada 401.
- Base y procesos temporales: eliminados tras las pruebas.
- Frontend SSR `/login`: respondió HTTP 200.

## Seguridad final

- Separación explícita autenticación/autorización.
- Defensa backend obligatoria para roles y propiedad.
- Sin secretos ni credenciales reales en el starter.
- Sin JWT innecesario ni credenciales persistidas en JavaScript.
- Protección CSRF, CORS exacto, cookie segura, expiración/revocación, fuerza bruta y rate limiting.
- Validación estricta, mass assignment evitado mediante DTOs y updates explícitos.
- Queries parametrizadas por GORM; búsqueda LIKE escapada.
- Sin uploads, SSRF, path traversal, WebSocket ni exposición de archivos en el alcance base.
- Errores sanitizados y logs sin datos de autenticación.

## Decisiones arquitectónicas

1. **Sesión opaca en lugar de JWT.** El sistema ya requiere PostgreSQL y necesita revocación inmediata; un token autosuficiente no aporta valor.
2. **SQL versionado en lugar de AutoMigrate.** Hace reproducible y auditable el esquema, sus constraints, índices y triggers.
3. **RBAC simple.** Dos constantes y middleware de grupo son suficientes; una matriz de permisos sería sobrearquitectura en este alcance.
4. **Sin flujo de email inventado.** Verificación y recuperación necesitan proveedor, remitente, UX y política definidos por cada producto.
5. **PrimeVue antes que wrappers propios.** Solo se envolvieron estados/layouts verdaderamente compartidos.
6. **Tokens CSS semánticos.** Permiten rebranding global sin acoplar componentes a una paleta concreta.

## Pendientes y límites conocidos

- Docker Desktop no estaba activo durante la verificación; se validaron ambos Compose con `docker compose config`, pero no se construyeron imágenes. Backend, frontend y PostgreSQL sí se ejecutaron directamente.
- El conector de navegador integrado falló antes de abrir la página por un error de metadatos de sandbox. Se verificó SSR y respuesta HTTP, pero no se declara una inspección visual automatizada completa.
- Nginx de ejemplo escucha HTTP para facilitar su colocación detrás de un terminador TLS. Antes de Internet público se debe configurar TLS real o un balanceador que lo termine.
- El rate limiter vive en memoria por proceso; varias réplicas requieren almacenamiento compartido como Redis.
- Recuperación de contraseña, verificación de correo, MFA, uploads y permisos finos deben diseñarse con requisitos/proveedores concretos; no se inventaron.
