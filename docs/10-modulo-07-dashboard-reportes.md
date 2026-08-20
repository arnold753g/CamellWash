# Módulo 7 — Dashboard, reportes y marca Camell Wash

Estado al 2026-08-20: implementado y verificado técnicamente; pendiente de aprobación expresa para iniciar el Módulo 8.

## Identidad

- Nombre completo: `Camell Wash`; nombre corto: `Camell`.
- Color principal: `#ff5500`, aplicado mediante tokens semánticos y variantes de interacción accesibles.
- Logo oficial proporcionado por el propietario, conservado sin modificaciones de contenido en `frontend/public/branding/camell-wash-logo.png`.
- Como la imagen ya incluye el nombre, la interfaz evita duplicarlo y mantiene un texto alternativo accesible.

## Fórmulas

Todos los rangos son fechas civiles inclusivas en `America/La_Paz`; el backend las transforma a límites UTC `[desde, día_siguiente_a_hasta)` y admite como máximo un año.

- Órdenes recibidas: `orders.received_at` dentro del rango.
- Entregadas/canceladas: filas de `order_status_history` cuya transición destino ocurre dentro del rango.
- En operación: estado actual `PENDING`, `IN_PROGRESS`, `QUALITY_CONTROL` o `FINISHED`.
- Listas para retiro: estado actual `FINISHED`.
- Con demora: orden activa cuya entrega estimada ya venció.
- Ciclo promedio: minutos entre recepción y primera transición a `DELIVERED` ocurrida en el rango.
- Servicios más solicitados: líneas de órdenes recibidas en el rango, excluyendo órdenes canceladas.
- Ingreso: suma de pagos con estado `VALID` y `paid_at` dentro del rango.
- Ticket promedio: promedio de esos pagos válidos.
- Anulaciones: cantidad y monto de pagos `VOIDED` cuyo `voided_at` cae en el rango; se muestran aparte y no se restan dos veces.

No existen tablas resumen ni contadores persistidos. La migración 10 añade cuatro índices para sostener las consultas sin duplicar fuentes de verdad.

## Permisos

| Recurso | Administrador | Encargado | Personal de lavado |
| --- | --- | --- | --- |
| Dashboard/reporte operativo | Sí | Sí | No |
| Ingresos, ticket, métodos y anulaciones agregadas | Sí | No | No |
| Exportación CSV operativa | Sí | Sí | No |
| Exportación CSV financiera | Sí | No | No |

El Encargado puede registrar y consultar pagos dentro de una orden por el Módulo 6, pero esa responsabilidad no le concede análisis financiero global.

## API e interfaz

- `GET /api/v1/reports/operations?date_from&date_to` — Administrador/Encargado.
- `GET /api/v1/admin/reports/financial?date_from&date_to` — solo Administrador.
- `GET /api/v1/admin/dashboard?date_from&date_to` — combinación administrativa.
- `/admin` y `/app` muestran el día; `/admin/reports` y `/app/reports` ofrecen periodo y CSV.

La interfaz incluye carga, error con reintento, vacío, tarjetas, barras con cifras, tablas alternativas, filtros etiquetados, contraste AA, cifras tabulares y breakpoints de 64/48/34 rem.

## Evidencia

- E2E aislado: migración repetida, fechas, totales diarios sin duplicación, anulaciones separadas, cuatro métodos, RBAC, cuatro índices y respuesta operativa de 55 ms aprobados.
- Calidad: Go test/vet/build, contrato de migración, 15/15 pruebas frontend, typecheck y build de producción aprobados.
- Revisión visual: login, logo oficial, dashboard Administrador, reporte por periodo y dashboard Encargado sin finanzas comprobados.
- PostgreSQL configurado en versión 10; backend HTTP saludable en 8080 y frontend Camell Wash activo en 3000.
- Los datos, usuarios, procesos y bases temporales de inspección fueron eliminados. Git continúa sin inicializar por decisión expresa.
