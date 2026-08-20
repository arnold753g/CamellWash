# Módulo 1 — Usuarios, roles y autenticación

Estado: **IMPLEMENTADO Y MIGRADO; PENDIENTE DE REVISIÓN VISUAL INTERACTIVA Y APROBACIÓN**.

Fuente funcional: [`GUIA_UNICA_SISTEMA_LAVADERO.md`](../GUIA_UNICA_SISTEMA_LAVADERO.md).

## Decisiones aplicadas

- El correo normalizado con `trim` y minúsculas es el login único.
- Solo el Administrador crea cuentas; se eliminaron la ruta y la pantalla de registro público.
- Roles técnicos: `admin`, `manager`, `wash_staff`; etiquetas: Administrador, Encargado y Personal de lavado.
- El usuario puede actualizar su perfil y cambiar su propia contraseña mediante el flujo existente.
- El seed usa únicamente `DEV_ADMIN_EMAIL` y `DEV_ADMIN_PASSWORD`, exige `SEED_DEVELOPMENT=true`, es idempotente y se niega en producción.
- No se añadió Bootstrap, Tailwind ni otra librería UI; se conservaron PrimeVue, Outfit, los tokens de `theme.css` y `AppShell`.

## Migración

`000002_lavadero_roles.sql` elimina el default legado y cambia el constraint a los tres roles del negocio. Por decisión expresa, clasifica únicamente `user@example.test` como Encargado (`manager`). Si detecta cualquier otra fila `user`, lanza una excepción y revierte la transacción.

La prueba en base vacía aplicó las migraciones 1 y 2 dos veces y confirmó idempotencia. El contrato automatizado comprueba la clasificación específica y que el bloqueo para otras cuentas heredadas siga presente.

La migración 2 se aplicó a la base configurada el 2026-08-13. Conservó las dos identidades existentes, cambió `user@example.test` de `user` a `manager` y dejó cero filas con el rol legado.

## API implementada

| Método y ruta | Permiso | Función |
| --- | --- | --- |
| `POST /api/v1/auth/login` | Cuenta activa | Inicia sesión con correo y contraseña. |
| `GET /api/v1/auth/session` | Autenticado | Restaura usuario, CSRF y vencimientos. |
| `POST /api/v1/auth/logout` | Autenticado | Revoca la sesión. |
| `GET/PATCH /api/v1/me` | Autenticado | Consulta/edita el perfil propio. |
| `POST /api/v1/me/password` | Autenticado | Cambia contraseña y revoca todas las sesiones. |
| `POST /api/v1/admin/users` | Administrador | Crea una cuenta con rol y estado explícitos. |
| `GET /api/v1/admin/users` | Administrador | Lista con búsqueda, rol, estado y paginación. |
| `GET /api/v1/admin/users/:id` | Administrador | Consulta detalle. |
| `PATCH /api/v1/admin/users/:id` | Administrador | Edita nombre, apellido y correo. |
| `PATCH /api/v1/admin/users/:id/role` | Administrador | Cambia rol y revoca sesiones. |
| `PATCH /api/v1/admin/users/:id/status` | Administrador | Activa/desactiva; desactivar revoca sesiones. |

`POST /api/v1/auth/register` ya no existe y responde 404.

## Seguridad

- Contraseñas bcrypt, 12–72 bytes, nunca serializadas ni registradas.
- Tokens opacos de 256 bits; solo el hash vive en PostgreSQL.
- Cookie HttpOnly, Secure obligatoria en producción y SameSite validado.
- CSRF HMAC para métodos inseguros y origen exacto configurado.
- Autorización administrativa aplicada en el subrouter backend.
- Login inexistente/contraseña incorrecta usa `INVALID_CREDENTIALS` genérico.
- Contador de fallos se actualiza atómicamente en PostgreSQL para reducir carreras.
- Cambio de correo, contraseña, rol o desactivación revoca sesiones según corresponda.
- DTO explícito evita asignación masiva de hash, actor u otros campos protegidos.

## Interfaz

- Login sin enlace de alta y con explicación de que las cuentas son administrativas.
- Administración con búsqueda, filtros de rol/estado, paginación, alta, edición y confirmación.
- Estados de carga, vacío, error, envío y permiso denegado existentes.
- Roles visibles con texto y badge, no solo color.
- Diálogos fluidos con ancho máximo y controles táctiles heredados de al menos 44 px.
- Navegación y encabezado muestran las etiquetas de los tres roles.

La revisión estática, typecheck y build fueron aprobados. La inspección visual interactiva sigue no ejecutada porque no hay navegador controlable disponible.

## Evidencia de pruebas

| Verificación | Resultado |
| --- | --- |
| Tests unitarios backend | Aprobados. |
| Matriz middleware: Administrador permitido, Encargado/Personal denegados | Aprobada. |
| Registro público ausente | 404, aprobado. |
| Login incorrecto | 401 genérico, aprobado. |
| Login Administrador/Encargado/Personal | Aprobado. |
| Alta y edición administrativa | 201/200, aprobado. |
| Duplicado normalizado | 409, aprobado. |
| Hash/contraseña fuera de respuesta | Aprobado. |
| Encargado y Personal contra `/admin/users` | 403/403, aprobado. |
| Sesión expirada y sesión revocada por desactivación | 401/401, aprobado. |
| Usuario inactivo | Login rechazado, aprobado. |
| Bloqueo tras intentos fallidos | 429, aprobado. |
| Migración vacía + segunda ejecución | Aprobada. |
| Migración de `user@example.test` | Clasificado como `manager`, según decisión aprobada. |
| Base configurada | Versión 2 aplicada; dos usuarios activos y cero roles `user`. |
| `go test ./...`, `go vet ./...`, `gofmt -l .` | Aprobados. |
| Builds `api`, `migrate`, `seed` | Aprobados; binarios generados eliminados al terminar. |
| `npm test`, `npm run typecheck`, `npm run build` | Aprobados; 3/3 tests frontend. |
| `npm audit --omit=dev --audit-level=high` | Aprobado, 0 vulnerabilidades reportadas. |
| Compose desarrollo/producción | Configuración válida. |

El script reproducible es `scripts/verify-module1.ps1`; crea y elimina exclusivamente una base temporal con prefijo validado.
