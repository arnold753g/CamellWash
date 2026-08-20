# Tareas

Fuente funcional: [`GUIA_UNICA_SISTEMA_LAVADERO.md`](GUIA_UNICA_SISTEMA_LAVADERO.md).

## Fase 0

- [x] Guardar la guía única en la raíz.
- [x] Confirmar la ruta del proyecto base.
- [x] Inventariar backend, frontend, datos, seguridad, pruebas e infraestructura.
- [x] Documentar el contrato visual real.
- [x] Ejecutar verificaciones estáticas, tests y builds disponibles.
- [x] Crear auditoría, arquitectura, modelo inicial y plan modular.
- [x] Recibir aprobación expresa de la Fase 0.
- [x] Usar correo normalizado como login y limitar el seed al Administrador de `.env`.
- [x] Clasificar `user@example.test` como Encargado por decisión expresa.

## Módulo 1 — aprobado

- [x] Implementar migración y tres roles en backend/frontend.
- [x] Retirar el registro público.
- [x] Implementar alta, edición, rol y activación/desactivación administrativa.
- [x] Probar autenticación, RBAC, sesiones, duplicados, seed y migraciones en PostgreSQL temporal.
- [x] Mantener PrimeVue y el contrato visual existente sin Bootstrap/Tailwind.
- [x] Aplicar la migración a la base configurada y comprobar que no queden roles `user`.
- [ ] Ejecutar inspección visual interactiva cuando haya navegador controlable.
- [x] Recibir aprobación del Módulo 1 mediante autorización expresa para iniciar el Módulo 2.

## Módulo 2 — aprobado

- [x] Crear tipos, categorías y servicios extensibles con baja lógica.
- [x] Proteger unicidad normalizada en aplicación y PostgreSQL.
- [x] Implementar administración exclusiva y consultas activas por rol.
- [x] Añadir búsqueda, filtros, paginación y confirmaciones.
- [x] Mantener PrimeVue y el contrato visual sin Bootstrap/Tailwind.
- [x] Probar migraciones, HTTP, permisos y conservación de inactivos.
- [x] Aplicar la migración 3 vacía a la base configurada.
- [x] Confirmar y cargar 6 tipos, 2 categorías y 6 servicios como registros editables y extensibles.
- [x] Ejecutar inspección visual interactiva de las tres pestañas y sus estados de carga, vacío y error.
- [x] Recibir aprobación del Módulo 2 mediante la solicitud expresa de concluir su desarrollo.

## Módulo 3 — aprobado

- [x] Crear `service_prices` con decimal exacto, precio positivo, FK restrictivas y combinación única.
- [x] Mantener la base sin tarifas precargadas para que el Administrador las cargue.
- [x] Implementar creación, actualización, matriz filtrada y faltantes dinámicos.
- [x] Limitar escritura al Administrador, consulta vigente al Encargado y negar acceso al Personal.
- [x] Probar precisión, formatos inválidos, precio cero y concurrencia HTTP/PostgreSQL.
- [x] Conservar precios históricos al desactivar y reactivar catálogos.
- [x] Documentar y probar el contrato de tarifa vigente para el Módulo 5.
- [x] Integrar `/admin/prices` y verificar matriz, modal, carga, vacío y error visualmente.
- [x] Aplicar la migración 5 vacía a la base configurada.
- [x] Recibir aprobación expresa del Módulo 3 antes de iniciar el Módulo 4.

## Módulo 4 — aprobado

- [x] Aprobar CI opcional: `NULL` persistido, `00000000` solo en pantalla/reporte y valor reservado.
- [x] Aprobar placa opcional: `NULL` persistido e ID interno temporal, sin placa ficticia.
- [x] Prohibir cambio de propietario en V1.
- [x] Crear migración, constraints, normalización e índices parciales únicos.
- [x] Implementar CRUD sin borrado físico para clientes y vehículos.
- [x] Limitar el módulo a Administrador y Encargado; bloquear Personal de lavado.
- [x] Enmascarar CI y teléfono en listados y devolverlos completos solo en la ficha autorizada.
- [x] Integrar `/admin/clients` y `/app/clients` mediante un componente compartido con estados de carga, vacío y error.
- [x] Añadir prueba que detecta paneles no renderizados o componentes no resolubles.
- [x] Probar migración doble, PostgreSQL, HTTP, permisos, duplicados y propietario inmutable.
- [x] Aplicar las migraciones 6–7 vacías a la base configurada y revisar visualmente sin insertar datos.
- [x] Recibir aprobación expresa antes de iniciar el Módulo 5.

