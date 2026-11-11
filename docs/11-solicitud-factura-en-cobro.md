# Solicitud de factura al cobrar

Cambio solicitado el 2026-10-31.

- El formulario de cobro incluye la casilla «Solicita factura», desmarcada por defecto.
- La migración 11 añade `payments.invoice_requested BOOLEAN NOT NULL DEFAULT FALSE`. Los pagos anteriores quedan sin solicitud registrada; esto no afirma que se haya emitido o no una factura fuera del sistema.
- El detalle del pago y su historial muestran la solicitud.
- Una corrección precarga la elección anterior y permite cambiarla con el motivo obligatorio ya existente. La API conserva el valor anterior si se omite el campo al corregir.
- Un reintento con la misma clave y una elección diferente produce conflicto.
- Se mantienen los permisos, el importe total y el flujo de entrega actuales.
- Este campo registra una solicitud, no una factura emitida. Emisión fiscal, SIAT, datos fiscales y PDF siguen fuera de este cambio.

Regresión: `scripts/verify-module7.ps1` incluye cobro con/sin solicitud, tipo inválido, lectura persistida, correcciones, reintentos e historial.

## Verificación del 2026-10-31

- `go test ./...` y `go vet ./...`: aprobados.
- Frontend: 15/15 pruebas, typecheck y build de producción aprobados; advertencias de dependencias DEP0155, sin errores.
- E2E PostgreSQL/HTTP: aprobado, incluidos permisos, tipos inválidos, elección true/false, corrección que conserva y cambia la solicitud, conflicto por reintentos e historial. La API se compila durante esta prueba.
- Migración ejecutada dos veces en la base temporal; esquema local confirmado en versión 11.
- Base temporal eliminada al finalizar.
- No se ejecutó revisión visual interactiva de la casilla.

## Ampliación del 2026-11-10

- La orden guarda `invoice_requested` desde su creación. La casilla comienza desmarcada y puede editarse mientras la orden esté pendiente y sin pago vigente.
- El cobro precarga la elección de la orden. Los clientes de API que omiten el campo heredan esa elección. Una selección explícita al cobrar o corregir actualiza la orden dentro de la misma transacción y conserva los pagos y documentos anteriores.
- El correo opcional del cliente se valida y normaliza. Vacío se guarda como NULL. Las actualizaciones anteriores que omiten el campo conservan el correo registrado.
- La contadora consulta y exporta las órdenes y sus solicitudes, y consulta los clientes mediante rutas GET independientes.
- Después de crear una orden se ofrece abrir la ficha de recepción de 80 mm. El navegador presenta el diálogo al solicitar la impresión.
- La migración `000014_order_invoice_client_email.sql` conserva el campo de pagos de la antigua migración local 11 e incorpora los nuevos campos. Reconcilia la restricción de roles cuando esa base ya tenía la migración local 11 aplicada.

La emisión fiscal continúa fuera del alcance de esta solicitud.
