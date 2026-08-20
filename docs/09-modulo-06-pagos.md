# Módulo 6 — Pagos y entrega

Estado al 2026-08-20: implementado, verificado y aprobado; el Módulo 7 se desarrolló posteriormente.

## Política aprobada

- El pago es siempre total: no existen anticipos, saldos ni pagos parciales.
- Puede registrarse desde `PENDING` hasta `FINISHED`. Cobrar al dejar el vehículo no cambia su estado operativo.
- Si se cobra al retirar, pago, `FINISHED → DELIVERED` e historial se escriben atómicamente.
- Una orden pagada con anterioridad puede entregarse después, únicamente desde `FINISHED`.
- El backend toma el monto congelado de la orden y rechaza campos monetarios enviados por el cliente.
- Métodos fijos: efectivo (`CASH`), QR, transferencia (`BANK_TRANSFER`) y tarjeta (`CARD`). La referencia externa es opcional.
- Administrador y Encargado registran pagos; Personal de lavado no consulta ni modifica datos de pago.

## Anulación, corrección y revisión posterior

Antes de entregar, Administrador o Encargado pueden anular el pago con motivo obligatorio. Una anulación del Encargado queda `PENDING_REVIEW`; el Administrador puede marcarla `REVIEWED` u `OBSERVED` posteriormente. La revisión no cambia el flujo operativo ni elimina registros.

Una orden entregada no permite quedar sin pago. Su corrección anula el pago vigente y crea el reemplazo dentro de una sola transacción. El pago anterior conserva actor, fecha, motivo y estado de revisión. Las devoluciones reales de dinero quedan fuera de V1.

## Persistencia y concurrencia

La migración `000009_payments.sql` crea:

- `payments`: orden, monto, método, referencia, idempotencia, cobrador, pago, anulación, revisión y reemplazo.
- `payment_events`: eventos `REGISTERED`, `VOIDED`, `CORRECTED`, `REVIEWED` y `OBSERVED` con actor e instante.

Un índice parcial único sobre `payments(order_id) WHERE status = 'VALID'` garantiza como máximo un pago vigente por orden. `idempotency_key` devuelve el resultado existente ante un reintento equivalente y conflicto si la clave se reutiliza con otros datos. Las operaciones bloquean la fila de la orden para serializar pago, anulación, cancelación y entrega. Los eventos tienen trigger append-only.

Mientras exista un pago válido, vehículo y servicios no pueden cambiar y la orden no puede cancelarse. Entrega estimada, personal y estados operativos previos continúan según las reglas del Módulo 5.

## API y pantallas

- `GET/POST /api/v1/orders/{id}/payments`: consulta histórica y registro del pago.
- `POST /api/v1/payments/{id}/void`: anulación previa a entrega.
- `POST /api/v1/payments/{id}/correct`: reemplazo atómico.
- `GET /api/v1/admin/payment-reviews`: cola administrativa.
- `POST /api/v1/admin/payment-reviews/{id}`: revisión conforme u observada.

El detalle de la orden muestra monto de solo lectura, método, referencia, cobrador, fecha, estado y pagos anulados. Los formularios tienen etiquetas visibles, mensajes junto al campo, confirmación clara, botones con carga y prevención de doble envío. `/admin/payment-reviews` ofrece carga, vacío, error/reintento, tabla y modal de revisión.

## Evidencia

- PostgreSQL/HTTP aislado: migración repetida, pago temprano, pago+entrega, exactitud, campos protegidos, orden cancelada, bloqueo comercial, roles, idempotencia 201/200, concurrencia 201/409, un pago válido, anulación, corrección, revisión y append-only aprobados.
- La base temporal E2E y el entorno visual con fixtures fueron eliminados después de comprobarlos.
- Base configurada: esquema versión 9, con cero órdenes, pagos y eventos de pago.
- Revisión visual: panel de pago, monto inmutable, selector, referencia, confirmación, historial anulado, cola y modal administrativo comprobados.
- Calidad final: Go test/vet/build, contrato de migración, 12/12 pruebas frontend, typecheck, build de producción y salud HTTP 200 aprobados.
- Git sigue sin inicializar por decisión expresa.

Al cierre del Módulo 6 no se implementaron pagos parciales, anticipos, descuentos, devoluciones, facturación, integraciones bancarias, dashboard ni reportes. Dashboard y reportes fueron resueltos posteriormente en el Módulo 7.