## Módulo 5 — aprobado

- [x] Confirmar aprobación del Módulo 4.
- [x] Aprobar máquina de estados, snapshots, límites de edición/cancelación y bloqueo temporal de entrega.
- [x] Crear la migración 8 con órdenes, servicios congelados, asignaciones e historial append-only.
- [x] Implementar creación atómica con total decimal exacto y precio vigente obligatorio.
- [x] Implementar listados, detalle, filtros diarios e historiales por cliente y vehículo.
- [x] Aplicar RBAC e IDOR: Personal de lavado solo consulta y avanza órdenes asignadas.
- [x] Proteger transiciones mediante versión optimista y registrar un único historial por cambio válido.
- [x] Limitar edición por estado y exigir motivo de cancelación en estados admitidos.
- [x] Bloquear `FINISHED → DELIVERED` hasta que exista un pago real en el Módulo 6.
- [x] Integrar `/admin/orders` y `/app/orders` con estados de carga, vacío y error exclusivos.
- [x] Añadir prueba de resolución/renderizado de paneles de órdenes.
- [x] Probar migración doble, PostgreSQL, HTTP, permisos, concurrencia y append-only en una base temporal.
- [x] Aplicar la migración 8 vacía a la base configurada y completar revisión visual sin insertar datos.
- [x] Recibir aprobación expresa antes de iniciar el Módulo 6.

## Módulo 6 — aprobado

- [x] Aprobar pago total al dejar el vehículo o al retirarlo, sin anticipos ni parciales.
- [x] Aprobar métodos fijos: efectivo, QR, transferencia y tarjeta; referencia externa opcional.
- [x] Implementar anulación del Encargado con revisión posterior y corrección atómica tras la entrega.
- [x] Crear migración 9 con un pago válido por orden, idempotencia y eventos append-only.
- [x] Tomar siempre el monto desde la orden y bloquear cambios de vehículo/servicios después del pago.
- [x] Integrar pago temprano, pago+entrega y entrega posterior de una orden ya pagada.
- [x] Proteger cobros y datos financieros para Administrador/Encargado; bloquear Personal de lavado.
- [x] Integrar panel de pago en la orden y cola `/admin/payment-reviews`.
- [x] Probar reintentos, doble clic, concurrencia, canceladas, fallos intermedios, anulación, corrección y revisión.
- [x] Aplicar migración 9 vacía a la base configurada y completar revisión visual con datos temporales eliminados.
- [x] Recibir aprobación expresa antes de iniciar el Módulo 7.

## Módulo 7 — implementado, pendiente de aprobación

- [x] Aprobar la marca `Camell Wash`, integrar el logo oficial proporcionado y cambiar el color principal a `#ff5500`.
- [x] Definir fechas inclusivas en `America/La_Paz` y limitar cada consulta a un año.
- [x] Implementar dashboard diario y reportes operativos para Administrador/Encargado.
- [x] Restringir ingresos, ticket promedio, medios y anulaciones agregadas al Administrador.
- [x] Evitar duplicación mediante agregados independientes sobre órdenes, transiciones y pagos válidos.
- [x] Añadir migración 10 con cuatro índices para consultas de reportes.
- [x] Integrar filtros, estados carga/vacío/error, barras con alternativa textual, tablas y exportación CSV.
- [x] Probar fórmulas, fechas, RBAC, rendimiento, PostgreSQL/HTTP, build y resolución de componentes.
- [x] Completar revisión visual con datos temporales y eliminar el entorno aislado.
- [ ] Recibir aprobación expresa antes de iniciar el Módulo 8.

## Módulo 8 — cierre

- [ ] Pendiente; no añadir funciones nuevas.
