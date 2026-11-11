# Rol de contador

El rol `accountant` consulta y exporta reportes operativos, financieros y el detalle contable de las órdenes desde `/accountant`. Consulta las fichas y vehículos de los clientes en `/accountant/clients`. Los accesos de escritura comercial, revisión de pagos y administración están reservados a los roles operativos correspondientes. Conserva las funciones de su propia cuenta.

## Permisos de reportes

Todos los endpoints siguientes se encuentran bajo `/api/v1` y requieren una sesión activa.

| Ruta GET | Administrador | Encargado | Personal de lavado | Contador |
|---|---|---|---|---|
| `/reports/operations` | Sí | Sí | No | Sí |
| `/reports/dashboard` | Sí | No | No | Sí |
| `/reports/financial` | Sí | No | No | Sí |
| `/reports/orders` | Sí | No | No | Sí |
| `/accountant/clients` y sus fichas/vehículos | Sí | No | No | Sí |
| `/admin/dashboard` | Sí | No | No | No |
| `/admin/reports/financial` | Sí | No | No | No |

Las rutas históricas de administración permanecen exclusivas del administrador. El reporte de órdenes incluye la solicitud de factura, nombre, CI, teléfono, correo opcional, placa, total y pago vigente. Su exportación recorre todas las páginas del filtro y neutraliza fórmulas en los campos del cliente. Las rutas de clientes contables admiten únicamente GET.

La protección se aplica mediante listas explícitas de roles en el servidor, además de los guardias de navegación. El servicio de órdenes rechaza roles no autorizados y sigue limitando al personal de lavado a sus órdenes asignadas. Los roles desconocidos se rechazan en la base de datos, la validación de usuarios y las sesiones.

## Alta y cambio de rol

Un administrador puede crear un usuario seleccionando **Contador**, o asignar ese rol a un usuario existente. El cambio de rol revoca sus sesiones existentes; el usuario debe volver a ingresar. No se crea una cuenta contable compartida ni se modifica la contraseña de ninguna cuenta existente.

La migración `000011_accountant_role.sql` amplía la restricción de roles y conserva los usuarios actuales. Estos reportes cubren ingresos y cobros; no incorporan gastos, balances contables ni conciliación bancaria.

## Validación y despliegue

Las pruebas de integración requieren `TEST_DATABASE_DSN` apuntando a una base desechable cuyo nombre comience con `camell_wash_security_test_`. Verifican autenticación real, permisos de lectura por rol, denegación de escrituras con CSRF válido, asignación y revocación de sesiones. Nunca deben usar la base de producción.

Antes de desplegar, comparar los archivos modificados con la versión actual del VPS y respaldar únicamente `/opt/apps/camell-wash` y su base de datos. Esta copia local procede de una revisión anterior y no debe reemplazar archivos de dependencias, secretos ni configuración de producción del VPS.

Aplicar la migración con el servicio `migrate`, reconstruir los servicios propios afectados y regenerar los hashes CSP con `scripts/configure-csp.py`. Validar y recargar únicamente el contenedor nginx de Camell Wash. Comprobar salud, acceso de los cuatro roles y que los contenedores y configuraciones de los demás proyectos permanezcan intactos.

No revertir la restricción de roles mientras existan usuarios `accountant`. Una reversión requiere reasignar esas cuentas y revocar sus sesiones, conservando siempre los datos.
