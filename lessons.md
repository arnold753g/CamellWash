# Lecciones técnicas

## Módulo 5 — órdenes

- Los estados de carga, error y vacío deben modelarse como ramas visuales mutuamente exclusivas; una prueba de presencia de componentes no reemplaza la inspección del comportamiento real.
- Una orden histórica no debe depender de nombres ni tarifas actuales: congelar ambos valores evita que una edición de catálogo reescriba el pasado.
- El total debe derivarse exclusivamente en servidor mediante centavos exactos; el total mostrado por el cliente es solo una vista previa.
- El control optimista por `version` debe abarcar el cambio de estado y su historial dentro de la misma transacción para evitar dobles transiciones.
- Filtrar órdenes asignadas desde la consulta, además de validar la mutación, reduce el riesgo de IDOR para Personal de lavado.
- Un historial append-only merece protección en base de datos, no solo una convención de aplicación.
- Las dependencias futuras deben fallar de forma explícita: bloquear entrega sin pago real es más seguro que simular o adelantar el Módulo 6.
- Los scripts HTTP que deben funcionar en Windows PowerShell necesitan serializar JSON a bytes UTF-8 explícitos; depender del encoding implícito falla con fixtures acentuados antes de llegar al handler.

## Módulo 6 — pagos

- Cobrar temprano obliga a congelar vehículo y servicios desde el primer pago válido; de otro modo el total pagado podría separarse del total de la orden.
- La clave idempotente resuelve reintentos idénticos y el bloqueo de fila más el índice parcial resuelven solicitudes concurrentes diferentes.
- Cobrar al retirar requiere que pago, transición e historial compartan una transacción; cualquier fallo debe revertir los tres.
- Una corrección después de entregar debe reemplazar el pago atómicamente para que nunca exista una orden entregada sin pago válido.
- La operación diaria y el control administrativo pueden desacoplarse: el Encargado anula con motivo y el Administrador revisa después, sin eliminar evidencia.
- Los fixtures con acentos deben construirse de forma independiente al encoding del archivo cuando se ejecutan con Windows PowerShell 5.1.

## Módulo 7 — dashboard y reportes

- Un reporte confiable deriva cada métrica de su evento temporal correcto: recepción, transición, pago o anulación; no del estado actual para todo.
- Preagregar cada fuente antes de combinarla evita multiplicar órdenes por servicios, personal o eventos relacionados.
- Los agregados financieros deben tener una frontera RBAC distinta a la operación: registrar un cobro no implica permiso para analizar toda la caja.
- Un rango civil inclusivo requiere convertir medianoches de `America/La_Paz` a límites UTC exclusivos; sumar 24 horas en UTC no expresa la regla de negocio.
- Las barras visuales deben acompañarse de cifras y tablas; el color no puede ser el único canal ni el gráfico la única forma de leer el dato.
- Un entorno visual aislado necesita origen CORS y host de cookie coherentes; esa ampliación debe quedar limitada a fixtures temporales.
