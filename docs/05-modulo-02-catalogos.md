# Módulo 2 — Catálogos

Estado: **IMPLEMENTADO, MIGRADO, VERIFICADO Y APROBADO**.

Fuente funcional: [`GUIA_UNICA_SISTEMA_LAVADERO.md`](../GUIA_UNICA_SISTEMA_LAVADERO.md).

## Resultado técnico

- `vehicle_types`, `service_categories` y `services` son tablas extensibles con timestamps y baja lógica mediante `active`.
- PostgreSQL genera `normalized_name` con recorte, colapso de espacios y minúsculas.
- Tipos y categorías son únicos globalmente; servicios son únicos dentro de su categoría.
- Un mismo nombre de servicio puede existir en categorías diferentes.
- La FK `services.category_id` usa `ON DELETE RESTRICT`; no existe API de borrado físico.
- Crear un servicio exige una categoría activa. Un servicio histórico puede conservar su categoría aunque luego sea desactivada.
- Los listados operativos devuelven solo elementos activos; un servicio también exige que su categoría esté activa.

La migración estructural `000003_catalogs.sql` y la migración de valores iniciales `000004_catalog_initial_values.sql` se aplicaron a la base configurada el 2026-08-13.

## Normalización

La aplicación conserva la escritura visible después de recortar extremos y convertir cualquier secuencia de espacios en uno. La comparación canónica de PostgreSQL aplica además minúsculas. Ejemplo: `"  Lavado   Premium "` y `"lavado premium"` ocupan la misma clave normalizada.

La disponibilidad se comprueba en aplicación para entregar un conflicto claro y se garantiza de nuevo con constraints únicos. La base de datos sigue siendo la protección final frente a solicitudes concurrentes.

## Permisos y endpoints

Todos los endpoints requieren sesión. Solo el Administrador accede a las rutas de administración.

| Área | Endpoints |
| --- | --- |
| Tipos | `GET/POST /admin/vehicle-types`, `GET/PATCH /admin/vehicle-types/:id`, `PATCH /admin/vehicle-types/:id/status` |
| Categorías | `GET/POST /admin/service-categories`, `GET/PATCH /admin/service-categories/:id`, `PATCH /admin/service-categories/:id/status` |
| Servicios | `GET/POST /admin/services`, `GET/PATCH /admin/services/:id`, `PATCH /admin/services/:id/status` |
| Consulta activa | `GET /catalogs/vehicle-types`, `GET /catalogs/service-categories`, `GET /catalogs/services` |

Los listados administrativos aceptan búsqueda, filtro activo/inactivo, página y límite. Servicios también filtra por categoría. Las consultas activas están disponibles a Administrador, Encargado y Personal de lavado; no exponen administración.

## Interfaz

- Ruta administrativa: `/admin/catalogs`.
- Una entrada nueva en el `AppShell` existente; no se cambió la navegación global.
- Tabs PrimeVue para tipos, categorías y servicios.
- Panel reutilizable para tipos/categorías y panel especializado para servicios.
- Búsqueda, filtros, paginación, carga, vacío, error, confirmación, doble envío bloqueado y feedback con texto.
- Formularios con labels visibles, ayuda de normalización, errores junto al campo y controles táctiles existentes.
- Tablas permanecen dentro de `.table-wrap`; filtros y acciones se apilan en móvil.
- No se agregó Bootstrap, Tailwind, paleta, fuente ni librería UI adicional.

Typecheck y build aprobaron. La inspección visual interactiva comprobó las tres pestañas con sus datos esperados, además de los estados de carga, vacío y error. Se corrigieron las importaciones explícitas de los componentes de catálogo y no se registraron nuevas advertencias de componentes no resueltos en la comprobación final.

## Valores iniciales aprobados

La migración 4 carga de forma idempotente seis tipos, dos categorías y seis servicios aprobados por el negocio. Usa las mismas claves normalizadas y `ON CONFLICT DO NOTHING`, por lo que una instalación no recibe duplicados.

Estos valores son registros ordinarios, no enums ni constantes. La administración permite agregar otros, renombrarlos, cambiar servicios de categoría y activar/desactivar cualquiera. La migración se ejecuta una sola vez por el ledger y nunca vuelve a sobrescribir modificaciones posteriores.

## Evidencia de pruebas

| Verificación | Resultado |
| --- | --- |
| Migraciones 1–3 desde base vacía y segunda ejecución | Aprobada, idempotente. |
| Migraciones 3 y 4 en base configurada | Aprobadas; 6 tipos, 2 categorías y 6 servicios activos. |
| Segunda ejecución de migraciones | Sin duplicados, aprobada. |
| Edición de un valor inicial y alta de uno adicional | 200/201, aprobada. |
| Duplicado normalizado de tipo/categoría | 409, aprobado. |
| Servicio duplicado dentro de categoría | 409, aprobado. |
| Mismo servicio en categoría diferente | 201, aprobado. |
| Servicio nuevo con categoría inactiva | 422, aprobado. |
| Administrador administra los tres catálogos | Aprobado. |
| Encargado/Personal intentan escribir | 403/403, aprobado. |
| Encargado/Personal consultan activos | 200/200, aprobado. |
| Inactivos fuera de listados operativos | Aprobado. |
| Baja lógica conserva la fila | Aprobado por consulta PostgreSQL. |
| Tests, formato, vet y build Go | Aprobados. |
| Tests frontend | 4/4 aprobados; incluye contrato de resolución de componentes de catálogo contra el registro generado por Nuxt. |
| Typecheck y build Nuxt | Aprobados; persisten dos avisos DEP0155 transitivos conocidos. |
| Inspección visual de tipos, categorías y servicios | Aprobada; 6/2/6 registros renderizados. |
| Estados de carga, vacío y error en las tres pestañas | Aprobados mediante señales controladas y reversibles, sin modificar datos comerciales. |

El script reproducible `scripts/verify-module2.ps1` crea una base temporal con prefijo validado y la elimina al terminar.

## Cierre

El Módulo 2 fue aprobado el 2026-08-20 mediante la solicitud expresa de concluir su desarrollo. No quedan pendientes dentro de su alcance.

El Módulo 3 comenzó después de definir una matriz sin precios precargados y bloquear `0.00`; su evidencia se registra en `docs/06-modulo-03-precios.md`.
