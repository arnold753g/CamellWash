# Revisión de seguridad de Camell Wash

Fecha: 21 de octubre de 2026, hora de Bolivia. Proyecto: `/opt/apps/camell-wash`. Sitio: https://camellwash.cloud.

El despliegue está operativo y superó las comprobaciones descritas abajo. Se corrigieron tres defectos de seguridad y se redujeron los privilegios de producción. Las credenciales del administrador entregadas al usuario siguen vigentes. El resto de los proyectos del VPS conserva sus contenedores, configuración y funcionamiento.

## Defectos corregidos y evidencia

| Problema | Antes | Resultado después de corregir |
| --- | --- | --- |
| El proxy interno agrupaba visitantes distintos bajo una sola cuota | De 20 IP de prueba, 6 recibían 401 y 14 recibían 503 | Las 20 IP reciben el 401 esperado sin sesión; una misma IP sigue limitada con 429 |
| Cambiar el correo no revocaba las sesiones anteriores | La cookie anterior seguía autorizada con 200 porque GORM modificaba el correo en memoria antes de compararlo | Cookie anterior rechazada con 401; cambiar solamente el nombre conserva la sesión |
| Una cuenta bloqueada seguía comprobando su contraseña real | Durante el bloqueo, contraseña incorrecta devolvía 401 y correcta devolvía 429 | Ambas devuelven 429, sin cambiar contador ni vencimiento; al vencer el bloqueo se permite el acceso válido |

Los dos defectos de autenticación tienen regresiones con el router real, PostgreSQL, TLS, cookies y CSRF. Se observó el fallo antes del parche y el éxito después. Las pruebas usaron una base separada llamada `camell_wash_security_test_review`, eliminada al terminar; no se crearon cuentas de prueba en la base de producción.

## Medidas aplicadas al proyecto

- Backend con un rol PostgreSQL separado: sin superusuario, creación de bases, creación de roles, replicación, bypass de RLS ni DDL. Tiene permisos de operación sobre los datos; no puede reescribir `schema_migrations`. La prohibición de crear tablas se comprobó directamente.
- Migraciones y asignación de permisos mediante servicios de una sola ejecución. El asignador de permisos pasó dos ejecuciones consecutivas para verificar su idempotencia.
- Backend y frontend con usuarios sin privilegios, sistema de archivos de solo lectura, `/tmp` temporal, capacidades eliminadas y `no-new-privileges`. Se añadieron límites de memoria y procesos del proyecto.
- PostgreSQL y backend sin puertos públicos. El único puerto publicado por Camell es `127.0.0.1:8090`, detrás del proxy HTTPS del VPS.
- Límites por IP saneada por el proxy del host. El host sobrescribe `X-Real-IP` y `X-Forwarded-For`; el proxy interno conserva esa frontera de confianza.
- CSP con hashes exactos de los scripts de Nuxt, comprobados en seis respuestas; restricciones de objetos, URL base, formularios e inclusión en marcos. No se habilitó `unsafe-inline` para JavaScript.
- Secretos fuera de Git: `.env` con permiso 600, `.deployment` con 700 y archivo privado de acceso con 600.

## Componentes actualizados

| Componente | Versión utilizada |
| --- | --- |
| Compilador Go | 1.26.10 |
| Nginx interno de Camell | 1.30.6 |
| Nuxt | 4.6.2 |
| Vue y server-renderer | 3.5.44 |
| devalue | 6.0.3 |
| undici | 8.12.3 |
| golang.org/x/text | 0.41.0 |

Node 22.23.3 y PostgreSQL 16.15 se mantuvieron. La rama Go 1.25 ya no recibía soporte según la [política oficial de Go](https://go.dev/doc/devel/release). Los cambios de Nuxt corresponden a su [versión oficial 4.6.2](https://github.com/nuxt/nuxt/releases/tag/v4.6.2); Nginx se contrastó con sus [publicaciones y avisos oficiales](https://nginx.org/en/security_advisories.html).

## Comprobaciones superadas

- Suite completa de Go, incluidas las regresiones con PostgreSQL aislado.
- 15 de 15 pruebas del frontend y comprobación TypeScript.
- `govulncheck`: cero vulnerabilidades alcanzables en el código analizado. El análisis conserva cuatro avisos de módulos cuyas funciones vulnerables no son llamadas por el proyecto.
- Acceso administrativo real, restauración de sesión y cierre de sesión.
- Cookies `Secure`, `HttpOnly`, `SameSite=Lax`, limitadas a `/api/v1`.
- Solicitudes sin sesión rechazadas; ausencia o falsificación de CSRF y origen no autorizado rechazados con 403.
- Escritura normal del perfil permitida con el rol de base de datos restringido.
- `/.env` y `/.git/HEAD` responden 404.
- HTTPS válido, redirección de `www` al dominio principal y renovación simulada del certificado superada durante la puesta en marcha; temporizador de renovación activo.
- Los otros diez contenedores conservaron identificadores, hora de inicio, reinicios, puertos, montajes y salud. Los cuatro virtual hosts ajenos conservaron sus hashes. Los otros sitios respondieron 200.

## Observaciones pendientes

1. **Dependencias de compilación:** el análisis npm pasó de 26 avisos —3 críticos, 22 altos y 1 moderado— a 11 altos y cero críticos. Los avisos restantes se propagan desde `braces` y `node-forge`, sin parche publicado en los avisos consultados. La inspección del artefacto `.output/server`, sus paquetes y sus importaciones no encontró esas librerías ni las herramientas afectadas. Se mantienen documentados para revisar cuando sus mantenedores publiquen correcciones. Referencias: [braces](https://github.com/advisories/GHSA-vfj7-8cjw-p6xm), [node-forge](https://github.com/advisories/GHSA-86w9-cpqp-85rv).
2. **Acceso global al VPS:** SSH admite `root` mediante contraseña. Migrar a llaves y restringir ese acceso requiere coordinación a nivel del servidor compartido. Esa configuración global quedó sin modificar para respetar el alcance de trabajar únicamente en Camell Wash.

Las evidencias, registros de pruebas y respaldo previo están en `/opt/apps/camell-wash/.deployment/security/`. Los procedimientos de actualización están documentados en `docs/deployment-vps.md`. Los hashes CSP deben recalcularse después de cada compilación del frontend con `python3 scripts/configure-csp.py`.
