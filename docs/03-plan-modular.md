# Plan modular

Fuente funcional y criterios de aceptación: [`GUIA_UNICA_SISTEMA_LAVADERO.md`](../GUIA_UNICA_SISTEMA_LAVADERO.md). Este documento solo organiza la ejecución técnica.

## Secuencia

| Fase | Trabajo técnico | Puerta de salida |
| --- | --- | --- |
| 0 | Auditoría, arquitectura, modelo inicial, contrato visual y decisiones. | Documentos revisados; decisiones de Módulo 1 aprobadas. |
| 1 | Tres roles, login, sesiones, alta/edición/desactivación administrativa y primer admin. | Matriz RBAC, migración limpia/actualizada, tests/build y revisión visual aprobados. |
| 2 | Tipos, categorías y servicios extensibles con baja lógica. | Unicidad y activos/inactivos probados; seeds comerciales confirmados. |
| 3 | Matriz decimal servicio × tipo y contrato de tarifa vigente. | Precisión, faltantes, unicidad y concurrencia probados. |
| 4 | Clientes y varios vehículos. | Políticas CI/placa aprobadas, privacidad y permisos probados. |
| 5 | Órdenes, snapshots, asignaciones, estados e historiales. | Flujo transaccional, IDOR y concurrencia probados; móvil verificado. |
| 6 | Pago total único e integración con entrega. | Idempotencia, exactitud, trazabilidad y roles probados. |
| 7 | Dashboard y reportes reales. | Fórmulas aprobadas, sin duplicación, fechas/roles/rendimiento probados. |
| 8 | Integración y cierre sin funciones nuevas. | E2E, seguridad, migración limpia, documentación y veredicto final. |

## Reglas de ejecución por módulo

1. Confirmar aprobación del módulo anterior y decisiones que afecten datos, permisos o dinero.
2. Inspeccionar archivos equivalentes y contrato visual antes de editar.
3. Añadir migración, backend, frontend y pruebas dentro del alcance exclusivo.
4. Ejecutar formato, análisis estático, tests, build y migración desde base limpia cuando aplique.
5. Verificar permisos por API y continuidad visual en los breakpoints acordados.
6. Actualizar `tasks.md`, `PROGRESS.md` y `lessons.md` con evidencia real.
7. Revisar todos los archivos modificados; mientras no exista Git, usar inventario/hash como sustituto explícito y declarar que no equivale a `git diff`.
8. Detenerse y solicitar aprobación; nunca avanzar automáticamente al módulo siguiente.

## Plan exacto propuesto para el Módulo 1

- [x] Aprobar correo normalizado como login único.
- [x] Definir `admin`, `manager` y `wash_staff` con etiquetas visibles en español.
- [x] Diseñar migración aditiva que clasifica únicamente `user@example.test` como Encargado y se detiene ante cualquier otro `user` no clasificado.
- [x] Eliminar registro público y su configuración frontend/backend.
- [x] Implementar creación y edición administrativa con contraseña inicial, validación, baja lógica y revocación de sesiones.
- [x] Adaptar middleware, modelos, filtros, navegación, guards y etiquetas a tres roles.
- [x] Limitar el seed al Administrador ya definido en `.env`, con confirmación y bloqueo productivo.
- [x] Endurecer/probar login, bloqueo concurrente, usuario inactivo, expiración, CSRF y revocación.
- [x] Añadir pruebas unitarias y verificación PostgreSQL/HTTP aislada.
- [x] Aplicar la migración a la base configurada con la clasificación aprobada.
- [ ] Completar inspección visual interactiva cuando haya navegador controlable.
- [x] Ejecutar la suite final y aprobar el Módulo 1 antes del Módulo 2.

Fuera del Módulo 1: catálogos, precios, clientes, vehículos, órdenes, pagos y reportes.

## Ejecución del Módulo 2

- [x] Crear tablas, índices, constraints, FK restrictiva y baja lógica.
- [x] Definir normalización y unicidad en aplicación y PostgreSQL.
- [x] Implementar administración exclusiva para tipos, categorías y servicios.
- [x] Implementar consultas activas para los tres roles.
- [x] Añadir búsqueda, filtros, paginación y relación servicio/categoría.
- [x] Integrar `/admin/catalogs` con PrimeVue y el contrato visual existente.
- [x] Probar duplicados, permisos, activos/inactivos, categorías y migración idempotente mediante PostgreSQL/HTTP.
- [x] Aplicar la migración 3 vacía a la base configurada.
- [x] Confirmar y cargar de forma idempotente la escritura comercial y los valores iniciales aprobados.
- [x] Completar inspección visual interactiva de las tres pestañas y sus estados de carga, vacío y error.
- [x] Recibir aprobación del Módulo 2 mediante la solicitud expresa de concluir su desarrollo.

Fuera del Módulo 2: precios, clientes, vehículos, órdenes, pagos y reportes.

## Ejecución del Módulo 3

- [x] Resolver que no habrá precios precargados y que `0.00` es inválido.
- [x] Crear tarifa decimal positiva con combinación única y FK restrictivas.
- [x] Implementar matriz activa, resumen, filtros y faltantes calculados dinámicamente.
- [x] Implementar alta/edición administrativa y consulta vigente para Administrador/Encargado.
- [x] Probar permisos, precisión exacta, validación, concurrencia e inactividad en PostgreSQL/HTTP.
- [x] Integrar `/admin/prices` con el contrato visual y estados de carga, vacío y error.
- [x] Aplicar la migración 5 vacía a la base configurada.
- [x] Documentar el contrato de congelamiento que consumirá el Módulo 5.
- [x] Recibir aprobación expresa antes del Módulo 4.

