# Recibos térmicos de 80 mm

Publicado y verificado el 4 de noviembre de 2026, con Windows, impresora USB y diálogo estándar del navegador. La salida contiene recibo de pago y ficha de recepción/retiro, ambos con placa. Una línea de corte separa los documentos; no se envían comandos de corte ni de cajón de dinero.

## Uso

Al guardar una orden nueva se ofrece **Imprimir orden**. La ficha de recepción se crea en la transacción de la orden. Administrador y encargado también pueden abrir **Ficha del cliente** en su detalle. Al registrar o corregir un pago se ofrece **Abrir recibo y ficha**; también queda disponible el botón **Recibo y ficha** en el pago. Los pagos anulados conservan acceso a su recibo histórico. Contador y personal de lavado no tienen acceso a estas páginas ni a sus endpoints.

En la vista previa, **Imprimir** registra la solicitud, actualiza el estado del pago y abre el diálogo del navegador. Seleccionar la impresora instalada en Windows, rollo de 80 mm, escala 100 % y desactivar encabezados y pies del navegador. El diseño usa 72 mm de contenido, márgenes de 4 mm y longitud calculada a partir del contenido. La compatibilidad final de márgenes, longitud y corte se comprueba con el equipo físico.

Los datos incluyen empresa, número de recibo/ficha, orden, cliente, vehículo y placa, servicios, total en BOB, fechas en America/La_Paz y, en el recibo, método, cajero y referencia de pago. La ficha conserva los datos al emitirla y no acredita pago. No se genera factura fiscal.

Nombre comercial por defecto: Camell Wash. El servidor admite `RECEIPT_BUSINESS_NAME`, `RECEIPT_BUSINESS_ADDRESS` y `RECEIPT_BUSINESS_PHONE` para los nuevos documentos; no se inventan dirección ni teléfono. El logo se adapta a blanco y negro sobre fondo blanco.

## Integridad y auditoría

- Migraciones nuevas `000012_thermal_documents.sql` y `000013_preserve_document_history.sql`.
- Un recibo por pago y una ficha por orden. Numeración persistente `R-000001` y `F-000002`; pueden existir saltos y los prefijos distinguen los documentos.
- El recibo se genera en la transacción del pago y reutiliza la ficha que fue emitida al crear la orden. Los pagos anteriores a esta función generan sus documentos al solicitarlos por primera vez.
- Los datos originales se guardan como una instantánea. Cambiar el cliente, vehículo o catálogo después no reescribe un documento emitido. La base de datos impide actualizar y eliminar documentos y solicitudes de impresión.
- Anular o corregir un pago conserva el original. Al consultar e imprimir se muestra su estado actual: anulado o reemplazado, con enlace numérico al recibo anterior/nuevo.
- La primera solicitud figura como original; las siguientes como reimpresión 1, 2, etc. Se registra usuario y fecha. Las solicitudes tienen clave de idempotencia y se serializan con el bloqueo de la orden para evitar numeración duplicada por concurrencia.
- No se registra otro pago al imprimir. Todas las escrituras requieren sesión, rol y CSRF válidos. No hay enlaces públicos ni acceso de contador a información operativa.

El navegador no confirma si salió papel. Se auditan **solicitudes de impresión**, incluidas las canceladas. Las copias adicionales elegidas dentro del diálogo del controlador no pueden contabilizarse por separado. La impresión desde la barra del navegador muestra una indicación para usar el botón de la aplicación; no permite acreditar una impresión física.

## Rutas

Todas requieren administrador o encargado y están bajo `/api/v1`:

- `POST /payments/{id}/documents/receipt`: preparar/recuperar recibo sin duplicarlo.
- `POST /reception/orders/{id}/document`: preparar/recuperar ficha.
- `GET /documents/{id}`: datos originales, estado actual e historial de solicitudes.
- `POST /documents/{id}/print-attempts`: solicitud idempotente con estado actualizado.

Vista web: `/print/orders/{order_id}` para ficha; agregar `?payment_id={payment_id}` para recibo y ficha. El pago debe corresponder a la orden de la URL.

## Validación y publicación

Las pruebas con PostgreSQL desechable comprueban instantáneas, reimpresión, idempotencia, cambios de pago, restricciones de roles y protección de historial. Las pruebas de navegador verifican el flujo de pago, documentos con placa, concurrencia, controles CSRF, estados anulado/reemplazado y salida para 80 mm.

La publicación se verificó el 4 de noviembre de 2026. La calibración con la impresora física sigue pendiente. Antes de subir, comprobar cambios actuales del VPS y respaldar únicamente Camell Wash. Aplicar migraciones y permisos con el propietario de su base, reconstruir backend/frontend y ejecutar el generador CSP, que ahora comprueba también la ruta de impresión. No modificar servicios ni configuraciones de otros proyectos.

La solicitud de factura figura en los nuevos documentos. Los documentos anteriores conservan su instantánea y muestran que ese dato no fue registrado. Una corrección conserva el valor original del pago anterior.
