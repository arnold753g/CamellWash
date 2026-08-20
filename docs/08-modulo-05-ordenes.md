# Módulo 5 — Órdenes, asignaciones e historial

Estado al 2026-08-20: aprobado; el usuario autorizó e inició el Módulo 6.

## Decisiones aprobadas

- Una orden requiere un vehículo activo, uno o más servicios activos con tarifa vigente y al menos un Personal de lavado activo.
- Cada línea conserva `service_name_snapshot` y `price_snapshot`; cambios posteriores de catálogo o precio no alteran el documento histórico.
- El total se calcula en servidor con centavos exactos y se persiste como `NUMERIC(12,2)`.
- La secuencia normal es `PENDING → IN_PROGRESS → QUALITY_CONTROL → FINISHED`.
- Personal de lavado solo consulta órdenes asignadas y ejecuta las tres transiciones operativas hasta control de calidad.
- Administrador y Encargado pueden ejecutar transiciones válidas y aprobar `QUALITY_CONTROL → FINISHED`.
- La cancelación solo se admite desde `PENDING`, `IN_PROGRESS` o `QUALITY_CONTROL`, solo para Administrador/Encargado y siempre con motivo.
- Vehículo y servicios solo pueden editarse en `PENDING`. Entrega estimada y asignaciones pueden editarse en `PENDING` e `IN_PROGRESS`.
- `FINISHED → DELIVERED` devuelve conflicto mientras no exista un pago real; esa integración pertenece al Módulo 6.
- Los tiempos se interpretan en `America/La_Paz` en la interfaz y se persisten como `TIMESTAMPTZ` UTC.

## Persistencia y transacciones

La migración `000008_orders.sql` crea:

- `orders`: cabecera, vehículo, registrador, fechas, estado, total y `version`.
- `order_services`: referencias al catálogo y snapshots inmutables de nombre/precio.
- `order_staff`: relación única orden–personal con instante de asignación.
- `order_status_history`: origen, destino, actor, instante y motivo.

Crear una orden escribe cabecera, líneas, asignaciones e historial inicial dentro de una sola transacción. Las transiciones realizan un `UPDATE` condicional por ID, estado y versión; una carrera produce un éxito y un conflicto, sin duplicar historial. Un trigger PostgreSQL impide actualizar o borrar el historial.

## Permisos y seguridad

| Acción | Administrador | Encargado | Personal de lavado |
| --- | --- | --- | --- |
| Crear/editar orden | Sí | Sí | No |
| Listar/ver detalle | Todas | Todas | Solo asignadas |
| Avanzar flujo operativo | Sí | Sí | Solo asignadas hasta control de calidad |
| Aprobar finalización | Sí | Sí | No |
| Cancelar con motivo | Sí | Sí | No |
| Entregar sin pago real | No | No | No |

La restricción de asignación se aplica en consultas y mutaciones para evitar IDOR. Las respuestas de orden incluyen el nombre operativo del cliente, pero no exponen CI ni teléfono.

## API e interfaz

- CRUD operativo: `GET/POST /api/v1/orders`, `GET/PATCH /api/v1/orders/{id}`.
- Estados: `POST /api/v1/orders/{id}/transitions` y `/cancel`.
- Opciones: `GET /api/v1/operations/wash-staff` y `/service-options`.
- Historiales: `GET /api/v1/clients/{id}/orders` y `/vehicles/{id}/orders`.
- Interfaces compartidas: `/admin/orders` y `/app/orders`.

La pantalla ofrece filtros de estado/fecha, creación y edición condicionada por estado, total previo, detalle con snapshots, personal e historial. Carga, vacío y error son estados mutuamente exclusivos; el error permite reintentar. El historial también se abre desde la ficha de cliente o vehículo.

## Evidencia

- E2E PostgreSQL/HTTP aislado: migración repetida, creación atómica, suma exacta, snapshots estables, validaciones, permisos, IDOR, concurrencia 200/409, append-only, cancelación, edición limitada, historiales y bloqueo de entrega aprobados.
- Backend: `go test ./...`, `go vet ./...` y builds de API/migrador aprobados.
- Frontend: 11/11 pruebas, typecheck, build y auditoría npm con 0 vulnerabilidades aprobados.
- Revisión visual: carga, vacío, error exclusivo, reintento, formulario, validación y layout responsive comprobados sin insertar datos comerciales.
- Base configurada: esquema en versión 8 con cero órdenes y cero filas relacionadas.

En este módulo no se implementaron pagos ni entrega efectiva; esa limitación histórica fue resuelta posteriormente por el Módulo 6. Dashboard y reportes siguen fuera de alcance. Git continúa sin inicializar por decisión expresa.
