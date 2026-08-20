# Modelo de datos inicial

Este documento registra la traducción técnica inicial de la guía; no reemplaza sus reglas funcionales.

## Convenciones

- PostgreSQL con `BIGSERIAL`/`BIGINT`, `TIMESTAMPTZ` en UTC y `NUMERIC(p,s)` para dinero.
- Claves foráneas y `ON DELETE` explícitos; entidades históricas no se eliminan físicamente.
- Unicidad normalizada mediante valor canónico, índice funcional o columna normalizada, según la decisión de cada módulo.
- Estados y roles protegidos por constraints SQL y validación de aplicación.
- Migraciones aditivas a partir de `000001_initial.sql`; no se reescribe una migración ya aplicable.

## Estado actual

`users` se adapta mediante `000002_lavadero_roles.sql`: conserva email único en minúsculas como login y admite `admin`, `manager` y `wash_staff`. Por decisión del negocio, la cuenta heredada `user@example.test` se clasifica como Encargado (`manager`); cualquier otra fila `user` detiene la migración. `auth_sessions` conserva hash de token, vencimientos, revocación y metadatos.

`vehicle_types`, `service_categories` y `services` fueron creadas por `000003_catalogs.sql`. Usan baja lógica, nombre canónico generado en PostgreSQL, unicidad global para tipos/categorías, unicidad por categoría para servicios y FK restrictiva. `000004_catalog_initial_values.sql` carga una sola vez los valores comerciales aprobados, que siguen siendo editables y extensibles.

`service_prices` fue creada por `000005_service_prices.sql` con `NUMERIC(12,2)`, precio estrictamente positivo, combinación servicio/tipo única y FK restrictivas. No contiene seeds: el Administrador carga las tarifas y una ausencia representa una celda faltante.

`clients` y `vehicles` fueron creadas por `000006_clients_vehicles.sql`; `000007_customer_identifier_format.sql` añadió constraints de formato ASCII al historial ya aplicado. El CI y la placa son opcionales y se persisten como `NULL`; sus valores reales usan columnas canónicas generadas e índices parciales únicos. `00000000` está reservado para representar un CI no proporcionado en pantalla/reporte y nunca se persiste como CI real. Los vehículos sin placa se identifican temporalmente por `vehicles.id`; `client_id` no puede cambiar mediante la API V1.

`orders`, `order_services`, `order_staff` y `order_status_history` fueron creadas por `000008_orders.sql`. Cada orden congela nombre y precio de sus servicios, calcula un total positivo exacto, exige personal activo de lavado y usa `version` para concurrencia optimista. El historial de estados es append-only mediante trigger; la entrega exige un pago válido integrado por el Módulo 6.

`payments` y `payment_events` fueron creadas por `000009_payments.sql`. El monto proviene del total de la orden, un índice parcial permite como máximo un pago `VALID` por orden y las claves de idempotencia evitan duplicados por reintento. Las anulaciones conservan actor, instante, motivo y revisión; las correcciones crean un pago reemplazante sin borrar el anterior. Los eventos son append-only.

`000010_report_indexes.sql` no crea datos ni tablas: añade índices por recepción de orden, transición/fecha, pago válido/fecha y anulación/fecha. Los reportes se calculan directamente desde los agregados transaccionales y no mantienen contadores duplicados.

## Esquema objetivo preliminar

| Agregado/tabla | Claves y tipos propuestos | Integridad/índices |
| --- | --- | --- |
| users | ID, login aprobado, password_hash, nombre, rol, activo, auditoría temporal | login canónico único; rol de tres valores; sin delete físico. |
| auth_sessions | user_id, token_hash, actividad, expiración, revocación | token único; consultas por usuario/activa y expiración. |
| vehicle_types | nombre, activo, timestamps | nombre normalizado único. |
| service_categories | nombre, activo, timestamps | nombre normalizado único. |
| services | category_id, nombre, activo, timestamps | único `(category_id, nombre_normalizado)`; FK restrictiva. |
| service_prices | service_id, vehicle_type_id, `NUMERIC(12,2)`, timestamps | combinación única; `price > 0`; sin seeds comerciales. |
| clients | nombre, CI nullable, teléfono, timestamps | CI real normalizado único mediante índice parcial; `00000000` reservado. |
| vehicles | client_id, vehicle_type_id, placa nullable, marca, modelo, color | placa real normalizada única; ID interno si no hay placa; propietario inmutable en V1; FKs restrictivas. |
| orders | vehicle_id, registered_by_user_id, received_at, estimated_delivery_at, status, `NUMERIC(12,2)`, version/timestamps | total positivo; estado controlado; concurrencia optimista. |
| order_services | order_id, service_id, service_name_snapshot, `price_snapshot NUMERIC(12,2)` | combinación única por orden; snapshots inmutables. |
| order_staff | order_id, user_id | combinación única; rol válido comprobado por servicio y/o constraint viable. |
| order_status_history | order_id, origen, destino, instante, actor_id, motivo | append-only; índice cronológico por orden. |
| payments | order_id, amount, method, reference, status, idempotency_key, actores/fechas de anulación y revisión | índice único parcial para un pago válido; monto positivo exacto; reemplazo trazable. |
| payment_events | payment_id, order_id, event_type, actor_user_id, occurred_at, reason | append-only; índice cronológico por pago. |

## Transacciones previstas

- Crear orden: cabecera + servicios + snapshots + asignaciones + total + primer historial.
- Cambiar estado: update condicional + fila de historial.
- Pagar/entregar: pago único + transición opcional relacionada, atómicos cuando se cobra al retirar.
- Cambiar contraseña/rol/desactivar usuario: cambio de identidad + revocación de sesiones.

## Decisiones pendientes que afectan el esquema

Las decisiones de pagos quedaron resueltas en el Módulo 6 y las fórmulas/permisos de reportes en el Módulo 7. El Módulo 8 no requiere cambios de esquema ni funciones nuevas.