Fuera del Módulo 3: clientes, vehículos, órdenes, pagos y reportes.

## Ejecución del Módulo 4

- [x] Definir CI opcional como `NULL`, reservar `00000000` para presentación y exigir unicidad solo a valores reales.
- [x] Definir placa opcional como `NULL`, usar el ID interno temporalmente y no inventar placas.
- [x] Mantener propietario inmutable en V1.
- [x] Crear `clients` y `vehicles` con constraints, FK restrictivas e índices parciales únicos.
- [x] Implementar API compartida para Administrador/Encargado, con Personal de lavado bloqueado.
- [x] Proteger datos personales en listados y mostrar la ficha completa solo dentro del módulo autorizado.
- [x] Integrar las rutas `/admin/clients` y `/app/clients` con PrimeVue y estados explícitos.
- [x] Probar migración idempotente, SQL `NULL`, duplicados normalizados, RBAC, contrato inmutable y ausencia de delete.
- [x] Aplicar las migraciones 6–7 vacías a la base configurada y completar revisión visual.
- [x] Recibir aprobación expresa antes del Módulo 5.

Fuera del Módulo 4: órdenes, snapshots, asignaciones de personal, estados operativos, pagos y reportes.

## Ejecución del Módulo 5

- [x] Aprobar snapshots de nombre y precio, permisos, límites de edición/cancelación y bloqueo temporal de entrega.
- [x] Crear `orders`, `order_services`, `order_staff` y `order_status_history` mediante la migración 8.
- [x] Implementar creación transaccional con servicios activos, tarifa vigente, personal activo y total decimal exacto.
- [x] Implementar la secuencia `PENDING → IN_PROGRESS → QUALITY_CONTROL → FINISHED`; reservar `DELIVERED` para el Módulo 6.
- [x] Permitir al Personal de lavado avanzar solo sus órdenes asignadas y proteger consultas contra IDOR.
- [x] Permitir a Administrador/Encargado operar transiciones válidas y cancelar con motivo obligatorio.
- [x] Implementar edición por estado y concurrencia optimista mediante `version`.
- [x] Integrar `/admin/orders` y `/app/orders`, historiales por cliente/vehículo y estados visuales explícitos.
- [x] Probar migración idempotente, atomicidad, snapshots, RBAC, IDOR, concurrencia, historial append-only y bloqueo de entrega.
- [x] Aplicar la migración 8 vacía a la base configurada y completar revisión visual responsive.
- [x] Recibir aprobación expresa antes del Módulo 6.

Fuera del Módulo 5: cobros, métodos de pago, entrega efectiva, dashboard y reportes.

## Ejecución del Módulo 6

- [x] Permitir el pago total desde `PENDING` hasta `FINISHED`, sin cambiar el estado cuando se cobra temprano.
- [x] Validar efectivo, QR, transferencia y tarjeta; mantener referencia externa opcional.
- [x] Crear `payments` y `payment_events` mediante la migración 9, con máximo un pago válido e idempotencia.
- [x] Integrar cobro al retiro con `FINISHED → DELIVERED` dentro de una transacción atómica.
- [x] Permitir entrega posterior cuando la orden ya tiene un pago válido.
- [x] Bloquear cancelación y cambios comerciales mientras exista un pago válido.
- [x] Permitir anulación previa a entrega; las anulaciones del Encargado quedan pendientes de revisión administrativa.
- [x] Implementar corrección por reemplazo atómico, incluso después de la entrega.
- [x] Integrar panel accesible de pago y cola administrativa de revisiones.
- [x] Probar exactitud, campos protegidos, roles, idempotencia, concurrencia, consistencia y eventos append-only.
- [x] Aplicar migración 9 vacía a la base configurada y completar revisión visual con entorno temporal aislado.
- [x] Recibir aprobación expresa antes del Módulo 7.

Fuera del Módulo 6: pagos parciales, anticipos, descuentos, devoluciones reales, facturación, dashboard y reportes.

## Ejecución del Módulo 7

- [x] Sustituir el dashboard genérico por indicadores reales de órdenes, transiciones y pagos.
- [x] Interpretar rangos inclusivos en `America/La_Paz`, con máximo de un año y persistencia UTC.
- [x] Ofrecer operación agregada a Administrador/Encargado y finanzas agregadas solo al Administrador.
- [x] Definir ingreso como suma de pagos `VALID` por `paid_at`; mostrar anulaciones aparte por `voided_at`.
- [x] Calcular recibidas por `received_at`, entregadas/canceladas por su transición y ciclo desde recepción a entrega.
- [x] Añadir índices de reporte sin crear tablas resumen ni contadores duplicados.
- [x] Integrar dashboard diario, reportes por periodo, CSV, estados exclusivos y tablas accesibles.
- [x] Aplicar marca `Camell Wash`, logo oficial proporcionado y color principal `#ff5500` mediante tokens.
- [x] Probar fórmulas, fechas, permisos, rendimiento, PostgreSQL/HTTP, build y revisión visual.
- [ ] Recibir aprobación expresa antes del Módulo 8.

Fuera del Módulo 7: facturación, devoluciones, integraciones bancarias, almacenamiento de reportes y funciones nuevas de cierre.
