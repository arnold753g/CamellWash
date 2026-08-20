# Módulo 4 — Clientes y vehículos

Estado: **IMPLEMENTADO, MIGRADO, VERIFICADO Y APROBADO**.

Fuente funcional: [`GUIA_UNICA_SISTEMA_LAVADERO.md`](../GUIA_UNICA_SISTEMA_LAVADERO.md).

## Decisiones aplicadas

- El CI es opcional. La ausencia se guarda como `NULL`; interfaz y futuros reportes muestran `00000000`.
- `00000000` está reservado y la API rechaza su ingreso como CI real.
- Los CI reales se comparan sin espacios/guiones y en mayúsculas; son únicos únicamente cuando existen.
- La placa es opcional y se guarda como `NULL`. Un vehículo sin placa se presenta como `Sin placa · ID #<id>`.
- Nunca se crea una placa ficticia, incluyendo vehículos de competición.
- El propietario (`client_id`) es inmutable en V1. El contrato PATCH no acepta ese campo y PostgreSQL conserva la FK restrictiva.
- No existen endpoints de borrado físico.

## Modelo e integridad

La migración `000006_clients_vehicles.sql` crea:

- `clients`: nombre, CI nullable, CI canónico generado, teléfono y timestamps.
- `vehicles`: propietario, tipo, placa nullable, placa canónica generada, marca/modelo/color y timestamps.
- Índices parciales únicos para CI y placa solo cuando su forma canónica no es `NULL`.
- Constraints de longitud, teléfono y valor CI reservado.
- FK `ON DELETE RESTRICT` hacia cliente y tipo de vehículo.
- Triggers comunes para `updated_at`.

La migración aditiva `000007_customer_identifier_format.sql` refuerza en PostgreSQL que las formas canónicas reales de CI y placa contengan exclusivamente letras ASCII mayúsculas y dígitos. Se añadió sin reescribir la migración 6 ya aplicada.

Los nombres se recortan y colapsan; teléfono, CI y placa se validan/canonicalizan en aplicación. La base mantiene las restricciones finales frente a concurrencia o escrituras ajenas a la API.

## Privacidad y permisos

Administrador y Encargado pueden listar, registrar y editar clientes/vehículos. Personal de lavado recibe `403`.

El listado enmascara CI y teléfono, conservando solo los cuatro caracteres finales. La ficha autorizada devuelve los valores completos. Un CI ausente se muestra `00000000` y se acompaña con `ci_provided=false` para evitar confundir el marcador con un documento real.

## Endpoints

| Método y ruta | Permiso | Función |
| --- | --- | --- |
| `GET /api/v1/clients` | Administrador/Encargado | Lista paginada y enmascarada con búsqueda. |
| `POST /api/v1/clients` | Administrador/Encargado | Registra cliente. |
| `GET /api/v1/clients/{id}` | Administrador/Encargado | Ficha completa. |
| `PATCH /api/v1/clients/{id}` | Administrador/Encargado | Edita datos del cliente. |
| `GET /api/v1/clients/{id}/vehicles` | Administrador/Encargado | Lista vehículos del propietario. |
| `POST /api/v1/clients/{id}/vehicles` | Administrador/Encargado | Agrega vehículo al propietario. |
| `GET /api/v1/vehicles/{id}` | Administrador/Encargado | Detalle de vehículo. |
| `PATCH /api/v1/vehicles/{id}` | Administrador/Encargado | Edita vehículo sin cambiar propietario. |

## Interfaz

- Administrador: `/admin/clients`.
- Encargado: `/app/clients`.
- Ambas rutas importan explícitamente `CustomerVehicleWorkspace.vue` para evitar paneles silenciosamente no renderizados.
- Incluye búsqueda, paginación, alta/edición de cliente, ficha, listado y alta/edición de vehículos.
- Explica CI/placa opcionales, muestra propietario deshabilitado y comunica carga, vacío, error y validaciones sin depender solo del color.
- La navegación del Encargado se incorpora solo después de hidratación para evitar diferencias SSR/cliente.

## Evidencia

| Verificación | Resultado |
| --- | --- |
| Migraciones 1–7 desde base vacía y segunda ejecución | Aprobada. |
| Base configurada | Versión 7; constraints de formato presentes; 0 clientes y 0 vehículos. |
| Dos clientes sin CI | `201`; dos `NULL`; ambos muestran `00000000`. |
| CI `00000000` / CI real duplicado normalizado | `422` / `409`. |
| Listado / ficha | Identificadores enmascarados / datos completos. |
| Dos vehículos sin placa del mismo cliente | `201`; ambos `plate IS NULL`. |
| Placa duplicada normalizada | `409`. |
| Cambio de propietario por PATCH | `400`; propietario conservado. |
| Tipo inactivo nuevo / vehículo histórico | `422` / edición conservando tipo permitida. |
| Encargado / Personal de lavado | Escritura `200` / acceso `403`. |
| DELETE de vehículo | Ruta inexistente (`404`). |
| Tests frontend | 8/8, incluidas rutas, panel, estados y resolución de componentes. |
| Tests Go y typecheck Nuxt | Aprobados. |
| Build Nuxt | Aprobado; solo avisos DEP0155 transitivos. |
| Revisión visual | Carga, error/reintento, recuperación a vacío, formulario y validación aprobados sin datos permanentes. |

`scripts/verify-module4.ps1` reproduce la integración en una base temporal con nombre validado y la elimina siempre al finalizar.

## Límite y próxima puerta

No se implementaron órdenes, servicios de una orden, snapshots, asignaciones, historiales, pagos ni reportes. El usuario aprobó el módulo el 2026-08-20 y solicitó continuar con el Módulo 5. Git sigue sin inicializar por decisión expresa.
