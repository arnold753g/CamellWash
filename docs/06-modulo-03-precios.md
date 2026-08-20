# Módulo 3 — Matriz de precios

Estado: **IMPLEMENTADO, MIGRADO, VERIFICADO Y APROBADO**.

Fuente funcional: [`GUIA_UNICA_SISTEMA_LAVADERO.md`](../GUIA_UNICA_SISTEMA_LAVADERO.md).

## Decisiones aplicadas

- La tabla comienza vacía. No existen seeds ni precios comerciales de ejemplo; el Administrador carga cada tarifa.
- Una celda vacía significa **sin precio** y permanece visible en la matriz.
- El precio cero se bloquea. Toda tarifa debe ser estrictamente mayor que `0.00`.
- La API recibe y devuelve el importe como texto decimal canónico con dos posiciones, por ejemplo `"25.00"`; Go y JavaScript no lo convierten a `float`/`number`.
- No se asumió símbolo ni código de moneda porque la guía no define todavía una moneda contractual.

## Modelo e integridad

La migración `000005_service_prices.sql` crea `service_prices` con:

- `price NUMERIC(12,2) NOT NULL` y `CHECK (price > 0)`.
- Unicidad `(service_id, vehicle_type_id)`.
- Claves foráneas restrictivas hacia `services` y `vehicle_types`.
- Timestamps y trigger común de actualización.
- Índice inverso por tipo y servicio para consultas futuras.

La creación toma bloqueos compartidos sobre servicio, categoría y tipo activos antes de insertar. La restricción única sigue siendo la protección definitiva: dos altas HTTP simultáneas producen exactamente un `201` y un `409`, nunca dos filas.

Las tarifas existentes no se eliminan cuando un catálogo queda inactivo. La combinación deja de aparecer en la matriz y no se considera vigente; al reactivarse recupera la tarifa conservada. No se permite crear ni actualizar una tarifa nueva mientras el servicio, su categoría o el tipo estén inactivos.

## Matriz y faltantes

La matriz se calcula mediante el producto de servicios activos y tipos activos, limitado a 10 000 celdas por consulta. Un `LEFT JOIN` con `service_prices` distingue configuradas y faltantes. Por ello, crear o reactivar un catálogo se refleja inmediatamente sin materializar filas vacías.

Admite filtros por categoría, servicio, tipo y estado de precio en API. La interfaz ofrece categoría, servicio, tipo y cobertura (`Todas`, `Con faltantes`, `Completas`), resumen 36/0/36 en la base configurada y texto/icono para cada hueco.

## Permisos y endpoints

| Método y ruta | Permiso | Función |
| --- | --- | --- |
| `GET /api/v1/admin/service-prices/matrix` | Administrador | Consulta matriz, resumen y faltantes filtrados. |
| `POST /api/v1/admin/service-prices` | Administrador | Crea una tarifa ausente. |
| `PATCH /api/v1/admin/service-prices/{service_id}/{vehicle_type_id}` | Administrador | Actualiza una tarifa existente. |
| `GET /api/v1/pricing/current?service_id=...&vehicle_type_id=...` | Administrador y Encargado | Obtiene la tarifa vigente exacta para operación futura. |

Personal de lavado recibe `403` incluso en la consulta de tarifa vigente. Encargado consulta, pero las rutas administrativas le devuelven `403`.

## Contrato para el Módulo 5

`services.PricingService.CurrentPrice(ctx, serviceID, vehicleTypeID)` es el caso de uso confiable para congelar precios en una orden:

1. Solo devuelve una fila si servicio, categoría y tipo están activos y la combinación tiene tarifa.
2. Devuelve el importe exacto como texto decimal con dos posiciones.
3. Una combinación faltante o inactiva produce `ErrPriceNotFound`; la API lo traduce a `404 CURRENT_PRICE_NOT_FOUND`.
4. El Módulo 5 debe crear `PricingService` con su transacción (`NewPricingService(tx)`), llamar `CurrentPrice` y copiar inmediatamente ese valor a `order_services.price_snapshot` dentro de la misma transacción de creación de orden.
5. Cambios posteriores en `service_prices` nunca deben recalcular snapshots existentes.

## Interfaz y verificación visual

- Ruta: `/admin/prices`; navegación integrada en `AppShell`.
- Conserva PrimeVue, tema oscuro, tokens, tarjetas, filtros, diálogo y feedback existentes.
- La tabla mantiene categoría y servicio visibles, permite desplazamiento horizontal y apila filtros en móvil.
- Cada celda comunica `Sin precio / Cargar` o importe exacto / `Editar`, sin depender solo del color.
- Se verificaron matriz 6 × 6, modal, validación, carga, vacío y error mediante navegador integrado, sin insertar tarifas en la base configurada.
- Persisten los avisos previos de hidratación del encabezado autenticado y el aviso transitorio del manifiesto en desarrollo; no nacen en este módulo ni afectan la matriz.

## Evidencia de pruebas

| Verificación | Resultado |
| --- | --- |
| Migraciones 1–5 desde base vacía y segunda ejecución | Aprobada. |
| Base configurada | Versión 5, `NUMERIC(12,2)`, cero tarifas precargadas. |
| Matriz inicial | 36 combinaciones, 0 configuradas, 36 faltantes. |
| Crear y actualizar precio | `201` y `200`; canonicalización exacta. |
| Precisión `9999999999.99` y actualización `0.10` | Devueltas sin pérdida. |
| Cero, negativo, exponente y más de dos decimales | `422`. |
| Encargado consulta/modifica | `200` / `403`. |
| Personal consulta tarifa | `403`. |
| Dos altas simultáneas | `201` / `409`; una sola fila PostgreSQL. |
| Alta/reactivación de servicio | Seis faltantes visibles inmediatamente. |
| Catálogo inactivo | No admite escritura ni tarifa vigente; fila conservada. |
| Tests frontend | 6/6 aprobados, incluida utilidad decimal y resolución de componentes. |
| Tests, formato, vet y build Go | Aprobados. |
| Typecheck y build Nuxt | Aprobados; dos avisos DEP0155 transitivos conocidos. |
| `npm audit --omit=dev --audit-level=high` | Aprobado, 0 vulnerabilidades. |
| Compose desarrollo/producción | Configuración válida con avisos de acceso al config local de Docker. |

El script reproducible `scripts/verify-module3.ps1` crea y elimina exclusivamente una base temporal con prefijo validado.

## Riesgos y límites

- La moneda visible sigue sin definirse; no se inventó `BOB`, símbolo ni formato localizado.
- El límite de 10 000 combinaciones exige filtrar si los catálogos crecen de forma extraordinaria.
- No se implementaron órdenes, clientes, vehículos, pagos ni reportes.
- Git continúa sin inicializar por decisión expresa; no existe diff ni rama verificable.

## Plan del Módulo 4

Antes de crear su migración deben aprobarse explícitamente CI obligatorio/único y la política para vehículos sin placa. Luego corresponde implementar clientes, varios vehículos, normalización, privacidad, permisos Administrador/Encargado y pruebas PostgreSQL/HTTP, sin iniciar órdenes.

## Cierre

El usuario aprobó expresamente el Módulo 3 el 2026-08-20 y solicitó continuar con el Módulo 4. No se realizó ninguna operación Git.
